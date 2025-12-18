package balldontlie

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFetchGamesHitsAPIAndMapsResponse(t *testing.T) {
	fixed := time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC) // should still yield 2024-01-01 in America/New_York
	var capturedAuth string
	var capturedQuery string

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/games" {
			t.Fatalf("expected /games path, got %s", req.URL.Path)
		}
		capturedQuery = req.URL.RawQuery
		capturedAuth = req.Header.Get("Authorization")

		body := `{
			"data": [
				{
					"id": 10,
					"date": "2024-01-02T15:00:00Z",
					"status": "Final",
					"home_team": { "id": 1, "full_name": "Home Team" },
					"visitor_team": { "id": 2, "full_name": "Away Team" },
					"home_team_score": 110,
					"visitor_team_score": 102,
					"season": 2023
				}
			]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(Config{
		BaseURL:    "http://example.com",
		APIKey:     "secret",
		HTTPClient: &http.Client{Transport: rt},
		Timezone:   "America/New_York",
	})
	client.now = func() time.Time { return fixed }

	games, err := client.FetchGames(context.Background(), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedAuth != "Bearer secret" {
		t.Fatalf("expected authorization header, got %s", capturedAuth)
	}
	q, err := url.ParseQuery(capturedQuery)
	if err != nil {
		t.Fatalf("failed parsing query %s: %v", capturedQuery, err)
	}
	if q.Get("per_page") != "100" {
		t.Fatalf("expected per_page=100, got %s", q.Get("per_page"))
	}
	if q.Get("dates[]") != "2024-01-01" {
		t.Fatalf("expected date=2024-01-01 in NY, got %s", q.Get("dates[]"))
	}
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}

	game := games[0]
	if game.ID != "balldontlie-10" || game.Provider != "balldontlie" {
		t.Fatalf("unexpected game identifiers %+v", game)
	}
	if game.Score.Home != 110 || game.Score.Away != 102 {
		t.Fatalf("unexpected scores %+v", game.Score)
	}
	if game.Status != "FINAL" {
		t.Fatalf("unexpected status %s", game.Status)
	}
	if game.Meta.UpstreamGameID != 10 || game.Meta.Season != "2023" {
		t.Fatalf("unexpected meta %+v", game.Meta)
	}
}

func TestFetchGamesHandlesNon200(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_ = req
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("boom")),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(Config{
		BaseURL:    "http://example.com",
		HTTPClient: &http.Client{Transport: rt},
	})
	client.now = func() time.Time { return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) }

	if _, err := client.FetchGames(context.Background(), ""); err == nil {
		t.Fatal("expected error on non-200 response")
	}
}

func TestFetchGamesHandlesDecodeError(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_ = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{bad json")),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(Config{
		BaseURL:    "http://example.com",
		HTTPClient: &http.Client{Transport: rt},
	})
	client.now = func() time.Time { return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) }

	if _, err := client.FetchGames(context.Background(), ""); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestNewClientSetsDefaultHTTPClient(t *testing.T) {
	c := NewClient(Config{})
	httpClient, ok := c.httpClient.(*http.Client)
	if !ok {
		t.Fatalf("expected default http client")
	}
	if httpClient.Timeout == 0 {
		t.Fatalf("expected timeout to be set on default http client")
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
