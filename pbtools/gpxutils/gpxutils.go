package gpxutils

import (
	"github.com/tkrajina/gpxgo/gpx"
)

// ExtendBounds extends the given GPX boundaries with the given decimal degree increment
func ExtendBounds(bounds gpx.GpxBounds, increment float64) gpx.GpxBounds {
	bounds.MinLatitude = bounds.MinLatitude - increment
	bounds.MaxLatitude = bounds.MaxLatitude + increment
	bounds.MinLongitude = bounds.MinLongitude - increment
	bounds.MaxLongitude = bounds.MaxLongitude + increment
	return bounds
}
