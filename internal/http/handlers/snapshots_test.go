package handlers

import (
	"net/http"
	"testing"
	"time"

	"nba-data-service/internal/app/games"
	"nba-data-service/internal/domain"
	"nba-data-service/internal/store"
	"nba-data-service/internal/testutil"
)

func TestGamesTodayFallsBackToSnapshotWhenCacheEmpty(t *testing.T) {
	ms := store.NewMemoryStore()
	svc := games.NewService(ms)
	snaps := &stubSnapshots{
		resp: domain.TodayResponse{
			Date:  "2024-02-01",
			Games: []domain.Game{{ID: "snapshot-game"}},
		},
	}

	h := NewHandler(svc, snaps, nil, nil)
	h.now = func() time.Time { return time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC) }

	rr := testutil.Serve(h, http.MethodGet, "/games/today", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
	var resp domain.TodayResponse
	testutil.DecodeJSON(t, rr, &resp)
	if resp.Date != "2024-02-01" || len(resp.Games) != 1 || resp.Games[0].ID != "snapshot-game" {
		t.Fatalf("unexpected snapshot response %+v", resp)
	}
}
