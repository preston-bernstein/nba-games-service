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

func TestResolveHTTPClientDefaultsTimeout(t *testing.T) {
	client := resolveHTTPClient(nil)
	httpClient, ok := client.(*http.Client)
	if !ok {
		t.Fatalf("expected *http.Client, got %T", client)
	}
	if httpClient.Timeout != defaultHTTPTimeout {
		t.Fatalf("expected timeout %s, got %s", defaultHTTPTimeout, httpClient.Timeout)
	}
}

func TestResolveHTTPClientUsesProvidedClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	client := resolveHTTPClient(custom)
	if client != custom {
		t.Fatalf("expected provided client to be used")
	}
}
