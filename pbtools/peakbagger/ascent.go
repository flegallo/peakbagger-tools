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
