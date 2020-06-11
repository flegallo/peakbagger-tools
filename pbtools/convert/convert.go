package convert

import (
	"math"
	"strconv"
	"time"
)

// seriously US?
const meterToFeet = 3.28084

const meterToMile = 0.0006213712

// ToFeet returns the given distance in meters to feet
// because we live in the US, and this country is still using the deprecated imperial
// system instead of the metric system like the rest of the world.
func ToFeet(meters float64) float64 {
	return meters * meterToFeet
}

// ToMiles returns the given distance in meters to miles
// because we live in the US, and this country is still using the deprecated imperial
// system instead of the metric system like the rest of the world.
func ToMiles(meters float64) float64 {
	return meters * meterToMile
}

// ToDaysHoursMin returns the number of days, hours, and minutes in a given duration
func ToDaysHoursMin(d time.Duration) (int, int, int) {
	if d <= 0 {
		return 0, 0, 0
	}
	r := d

	days := int(r.Hours() / 24)
	r -= time.Duration(days) * 24 * time.Hour

	hours := int(r.Hours())
	r -= time.Duration(hours) * time.Hour

	minutes := int(r.Minutes())
	r -= time.Duration(minutes) * time.Minute

	return days, hours, minutes
}

// Ftoan converts a float to a non decimal string
func Ftoan(n float64) string {
	return strconv.Itoa(int(math.Round(n)))
}
