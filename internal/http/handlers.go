package http

import (
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"nba-games-service/internal/domain"
	"nba-games-service/internal/providers"
)

type nowFunc func() time.Time

// Handler wires HTTP routes to the domain service.
type Handler struct {
	svc      *domain.Service
	logger   *slog.Logger
	now      nowFunc
	provider providers.GameProvider
}

// NewHandler constructs a Handler with defaults.
func NewHandler(svc *domain.Service, logger *slog.Logger, provider providers.GameProvider) *Handler {
	return &Handler{
		svc:      svc,
		logger:   logger,
		now:      time.Now,
		provider: provider,
	}
}

// Health reports the service health.
func (h *Handler) Health(w nethttp.ResponseWriter, r *nethttp.Request) {
	resp := map[string]string{"status": "ok"}
	h.writeJSON(w, nethttp.StatusOK, resp)
}

// GamesToday returns the current snapshot of games.
func (h *Handler) GamesToday(w nethttp.ResponseWriter, r *nethttp.Request) {
	dateParam := r.URL.Query().Get("date")
	games := h.svc.Games()
	date := h.now().Format("2006-01-02")

	if dateParam != "" && h.provider != nil {
		fetched, err := h.provider.FetchGames(r.Context(), dateParam)
		if err != nil {
			h.writeError(w, nethttp.StatusBadGateway, "failed to fetch games")
			return
		}
		games = fetched
		date = dateParam
	}

	payload := domain.TodayResponse{
		Date:  date,
		Games: games,
	}
	h.writeJSON(w, nethttp.StatusOK, payload)
}

// GameByID returns a specific game if present.
func (h *Handler) GameByID(w nethttp.ResponseWriter, r *nethttp.Request) {
	// Expect path: /games/{id}
	id := strings.TrimPrefix(r.URL.Path, "/games/")
	if id == "" || id == "games" {
		h.writeError(w, nethttp.StatusBadRequest, "missing game id")
		return
	}

	game, ok := h.svc.GameByID(id)
	if !ok {
		h.writeError(w, nethttp.StatusNotFound, "game not found")
		return
	}

	h.writeJSON(w, nethttp.StatusOK, game)
}

func (h *Handler) writeJSON(w nethttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil && h.logger != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) writeError(w nethttp.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
