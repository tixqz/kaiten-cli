package config

import "testing"

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
