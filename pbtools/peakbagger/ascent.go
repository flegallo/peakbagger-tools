package peakbagger

import (
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

// Ascent represents a peak ascent in peakbagger.com
type Ascent struct {
	PeakID string

	Date       *time.Time
	Gpx        *gpx.GPX
	TripReport string

	NetGain        float64       // Net elevation change on the way up (in meters)
	ExtraGainUp    float64       // Extra elevation gain on the way up (in meters)
	StartElevation float64       // Start elevation (in meters)
	DistanceUp     float64       // Distance up (in meters)
	TimeUp         time.Duration // Duration up

	NetLoss       float64       // Net elevation change on the way down (in meters)
	ExtraLossDown float64       // Extra elevation loss on the way up (in meters)
	EndElevation  float64       // End elevation (in meters)
	DistanceDown  float64       // Distance down (in meters)
	TimeDown      time.Duration // Duration down
}

// AscentSummary represents a short version of a peak ascent in peakbagger.com
type AscentSummary struct {
	AscentID  string
	PeakID    string
	PeakName  string
	Date      *time.Time
	Elevation float64
	Location  string
}

// ClimberAscents represents a list of ascents
type ClimberAscents []AscentSummary

// Has returns true if the ascent already exists for the given peak ID and date, false otherwise
func (as *ClimberAscents) Has(peakID string, date *time.Time) bool {
	for _, a := range *as {
		if dateEqual(*a.Date, *date) && peakID == a.PeakID {
			return true
		}
	}

	return false
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
