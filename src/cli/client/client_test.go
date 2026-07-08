package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestClientListGames(t *testing.T) {
	var gotPath string
	var gotQuery string
	apiClient := New("http://example.test")
	apiClient.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(map[string]any{
			"success": true,
			"games": []map[string]any{
				{
					"game_id":      "game-1",
					"status":       "waiting",
					"current_turn": "red",
					"red_player":   nil,
					"black_player": nil,
					"created_at":   "2026-07-08T12:00:00Z",
				},
			},
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(&body),
		}, nil
	})}

	response, raw, err := apiClient.ListGames(context.Background(), "waiting", "player-1")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("raw response is empty")
	}
	if gotPath != "/api/v1/games" {
		t.Fatalf("path = %q, want /api/v1/games", gotPath)
	}
	if gotQuery != "player_id=player-1&status=waiting" {
		t.Fatalf("query = %q, want player_id=player-1&status=waiting", gotQuery)
	}
	if len(response.Games) != 1 || response.Games[0].GameID != "game-1" {
		t.Fatalf("games = %#v, want game-1", response.Games)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
