package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nba-data-service/internal/domain"
	"nba-data-service/internal/http/middleware"
	"nba-data-service/internal/poller"
	"nba-data-service/internal/testutil"
)

type stubSnapshots struct {
	resp domain.TodayResponse
	err  error
}

func (s *stubSnapshots) LoadGames(date string) (domain.TodayResponse, error) {
	_ = date
	return s.resp, s.err
}

func TestHealth(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/health", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp map[string]string
	testutil.DecodeJSON(t, rr, &resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
}

func TestHealthShuttingDownReturnsServiceUnavailable(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	rr := testutil.ServeRequest(http.HandlerFunc(h.Health), req)

	testutil.AssertStatus(t, rr, http.StatusServiceUnavailable)
	var resp map[string]string
	testutil.DecodeJSON(t, rr, &resp)
	if resp["error"] != "shutting down" {
		t.Fatalf("unexpected error %q", resp["error"])
	}
}

func TestGamesToday(t *testing.T) {
	game := testutil.SampleGame("game-1")
	game.StartTime = time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	h := NewHandler(testutil.NewServiceWithGames([]domain.Game{game}), nil, nil, nil)
	fixedNow := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return fixedNow }

	rr := testutil.Serve(h, http.MethodGet, "/games/today", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp domain.TodayResponse
	testutil.DecodeJSON(t, rr, &resp)

	if resp.Date != "2024-01-02" {
		t.Fatalf("expected date 2024-01-02, got %s", resp.Date)
	}

	if len(resp.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(resp.Games))
	}

	if resp.Games[0].ID != "game-1" {
		t.Fatalf("unexpected game id %s", resp.Games[0].ID)
	}
}

func TestGamesTodayWithDateUsesProvider(t *testing.T) {
	snaps := &stubSnapshots{
		resp: testutil.SampleTodayResponse("2024-02-01", "snapshot-game"),
	}

	h := NewHandler(testutil.NewServiceWithGames(nil), snaps, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/games?date=2024-02-01", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp domain.TodayResponse
	testutil.DecodeJSON(t, rr, &resp)

	if resp.Date != "2024-02-01" {
		t.Fatalf("expected date to reflect query param, got %s", resp.Date)
	}
	if len(resp.Games) != 1 || resp.Games[0].ID != "snapshot-game" {
		t.Fatalf("expected provider games, got %+v", resp.Games)
	}
}

func TestGamesTodayWithInvalidDateReturnsBadRequest(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/games?date=not-a-date", nil)
	testutil.AssertStatus(t, rr, http.StatusBadRequest)
}

func TestGamesTodayLogsUpstreamErrors(t *testing.T) {
	snaps := &stubSnapshots{
		err: errors.New("missing snapshot"),
	}

	h := NewHandler(testutil.NewServiceWithGames(nil), snaps, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/games?date=2024-02-01", nil)
	testutil.AssertStatus(t, rr, http.StatusBadGateway)
}

func TestGamesTodayInvalidTimezoneFallsBack(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	fixedNow := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return fixedNow }

	rr := testutil.Serve(h, http.MethodGet, "/games/today?tz=invalid-timezone", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp domain.TodayResponse
	testutil.DecodeJSON(t, rr, &resp)

	if resp.Date != "2024-01-02" {
		t.Fatalf("expected date 2024-01-02, got %s", resp.Date)
	}
}

func TestGameByID(t *testing.T) {
	game := testutil.SampleGame("id-1")
	game.StartTime = time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC).Format(time.RFC3339)
	h := NewHandler(testutil.NewServiceWithGames([]domain.Game{game}), nil, nil, nil)

	rr := testutil.Serve(http.HandlerFunc(h.GameByID), http.MethodGet, "/games/id-1", nil)

	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp domain.Game
	testutil.DecodeJSON(t, rr, &resp)

	if resp.ID != "id-1" {
		t.Fatalf("expected game id id-1, got %s", resp.ID)
	}
}

func TestGameByIDInvalid(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(http.HandlerFunc(h.GameByID), http.MethodGet, "/games", nil)

	testutil.AssertStatus(t, rr, http.StatusBadRequest)
}

func TestGameByIDNotFound(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(http.HandlerFunc(h.GameByID), http.MethodGet, "/games/unknown", nil)

	testutil.AssertStatus(t, rr, http.StatusNotFound)
}

func TestMethodNotAllowedHandlers(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	tests := []struct {
		name string
		path string
		fn   func(w http.ResponseWriter, r *http.Request)
	}{
		{"health", "/health", h.Health},
		{"ready", "/ready", h.Ready},
		{"gamesToday", "/games/today", h.GamesToday},
		{"gameByID", "/games/id", h.GameByID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testutil.Serve(http.HandlerFunc(tt.fn), http.MethodPost, tt.path, nil)
			testutil.AssertStatus(t, rr, http.StatusMethodNotAllowed)
		})
	}
}

