package dto

import (
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
)

// ErrorResponse is the standard JSON error payload.
type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

// GameResponse contains the complete game state.
type GameResponse struct {
	Success   bool       `json:"success"`
	GameID    string     `json:"game_id"`
	GameState *GameState `json:"game_state"`
}

// PlayerGameResponse includes the assigned player and optional session.
type PlayerGameResponse struct {
	Success   bool             `json:"success"`
	GameID    string           `json:"game_id"`
	Player    *PlayerResponse  `json:"player"`
	Session   *SessionResponse `json:"session,omitempty"`
	GameState *GameState       `json:"game_state"`
}

// MoveResponse reports a successful move and the updated game state.
type MoveResponse struct {
	Success   bool             `json:"success"`
	Move      *MoveResponseDTO `json:"move"`
	GameState *GameState       `json:"game_state"`
}

// MoveHistoryResponse contains all moves for a game.
type MoveHistoryResponse struct {
	Success bool              `json:"success"`
	Moves   []MoveResponseDTO `json:"moves"`
}

// GameState is a JSON-friendly representation of a game.
type GameState struct {
	ID          string          `json:"id"`
	Board       [][]interface{} `json:"board"`
	RedPlayer   *PlayerResponse `json:"red_player,omitempty"`
	BlackPlayer *PlayerResponse `json:"black_player,omitempty"`
	CurrentTurn string          `json:"current_turn"`
	Status      string          `json:"status"`
	Result      *game.Result    `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PlayerResponse is a JSON-friendly player representation.
type PlayerResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

// SessionResponse is returned when a REST call creates a player session.
type SessionResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MoveResponseDTO is a JSON-friendly move representation.
type MoveResponseDTO struct {
	From      interface{}   `json:"from"`
	To        interface{}   `json:"to"`
	PlayerID  string        `json:"player_id"`
	Timestamp time.Time     `json:"timestamp"`
	Captured  []interface{} `json:"captured,omitempty"`
	Promoted  bool          `json:"promoted"`
}

// NewGameState converts a domain game into a REST DTO.
func NewGameState(g *game.Game) *GameState {
	if g == nil {
		return nil
	}
	return &GameState{
		ID:          g.ID,
		Board:       g.Board.ToJSON(),
		RedPlayer:   NewPlayerResponse(g.RedPlayer),
		BlackPlayer: NewPlayerResponse(g.BlackPlayer),
		CurrentTurn: g.CurrentTurn.String(),
		Status:      g.Status.String(),
		Result:      g.Result,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// NewPlayerResponse converts a domain player into a REST DTO.
func NewPlayerResponse(p *game.Player) *PlayerResponse {
	if p == nil {
		return nil
	}
	return &PlayerResponse{
		ID:    p.ID,
		Name:  p.Name,
		Color: p.Color.String(),
		Type:  p.Type,
	}
}

// NewSessionResponse converts a domain session into a REST DTO.
func NewSessionResponse(s *session.Session) *SessionResponse {
	if s == nil {
		return nil
	}
	return &SessionResponse{
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
	}
}

// NewMoveResponse converts a domain move into a REST DTO.
func NewMoveResponse(m game.Move) MoveResponseDTO {
	captured := make([]interface{}, 0, len(m.Captured))
	for _, pos := range m.Captured {
		captured = append(captured, pos)
	}
	return MoveResponseDTO{
		From:      m.From,
		To:        m.To,
		PlayerID:  m.PlayerID,
		Timestamp: m.Timestamp,
		Captured:  captured,
		Promoted:  m.Promoted,
	}
}

// NewMoveHistoryResponse converts domain moves into a REST DTO.
func NewMoveHistoryResponse(moves []game.Move) []MoveResponseDTO {
	result := make([]MoveResponseDTO, len(moves))
	for i, move := range moves {
		result[i] = NewMoveResponse(move)
	}
	return result
}

// ValidMovesResponse contains the valid destination squares for a piece.
type ValidMovesResponse struct {
	Success bool             `json:"success"`
	Moves   []board.Position `json:"moves"`
}
