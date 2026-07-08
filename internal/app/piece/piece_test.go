package piece

import (
	"testing"
)

func TestColorString(t *testing.T) {
	tests := []struct {
		color    Color
		expected string
	}{
		{Red, "red"},
		{Black, "black"},
		{None, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.color.String(); got != tt.expected {
				t.Errorf("Color.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	p := New("r1", Red)
	if p.ID != "r1" {
		t.Errorf("New().ID = %v, want r1", p.ID)
	}
	if p.Color != Red {
		t.Errorf("New().Color = %v, want Red", p.Color)
	}
	if p.IsKing {
		t.Error("New().IsKing should be false")
	}
}

func TestPromote(t *testing.T) {
	p := New("r1", Red)
	if p.IsKing {
		t.Error("piece should not be king initially")
	}
	p.Promote()
	if !p.IsKing {
		t.Error("piece should be king after Promote()")
	}
	// Promoting again should be idempotent
	p.Promote()
	if !p.IsKing {
		t.Error("piece should still be king")
	}
}

func TestCanPromote(t *testing.T) {
	tests := []struct {
		name     string
		piece    *Piece
		row      int
		expected bool
	}{
		{
			name:     "red piece at row 7 can promote",
			piece:    New("r1", Red),
			row:      7,
			expected: true,
		},
		{
			name:     "red piece at row 0 cannot promote",
			piece:    New("r1", Red),
			row:      0,
			expected: false,
		},
		{
			name:     "red piece at row 5 cannot promote",
			piece:    New("r1", Red),
			row:      5,
			expected: false,
		},
		{
			name:     "black piece at row 0 can promote",
			piece:    New("b1", Black),
			row:      0,
			expected: true,
		},
		{
			name:     "black piece at row 7 cannot promote",
			piece:    New("b1", Black),
			row:      7,
			expected: false,
		},
		{
			name:     "king cannot promote again",
			piece:    &Piece{ID: "k1", Color: Red, IsKing: true},
			row:      7,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.piece.CanPromote(tt.row); got != tt.expected {
				t.Errorf("CanPromote(%d) = %v, want %v", tt.row, got, tt.expected)
			}
		})
	}
}

func TestSymbol(t *testing.T) {
	tests := []struct {
		name     string
		piece    *Piece
		expected string
	}{
		{
			name:     "nil piece",
			piece:    nil,
			expected: " ",
		},
		{
			name:     "red piece",
			piece:    New("r1", Red),
			expected: "○",
		},
		{
			name:     "black piece",
			piece:    New("b1", Black),
			expected: "●",
		},
		{
			name:     "red king",
			piece:    &Piece{ID: "rk1", Color: Red, IsKing: true},
			expected: "♚",
		},
		{
			name:     "black king",
			piece:    &Piece{ID: "bk1", Color: Black, IsKing: true},
			expected: "♛",
		},
		{
			name:     "none color",
			piece:    &Piece{ID: "n1", Color: None, IsKing: false},
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.piece.Symbol(); got != tt.expected {
				t.Errorf("Symbol() = %v, want %v", got, tt.expected)
			}
		})
	}
}
