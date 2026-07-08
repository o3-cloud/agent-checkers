// Package dto contains REST API request and response payloads.
package dto

import "github.com/stackable-specs/agent-checkers/internal/app/board"

// CreateGameRequest registers the first player for a new game.
type CreateGameRequest struct {
	PlayerName string `json:"player_name"`
	PlayerType string `json:"player_type"`
}

// JoinGameRequest registers a player for an existing waiting game.
type JoinGameRequest struct {
	PlayerName string `json:"player_name"`
	PlayerType string `json:"player_type"`
}

// MoveRequest executes a move for a player.
type MoveRequest struct {
	PlayerID string         `json:"player_id"`
	From     board.Position `json:"from"`
	To       board.Position `json:"to"`
}

// PlayerActionRequest identifies the player performing a game-level action.
type PlayerActionRequest struct {
	PlayerID string `json:"player_id"`
}
