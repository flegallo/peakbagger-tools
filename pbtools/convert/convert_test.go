package convert_test

import (
	"peakbagger-tools/pbtools/convert"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToFeet(t *testing.T) {
	require := require.New(t)

	tests := map[string]struct {
		input float64
		want  float64
	}{
		"simple": {input: 1000, want: 3280.84},
		"zero":   {input: 0, want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			feet := convert.ToFeet(tc.input)
			require.Equal(tc.want, feet)
		})
	}
}

func TestToMiles(t *testing.T) {
	require := require.New(t)

	tests := map[string]struct {
		input float64
		want  float64
	}{
		"simple": {input: 1000, want: 0.6213712},
		"zero":   {input: 0, want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			feet := convert.ToMiles(tc.input)
			require.Equal(tc.want, feet)
		})
	}
}

func TestToHoursMinSec(t *testing.T) {
	require := require.New(t)

	tests := map[string]struct {
		input time.Duration
		want  []int
	}{
		"days":    {input: 2 * 24 * time.Hour, want: []int{2, 0, 0}},
		"hours":   {input: 18 * time.Hour, want: []int{0, 18, 0}},
		"minutes": {input: 24 * time.Minute, want: []int{0, 0, 24}},
		"mix":     {input: (1440 + 600 + 43) * time.Minute, want: []int{1, 10, 43}},
		"zero":    {input: 0, want: []int{0, 0, 0}},
		"invalid": {input: -5 * time.Minute, want: []int{0, 0, 0}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d, h, m := convert.ToDaysHoursMin(tc.input)
			require.Equal(tc.want[0], d)
			require.Equal(tc.want[1], h)
			require.Equal(tc.want[2], m)
		})
	}
}

func TestFtoan(t *testing.T) {
	require := require.New(t)

	tests := map[string]struct {
		input float64
		want  string
	}{
		"floor":       {input: 2524.13242435, want: "2524"},
		"ceil":        {input: 2524.72342341, want: "2525"},
		"almost_zero": {input: 0.11, want: "0"},
		"zero":        {input: 0.0, want: "0"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := convert.Ftoan(tc.input)
			require.Equal(tc.want, s)
		})
	}
}
