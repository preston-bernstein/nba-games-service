package balldontlie

import (
	"net/http"
	"testing"
	"time"
)

func TestNormalizeBaseURLTrimsTrailingSlashAndDefaults(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", defaultBaseURL},
		{"https://api.example.com/", "https://api.example.com"},
		{"https://api.example.com", "https://api.example.com"},
	}

	for _, c := range cases {
		if got := normalizeBaseURL(c.input); got != c.expected {
			t.Fatalf("expected %s, got %s", c.expected, got)
		}
	}
}

func TestResolveHTTPClientTimeouts(t *testing.T) {
	cases := []struct {
		name        string
		client      *http.Client
		timeout     time.Duration
		expected    time.Duration
		expectSame  bool
	}{
		{
			name:     "defaults",
			client:   nil,
			timeout:  0,
			expected: defaultHTTPTimeout,
		},
		{
			name:       "uses provided client",
			client:     &http.Client{Timeout: 5 * time.Second},
			timeout:    0,
			expectSame: true,
		},
		{
			name:     "overrides timeout",
			client:   nil,
			timeout:  2 * time.Second,
			expected: 2 * time.Second,
		},
	}

	for _, tc := range cases {
		c := resolveHTTPClient(tc.client, tc.timeout)
		if tc.expectSame {
			if c != tc.client {
				t.Fatalf("%s: expected provided client to be used", tc.name)
			}
			continue
		}
		httpClient, ok := c.(*http.Client)
		if !ok {
			t.Fatalf("%s: expected *http.Client, got %T", tc.name, c)
		}
		if httpClient.Timeout != tc.expected {
			t.Fatalf("%s: expected timeout %s, got %s", tc.name, tc.expected, httpClient.Timeout)
		}
	}
}

func TestResolveLocationDefaultsAndFallback(t *testing.T) {
	loc := resolveLocation("")
	if loc.String() != defaultTimezone {
		t.Fatalf("expected default timezone %s, got %s", defaultTimezone, loc.String())
	}

	utc := resolveLocation("UTC")
	if utc.String() != "UTC" {
		t.Fatalf("expected UTC, got %s", utc.String())
	}

	if fallback := resolveLocation("Not/AZone"); fallback.String() != "UTC" {
		t.Fatalf("expected fallback to UTC, got %s", fallback.String())
	}
}

func TestResolveMaxPages(t *testing.T) {
	if got := resolveMaxPages(0); got != defaultMaxPages {
		t.Fatalf("expected default max pages %d, got %d", defaultMaxPages, got)
	}
	if got := resolveMaxPages(3); got != 3 {
		t.Fatalf("expected max pages 3, got %d", got)
	}
}