func TestRequestIDPropagatesThroughMiddleware(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/games/", h.GameByID)
	wrapped := middleware.LoggingMiddleware(nil, nil, mux)

	req := httptest.NewRequest(http.MethodGet, "/games/missing", nil)
	req.Header.Set("X-Request-ID", "abc123")
	rr := testutil.ServeRequest(wrapped, req)

	testutil.AssertStatus(t, rr, http.StatusNotFound)

	var resp map[string]string
	testutil.DecodeJSON(t, rr, &resp)
	if resp["requestId"] != "abc123" {
		t.Fatalf("expected requestId propagated, got %s", resp["requestId"])
	}
	if resp["error"] == "" {
		t.Fatalf("expected error field in response")
	}
}

func TestReady(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(http.HandlerFunc(h.Ready), http.MethodGet, "/ready", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
}

func TestReadyWithStatus(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, func() poller.Status {
		return poller.Status{
			LastSuccess: time.Now(),
		}
	})

	rr := testutil.Serve(http.HandlerFunc(h.Ready), http.MethodGet, "/ready", nil)

	testutil.AssertStatus(t, rr, http.StatusOK)
}

func TestReadyNotReady(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, func() poller.Status {
		return poller.Status{}
	})

	rr := testutil.Serve(http.HandlerFunc(h.Ready), http.MethodGet, "/ready", nil)

	testutil.AssertStatus(t, rr, http.StatusServiceUnavailable)
}

func TestGamesTodayHonorsTimezone(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	h.now = func() time.Time {
		return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).In(loc)
	}

	rr := testutil.Serve(h, http.MethodGet, "/games/today?tz=America/New_York", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)

	var resp domain.TodayResponse
	testutil.DecodeJSON(t, rr, &resp)

	if resp.Date != "2024-01-01" {
		t.Fatalf("expected date 2024-01-01 for America/New_York, got %s", resp.Date)
	}
}

func TestGamesTodayLogsCacheHitsWhenNoDateParam(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/games", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
}

func TestServeHTTPNotFound(t *testing.T) {
	h := NewHandler(testutil.NewServiceWithGames(nil), nil, nil, nil)
	rr := testutil.Serve(h, http.MethodGet, "/unknown", nil)
	testutil.AssertStatus(t, rr, http.StatusNotFound)
}

func TestServeHTTPRoutes(t *testing.T) {
	game := testutil.SampleGame("id-1")
	game.StartTime = time.Now().Format(time.RFC3339)
	svc := testutil.NewServiceWithGames([]domain.Game{game})
	h := NewHandler(svc, &stubSnapshots{resp: domain.TodayResponse{Date: "2024-01-01"}}, nil, func() poller.Status { return poller.Status{LastSuccess: time.Now()} })

	tests := []struct {
		path   string
		status int
	}{
		{"/health", http.StatusOK},
		{"/ready", http.StatusOK},
		{"/games/today", http.StatusOK},
		{"/games/id-1", http.StatusOK},
	}

	for _, tt := range tests {
		rr := testutil.Serve(h, http.MethodGet, tt.path, nil)
		testutil.AssertStatus(t, rr, tt.status)
	}
}

func TestGamesTodayUpstreamErrorsReturnBadGateway(t *testing.T) {
	snaps := &stubSnapshots{
		err: errors.New("boom"),
	}

	h := NewHandler(testutil.NewServiceWithGames(nil), snaps, nil, nil)

	rr := testutil.Serve(h, http.MethodGet, "/games?date=2024-02-01", nil)
	testutil.AssertStatus(t, rr, http.StatusBadGateway)
}

func TestWriteJSONErrorPath(t *testing.T) {
	rr := httptest.NewRecorder()
	// channels cannot be JSON encoded; triggers the error branch.
	writeJSON(rr, http.StatusOK, make(chan int), nil)
	// Status is still written even on encode error.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 despite encode error, got %d", rr.Code)
	}
}
