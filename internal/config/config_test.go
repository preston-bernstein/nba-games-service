package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %s, got %s", defaultPort, cfg.Port)
	}
	if cfg.PollInterval != defaultPollInterval {
		t.Fatalf("expected default poll interval %s, got %s", defaultPollInterval, cfg.PollInterval)
	}
	if cfg.Provider != defaultProvider {
		t.Fatalf("expected default provider %s, got %s", defaultProvider, cfg.Provider)
	}
	if cfg.Balldontlie.BaseURL != defaultBdlBaseURL {
		t.Fatalf("expected default balldontlie base url %s, got %s", defaultBdlBaseURL, cfg.Balldontlie.BaseURL)
	}
	if cfg.Balldontlie.APIKey != "" {
		t.Fatalf("expected empty balldontlie api key by default, got %s", cfg.Balldontlie.APIKey)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv(envPort, "5000")
	t.Setenv(envPollInterval, "45s")
	t.Setenv(envProvider, "balldontlie")
	t.Setenv(envBdlBaseURL, "http://example.com/api")
	t.Setenv(envBdlAPIKey, "secret-key")

	cfg := Load()

	if cfg.Port != "5000" {
		t.Fatalf("expected port 5000, got %s", cfg.Port)
	}
	if cfg.PollInterval != 45*time.Second {
		t.Fatalf("expected poll interval 45s, got %s", cfg.PollInterval)
	}
	if cfg.Provider != "balldontlie" {
		t.Fatalf("expected provider balldontlie, got %s", cfg.Provider)
	}
	if cfg.Balldontlie.BaseURL != "http://example.com/api" {
		t.Fatalf("expected balldontlie base url override, got %s", cfg.Balldontlie.BaseURL)
	}
	if cfg.Balldontlie.APIKey != "secret-key" {
		t.Fatalf("expected balldontlie api key override, got %s", cfg.Balldontlie.APIKey)
	}
}

func TestLoadInvalidDurationFallsBack(t *testing.T) {
	t.Setenv(envPollInterval, "not-a-duration")

	cfg := Load()

	if cfg.PollInterval != defaultPollInterval {
		t.Fatalf("expected default poll interval on invalid value, got %s", cfg.PollInterval)
	}
}

func TestLoadNonPositiveDurationFallsBack(t *testing.T) {
	t.Setenv(envPollInterval, "0s")

	cfg := Load()

	if cfg.PollInterval != defaultPollInterval {
		t.Fatalf("expected default poll interval on non-positive value, got %s", cfg.PollInterval)
	}
}
