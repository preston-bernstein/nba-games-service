package config

import (
	"os"
	"time"
)

// Duration wraps time.Duration for clearer type usage in Config.
type Duration = time.Duration

func envOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func durationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
