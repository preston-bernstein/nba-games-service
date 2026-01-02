package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"nba-data-service/internal/app/games"
	"nba-data-service/internal/domain"
	"nba-data-service/internal/providers"
	"nba-data-service/internal/snapshots"
)

// AdminHandler exposes admin-only endpoints (e.g., snapshot refresh).
type AdminHandler struct {
	app      *games.Service
	writer   *snapshots.Writer
	provider providers.GameProvider
	token    string
	logger   *slog.Logger
}

// NewAdminHandler constructs an AdminHandler.
func NewAdminHandler(app *games.Service, writer *snapshots.Writer, provider providers.GameProvider, token string, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		app:      app,
		writer:   writer,
		provider: provider,
		token:    token,
		logger:   logger,
	}
}

// RefreshSnapshots writes a games snapshot for the requested date (defaults to today).
// Guarded by ADMIN_TOKEN env; returns 401 if missing/invalid.
func (h *AdminHandler) RefreshSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", h.logger)
		return
	}
	if h.token == "" || r.Header.Get("Authorization") != "Bearer "+h.token {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}
	if h.provider == nil || h.writer == nil {
		writeError(w, r, http.StatusServiceUnavailable, "snapshot writer not configured", h.logger)
		return
	}

	logger := loggerFromContext(r, h.logger)
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	// Validate date format.
	if _, err := time.Parse("2006-01-02", date); err != nil {
		if logger != nil {
			logger.Warn("admin snapshot invalid date", slog.String("date", date))
		}
		writeError(w, r, http.StatusBadRequest, "invalid date format", logger)
		return
	}
	// Fetch games from provider for the date; no tz support here (keep simple).
	tz := strings.TrimSpace(r.URL.Query().Get("tz"))
	if tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			if logger != nil {
				logger.Warn("admin snapshot invalid tz", slog.String("tz", tz))
			}
			writeError(w, r, http.StatusBadRequest, "invalid timezone", logger)
			return
		}
	}
	games, err := h.provider.FetchGames(r.Context(), date, tz)
	if err != nil {
		if logger != nil {
			logger.Warn("admin snapshot fetch failed", slog.String("date", date), slog.String("tz", tz), slog.Any("err", err))
		}
		writeError(w, r, http.StatusBadGateway, "failed to fetch games", logger)
		return
	}
	if len(games) == 0 {
		if logger != nil {
			logger.Warn("admin snapshot no games", slog.String("date", date))
		}
		writeError(w, r, http.StatusBadRequest, "no games to snapshot", logger)
		return
	}

	snap := domain.TodayResponse{
		Date:  date,
		Games: games,
	}
	if err := h.writer.WriteGamesSnapshot(date, snap); err != nil {
		if logger != nil {
			logger.Warn("admin snapshot write failed", slog.String("date", date), slog.Any("err", err))
		}
		writeError(w, r, http.StatusInternalServerError, "failed to write snapshot", logger)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"date":      date,
		"snapshots": len(games),
		"status":    "ok",
	}, logger)
	if logger != nil {
		logger.Info("admin snapshot written", slog.String("date", date), slog.Int("count", len(games)))
	}
}

// AdminTokenFromEnv reads ADMIN_TOKEN (optional).
func AdminTokenFromEnv() string {
	return os.Getenv("ADMIN_TOKEN")
}
