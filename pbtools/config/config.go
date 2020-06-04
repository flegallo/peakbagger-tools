package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"

	"github.com/github/go-config"
)

// Config holds application configuration, including statting and tracing configs.
type Config struct {
	HTTPPort         int
	StravaActivityID int64

	StravaClientID int    `config:"<YOUR_CLIENT_ID>,env=STRAVA_CLIENT_ID"`
	StravaSecretID string `config:"<YOUR_SECRET_ID>,env=STRAVA_SECRET_ID"`

	PeakBaggerUsername string
	PeakBaggerPassword string
}

// Load parses configuration from the environment and places it in a newly
// allocated Config struct.
func Load() (*Config, error) {
	port := flag.Int("port", 8080, "port number to run http server on")
	peakBaggerUsername := flag.String("pbUser", "", "Peakbagger username")
	peakBaggerPassword := flag.String("pbPwd", "", "Peakbagger password")
	activity := flag.String("activity", "", "Strava activity link")

	flag.Parse()

	activityID, err := parseActivityID(*activity)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPPort:           *port,
		StravaActivityID:   activityID,
		PeakBaggerUsername: *peakBaggerUsername,
		PeakBaggerPassword: *peakBaggerPassword,
	}

	if err := config.Load(cfg); err != nil {
		return nil, err
	}

	if cfg.StravaClientID <= 0 || cfg.StravaSecretID == "" {
		return nil, errors.New("please provide your Strava's client_id and client_secret")
	}

	return cfg, nil
}

func parseActivityID(activityLink string) (int64, error) {
	parts := strings.Split(activityLink, "/")
	if len(parts) > 0 {
		activityID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
		if err != nil {
			return -1, errors.New("wrong activity link format")
		}

		return activityID, nil
	}

	return -1, errors.New("wrong activity link format")
}
