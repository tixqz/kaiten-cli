package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Token  string // KAITEN_API_TOKEN — API bearer token
	URL    string // KAITEN_URL — e.g. "https://mycompany.kaiten.ru"
	DBPath string // KAITEN_DB_PATH — path to SQLite database
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

	dbPath, err := defaultDBPath()
	if err != nil {
		return nil, err
	}

	return &Config{
		Token:  token,
		URL:    url,
		DBPath: dbPath,
	}, nil
}

// LoadLocal reads configuration needed by local-only commands.
// It does not require API credentials because db/search commands use SQLite only.
func LoadLocal() (*Config, error) {
	dbPath, err := defaultDBPath()
	if err != nil {
		return nil, err
	}

	return &Config{
		Token:  os.Getenv("KAITEN_API_TOKEN"),
		URL:    os.Getenv("KAITEN_URL"),
		DBPath: dbPath,
	}, nil
}

func defaultDBPath() (string, error) {
	dbPath := os.Getenv("KAITEN_DB_PATH")
	if dbPath != "" {
		return dbPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(homeDir, ".kaiten", "kaiten.db"), nil
}
