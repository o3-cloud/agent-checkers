// Package client wraps the agent-checkers REST API for CLI commands.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// Client sends HTTP requests to the checkers REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a REST API client.
func New(serverURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateGame creates a new waiting game.
func (c *Client) CreateGame(ctx context.Context, name, playerType string) (*dto.PlayerGameResponse, []byte, error) {
	var response dto.PlayerGameResponse
	raw, err := c.do(ctx, http.MethodPost, "/api/v1/games", dto.CreateGameRequest{
		PlayerName: name,
		PlayerType: playerType,
	}, &response)
	return &response, raw, err
}

// JoinGame joins an existing waiting game.
func (c *Client) JoinGame(ctx context.Context, gameID, name, playerType string) (*dto.PlayerGameResponse, []byte, error) {
	var response dto.PlayerGameResponse
	raw, err := c.do(ctx, http.MethodPost, "/api/v1/games/"+url.PathEscape(gameID)+"/join", dto.JoinGameRequest{
		PlayerName: name,
		PlayerType: playerType,
	}, &response)
	return &response, raw, err
}

// GetGame returns the current game state.
func (c *Client) GetGame(ctx context.Context, gameID string) (*dto.GameResponse, []byte, error) {
	var response dto.GameResponse
	raw, err := c.do(ctx, http.MethodGet, "/api/v1/games/"+url.PathEscape(gameID), nil, &response)
	return &response, raw, err
}

// MakeMove executes a move.
func (c *Client) MakeMove(ctx context.Context, gameID, playerID string, from, to board.Position) (*dto.MoveResponse, []byte, error) {
	var response dto.MoveResponse
	raw, err := c.do(ctx, http.MethodPost, "/api/v1/games/"+url.PathEscape(gameID)+"/moves", dto.MoveRequest{
		PlayerID: playerID,
		From:     from,
		To:       to,
	}, &response)
	return &response, raw, err
}

// Resign resigns the current game.
func (c *Client) Resign(ctx context.Context, gameID, playerID string) (*dto.GameResponse, []byte, error) {
	var response dto.GameResponse
	raw, err := c.do(ctx, http.MethodDelete, "/api/v1/games/"+url.PathEscape(gameID), dto.PlayerActionRequest{
		PlayerID: playerID,
	}, &response)
	return &response, raw, err
}

// Draw offers or accepts a draw.
func (c *Client) Draw(ctx context.Context, gameID, playerID string) (*dto.GameResponse, []byte, error) {
	var response dto.GameResponse
	raw, err := c.do(ctx, http.MethodPost, "/api/v1/games/"+url.PathEscape(gameID)+"/draw", dto.PlayerActionRequest{
		PlayerID: playerID,
	}, &response)
	return &response, raw, err
}

// ValidMoves returns legal destinations for a piece.
func (c *Client) ValidMoves(ctx context.Context, gameID string, pos board.Position) ([]board.Position, []byte, error) {
	var response dto.ValidMovesResponse
	path := fmt.Sprintf("/api/v1/games/%s/valid-moves?row=%d&col=%d", url.PathEscape(gameID), pos.Row, pos.Col)
	raw, err := c.do(ctx, http.MethodGet, path, nil, &response)
	return response.Moves, raw, err
}

func (c *Client) do(ctx context.Context, method, path string, body any, target any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}

	raw, err := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close response body: %w", closeErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return raw, apiError(response.StatusCode, raw)
	}
	if target != nil {
		if err := json.Unmarshal(raw, target); err != nil {
			return raw, fmt.Errorf("decode response: %w", err)
		}
	}
	return raw, nil
}

func apiError(status int, raw []byte) error {
	var response dto.ErrorResponse
	if err := json.Unmarshal(raw, &response); err == nil && response.Error != "" {
		return fmt.Errorf("api error %d: %s", status, response.Error)
	}
	return fmt.Errorf("api error %d: %s", status, strings.TrimSpace(string(raw)))
}
