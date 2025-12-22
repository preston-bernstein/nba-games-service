package http

import "testing"

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"/health", "/health"},
		{"/games", "/games/today"},
		{"/games/today", "/games/today"},
		{"/games/123", "/games/:id"},
		{"/games/123?foo=bar", "/games/:id"},
		{"/other", "/other"},
		{"", ""},
	}

	for _, c := range cases {
		if got := normalizePath(c.in); got != c.want {
			t.Fatalf("normalizePath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeQuery(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"page=1&limit=10", "page=1&limit=10"},
		{"api_key=secret", "[redacted]"},
		{"token=abc123", "[redacted]"},
		{"password=hunter2", "[redacted]"},
	}

	for _, c := range cases {
		if got := sanitizeQuery(c.raw); got != c.want {
			t.Fatalf("sanitizeQuery(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}
