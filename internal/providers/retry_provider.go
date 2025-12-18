package providers

import (
	"context"
	"log/slog"
	"time"

	"nba-games-service/internal/domain"
	"nba-games-service/internal/logging"
)

const (
	defaultRetryAttempts = 3
	defaultBackoff       = 200 * time.Millisecond
)

type backoffFunc func(attempt int) time.Duration

// retryingProvider wraps a GameProvider with retry/backoff behavior.
type retryingProvider struct {
	inner       GameProvider
	logger      *slog.Logger
	maxAttempts int
	backoffFn   backoffFunc
}

// NewRetryingProvider wraps the given provider with retries. If maxAttempts/backoff are <= 0, defaults are used.
func NewRetryingProvider(inner GameProvider, logger *slog.Logger, maxAttempts int, backoff time.Duration) GameProvider {
	if maxAttempts <= 0 {
		maxAttempts = defaultRetryAttempts
	}
	if backoff <= 0 {
		backoff = defaultBackoff
	}
	return &retryingProvider{
		inner:       inner,
		logger:      logger,
		maxAttempts: maxAttempts,
		backoffFn: func(attempt int) time.Duration {
			return time.Duration(attempt) * backoff
		},
	}
}

func (r *retryingProvider) FetchGames(ctx context.Context, date string, tz string) ([]domain.Game, error) {
	var lastErr error

	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		games, err := r.inner.FetchGames(ctx, date, tz)
		if err == nil {
			return games, nil
		}
		lastErr = err

		if attempt == r.maxAttempts {
			break
		}

		r.logWarn(ctx, "provider fetch retry", "attempt", attempt, "max_attempts", r.maxAttempts, "err", err)

		// backoff with context awareness
		delay := r.backoffFn(attempt)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	r.logWarn(ctx, "provider fetch failed", "attempts", r.maxAttempts, "err", lastErr)
	return nil, lastErr
}

func (r *retryingProvider) logWarn(ctx context.Context, msg string, args ...any) {
	logger := logging.FromContext(ctx, r.logger)
	if logger != nil {
		logger.Warn(msg, args...)
	}
}
