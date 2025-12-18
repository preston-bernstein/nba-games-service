package config

import "time"

const (
	envPort         = "PORT"
	envPollInterval = "POLL_INTERVAL"
	envProvider     = "PROVIDER"

	defaultPort         = "4000"
	defaultPollInterval = 30 * Duration(time.Second)
	defaultProvider     = "fixture"
)
