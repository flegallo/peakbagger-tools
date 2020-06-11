package track

import (
	"math"
	"time"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/tkrajina/gpxgo/gpx"
)

// Track represents a gps track made of a serie of points in a 3D dimension
// in order to get the most accurate distances and positioning on the Earth.
type Track struct {
	Points []Point

	polyline *s2.Polyline
	segment  gpx.GPXTrackSegment
}

// LatLng latlng
type LatLng interface {
	Lat() float64
	Lng() float64
}

// Stats track statistics
type Stats struct {
	Duration       time.Duration
	ElevationGain  float64
	ElevationLoss  float64
	StartElevation float64
	EndElevation   float64
	Distance       float64
}

const earthRadius = 6378100
const elevationChangeThreshold = 18

// New Create a track from the given gpx points
func New(pts *[]gpx.GPXPoint) *Track {

	pPts := make([]s2.LatLng, len(*pts))
	gPts := make([]Point, len(*pts))
	for i, p := range *pts {
		point := Point{
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
			Elevation: p.Elevation.Value(),
			Time:      p.Timestamp,
		}
		pPts[i] = toS2LatLng(point)
		gPts[i] = point
	}

	p := s2.PolylineFromLatLngs(pPts)

	return &Track{
		polyline: p,
		Points:   gPts,
		segment:  gpx.GPXTrackSegment{Points: *pts},
	}
}

// GetClosestPoint returns the closest gpx point(along with its position on the track) on the track
func (t *Track) GetClosestPoint(pt LatLng) (Point, int) {
	p := s2.PointFromLatLng(toS2LatLng(pt))

	projectedPt, index := t.polyline.Project(p)
	l := len(t.Points)

	if index >= l {
		return t.Points[l-1], l - 1
	}

	var closestPtIndex = index - 1
	pts := *t.polyline
	if projectedPt.Distance(pts[index]) < projectedPt.Distance(pts[index-1]) {
		closestPtIndex = index
	}

	return t.Points[closestPtIndex], closestPtIndex
}

// GetShortestDistanceFromPoint returns the shortest distance from the given point to the track
func (t *Track) GetShortestDistanceFromPoint(pt LatLng) float64 {
	ptLatLng := toS2LatLng(pt)
	p := s2.PointFromLatLng(ptLatLng)

	projectedPoint, _ := t.polyline.Project(p)
	projectedLatLng := s2.LatLngFromPoint(projectedPoint)

	d := ptLatLng.Distance(projectedLatLng)

	return d.Radians() * earthRadius
}

// Stats retrieves statistics from the track
func (t *Track) Stats() Stats {
	tb := t.segment.TimeBounds()
	gain, loss := t.elevationGainLoss(elevationChangeThreshold)

	return Stats{
		Duration:       tb.EndTime.Sub(tb.StartTime),
		ElevationGain:  gain,
		ElevationLoss:  loss,
		StartElevation: t.Points[0].Elevation,
		EndElevation:   t.Points[len(t.Points)-1].Elevation,
		Distance:       t.segment.Length3D(),
	}
}

// Split splits the track in two tracks from the given point. Point at index `index` will remain in first part.
func (t *Track) Split(index int) (*Track, *Track) {
	s1, s2 := t.segment.Split(index)

	t1 := New(&s1.Points)
	t2 := New(&s2.Points)

	return t1, t2
}

// ReduceTrackPoints reduce number of points in the track
func (t *Track) ReduceTrackPoints(maxPoints float64) *Track {
	minDistance := math.Ceil(float64(len(t.Points)) / float64(maxPoints))
	s := t.segment
	s.ReduceTrackPoints(minDistance)
	return t
}

// Bounds returns the boundaries of the track
func (t *Track) Bounds() Bounds {
	b := t.segment.Bounds()
	return Bounds{
		MinLat: b.MinLatitude,
		MinLng: b.MinLongitude,
		MaxLat: b.MaxLatitude,
		MaxLng: b.MaxLongitude,
	}
}

func (t *Track) elevationGainLoss(threshold float64) (float64, float64) {
	elevations := t.segment.Elevations()
	selectedElevations := []float64{}
	i := 0
	for _, e := range elevations {
		if e.NotNull() {
			if i == 0 || math.Abs(e.Value()-selectedElevations[i-1]) > threshold {
				selectedElevations = append(selectedElevations, e.Value())
				i++
			}
		}
	}

	var gain float64
	var loss float64

	for i := 1; i < len(selectedElevations); i++ {
		d := selectedElevations[i] - selectedElevations[i-1]
		if d > 0.0 {
			gain += d
		} else {
			loss -= d
		}
	}

	return gain, loss
}

func toS2LatLng(p LatLng) s2.LatLng {
	return s2.LatLng{
		Lat: s1.Angle(p.Lat()) * s1.Degree,
		Lng: s1.Angle(p.Lng()) * s1.Degree,
	}
}
