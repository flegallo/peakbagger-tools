package track

import (
	"time"
)

// Point point in 2D coordinates
type Point struct {
	Latitude, Longitude float64
	Elevation           float64
	Time                time.Time
}

// Lat returns the latitude in degrees
func (p Point) Lat() float64 {
	return p.Latitude
}

// Lng returns the longitude in degrees
func (p Point) Lng() float64 {
	return p.Longitude
}
