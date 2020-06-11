package peakbagger

// Peak represents a peak in peakbagger.com
type Peak struct {
	PeakID    string
	Latitude  float64
	Longitude float64
	Name      string
}

// Lat returns latitude in degrees
func (p Peak) Lat() float64 {
	return p.Latitude
}

// Lng returns longitude in degrees
func (p Peak) Lng() float64 {
	return p.Longitude
}
