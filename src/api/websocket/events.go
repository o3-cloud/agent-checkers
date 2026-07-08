// Package websocket provides real-time game update delivery.
package websocket

import (
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// EventType identifies a WebSocket event kind.
type EventType string

const (
	// EventTypeGameState delivers the current game state after connection.
	EventTypeGameState EventType = "game_state"
	// EventTypeGameStarted announces that the second player joined a game.
	EventTypeGameStarted EventType = "game_started"
	// EventTypeMoveMade announces a completed move.
	EventTypeMoveMade EventType = "move_made"
	// EventTypeTurnChanged announces the current player after a move.
	EventTypeTurnChanged EventType = "turn_changed"
	// EventTypeGameEnded announces a completed or drawn game.
	EventTypeGameEnded EventType = "game_ended"
	// EventTypePing is the server heartbeat event.
	EventTypePing EventType = "ping"
	// EventTypePong is the client heartbeat acknowledgement event.
	EventTypePong EventType = "pong"
)

// Event is the envelope used for every WebSocket message.
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// GameStatePayload contains the current game state sent after connection.
type GameStatePayload struct {
	GameState *dto.GameState `json:"game_state"`
}

// GameStartedPayload contains game start details.
type GameStartedPayload struct {
	GameID  string                `json:"game_id"`
	Players []*dto.PlayerResponse `json:"players"`
}

// MoveMadePayload contains details for a completed move.
type MoveMadePayload struct {
	From     board.Position   `json:"from"`
	To       board.Position   `json:"to"`
	Captured []board.Position `json:"captured"`
	Board    [][]interface{}  `json:"board"`
	Turn     string           `json:"turn"`
}

// TurnChangedPayload identifies the player whose turn is current.
type TurnChangedPayload struct {
	CurrentPlayer string `json:"current_player"`
}

// GameEndedPayload contains the winner and reason for game completion.
type GameEndedPayload struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// PingPayload contains the server heartbeat timestamp.
type PingPayload struct {
	Timestamp time.Time `json:"timestamp"`
}

// PongPayload contains the client heartbeat timestamp.
type PongPayload struct {
	Timestamp time.Time `json:"timestamp"`
}
