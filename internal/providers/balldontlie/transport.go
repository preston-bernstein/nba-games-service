package balldontlie

import (
	"net/http"
	"strings"
	"time"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func resolveHTTPClient(client *http.Client) httpDoer {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: defaultHTTPTimeout}
}

func normalizeBaseURL(raw string) string {
	if raw == "" {
		raw = defaultBaseURL
	}
	return strings.TrimSuffix(raw, "/")
}

func resolveLocation(name string) *time.Location {
	if name == "" {
		name = defaultTimezone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
