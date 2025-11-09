package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	AccessSecret  []byte
	RefreshSecret []byte
}

func LoadConfig() (*Config, error) {
	accessSecret := []byte(os.Getenv("ACCESS_TOKEN_SECRET"))
	refreshSecret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

	if len(accessSecret) == 0 || len(refreshSecret) == 0 {
		return nil, fmt.Errorf("ACCESS_TOKEN_SECRET and REFRESH_TOKEN_SECRET must be set")
	}

	return &Config{
		AccessTTL:     24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
	}, nil
}
