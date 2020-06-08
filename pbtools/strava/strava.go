package strava

import (
	"errors"
	"fmt"
	"time"

	strava "github.com/strava/go.strava"
	"github.com/tkrajina/gpxgo/gpx"
)

// GpxVersion GPX version
const GpxVersion = "1.1"

const gpxXMLNs = "http://www.topografix.com/GPX/1/1"
const gpsXMLNsXsi = "http://www.w3.org/2001/XMLSchema-instance"

// Strava represents a Strava API client
type Strava struct {
	HTTPPort     int
	ClientID     int
	ClientSecret string

	CurrentToken *AuthToken
}

// NewClient creates a new Strava API client
func NewClient(httpPort int, clientID int, clientSecret string) *Strava {
	strava.ClientId = clientID
	strava.ClientSecret = clientSecret

	return &Strava{
		HTTPPort:     httpPort,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// RetrieveAuthToken retrieves an authorization token to ensure we can query the APIs
func (s *Strava) RetrieveAuthToken() error {
	token, err := GetAccessToken(s.HTTPPort)
	if err != nil {
		return err
	}

	s.CurrentToken = token

	return nil
}

// GetActivityLink builds strava activity url from activity ID
func (s *Strava) GetActivityLink(activityID int64) string {
	return fmt.Sprintf("http://strava.com/%d", activityID)
}

// DownloadGPX downloads a gpx from a given Strava activity
func (s *Strava) DownloadGPX(activityID int64) (*gpx.GPX, error) {
	if s.CurrentToken == nil {
		return nil, errors.New("No auth token found, call RetrieveAuthToken() first")
	}

	client := strava.NewClient(s.CurrentToken.Token)
	aService := strava.NewActivitiesService(client)
	asService := strava.NewActivityStreamsService(client)

	activity, err := aService.Get(activityID).
		IncludeAllEfforts().
		Do()
	if err != nil {
		return nil, err
	}

	stream, err := asService.Get(activityID, []strava.StreamType{strava.StreamTypes.Location, strava.StreamTypes.Elevation, strava.StreamTypes.Time}).Do()
	if err != nil {
		return nil, err
	}

	points := make([]gpx.GPXPoint, len(stream.Location.Data))
	for i := 0; i < len(stream.Location.Data); i++ {
		points[i] = gpx.GPXPoint{
			Point: gpx.Point{
				Latitude:  stream.Location.Data[i][0],
				Longitude: stream.Location.Data[i][1],
				Elevation: *gpx.NewNullableFloat64(stream.Elevation.Data[i]),
			},
			Timestamp: activity.StartDateLocal.Add(time.Second * time.Duration(stream.Time.Data[i])),
		}
	}

	segment := gpx.GPXTrackSegment{
		Points: points,
	}

	track := gpx.GPXTrack{
		Name:     activity.Name,
		Segments: []gpx.GPXTrackSegment{segment},
	}

	g := gpx.GPX{
		XMLNs:        gpxXMLNs,
		XmlNsXsi:     gpsXMLNsXsi,
		XmlSchemaLoc: gpxXMLNs,

		Version: GpxVersion,
		Creator: "peakbagger-tools",
		Name:    activity.Name,
		Time:    &activity.StartDate,
		Tracks:  []gpx.GPXTrack{track},
	}

	return &g, nil

}
