package snapshots

import (
	"context"
	"testing"
	"time"

	"nba-data-service/internal/domain"
)

type fakeProvider struct {
	dates []string
}

func (p *fakeProvider) FetchGames(ctx context.Context, date string, _ string) ([]domain.Game, error) {
	p.dates = append(p.dates, date)
	return []domain.Game{
		{ID: date + "-1", Provider: "stub"},
	}, nil
}

func TestSyncerBackfillsPastAndFuture(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := NewWriter(t.TempDir(), 10000)
	now := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	provider := &fakeProvider{}
	cfg := SyncConfig{
		Enabled:    true,
		Days:       3,
		FutureDays: 2,
		Interval:   time.Nanosecond,
	}

	// Seed snapshots: yesterday (will still refresh), 2 days back (should skip),
	// and future +2 (should skip).
	writeSimpleSnapshot(t, writer, "2024-01-09")
	writeSimpleSnapshot(t, writer, "2024-01-08")
	writeSimpleSnapshot(t, writer, "2024-01-12")

	syncer := NewSyncer(provider, writer, cfg, nil)
	syncer.now = func() time.Time { return now }

	syncer.Run(ctx)
	cancel()

	expected := []string{"2024-01-10", "2024-01-09", "2024-01-11"}

	assertDatesEqual(t, provider.dates, expected)
	for _, date := range expected {
		requireSnapshotExists(t, writer, date)
	}
	// Ensure previously existing snapshots remain.
	requireSnapshotExists(t, writer, "2024-01-08")
	requireSnapshotExists(t, writer, "2024-01-12")
}
