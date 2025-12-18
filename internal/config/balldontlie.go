package config

import "time"

const (
	envBdlBaseURL = "BALDONTLIE_BASE_URL"
	envBdlAPIKey  = "BALDONTLIE_API_KEY"

	defaultBdlBaseURL = "https://www.balldontlie.io/api/v1"
)

// BalldontlieConfig controls how we talk to the balldontlie API.
type BalldontlieConfig struct {
	BaseURL string
	APIKey  string
}

func loadBalldontlie() BalldontlieConfig {
	return BalldontlieConfig{
		BaseURL: envOrDefault(envBdlBaseURL, defaultBdlBaseURL),
		APIKey:  envOrDefault(envBdlAPIKey, ""),
	}
}

// Ensure Duration alias is used to avoid unused import of time in constants.
var _ = time.Second
