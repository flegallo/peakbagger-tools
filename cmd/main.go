package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"peakbagger-tools/pbtools/gpxutils"

	"peakbagger-tools/pbtools/peakbagger"

	"github.com/tkrajina/gpxgo/gpx"

	"peakbagger-tools/pbtools/config"
	"peakbagger-tools/pbtools/strava"
)

func main() {
	if err := realMain(); err != nil {
		fmt.Printf("failed to run service. %s", err)
		os.Exit(1)
	}
}

func realMain() error {

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	strava := strava.NewClient(cfg.HTTPPort, cfg.StravaClientID, cfg.StravaSecretID)
	pb := peakbagger.NewClient(cfg.PeakBaggerUsername, cfg.PeakBaggerPassword)

	// get auth token to query Strava
	err = strava.RetrieveAuthToken()
	if err != nil {
		return err
	}

	// get GPX from Strava
	g, err := strava.DownloadGPX(cfg.StravaActivityID)
	if err != nil {
		return err
	}

	// peakbagger limits gpx to 3000 points
	g.ReduceTrackPoints(3000, 0)

	// login to peakbagger
	_, err = pb.Login()
	if err != nil {
		return err
	}

	// find peaks within gpx boundaries
	bounds := gpxutils.ExtendBounds(g.Bounds(), 0.01)
	peaks, err := pb.FindPeaks(&bounds)
	if err != nil {
		return err
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
	positions := g.GetLocationsPositionsOnTrack(1000, locations...)
	for i, p := range positions {
		if len(p) > 0 {
			peaksOnTrack = append(peaksOnTrack, peaks[i])
		}
	}

	// no peaks, on track, quit for now
	// TODO propose to the user to manually enter a peak
	if len(peaksOnTrack) == 0 {
		return errors.New("no peaks found on track to add to Peakbagger")
	}

	// confirm with the user which peaks he summited
	// TODO propose the user to edit the list and add failed attempts
	fmt.Printf("Found %d peak(s) on track:\n", len(peaksOnTrack))
	for _, p := range peaksOnTrack {
		fmt.Printf(" - %s\n", p.Name)
	}
	fmt.Print("Is that correct? (y/n)")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if input.Text() != "y" && input.Text() != "yes" {
		return nil
	}

	// add new ascents to peakbagger
	fmt.Println("Adding ascents to peakbagger!")
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

		fmt.Printf("Adding ascent of '%s' to peakbagger\n", p.Name)
		_, err := pb.AddAscent(ascent)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Done, %d peak(s) added to peakbagger\n", len(peaksOnTrack))

	return nil
}
