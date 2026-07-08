// Package piece defines the checkers piece type and its properties.
package piece

// Color represents the color of a checkers piece.
type Color int

const (
	// None indicates no piece (empty square).
	None Color = iota
	// Red is the red player's pieces.
	Red
	// Black is the black player's pieces.
	Black
)

// String returns the string representation of a color.
func (c Color) String() string {
	switch c {
	case Red:
		return "red"
	case Black:
		return "black"
	default:
		return "none"
	}
}

// Piece represents a checkers piece on the board.
type Piece struct {
	// ID is a unique identifier for the piece.
	ID string `json:"id"`
	// Color is the piece's color (Red or Black).
	Color Color `json:"color"`
	// IsKing indicates whether the piece has been crowned.
	IsKing bool `json:"is_king"`
}

// New creates a new piece with the given color.
func New(id string, color Color) *Piece {
	return &Piece{
		ID:     id,
		Color:  color,
		IsKing: false,
	}
}

// Promote promotes the piece to a king.
func (p *Piece) Promote() {
	p.IsKing = true
}

// CanPromote returns true if the piece can be promoted to a king.
// A piece can be promoted when it reaches the opposite end of the board.
// row is the current row position (0-7).
// For red pieces, promotion happens at row 7 (black's home row).
// For black pieces, promotion happens at row 0 (red's home row).
func (p *Piece) CanPromote(row int) bool {
	if p.IsKing {
		return false
	}
	if p.Color == Red && row == 7 {
		return true
	}
	if p.Color == Black && row == 0 {
		return true
	}
	return false
}

// Symbol returns a visual representation of the piece for CLI display.
func (p *Piece) Symbol() string {
	if p == nil {
		return " "
	}
	switch {
	case p.Color == Red && p.IsKing:
		return "♚" // Red king
	case p.Color == Red:
		return "○" // Red piece
	case p.Color == Black && p.IsKing:
		return "♛" // Black king
	case p.Color == Black:
		return "●" // Black piece
	default:
		return " "
	}
}
