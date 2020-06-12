package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"peakbagger-tools/pbtools/config"
	"peakbagger-tools/pbtools/peakbagger"
	"peakbagger-tools/pbtools/strava"
	"peakbagger-tools/pbtools/terminal"
	"peakbagger-tools/pbtools/track"

	"github.com/google/subcommands"
)

type addCmd struct {
	stravaActivity string
}

// MaxGpxPoints Maximum number of points a gpx is allowed to be uploaded on PeakBaggers
const MaxGpxPoints = 3000

// DistanceToPeakThreshold is the maximum distance in meters from the peak coordinates
// after which we consider the peak to be summited.
const DistanceToPeakThreshold = 25

func (*addCmd) Name() string     { return "add" }
func (*addCmd) Synopsis() string { return "Add ascents to peakbagger.com from a Strava activity." }
func (*addCmd) Usage() string {
	return `add [-activity] <url>
	Register climbed peaks from Strava activity to peakbagger.
  `
}

func (c *addCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.stravaActivity, "activity", "", "Strava activity link")
}

func (c *addCmd) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	cfg := args[0].(*config.Config)

	activityID, err := strava.ParseActivityID(c.stravaActivity)
	if err != nil {
		terminal.Error(err, "Couldn't parse Strava activity id")
		return 1
	}

	strava := strava.NewClient(cfg.HTTPPort, cfg.StravaClientID, cfg.StravaSecretID)
	pb := peakbagger.NewClient(cfg.PeakBaggerUsername, cfg.PeakBaggerPassword)

	// get auth token to query Strava
	err = strava.RetrieveAuthToken()
	if err != nil {
		terminal.Error(err, "Something went wrong while trying to fetch auth token")
		return 1
	}

	// download GPX on Strava
	o := terminal.NewOperation("Downloading GPX from Strava")
	g, err := strava.DownloadGPX(activityID)
	if err != nil {
		o.Error(err, "Failed to download GPX from Strava")
		return 1
	}
	nbPoints := g.GetTrackPointsNo()
	t := track.New(&g.Tracks[0].Segments[0].Points)
	o.Success("GPX downloaded from Strava (%d points)", nbPoints)

	// login to peakbagger
	o = terminal.NewOperation("Login to peakbagger.com with username '%s'", pb.Username)
	_, err = pb.Login()
	if err != nil {
		o.Error(err, "Failed to login to peakbagger.com")
		return 1
	}
	o.Success("Successfully logged in as '%s'", pb.Username)

	// fetch climber ascents
	o = terminal.NewOperation("Retrieving climber ascents from peakbagger.com")
	ascents, err := pb.ListAscents()
	if err != nil {
		o.Error(err, "Failed to retrieve climber ascents from peakbagger.com")
		return 1
	}
	o.Success("Successfully fetched %d ascents", len(ascents))

	// find peaks within gpx boundaries
	o = terminal.NewOperation("Searching for peaks on GPX track")
	bounds := t.Bounds().Extend(0.01)
	peaks, err := pb.FindPeaks(&bounds)
	if err != nil {
		o.Error(err, "Failed to find peaks around GPX boundaries")
		return 1
	}

	// check which peaks are on the track
	peaksOnTrack := []peakbagger.Peak{}
	for i, p := range peaks {
		d := t.GetShortestDistanceFromPoint(p)
		if d < DistanceToPeakThreshold {
			peaksOnTrack = append(peaksOnTrack, peaks[i])
		}
	}
	if len(peaksOnTrack) > 0 {
		o.Success("Found %d peaks on GPX track", len(peaksOnTrack))
	} else {
		o.Error(nil, "No peaks found on GPX track")
		return 1
	}

	// confirm with the user which peaks he summited
	// TODO propose the user to edit the list and add failed attempts
	fmt.Println("")
	fmt.Println("   List of peak(s) on track:")
	for i, p := range peaksOnTrack {
		fmt.Printf("    (%d) %s\n", i+1, p.Name)
	}
	fmt.Println("")
	fmt.Print("Is that correct? (y/n)")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if input.Text() != "y" && input.Text() != "yes" {
		return 1
	}

	fmt.Println("")

	// peakbagger limits gpx to a certain nb of points
	if nbPoints > MaxGpxPoints {
		o = terminal.NewOperation("Reducing GPX to %d points", MaxGpxPoints)
		g.ReduceTrackPoints(MaxGpxPoints, 0)
		o.Success("GPX reduced to %d points", MaxGpxPoints)
	}

	// add new ascents to peakbagger
	fullStats := len(peaksOnTrack) == 1 // add up and down stats only if the track countains only 1 ascent
	for _, p := range peaksOnTrack {
		o = terminal.NewOperation("Adding ascent of '%s' to peakbagger", p.Name)
		closestPoint, index := t.GetClosestPoint(p)

		if ascents.Has(p.PeakID, &closestPoint.Time) {
			o.Error(err, "Ascent of '%s' on %s already exists on peakbagger", p.Name, closestPoint.Time.Format("Jan 2, 2006"))
			break
		}

		ascent := peakbagger.Ascent{
			PeakID:         p.PeakID,
			Date:           &closestPoint.Time,
			Gpx:            g,
			TripReport:     strava.GetActivityLink(activityID),
			StartElevation: t.Points[0].Elevation,
			EndElevation:   t.Points[len(t.Points)-1].Elevation,
		}

		// Add up and down stats
		if fullStats {
			t1, t2 := t.Split(index)
			s1 := t1.Stats()
			s2 := t2.Stats()

			ascent.NetGain = s1.EndElevation - s1.StartElevation
			ascent.NetLoss = s2.EndElevation - s2.StartElevation
			ascent.ExtraGainUp = s1.ElevationLoss
			ascent.ExtraLossDown = s2.ElevationGain
			ascent.DistanceUp = s1.Distance
			ascent.DistanceDown = s2.Distance
			ascent.TimeUp = s1.Duration
			ascent.TimeDown = s2.Duration
		}

		_, err := pb.AddAscent(ascent)
		if err != nil {
			o.Error(err, "Failed to add ascent of '%s' to peakbagger", p.Name)
			break
		}
		o.Success("Added ascent of '%s' to peakbagger!", p.Name)
	}

	return 0
}
