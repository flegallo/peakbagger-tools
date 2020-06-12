package config

import (
	"errors"
	"flag"

	"github.com/github/go-config"
)

// Config holds application configuration, including statting and tracing configs.
type Config struct {
	HTTPPort int

	StravaClientID int    `config:"<YOUR_CLIENT_ID>,env=STRAVA_CLIENT_ID"`
	StravaSecretID string `config:"<YOUR_CLIENT_SECRET>,env=STRAVA_SECRET_ID"`

	PeakBaggerUsername string
	PeakBaggerPassword string
}

// Load parses configuration from the environment and places it in a newly
// allocated Config struct.
func Load() (*Config, error) {
	port := flag.Int("port", 8080, "port number to run http server on")

	flag.Parse()

	cfg := &Config{
		HTTPPort: *port,
	}

	if err := config.Load(cfg); err != nil {
		return nil, err
	}

	if cfg.StravaClientID <= 0 || cfg.StravaSecretID == "" {
		return nil, errors.New("please provide your Strava's client_id and client_secret")
	}

	return cfg, nil
}
