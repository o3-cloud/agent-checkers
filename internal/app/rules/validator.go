// Package rules implements checkers move validation and capture logic.
package rules

import (
	"errors"
	"fmt"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

// ValidationError represents a move validation error with details.
type ValidationError struct {
	Message    string           `json:"message"`
	ValidMoves []board.Position `json:"valid_moves,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Validator validates moves according to checkers rules.
type Validator struct{}

// NewValidator creates a new move validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMove checks if a move is valid according to checkers rules.
// Returns nil if valid, or an error describing why the move is invalid.
func (v *Validator) ValidateMove(g *game.Game, from, to board.Position) error {
	// Check game is active
	if g.Status != game.StatusActive {
		return &ValidationError{Message: "game is not active"}
	}

	// Check positions are valid
	if !from.IsValid() {
		return &ValidationError{Message: fmt.Sprintf("invalid source position: %v", from)}
	}
	if !to.IsValid() {
		return &ValidationError{Message: fmt.Sprintf("invalid destination position: %v", to)}
	}

	// Check source has a piece
	p := g.Board.GetPiece(from)
	if p == nil {
		return &ValidationError{Message: "no piece at source position"}
	}

	// Check destination is empty
	if g.Board.GetPiece(to) != nil {
		return &ValidationError{Message: "destination position is not empty"}
	}

	// Check destination is a playable square (dark square)
	if !to.IsPlayable() {
		return &ValidationError{Message: "can only move to playable (dark) squares"}
	}

	// Check it's the correct player's turn
	if p.Color != g.CurrentTurn {
		return &ValidationError{Message: fmt.Sprintf("it is %s's turn, not %s's", g.CurrentTurn, p.Color)}
	}

	// Check if move is diagonal
	rowDiff := to.Row - from.Row
	colDiff := to.Col - from.Col
	absRowDiff := abs(rowDiff)
	absColDiff := abs(colDiff)

	if absRowDiff != absColDiff {
		return &ValidationError{Message: "moves must be diagonal"}
	}

	// Check move distance (1 for simple move, 2 for capture)
	if absRowDiff == 0 {
		return &ValidationError{Message: "must move at least one square"}
	}

	// Check direction for non-king pieces
	if !p.IsKing {
		direction := 1 // Red moves up (increasing row)
		if p.Color == piece.Black {
			direction = -1 // Black moves down (decreasing row)
		}
		if rowDiff*direction < 0 {
			return &ValidationError{Message: "non-king pieces can only move forward"}
		}
	}

	// Check for capture
	if absRowDiff == 2 {
		// This is a capture move
		capturedPos := board.Position{
			Row: from.Row + rowDiff/2,
			Col: from.Col + colDiff/2,
		}
		captured := g.Board.GetPiece(capturedPos)
		if captured == nil {
			return &ValidationError{Message: "capture move requires an opponent piece to jump over"}
		}
		if captured.Color == p.Color {
			return &ValidationError{Message: "cannot capture your own piece"}
		}
	} else if absRowDiff == 1 {
		// Simple move - check if captures are available (mandatory capture rule)
		if v.HasCaptures(g, g.CurrentTurn) {
			return &ValidationError{Message: "a capture is available, you must capture"}
		}
	} else {
		// Distance > 2 is invalid
		return &ValidationError{Message: "invalid move distance"}
	}

	return nil
}

// GetValidMoves returns all valid moves for a piece at the given position.
func (v *Validator) GetValidMoves(g *game.Game, pos board.Position) ([]board.Position, error) {
	p := g.Board.GetPiece(pos)
	if p == nil {
		return nil, &ValidationError{Message: "no piece at position"}
	}

	// If captures are available for this color, only return capture moves
	hasCaptures := v.HasCaptures(g, p.Color)

	var validMoves []board.Position
	directions := v.getDirections(p)

	for _, dir := range directions {
		// Check simple moves (distance 1)
		if !hasCaptures {
			dest := board.Position{Row: pos.Row + dir.row, Col: pos.Col + dir.col}
			if dest.IsValid() && dest.IsPlayable() && g.Board.GetPiece(dest) == nil {
				err := v.ValidateMove(g, pos, dest)
				if err == nil {
					validMoves = append(validMoves, dest)
				}
			}
		}

		// Check capture moves (distance 2)
		dest := board.Position{Row: pos.Row + dir.row*2, Col: pos.Col + dir.col*2}
		if dest.IsValid() && dest.IsPlayable() && g.Board.GetPiece(dest) == nil {
			err := v.ValidateMove(g, pos, dest)
			if err == nil {
				validMoves = append(validMoves, dest)
			}
		}
	}

	return validMoves, nil
}

// GetAllValidMoves returns all valid moves for the current player.
func (v *Validator) GetAllValidMoves(g *game.Game) (map[board.Position][]board.Position, error) {
	moves := make(map[board.Position][]board.Position)

	// Iterate through all pieces of current player
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			pos := board.Position{Row: row, Col: col}
			p := g.Board.GetPiece(pos)
			if p != nil && p.Color == g.CurrentTurn {
				validMoves, err := v.GetValidMoves(g, pos)
				if err == nil && len(validMoves) > 0 {
					moves[pos] = validMoves
				}
			}
		}
	}

	return moves, nil
}

// HasCaptures returns true if the given color has any capture moves available.
func (v *Validator) HasCaptures(g *game.Game, color piece.Color) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			pos := board.Position{Row: row, Col: col}
			p := g.Board.GetPiece(pos)
			if p != nil && p.Color == color {
				if v.HasCapturesForPiece(g, pos) {
					return true
				}
			}
		}
	}
	return false
}

// HasCapturesForPiece returns true if the piece at the given position has capture moves.
func (v *Validator) HasCapturesForPiece(g *game.Game, pos board.Position) bool {
	p := g.Board.GetPiece(pos)
	if p == nil {
		return false
	}

	directions := v.getDirections(p)

	for _, dir := range directions {
		// Check if there's an opponent piece to capture
		capturePos := board.Position{Row: pos.Row + dir.row, Col: pos.Col + dir.col}
		landingPos := board.Position{Row: pos.Row + dir.row*2, Col: pos.Col + dir.col*2}

		if !capturePos.IsValid() || !landingPos.IsValid() {
			continue
		}

		captured := g.Board.GetPiece(capturePos)
		if captured == nil || captured.Color == p.Color {
			continue
		}

		if !landingPos.IsPlayable() {
			continue
		}

		if g.Board.GetPiece(landingPos) != nil {
			continue
		}

		// Found a valid capture
		return true
	}

	return false
}

// ExecuteMove executes a move and returns the captured pieces.
// This should only be called after ValidateMove returns nil.
func (v *Validator) ExecuteMove(g *game.Game, from, to board.Position) ([]board.Position, error) {
	p := g.Board.GetPiece(from)
	if p == nil {
		return nil, errors.New("no piece at source position")
	}

	// Calculate captured position if this is a capture move
	var captured []board.Position
	rowDiff := to.Row - from.Row
	colDiff := to.Col - from.Col

	if abs(rowDiff) == 2 {
		// Capture move
		capturePos := board.Position{
			Row: from.Row + rowDiff/2,
			Col: from.Col + colDiff/2,
		}
		captured = append(captured, capturePos)
		g.Board.RemovePiece(capturePos)
	}

	// Move the piece
	err := g.Board.MovePiece(from, to)
	if err != nil {
		return nil, err
	}

	// Check for promotion
	if p.CanPromote(to.Row) {
		p.Promote()
	}

	// Record the move
	move := game.Move{
		From:     from,
		To:       to,
		PlayerID: g.CurrentPlayer().ID,
		Captured: captured,
		Promoted: p.IsKing && (to.Row == 0 || to.Row == 7),
	}
	g.RecordMove(move)

	return captured, nil
}

// getDirections returns the valid move directions for a piece.
// Kings can move in all 4 diagonal directions.
// Non-king pieces can only move forward.
func (v *Validator) getDirections(p *piece.Piece) []struct{ row, col int } {
	allDirections := []struct{ row, col int }{
		{-1, -1}, // Up-left
		{-1, 1},  // Up-right
		{1, -1},  // Down-left
		{1, 1},   // Down-right
	}

	if p.IsKing {
		return allDirections
	}

	// Non-king pieces can only move forward
	if p.Color == piece.Red {
		// Red moves up (increasing row)
		return []struct{ row, col int }{
			{1, -1}, // Up-left
			{1, 1},  // Up-right
		}
	} else {
		// Black moves down (decreasing row)
		return []struct{ row, col int }{
			{-1, -1}, // Down-left
			{-1, 1},  // Down-right
		}
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
