package balldontlie

import (
	"net/http"
	"strings"
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
