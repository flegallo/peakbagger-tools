package gpxutils

import (
	"math"

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

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(p1, p2 gpx.Location) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = p1.GetLatitude() * math.Pi / 180
	lo1 = p1.GetLongitude() * math.Pi / 180
	la2 = p2.GetLatitude() * math.Pi / 180
	lo2 = p2.GetLongitude() * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// GetClosestPoint returns the closest point from a point to a segment line
// Points are GPS coordinates. This only works if the 3 points are at close distance
// from each other as we don't take into account the Earth curvature. In the context of a GPX,
// 2 following points are meters away from each other, so it's fair to assume that the straight line between
// each other is the shortest distance.
//
//  A ----I-------- B     		      	A/I ------------ B             A ------------ B/I
//
//        P					    P																P
//
//
// A: Line segment point A
// B: Line segment point B
// P: point outside line
// I: Closest point to P on the AB segment
//
func GetClosestPoint(la, lb, p gpx.Location) gpx.Location {

	lax := la.GetLongitude()
	lay := la.GetLatitude()
	lbx := lb.GetLongitude()
	lby := lb.GetLatitude()
	px := p.GetLongitude()
	py := p.GetLatitude()

	atp := []float64{px - lax, py - lay}   // Storing vector A->P
	atb := []float64{lbx - lax, lby - lay} // Storing vector A->B

	atb2 := math.Pow(atb[0], 2) + math.Pow(atb[1], 2)

	if atb2 == 0.0 {
		return la
	}

	atpDotAtb := atp[0]*atb[0] + atp[1]*atb[1] // dot product of the 2 vectors

	t := atpDotAtb / atb2 // normalized distance

	t = math.Max(0, math.Min(1, t))

	return &gpx.Point{
		Longitude: lax + atb[0]*t,
		Latitude:  lay + atb[1]*t,
	}
}

// GetShortestDistanceFromPointToLine returns the shortest distance in meters from a point to a given line segment.
//
// Points are GPS coordinates. This only works if the 3 points are at close distance
// from each other as we don't take into account the Earth curvature. In the context of a GPX,
// 2 following points are meters away from each other, so it's fair to assume that the straight line between
// each other is the shortest distance.
//
func GetShortestDistanceFromPointToLine(la, lb, p gpx.Location) float64 {
	closestPoint := GetClosestPoint(la, lb, p)
	return Distance(closestPoint, p)
}

// FindShortestDistanceToTrack returns the distance in meters to the closest point on the track
// This function can search the distance of multiple points
func FindShortestDistanceToTrack(g *gpx.GPX, pts []gpx.Location) []float64 {

	results := make([]float64, len(pts))
	for i := range results {
		results[i] = math.MaxFloat64
	}

	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			for j, point := range segment.Points {
				for i, p := range pts {
					if j > 0 {
						currentDistance := GetShortestDistanceFromPointToLine(&segment.Points[j-1], &point, p)
						if currentDistance < results[i] {
							results[i] = currentDistance
						}
					}
				}
			}
		}
	}

	return results
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
