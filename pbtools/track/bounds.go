package track

// Bounds represents track coordinate boundaries
type Bounds struct {
	MinLat, MinLng float64
	MaxLat, MaxLng float64
}

// Extend extends boundaries from given decimal degrees
func (b Bounds) Extend(inc float64) Bounds {
	b.MinLat -= inc
	b.MinLng -= inc
	b.MaxLat += inc
	b.MaxLng += inc
	return b
}
