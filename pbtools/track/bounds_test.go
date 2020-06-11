package track_test

import (
	"peakbagger-tools/pbtools/track"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendBounds(t *testing.T) {
	require := require.New(t)
	bounds := track.Bounds{
		MinLat: 48.1344333,
		MaxLat: 48.2714123,
		MinLng: -121.8064235,
		MaxLng: -121.6771830,
	}

	newBounds := bounds.Extend(0.01)

	require.Equal(48.1244333, newBounds.MinLat)
	require.Equal(48.2814123, newBounds.MaxLat)
	require.Equal(-121.8164235, newBounds.MinLng)
	require.Equal(-121.6671830, newBounds.MaxLng)
}
