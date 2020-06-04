package convert

// seriously US?
const feetToMeter = 3.28084

// ToFeet returns the given distance in meters to feet
// because we live in the US, and this country is still using the deprecated imperial
// system instead of the metric system like the rest of the world.
func ToFeet(meters float64) float64 {
	return meters * feetToMeter
}
