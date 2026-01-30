package teststubs

import (
	"context"
	"errors"
	"testing"

	domaingames "github.com/preston-bernstein/nba-data-service/internal/domain/games"
)

func TestStubProviderTracksCalls(t *testing.T) {
	err := errors.New("boom")
	p := &StubProvider{Games: []domaingames.Game{{ID: "g1"}}, Err: err}
	if _, got := p.FetchGames(context.Background(), "2024-01-01", ""); !errors.Is(got, err) {
		t.Fatalf("expected error passthrough, got %v", got)
	}
	if p.Calls.Load() != 1 {
		t.Fatalf("expected call count 1, got %d", p.Calls.Load())
	}
}

func TestStubSnapshotStore(t *testing.T) {
	date := "2024-01-01"
	s := &StubSnapshotStore{
		Games: map[string]domaingames.TodayResponse{
			date: domaingames.NewTodayResponse(date, []domaingames.Game{{ID: "g1"}}),
		},
	}

	resp, err := s.LoadGames(date)
	if err != nil || resp.Date != date {
		t.Fatalf("expected loaded games, got %v err %v", resp, err)
	}

	game, ok := s.FindGameByID(date, "g1")
	if !ok || game.ID != "g1" {
		t.Fatalf("expected game found, got %v ok=%v", game, ok)
	}

	_, ok = s.FindGameByID(date, "missing")
	if ok {
		t.Fatalf("expected game not found")
	}
}

func TestStubSnapshotWriter(t *testing.T) {
	date := "2024-01-01"
	w := &StubSnapshotWriter{}
	err := w.WriteGamesSnapshot(date, domaingames.NewTodayResponse(date, []domaingames.Game{{ID: "g1"}}))
	if err != nil {
		t.Fatalf("expected write success, got %v", err)
	}
	if len(w.Written) != 1 {
		t.Fatalf("expected one written entry, got %d", len(w.Written))
	}

	w.Err = errors.New("write error")
	err = w.WriteGamesSnapshot("2024-01-02", domaingames.NewTodayResponse("2024-01-02", []domaingames.Game{{ID: "g2"}}))
	if err == nil {
		t.Fatalf("expected write error")
	}
}
