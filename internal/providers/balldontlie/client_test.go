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
	var capturedQueries []string

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/games" {
			t.Fatalf("expected /games path, got %s", req.URL.Path)
		}
		capturedQueries = append(capturedQueries, req.URL.RawQuery)
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
		if len(capturedQueries) == 1 {
			body = `{
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
				],
				"meta": {
					"total_pages": 2
				}
			}`
		} else {
			body = `{
				"data": [
					{
						"id": 11,
						"date": "2024-01-03T15:00:00Z",
						"status": "Final",
						"home_team": { "id": 3, "full_name": "Another Team" },
						"visitor_team": { "id": 4, "full_name": "Away Team 2" },
						"home_team_score": 120,
						"visitor_team_score": 115,
						"season": 2023
					}
				],
				"meta": {
					"total_pages": 2
				}
			}`
		}
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
		MaxPages:   2,
	})
	client.now = func() time.Time { return fixed }

	games, err := client.FetchGames(context.Background(), "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedAuth != "Bearer secret" {
		t.Fatalf("expected authorization header, got %s", capturedAuth)
	}
	if len(capturedQueries) != 2 {
		t.Fatalf("expected 2 requests (pagination), got %d", len(capturedQueries))
	}
	q, err := url.ParseQuery(capturedQueries[0])
	if err != nil {
		t.Fatalf("failed parsing query %s: %v", capturedQueries[0], err)
	}
	if q.Get("per_page") != "100" {
		t.Fatalf("expected per_page=100, got %s", q.Get("per_page"))
	}
	if q.Get("dates[]") != "2024-01-01" {
		t.Fatalf("expected date=2024-01-01 in NY, got %s", q.Get("dates[]"))
	}
	if q.Get("page") != "1" {
		t.Fatalf("expected page=1, got %s", q.Get("page"))
	}
	if len(games) != 2 {
		t.Fatalf("expected games from both pages, got %d", len(games))
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

	if _, err := client.FetchGames(context.Background(), "", ""); err == nil {
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

	if _, err := client.FetchGames(context.Background(), "", ""); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestFetchGamesRespectsMaxPagesCap(t *testing.T) {
	calls := 0
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		body := `{
			"data": [
				{
					"id": 1,
					"date": "2024-01-01T00:00:00Z",
					"status": "Final",
					"home_team": { "id": 1, "full_name": "Home" },
					"visitor_team": { "id": 2, "full_name": "Away" },
					"home_team_score": 10,
					"visitor_team_score": 5,
					"season": 2023
				}
			],
			"meta": { "total_pages": 10 }
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(Config{
		BaseURL:    "http://example.com",
		HTTPClient: &http.Client{Transport: rt},
		MaxPages:   1,
	})

	games, err := client.FetchGames(context.Background(), "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}
	if calls != 1 {
		t.Fatalf("expected to stop after max pages, got %d calls", calls)
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
