package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigBaseURL(t *testing.T) {
	cfg := &Config{URL: "https://acme.kaiten.ru"}

	got := cfg.BaseURL()
	want := "https://acme.kaiten.ru/api/v1"

	if got != want {
		t.Fatalf("BaseURL() = %q, want %q", got, want)
	}
}

func TestConfigBaseURLTrailingSlash(t *testing.T) {
	cfg := &Config{URL: "https://acme.kaiten.ru/"}

	got := cfg.BaseURL()
	want := "https://acme.kaiten.ru/api/v1"

	if got != want {
		t.Fatalf("BaseURL() = %q, want %q", got, want)
	}
}

func TestLoadSuccess(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "token-value")
	t.Setenv("KAITEN_URL", "https://acme.kaiten.ru")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Token != "token-value" {
		t.Errorf("Token = %q, want %q", cfg.Token, "token-value")
	}
	if cfg.URL != "https://acme.kaiten.ru" {
		t.Errorf("URL = %q, want %q", cfg.URL, "https://acme.kaiten.ru")
	}
}

func TestLoadMissingToken(t *testing.T) {
	t.Setenv("KAITEN_URL", "https://acme.kaiten.ru")
	t.Setenv("KAITEN_API_TOKEN", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error, got nil")
	}
	if err.Error() != "KAITEN_API_TOKEN environment variable is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadMissingURL(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "token-value")
	t.Setenv("KAITEN_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error, got nil")
	}
	if err.Error() != "KAITEN_URL environment variable is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadDefaultDBPath(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "token")
	t.Setenv("KAITEN_URL", "https://example.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath == "" {
		t.Fatal("DBPath is empty")
	}
	if !strings.HasSuffix(cfg.DBPath, filepath.Join(".kaiten", "kaiten.db")) {
		t.Fatalf("DBPath = %q, want path ending with .kaiten/kaiten.db", cfg.DBPath)
	}
}

func TestLoadCustomDBPath(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "token")
	t.Setenv("KAITEN_URL", "https://example.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", "/tmp/custom-kaiten.db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath != "/tmp/custom-kaiten.db" {
		t.Fatalf("DBPath = %q", cfg.DBPath)
	}
}

func TestLoadLocalDoesNotRequireAPICredentials(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "")
	t.Setenv("KAITEN_URL", "")
	t.Setenv("KAITEN_DB_PATH", "/tmp/local-only-kaiten.db")

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error = %v", err)
	}
	if cfg.Token != "" {
		t.Fatalf("Token = %q, want empty", cfg.Token)
	}
	if cfg.URL != "" {
		t.Fatalf("URL = %q, want empty", cfg.URL)
	}
	if cfg.DBPath != "/tmp/local-only-kaiten.db" {
		t.Fatalf("DBPath = %q", cfg.DBPath)
	}
}
