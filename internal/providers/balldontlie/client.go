package balldontlie

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nba-games-service/internal/domain"
)

// Config controls how the balldontlie client reaches the upstream API.
type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timezone   string
}

// Client fetches games from the balldontlie API and maps them to domain models.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient httpDoer
	now        func() time.Time
	loc        *time.Location
}

// NewClient constructs a balldontlie client with the provided configuration.
func NewClient(cfg Config) *Client {
	return &Client{
		baseURL:    normalizeBaseURL(cfg.BaseURL),
		apiKey:     cfg.APIKey,
		httpClient: resolveHTTPClient(cfg.HTTPClient),
		now:        time.Now,
		loc:        resolveLocation(cfg.Timezone),
	}
}

// FetchGames retrieves today's games from balldontlie.
func (c *Client) FetchGames(ctx context.Context, date string) ([]domain.Game, error) {
	req, err := c.buildRequest(ctx, date)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("balldontlie: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload gamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	games := make([]domain.Game, 0, len(payload.Data))
	for _, g := range payload.Data {
		games = append(games, mapGame(g))
	}

	return games, nil
}

func (c *Client) buildRequest(ctx context.Context, date string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/games", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("dates[]", c.resolveDate(date))
	q.Set("per_page", strconv.Itoa(defaultPerPage))
	req.URL.RawQuery = q.Encode()

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	return req, nil
}

func (c *Client) resolveDate(date string) string {
	if date != "" {
		if _, err := time.Parse("2006-01-02", date); err == nil {
			return date
		}
	}
	return c.now().In(c.loc).Format("2006-01-02")
}
