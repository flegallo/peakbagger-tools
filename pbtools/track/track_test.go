package track_test

import (
	"math"
	"peakbagger-tools/pbtools/track"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tkrajina/gpxgo/gpx"
)

func TestGetClosestPoint(t *testing.T) {
	require := require.New(t)

	pts := []float64{47.58358925699506, -121.95062398910524, 47.58878498470957, -121.94446563720703, 47.58622336725498, -121.9381356239319, 47.59581793370288, -121.93571090698244}
	tr := getTrack(pts)

	tests := map[string]struct {
		input []float64
		want  []float64
	}{
		"point_between_mid_seg":     {input: []float64{47.5896, -121.9411}, want: []float64{47.58878498470957, -121.94446563720703, 1}},
		"point_between_first_seg":   {input: []float64{47.5898, -121.9519}, want: []float64{47.58878498470957, -121.94446563720703, 1}},
		"point_outside_all_segment": {input: []float64{47.5980, -121.9299}, want: []float64{47.59581793370288, -121.93571090698244, 3}},
		"far_point":                 {input: []float64{47.5733, -121.8346}, want: []float64{47.58622336725498, -121.9381356239319, 2}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, i := tr.GetClosestPoint(track.Point{Latitude: tc.input[0], Longitude: tc.input[1]})
			require.Equal(tc.want[0], p.Latitude)
			require.Equal(tc.want[1], p.Longitude)
			require.Equal(int(tc.want[2]), i)
		})
	}
}

func TestFindShortestDistanceFromPoint(t *testing.T) {
	require := require.New(t)

	pts := []float64{47.58358925699506, -121.95062398910524, 47.58878498470957, -121.94446563720703, 47.58622336725498, -121.9381356239319, 47.59581793370288, -121.93571090698244}
	tr := getTrack(pts)

	tests := map[string]struct {
		input []float64
		want  int
	}{
		"point_between_mid_seg":     {input: []float64{47.5896, -121.9411}, want: 208},
		"point_between_first_seg":   {input: []float64{47.5898, -121.9519}, want: 507},
		"point_outside_all_segment": {input: []float64{47.5980, -121.9299}, want: 499},
		"far_point":                 {input: []float64{47.5733, -121.8346}, want: 7907},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := tr.GetShortestDistanceFromPoint(track.Point{Latitude: tc.input[0], Longitude: tc.input[1]})
			require.Equal(tc.want, int(math.Round(d)))
		})
	}
}

func TestSplit(t *testing.T) {
	require := require.New(t)

	pts := []float64{47.58358925699506, -121.95062398910524, 47.58878498470957, -121.94446563720703, 47.58622336725498, -121.9381356239319, 47.59581793370288, -121.93571090698244}
	tr := getTrack(pts)

	tr1, tr2 := tr.Split(1)

	require.Equal(2, len(*tr1.Points()))
	require.Equal(2, len(*tr2.Points()))
	require.Equal(47.58358925699506, (*tr1.Points())[0].Latitude)
	require.Equal(-121.95062398910524, (*tr1.Points())[0].Longitude)
	require.Equal(47.58622336725498, (*tr2.Points())[0].Latitude)
	require.Equal(-121.9381356239319, (*tr2.Points())[0].Longitude)
}

func TestSimplify(t *testing.T) {
	require := require.New(t)

	pts := []float64{
		47.57766745244047, -121.9222328066826,
		47.57766745244047, -121.92229181528094,
		47.57766202426469, -121.9223776459694,
		47.57766926183228, -121.92240983247758,
		47.57767107122401, -121.92243933677675,
		47.57766564304861, -121.92247420549394,
		47.5776638336567, -121.92250907421113,
		47.57766926183228, -121.92255467176439,
		47.57783753499646, -121.92247152328491,
		47.577676499398855, -121.92256540060045,
		47.57767830879033, -121.92258685827257,
		47.57766564304861, -121.9226109981537,
		47.57764393034137, -121.92261904478075,
		47.577622217625134, -121.92261636257173,
	}
	tr := getTrack(pts)

	tr2 := tr.Simplify(15)

	require.Equal(14, len(*tr.Points()))
	require.True(len(*tr2.Points()) < len(*tr.Points()))
}

func TestBounds(t *testing.T) {
	require := require.New(t)

	pts := []float64{47.58358925699506, -121.95062398910524, 47.58878498470957, -121.94446563720703, 47.58622336725498, -121.9381356239319, 47.59581793370288, -121.93571090698244}
	tr := getTrack(pts)

	b := tr.Bounds()

	require.Equal(47.58358925699506, b.MinLat)
	require.Equal(47.59581793370288, b.MaxLat)
	require.Equal(-121.95062398910524, b.MinLng)
	require.Equal(-121.93571090698244, b.MaxLng)
}

func TestStats(t *testing.T) {
	require := require.New(t)

	points := []track.Point{
		track.Point{
			Latitude:  47.58358925699506,
			Longitude: -121.95062398910524,
			Elevation: 100,
			Time:      time.Date(2009, time.November, 10, 14, 20, 0, 0, time.UTC),
		},
		track.Point{
			Latitude:  47.58878498470957,
			Longitude: -121.94446563720703,
			Elevation: 200,
			Time:      time.Date(2009, time.November, 10, 14, 21, 0, 0, time.UTC),
		},
		track.Point{
			Latitude:  47.58622336725498,
			Longitude: -121.9381356239319,
			Elevation: 300,
			Time:      time.Date(2009, time.November, 10, 14, 22, 0, 0, time.UTC),
		},
		track.Point{
			Latitude:  47.59581793370288,
			Longitude: -121.93571090698244,
			Elevation: 400,
			Time:      time.Date(2009, time.November, 10, 14, 23, 0, 0, time.UTC),
		},
		track.Point{
			Latitude:  47.62581793370288,
			Longitude: -121.93371090698244,
			Elevation: 300,
			Time:      time.Date(2009, time.November, 10, 14, 24, 0, 0, time.UTC),
		},
	}

	tr := getTrack2(points)
	stats := tr.Stats()

	require.Equal(100.0, stats.StartElevation)
	require.Equal(300.0, stats.EndElevation)
	require.Equal(300.0, stats.ElevationGain)
	require.Equal(100.0, stats.ElevationLoss)
	require.Equal(4*time.Minute, stats.Duration)
	require.Equal(5733, int(math.Round(stats.Distance)))
}

func getTrack(pts []float64) track.Track {
	ln := len(pts)
	if ln%2 != 0 {
		panic("must provide a pair number of points")
	}

	gPts := make([]gpx.GPXPoint, ln/2)
	j := 0
	for i := 0; i < ln-1; i += 2 {
		gPts[j] = gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  pts[i],
				Longitude: pts[i+1],
			},
		}
		j++
	}

	return track.New(&gPts)
}

func getTrack2(pts []track.Point) track.Track {

	gPts := make([]gpx.GPXPoint, len(pts))
	for i, p := range pts {
		gPts[i] = gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				Elevation: *gpx.NewNullableFloat64(p.Elevation),
			},
			Timestamp: p.Time,
		}
	}

	return track.New(&gPts)
}
