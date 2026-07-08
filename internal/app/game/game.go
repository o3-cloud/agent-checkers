// Package game defines the game state management for checkers.
package game

import (
	"errors"
	"fmt"
	"strings"
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

// ParseStatus converts a status string into a game status.
func ParseStatus(value string) (Status, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "waiting":
		return StatusWaiting, true
	case "active":
		return StatusActive, true
	case "completed":
		return StatusCompleted, true
	case "draw":
		return StatusDraw, true
	default:
		return StatusWaiting, false
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

	if g.Status == StatusActive && !g.hasLegalMove(g.CurrentTurn) {
		return true, &Result{Winner: opponent(g.CurrentTurn).String(), Reason: "no_legal_moves"}
	}

	return false, nil
}

func (g *Game) hasLegalMove(color piece.Color) bool {
	mustCapture := g.hasCapture(color)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			from := board.Position{Row: row, Col: col}
			p := g.Board.GetPiece(from)
			if p == nil || p.Color != color {
				continue
			}

			for _, direction := range moveDirections(p) {
				if g.hasLegalCapture(from, direction) {
					return true
				}
				if !mustCapture && g.hasLegalSimpleMove(from, direction) {
					return true
				}
			}
		}
	}

	return false
}

func (g *Game) hasCapture(color piece.Color) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			pos := board.Position{Row: row, Col: col}
			p := g.Board.GetPiece(pos)
			if p == nil || p.Color != color {
				continue
			}
			for _, direction := range moveDirections(p) {
				if g.hasLegalCapture(pos, direction) {
					return true
				}
			}
		}
	}

	return false
}

func (g *Game) hasLegalSimpleMove(from board.Position, direction moveDirection) bool {
	to := board.Position{Row: from.Row + direction.row, Col: from.Col + direction.col}
	return to.IsValid() && to.IsPlayable() && g.Board.GetPiece(to) == nil
}

func (g *Game) hasLegalCapture(from board.Position, direction moveDirection) bool {
	p := g.Board.GetPiece(from)
	if p == nil {
		return false
	}

	capturedPos := board.Position{Row: from.Row + direction.row, Col: from.Col + direction.col}
	landingPos := board.Position{Row: from.Row + direction.row*2, Col: from.Col + direction.col*2}
	if !capturedPos.IsValid() || !landingPos.IsValid() || !landingPos.IsPlayable() {
		return false
	}

	captured := g.Board.GetPiece(capturedPos)
	return captured != nil && captured.Color != p.Color && g.Board.GetPiece(landingPos) == nil
}

type moveDirection struct {
	row int
	col int
}

func moveDirections(p *piece.Piece) []moveDirection {
	if p.IsKing {
		return []moveDirection{
			{row: 1, col: 1},
			{row: 1, col: -1},
			{row: -1, col: 1},
			{row: -1, col: -1},
		}
	}

	if p.Color == piece.Black {
		return []moveDirection{
			{row: -1, col: -1},
			{row: -1, col: 1},
		}
	}

	return []moveDirection{
		{row: 1, col: -1},
		{row: 1, col: 1},
	}
}

func opponent(color piece.Color) piece.Color {
	if color == piece.Red {
		return piece.Black
	}
	return piece.Red
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
