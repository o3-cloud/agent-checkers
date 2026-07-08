// Package game defines the game state management for checkers.
package game

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

// Status represents the current state of a game.
type Status int

const (
	// StatusWaiting indicates the game is waiting for players.
	StatusWaiting Status = iota
	// StatusActive indicates the game is in progress.
	StatusActive
	// StatusCompleted indicates the game has ended.
	StatusCompleted
	// StatusDraw indicates the game ended in a draw.
	StatusDraw
)

// String returns the string representation of the game status.
func (s Status) String() string {
	switch s {
	case StatusWaiting:
		return "waiting"
	case StatusActive:
		return "active"
	case StatusCompleted:
		return "completed"
	case StatusDraw:
		return "draw"
	default:
		return "unknown"
	}
}

// Result represents the outcome of a completed game.
type Result struct {
	Winner   string `json:"winner"` // Player ID of the winner
	Reason   string `json:"reason"` // Reason for game end
	DrawOffe string `json:"draw_offer,omitempty"`
}

// Move represents a single move in the game.
type Move struct {
	From      board.Position   `json:"from"`
	To        board.Position   `json:"to"`
	PlayerID  string           `json:"player_id"`
	Timestamp time.Time        `json:"timestamp"`
	Captured  []board.Position `json:"captured,omitempty"` // Positions of captured pieces
	Promoted  bool             `json:"promoted"`           // Whether the piece was promoted
}

// Player represents a checkers player.
type Player struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Color piece.Color `json:"color"`
	Type  string      `json:"type"` // "human" or "ai"
}

