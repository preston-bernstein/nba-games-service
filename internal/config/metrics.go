package config

// MetricsConfig controls telemetry export settings.
type MetricsConfig struct {
	Enabled      bool
	Port         string
	OtlpEndpoint string
	ServiceName  string
}

func loadMetrics() MetricsConfig {
	return MetricsConfig{
		Enabled:      boolEnvOrDefault(envMetricsOn, true),
		Port:         envOrDefault(envMetricsPort, defaultMetricsPort),
		OtlpEndpoint: envOrDefault(envOtelEndpoint, ""),
		ServiceName:  envOrDefault(envOtelService, "nba-games-service"),
	}
}
