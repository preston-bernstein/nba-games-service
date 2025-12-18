package config

import "time"

const (
	envBdlBaseURL  = "BALDONTLIE_BASE_URL"
	envBdlAPIKey   = "BALDONTLIE_API_KEY"
	envBdlTimezone = "BALDONTLIE_TIMEZONE"

	defaultBdlBaseURL  = "https://api.balldontlie.io/v1"
	defaultBdlTimezone = "America/New_York"
)

// BalldontlieConfig controls how we talk to the balldontlie API.
type BalldontlieConfig struct {
	BaseURL  string
	APIKey   string
	Timezone string
}

func loadBalldontlie() BalldontlieConfig {
	return BalldontlieConfig{
		BaseURL:  envOrDefault(envBdlBaseURL, defaultBdlBaseURL),
		APIKey:   envOrDefault(envBdlAPIKey, ""),
		Timezone: envOrDefault(envBdlTimezone, defaultBdlTimezone),
	}
}

// Ensure Duration alias is used to avoid unused import of time in constants.
var _ = time.Second
