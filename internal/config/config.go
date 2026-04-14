package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Token string // KAITEN_API_TOKEN — API bearer token
	URL   string // KAITEN_URL — e.g. "https://mycompany.kaiten.ru"
}

// BaseURL returns the full API base URL.
func (c *Config) BaseURL() string {
	return strings.TrimRight(c.URL, "/") + "/api/v1"
}

// Load reads configuration from environment variables.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	token := os.Getenv("KAITEN_API_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("KAITEN_API_TOKEN environment variable is required")
	}

	url := os.Getenv("KAITEN_URL")
	if url == "" {
		return nil, fmt.Errorf("KAITEN_URL environment variable is required")
	}

	return &Config{
		Token: token,
		URL:   url,
	}, nil
}
