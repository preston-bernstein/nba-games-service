package providers

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"nba-games-service/internal/domain"
)

type flakeyProvider struct {
	failures int
	calls    int
}

func (f *flakeyProvider) FetchGames(ctx context.Context, date string, tz string) ([]domain.Game, error) {
	_ = ctx
	_ = date
	_ = tz
	f.calls++
	if f.calls <= f.failures {
		return nil, errors.New("boom")
	}
	return []domain.Game{{ID: "ok"}}, nil
}

func TestRetryingProviderRetriesAndSucceeds(t *testing.T) {
	fp := &flakeyProvider{failures: 2}
	rp := NewRetryingProvider(fp, slog.Default(), 3, 1*time.Millisecond)

	games, err := rp.FetchGames(context.Background(), "", "")
	if err != nil {
		t.Fatalf("expected success, got error %v", err)
	}
	if len(games) != 1 || games[0].ID != "ok" {
		t.Fatalf("unexpected games %+v", games)
	}
	if fp.calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", fp.calls)
	}
}

func TestRetryingProviderStopsAfterMaxAttempts(t *testing.T) {
	fp := &flakeyProvider{failures: 5}
	rp := NewRetryingProvider(fp, nil, 2, 1*time.Millisecond)

	_, err := rp.FetchGames(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if fp.calls != 2 {
		t.Fatalf("expected 2 attempts, got %d", fp.calls)
	}
}

func TestRetryingProviderRespectsContextCancel(t *testing.T) {
	fp := &flakeyProvider{failures: 5}
	rp := NewRetryingProvider(fp, nil, 3, time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rp.FetchGames(ctx, "", "")
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestRetryingProviderUsesCustomBackoff(t *testing.T) {
	fp := &flakeyProvider{failures: 1}
	rp := NewRetryingProvider(fp, nil, 2, time.Hour).(*retryingProvider)

	calls := 0
	rp.backoffFn = func(attempt int) time.Duration {
		calls++
		return 0
	}

	_, _ = rp.FetchGames(context.Background(), "", "")

	if calls == 0 {
		t.Fatalf("expected custom backoff to be invoked")
	}
}
