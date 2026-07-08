package mcp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stackable-specs/agent-checkers/internal/app/discovery"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

type toolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type listGamesArgs struct {
	Status   string `json:"status,omitempty"`
	PlayerID string `json:"player_id,omitempty"`
}

func listGamesTool() toolDefinition {
	return toolDefinition{
		Name:        "list_games",
		Description: "List games filtered by status and player. Defaults to active and waiting games.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"status": map[string]any{
					"type": "string",
					"enum": []string{"waiting", "active", "completed", "draw", "all"},
				},
				"player_id": map[string]any{"type": "string"},
			},
		},
	}
}

func (s *Server) handleToolCall(raw json.RawMessage) (any, error) {
	var params toolCallParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, errors.New("invalid tool call parameters")
	}
	if params.Name != "list_games" {
		return nil, fmt.Errorf("unknown tool %q", params.Name)
	}

	args := listGamesArgs{}
	if value, ok := params.Arguments["status"].(string); ok {
		args.Status = value
	}
	if value, ok := params.Arguments["player_id"].(string); ok {
		args.PlayerID = value
	}
	return s.ListGames(args.Status, args.PlayerID)
}

// ListGames runs the list_games tool directly.
func (s *Server) ListGames(status, playerID string) (*dto.ListGamesResponse, error) {
	games, err := discovery.ListGames(s.store, status, playerID)
	if err != nil {
		if errors.Is(err, discovery.ErrInvalidStatus) {
			return nil, errors.New("invalid status")
		}
		return nil, err
	}
	return &dto.ListGamesResponse{
		Success: true,
		Games:   dto.NewGameSummaries(games),
	}, nil
}