// Game represents a checkers game session.
type Game struct {
	ID          string       `json:"id"`
	Board       *board.Board `json:"board"`
	RedPlayer   *Player      `json:"red_player,omitempty"`
	BlackPlayer *Player      `json:"black_player,omitempty"`
	CurrentTurn piece.Color  `json:"current_turn"`
	Status      Status       `json:"status"`
	Moves       []Move       `json:"moves"`
	Result      *Result      `json:"result,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewGame creates a new game with an initial board position.
func NewGame() *Game {
	return &Game{
		ID:          uuid.New().String(),
		Board:       board.New(),
		CurrentTurn: piece.Red, // Red always moves first
		Status:      StatusWaiting,
		Moves:       []Move{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddPlayer adds a player to the game.
// Returns an error if the game already has two players.
func (g *Game) AddPlayer(p *Player) error {
	if g.Status != StatusWaiting {
		return errors.New("game is not waiting for players")
	}

	if g.RedPlayer == nil {
		p.Color = piece.Red
		g.RedPlayer = p
	} else if g.BlackPlayer == nil {
		p.Color = piece.Black
		g.BlackPlayer = p
		g.Status = StatusActive
	} else {
		return errors.New("game already has two players")
	}

	g.UpdatedAt = time.Now()
	return nil
}

// GetPlayer returns the player with the given ID.
func (g *Game) GetPlayer(playerID string) *Player {
	if g.RedPlayer != nil && g.RedPlayer.ID == playerID {
		return g.RedPlayer
	}
	if g.BlackPlayer != nil && g.BlackPlayer.ID == playerID {
		return g.BlackPlayer
	}
	return nil
}

// CurrentPlayer returns the player whose turn it is.
func (g *Game) CurrentPlayer() *Player {
	if g.CurrentTurn == piece.Red {
		return g.RedPlayer
	}
	return g.BlackPlayer
}

// MakeMove attempts to execute a move.
// This is a convenience method that delegates to the rules package.
// Returns an error if the move is invalid.
func (g *Game) MakeMove(playerID string, from, to board.Position) error {
	if g.Status != StatusActive {
		return errors.New("game is not active")
	}

	player := g.GetPlayer(playerID)
	if player == nil {
		return errors.New("player not in this game")
	}

	if player.Color != g.CurrentTurn {
		return fmt.Errorf("it is not %s's turn", g.CurrentTurn)
	}

	// Note: Actual move validation will be done by rules package
	// This is a placeholder for the game state transition
	return nil
}

// SwitchTurn changes the current turn to the other player.
func (g *Game) SwitchTurn() {
	if g.CurrentTurn == piece.Red {
		g.CurrentTurn = piece.Black
	} else {
		g.CurrentTurn = piece.Red
	}
	g.UpdatedAt = time.Now()
}

// RecordMove adds a move to the game history.
func (g *Game) RecordMove(m Move) {
	m.Timestamp = time.Now()
	g.Moves = append(g.Moves, m)
	g.UpdatedAt = time.Now()
}

// IsGameOver checks if the game has ended.
// Returns true if game is over and the result if applicable.
func (g *Game) IsGameOver() (bool, *Result) {
	redCount, blackCount := g.Board.CountPieces()

	// Win by capturing all opponent pieces
	if redCount == 0 {
		return true, &Result{Winner: "black", Reason: "all_pieces_captured"}
	}
	if blackCount == 0 {
		return true, &Result{Winner: "red", Reason: "all_pieces_captured"}
	}

	// TODO: Check for blocked moves (no valid moves available)
	// This requires the rules package to check all valid moves

	return false, nil
}

// EndGame ends the game with the given result.
func (g *Game) EndGame(result *Result) {
	g.Status = StatusCompleted
	g.Result = result
	g.UpdatedAt = time.Now()
}

// OfferDraw records a draw offer.
func (g *Game) OfferDraw(playerID string) error {
	player := g.GetPlayer(playerID)
	if player == nil {
		return errors.New("player not in this game")
	}

	// Store draw offer in result
	if g.Result == nil {
		g.Result = &Result{}
	}
	g.Result.DrawOffe = playerID
	g.UpdatedAt = time.Now()
	return nil
}

// AcceptDraw accepts a draw offer and ends the game.
func (g *Game) AcceptDraw(playerID string) error {
	player := g.GetPlayer(playerID)
	if player == nil {
		return errors.New("player not in this game")
	}

	if g.Result == nil || g.Result.DrawOffe == "" {
		return errors.New("no draw offer to accept")
	}

	if g.Result.DrawOffe == playerID {
		return errors.New("cannot accept your own draw offer")
	}

	g.Status = StatusDraw
	g.Result = &Result{Reason: "draw_agreement"}
	g.UpdatedAt = time.Now()
	return nil
}

// Resign ends the game with the other player as winner.
func (g *Game) Resign(playerID string) error {
	player := g.GetPlayer(playerID)
	if player == nil {
		return errors.New("player not in this game")
	}

	if g.Status != StatusActive {
		return errors.New("game is not active")
	}

	var winner string
	if player.Color == piece.Red {
		winner = "black"
	} else {
		winner = "red"
	}

	g.EndGame(&Result{
		Winner: winner,
		Reason: "resignation",
	})
	return nil
}

// Clone creates a deep copy of the game.
func (g *Game) Clone() *Game {
	clone := &Game{
		ID:          g.ID,
		Board:       g.Board.Clone(),
		CurrentTurn: g.CurrentTurn,
		Status:      g.Status,
		Moves:       make([]Move, len(g.Moves)),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}

	for i, move := range g.Moves {
		clone.Moves[i] = move
		if move.Captured != nil {
			clone.Moves[i].Captured = append([]board.Position(nil), move.Captured...)
		}
	}

	if g.RedPlayer != nil {
		clone.RedPlayer = &Player{
			ID:    g.RedPlayer.ID,
			Name:  g.RedPlayer.Name,
			Color: g.RedPlayer.Color,
			Type:  g.RedPlayer.Type,
		}
	}

	if g.BlackPlayer != nil {
		clone.BlackPlayer = &Player{
			ID:    g.BlackPlayer.ID,
			Name:  g.BlackPlayer.Name,
			Color: g.BlackPlayer.Color,
			Type:  g.BlackPlayer.Type,
		}
	}

	if g.Result != nil {
		clone.Result = &Result{
			Winner:   g.Result.Winner,
			Reason:   g.Result.Reason,
			DrawOffe: g.Result.DrawOffe,
		}
	}

	return clone
}
