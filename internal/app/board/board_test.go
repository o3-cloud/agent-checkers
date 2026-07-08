package board

import (
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

func TestPositionString(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected string
	}{
		{"a1", Position{Row: 0, Col: 0}, "a1"},
		{"h8", Position{Row: 7, Col: 7}, "h8"},
		{"e4", Position{Row: 3, Col: 4}, "e4"},
		{"invalid negative", Position{Row: -1, Col: 0}, "invalid"},
		{"invalid over", Position{Row: 8, Col: 0}, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.String(); got != tt.expected {
				t.Errorf("Position.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionIsValid(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected bool
	}{
		{"valid min", Position{Row: 0, Col: 0}, true},
		{"valid max", Position{Row: 7, Col: 7}, true},
		{"valid center", Position{Row: 4, Col: 4}, true},
		{"invalid negative row", Position{Row: -1, Col: 0}, false},
		{"invalid negative col", Position{Row: 0, Col: -1}, false},
		{"invalid over row", Position{Row: 8, Col: 0}, false},
		{"invalid over col", Position{Row: 0, Col: 8}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsValid(); got != tt.expected {
				t.Errorf("Position.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPositionIsPlayable(t *testing.T) {
	tests := []struct {
		name     string
		pos      Position
		expected bool
	}{
		// Row 0: playable at odd columns
		{"row 0 col 0", Position{Row: 0, Col: 0}, false},
		{"row 0 col 1", Position{Row: 0, Col: 1}, true},
		{"row 0 col 7", Position{Row: 0, Col: 7}, true},
		// Row 1: playable at even columns
		{"row 1 col 0", Position{Row: 1, Col: 0}, true},
		{"row 1 col 1", Position{Row: 1, Col: 1}, false},
		// Invalid positions
		{"invalid row", Position{Row: -1, Col: 0}, false},
		{"invalid col", Position{Row: 0, Col: 8}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.IsPlayable(); got != tt.expected {
				t.Errorf("Position.IsPlayable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsePosition(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Position
		shouldError bool
	}{
		{"a1", "a1", Position{Row: 0, Col: 0}, false},
		{"h8", "h8", Position{Row: 7, Col: 7}, false},
		{"e4", "e4", Position{Row: 3, Col: 4}, false},
		{"invalid short", "a", Position{}, true},
		{"invalid long", "abc", Position{}, true},
		{"invalid col", "i1", Position{}, true},
		{"invalid row", "a9", Position{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePosition(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Error("ParsePosition() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ParsePosition() unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("ParsePosition() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	b := New()

	// Check piece counts
	redCount, blackCount := b.CountPieces()
	if redCount != 12 {
		t.Errorf("New() red piece count = %d, want 12", redCount)
	}
	if blackCount != 12 {
		t.Errorf("New() black piece count = %d, want 12", blackCount)
	}

	// Check that pieces are on playable squares
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b.GetPiece(Position{Row: row, Col: col})
			if p != nil {
				pos := Position{Row: row, Col: col}
				if !pos.IsPlayable() {
					t.Errorf("piece found on non-playable square at %v", pos)
				}
			}
		}
	}

	// Check red pieces are on rows 0-2
	for row := 0; row < 3; row++ {
		for col := 0; col < 8; col++ {
			p := b.GetPiece(Position{Row: row, Col: col})
			if p != nil && p.Color != piece.Red {
				t.Errorf("expected red piece at row %d col %d, got %v", row, col, p.Color)
			}
		}
	}

	// Check black pieces are on rows 5-7
	for row := 5; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b.GetPiece(Position{Row: row, Col: col})
			if p != nil && p.Color != piece.Black {
				t.Errorf("expected black piece at row %d col %d, got %v", row, col, p.Color)
			}
		}
	}
}

func TestNewEmpty(t *testing.T) {
	b := NewEmpty()

	redCount, blackCount := b.CountPieces()
	if redCount != 0 {
		t.Errorf("NewEmpty() red piece count = %d, want 0", redCount)
	}
	if blackCount != 0 {
		t.Errorf("NewEmpty() black piece count = %d, want 0", blackCount)
	}
}

func TestSetAndGetPiece(t *testing.T) {
	b := NewEmpty()
	pos := Position{Row: 3, Col: 4}
	p := piece.New("test", piece.Red)

	b.SetPiece(pos, p)
	got := b.GetPiece(pos)

	if got == nil {
		t.Fatal("GetPiece() returned nil, expected piece")
	}
	if got.ID != "test" {
		t.Errorf("GetPiece().ID = %v, want test", got.ID)
	}
	if got.Color != piece.Red {
		t.Errorf("GetPiece().Color = %v, want Red", got.Color)
	}
}

func TestRemovePiece(t *testing.T) {
	b := NewEmpty()
	pos := Position{Row: 3, Col: 4}
	p := piece.New("test", piece.Red)

	b.SetPiece(pos, p)
	removed := b.RemovePiece(pos)

	if removed == nil {
		t.Fatal("RemovePiece() returned nil, expected piece")
	}
	if removed.ID != "test" {
		t.Errorf("RemovePiece().ID = %v, want test", removed.ID)
	}
	if b.GetPiece(pos) != nil {
		t.Error("RemovePiece() did not remove the piece")
	}
}

func TestMovePiece(t *testing.T) {
	b := NewEmpty()
	from := Position{Row: 2, Col: 3}
	to := Position{Row: 3, Col: 4}
	p := piece.New("test", piece.Red)

	b.SetPiece(from, p)

	err := b.MovePiece(from, to)
	if err != nil {
		t.Errorf("MovePiece() unexpected error: %v", err)
	}

	if b.GetPiece(from) != nil {
		t.Error("MovePiece() did not remove piece from source")
	}
	got := b.GetPiece(to)
	if got == nil {
		t.Fatal("MovePiece() did not place piece at destination")
	}
	if got.ID != "test" {
		t.Errorf("MovePiece() piece ID = %v, want test", got.ID)
	}
}

func TestMovePieceErrors(t *testing.T) {
	tests := []struct {
		name       string
		setupBoard func() *Board
		from       Position
		to         Position
		shouldErr  bool
	}{
		{
			name:       "move from empty square",
			setupBoard: func() *Board { return NewEmpty() },
			from:       Position{Row: 0, Col: 0},
			to:         Position{Row: 1, Col: 1},
			shouldErr:  true,
		},
		{
			name: "move to occupied square",
			setupBoard: func() *Board {
				b := NewEmpty()
				b.SetPiece(Position{Row: 2, Col: 3}, piece.New("p1", piece.Red))
				b.SetPiece(Position{Row: 3, Col: 4}, piece.New("p2", piece.Black))
				return b
			},
			from:      Position{Row: 2, Col: 3},
			to:        Position{Row: 3, Col: 4},
			shouldErr: true,
		},
		{
			name:       "invalid source position",
			setupBoard: func() *Board { return NewEmpty() },
			from:       Position{Row: -1, Col: 0},
			to:         Position{Row: 0, Col: 1},
			shouldErr:  true,
		},
		{
			name:       "invalid destination position",
			setupBoard: func() *Board { return NewEmpty() },
			from:       Position{Row: 0, Col: 1},
			to:         Position{Row: 8, Col: 0},
			shouldErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setupBoard()
			err := b.MovePiece(tt.from, tt.to)
			if tt.shouldErr && err == nil {
				t.Error("MovePiece() expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("MovePiece() unexpected error: %v", err)
			}
		})
	}
}

func TestClone(t *testing.T) {
	b := New()
	clone := b.Clone()

	// Verify counts are same
	redOriginal, blackOriginal := b.CountPieces()
	redClone, blackClone := clone.CountPieces()

	if redOriginal != redClone {
		t.Errorf("Clone() red count = %d, want %d", redClone, redOriginal)
	}
	if blackOriginal != blackClone {
		t.Errorf("Clone() black count = %d, want %d", blackClone, blackOriginal)
	}

	// Modify clone and verify original is unchanged
	clone.RemovePiece(Position{Row: 0, Col: 1})

	redOriginalAfter, _ := b.CountPieces()
	redCloneAfter, _ := clone.CountPieces()

	if redOriginalAfter != redOriginal {
		t.Error("Clone() modifications affected original board")
	}
	if redCloneAfter != redOriginal-1 {
		t.Errorf("Clone() piece count after modification = %d, want %d", redCloneAfter, redOriginal-1)
	}
}

func TestString(t *testing.T) {
	b := New()
	s := b.String()

	if s == "" {
		t.Error("String() returned empty string")
	}

	// Check that the string contains expected elements
	if s[0:4] != "    " {
		t.Error("String() format unexpected")
	}
}

func TestToJSON(t *testing.T) {
	b := New()
	json := b.ToJSON()

	if len(json) != 8 {
		t.Errorf("ToJSON() rows = %d, want 8", len(json))
	}
	for row := 0; row < 8; row++ {
		if len(json[row]) != 8 {
			t.Errorf("ToJSON() cols in row %d = %d, want 8", row, len(json[row]))
		}
	}

	// Check a piece is represented correctly
	piece := json[0][1]
	if piece == nil {
		t.Error("ToJSON() expected piece at row 0, col 1")
	} else {
		pieceMap, ok := piece.(map[string]interface{})
		if !ok {
			t.Error("ToJSON() piece is not a map")
		} else {
			if pieceMap["color"] != "red" {
				t.Errorf("ToJSON() piece color = %v, want red", pieceMap["color"])
			}
		}
	}
}
