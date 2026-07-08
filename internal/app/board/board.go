// Package board defines the checkers board representation and position handling.
package board

import (
	"errors"
	"fmt"

	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

// Position represents a square on the 8x8 board.
// Row 0 is black's home row, Row 7 is red's home row.
// Col 0 is the leftmost column from player's view.
type Position struct {
	Row int `json:"row"` // 0-7
	Col int `json:"col"` // 0-7
}

// String returns algebraic notation for the position (e.g., "a1", "h8").
func (p Position) String() string {
	if !p.IsValid() {
		return "invalid"
	}
	// Convert to algebraic notation: col 0 = 'a', row 0 = '1'
	col := string(rune('a' + p.Col))
	row := fmt.Sprintf("%d", p.Row+1)
	return col + row
}

// IsValid returns true if the position is within the board bounds.
func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row < 8 && p.Col >= 0 && p.Col < 8
}

// IsPlayable returns true if the position is a playable (dark) square.
// In checkers, only dark squares are used. On an 8x8 board:
// - Row 0: playable at cols 1, 3, 5, 7 (odd columns)
// - Row 1: playable at cols 0, 2, 4, 6 (even columns)
// This alternates by row.
func (p Position) IsPlayable() bool {
	if !p.IsValid() {
		return false
	}
	// Playable squares are where (row + col) is odd
	return (p.Row+p.Col)%2 == 1
}

// ParsePosition parses algebraic notation (e.g., "a1", "h8") into a Position.
func ParsePosition(s string) (Position, error) {
	if len(s) != 2 {
		return Position{}, errors.New("position must be 2 characters")
	}
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	pos := Position{Row: row, Col: col}
	if !pos.IsValid() {
		return Position{}, fmt.Errorf("invalid position: %s", s)
	}
	return pos, nil
}

// Board represents an 8x8 checkers board.
type Board struct {
	squares [8][8]*piece.Piece
}

// New creates a new board with all pieces in their starting positions.
// Red pieces are on rows 0-2 (but only playable squares).
// Black pieces are on rows 5-7 (but only playable squares).
func New() *Board {
	b := &Board{}
	b.setupInitialPosition()
	return b
}

// NewEmpty creates an empty board with no pieces.
func NewEmpty() *Board {
	return &Board{}
}

// setupInitialPosition places pieces in their starting positions.
func (b *Board) setupInitialPosition() {
	// Red pieces on rows 0, 1, 2 (playable squares only)
	// Red moves "up" the board (increasing row number)
	for row := 0; row < 3; row++ {
		for col := 0; col < 8; col++ {
			pos := Position{Row: row, Col: col}
			if pos.IsPlayable() {
				id := fmt.Sprintf("r%d%d", row, col)
				b.squares[row][col] = piece.New(id, piece.Red)
			}
		}
	}

	// Black pieces on rows 5, 6, 7 (playable squares only)
	// Black moves "down" the board (decreasing row number)
	for row := 5; row < 8; row++ {
		for col := 0; col < 8; col++ {
			pos := Position{Row: row, Col: col}
			if pos.IsPlayable() {
				id := fmt.Sprintf("b%d%d", row, col)
				b.squares[row][col] = piece.New(id, piece.Black)
			}
		}
	}
}

// GetPiece returns the piece at the given position, or nil if empty.
func (b *Board) GetPiece(pos Position) *piece.Piece {
	if !pos.IsValid() {
		return nil
	}
	return b.squares[pos.Row][pos.Col]
}

// SetPiece places a piece at the given position.
func (b *Board) SetPiece(pos Position, p *piece.Piece) {
	if pos.IsValid() {
		b.squares[pos.Row][pos.Col] = p
	}
}

// RemovePiece removes and returns the piece at the given position.
func (b *Board) RemovePiece(pos Position) *piece.Piece {
	if !pos.IsValid() {
		return nil
	}
	p := b.squares[pos.Row][pos.Col]
	b.squares[pos.Row][pos.Col] = nil
	return p
}

// MovePiece moves a piece from one position to another.
// Returns an error if there's no piece at the source position
// or if the destination is not empty.
func (b *Board) MovePiece(from, to Position) error {
	if !from.IsValid() {
		return fmt.Errorf("invalid source position: %v", from)
	}
	if !to.IsValid() {
		return fmt.Errorf("invalid destination position: %v", to)
	}

	p := b.GetPiece(from)
	if p == nil {
		return fmt.Errorf("no piece at position %v", from)
	}

	if b.GetPiece(to) != nil {
		return fmt.Errorf("position %v is not empty", to)
	}

	b.RemovePiece(from)
	b.SetPiece(to, p)
	return nil
}

// CountPieces returns the count of pieces for each color.
func (b *Board) CountPieces() (red, black int) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b.squares[row][col]
			if p != nil {
				if p.Color == piece.Red {
					red++
				} else if p.Color == piece.Black {
					black++
				}
			}
		}
	}
	return
}

// Clone creates a deep copy of the board.
func (b *Board) Clone() *Board {
	newBoard := &Board{}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b.squares[row][col]
			if p != nil {
				// Create a new piece with the same properties
				newBoard.squares[row][col] = &piece.Piece{
					ID:     p.ID,
					Color:  p.Color,
					IsKing: p.IsKing,
				}
			}
		}
	}
	return newBoard
}

// String returns a string representation of the board for debugging.
func (b *Board) String() string {
	var result string
	result += "    0   1   2   3   4   5   6   7\n"
	result += "  +---+---+---+---+---+---+---+---+\n"
	for row := 0; row < 8; row++ {
		result += fmt.Sprintf("%d |", row)
		for col := 0; col < 8; col++ {
			p := b.squares[row][col]
			if p == nil {
				result += "   |"
			} else {
				result += fmt.Sprintf(" %s |", p.Symbol())
			}
		}
		result += "\n"
		result += "  +---+---+---+---+---+---+---+---+\n"
	}
	return result
}

// ToJSON returns a JSON-serializable representation of the board.
func (b *Board) ToJSON() [][]interface{} {
	result := make([][]interface{}, 8)
	for row := 0; row < 8; row++ {
		result[row] = make([]interface{}, 8)
		for col := 0; col < 8; col++ {
			p := b.squares[row][col]
			if p != nil {
				result[row][col] = map[string]interface{}{
					"id":      p.ID,
					"color":   p.Color.String(),
					"is_king": p.IsKing,
				}
			}
		}
	}
	return result
}
