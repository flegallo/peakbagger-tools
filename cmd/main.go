package main

import (
	"bufio"
	"fmt"
	"os"
	"peakbagger-tools/pbtools/terminal"

	"peakbagger-tools/pbtools/gpxutils"

	"peakbagger-tools/pbtools/peakbagger"

	"github.com/tkrajina/gpxgo/gpx"

	"peakbagger-tools/pbtools/config"
	"peakbagger-tools/pbtools/strava"
)

// MaxGpxPoints Maximum number of points a gpx is allowed to be uploaded on PeakBaggers
const MaxGpxPoints = 3000

// DistanceToPeakThreshold is the maximum distance in meters from the peak coordinates
// after which we consider the peak to be summited.
const DistanceToPeakThreshold = 25

func main() {
	if !realMain() {
		os.Exit(1)
	}
}

func realMain() bool {

	cfg, err := config.Load()
	if err != nil {
		terminal.Error(nil, "Failed to load config")
	}

	strava := strava.NewClient(cfg.HTTPPort, cfg.StravaClientID, cfg.StravaSecretID)
	pb := peakbagger.NewClient(cfg.PeakBaggerUsername, cfg.PeakBaggerPassword)

	// get auth token to query Strava
	err = strava.RetrieveAuthToken()
	if err != nil {
		terminal.Error(err, "Something went wrong while trying to fetch auth token")
		return false
	}

	// download GPX on Strava
	o := terminal.NewOperation("Downloading GPX from Strava")
	g, err := strava.DownloadGPX(cfg.StravaActivityID)
	if err != nil {
		o.Error(err, "Failed to download GPX from Strava")
		return false
	}
	nbPoints := g.GetTrackPointsNo()
	o.Success("GPX downloaded from Strava (%d points)", nbPoints)

	// login to peakbagger
	o = terminal.NewOperation("Login to peakbagger.com with username '%s'", pb.Username)
	_, err = pb.Login()
	if err != nil {
		o.Error(err, "Failed to login to peakbagger.com")
		return false
	}
	o.Success("Successfully logged in as '%s'", pb.Username)

	// find peaks within gpx boundaries
	o = terminal.NewOperation("Searching for peaks on GPX track")
	bounds := gpxutils.ExtendBounds(g.Bounds(), 0.01)
	peaks, err := pb.FindPeaks(&bounds)
	if err != nil {
		o.Error(err, "Failed to find peaks around GPX boundaries")
		return false
	}

	// check which peaks are on the track
	locations := make([]gpx.Location, len(peaks))
	for i, p := range peaks {

		e := gpx.NewNullableFloat64(-1)
		e.SetNull()

		locations[i] = &gpx.Point{
			Longitude: p.Longitude,
			Latitude:  p.Latitude,
			Elevation: *e,
		}
	}
	peaksOnTrack := []peakbagger.Peak{}
	distances := gpxutils.FindShortestDistanceToTrack(g, locations)
	for i, d := range distances {
		if d < DistanceToPeakThreshold {
			peaksOnTrack = append(peaksOnTrack, peaks[i])
		}
	}
	if len(peaksOnTrack) > 0 {
		o.Success("Found %d peaks on GPX track", len(peaksOnTrack))
	} else {
		o.Error(nil, "No peaks found on GPX track")
		return false
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
		return false
	}

	fmt.Println("")

	// peakbagger limits gpx to a certain nb of points
	if nbPoints > MaxGpxPoints {
		o = terminal.NewOperation("Reducing GPX to %d points", MaxGpxPoints)
		g.ReduceTrackPoints(MaxGpxPoints, 0)
		o.Success("GPX reduced to %d points", MaxGpxPoints)
	}

	// add new ascents to peakbagger
	gpsPoints := g.Tracks[0].Segments[0].Points
	for _, p := range peaksOnTrack {
		ascent := peakbagger.Ascent{
			PeakID:         p.PeakID,
			Date:           g.Time,
			Gpx:            g,
			TripReport:     strava.GetActivityLink(cfg.StravaActivityID),
			StartElevation: gpsPoints[0].Elevation.Value(),
			EndElevation:   gpsPoints[len(gpsPoints)-1].Elevation.Value(),
		}

		o = terminal.NewOperation("Adding ascent of '%s' to peakbagger", p.Name)
		_, err := pb.AddAscent(ascent)
		if err != nil {
			o.Error(err, "Failed to add ascenf of '%s' to peakbagger", p.Name)
			return false
		}
		o.Success("Added ascent of '%s' to peakbagger!", p.Name)
	}

	return true
}
