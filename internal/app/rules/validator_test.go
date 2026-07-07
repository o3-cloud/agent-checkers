package rules

import (
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

func TestValidateMove_SimpleMove(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	// Red's first move - valid simple move
	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 3, Col: 4}

	err := v.ValidateMove(g, from, to)
	if err != nil {
		t.Errorf("ValidateMove() unexpected error: %v", err)
	}
}

func TestValidateMove_WrongTurn(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	// Try to move black piece on red's turn
	from := board.Position{Row: 5, Col: 2}
	to := board.Position{Row: 4, Col: 3}

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should error when moving on wrong turn")
	}
}

func TestValidateMove_InvalidDirection(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	// Try to move non-king backward (red piece moving "down")
	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 1, Col: 4}

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should error when moving backward")
	}
}

func TestValidateMove_KingCanMoveBackward(t *testing.T) {
	v := NewValidator()
	g := newTestGameWithKing()

	// King can move backward
	from := board.Position{Row: 4, Col: 3}
	to := board.Position{Row: 3, Col: 4}

	err := v.ValidateMove(g, from, to)
	if err != nil {
		t.Errorf("ValidateMove() king should be able to move backward: %v", err)
	}
}

func TestValidateMove_Capture(t *testing.T) {
	v := NewValidator()
	g := newTestGameWithCapture()

	// Red captures black piece
	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 4, Col: 5}

	err := v.ValidateMove(g, from, to)
	if err != nil {
		t.Errorf("ValidateMove() capture should be valid: %v", err)
	}
}

func TestValidateMove_MandatoryCapture(t *testing.T) {
	v := NewValidator()
	g := newTestGameWithCapture()

	// A simple move when capture is available should be rejected
	from := board.Position{Row: 2, Col: 1}
	to := board.Position{Row: 3, Col: 2}

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should reject simple move when capture is available")
	}

	// Check error message mentions capture
	if err != nil {
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Error("Error should be ValidationError")
		} else if ve.Message != "a capture is available, you must capture" {
			t.Errorf("Error message = %v, want mandatory capture message", ve.Message)
		}
	}
}

func TestValidateMove_EmptySource(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	from := board.Position{Row: 3, Col: 3} // Empty square
	to := board.Position{Row: 4, Col: 4}

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should error when source is empty")
	}
}

func TestValidateMove_OccupiedDestination(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 5, Col: 2} // Occupied by black

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should error when destination is occupied")
	}
}

func TestValidateMove_NonDiagonalMove(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 3, Col: 3} // Same column

	err := v.ValidateMove(g, from, to)
	if err == nil {
		t.Error("ValidateMove() should error for non-diagonal move")
	}
}

func TestGetValidMoves(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	// Get valid moves for a red piece at starting position
	pos := board.Position{Row: 2, Col: 3}
	moves, err := v.GetValidMoves(g, pos)

	if err != nil {
		t.Errorf("GetValidMoves() error: %v", err)
	}

	// At start, red piece at row 2 can move to row 3 diagonally
	if len(moves) < 1 {
		t.Error("GetValidMoves() should return at least one valid move")
	}
}

func TestGetAllValidMoves(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	moves, err := v.GetAllValidMoves(g)
	if err != nil {
		t.Errorf("GetAllValidMoves() error: %v", err)
	}

	// Red should have valid moves at the start
	if len(moves) == 0 {
		t.Error("GetAllValidMoves() should return moves for red at game start")
	}
}

func TestHasCaptures(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	// No captures available at start
	if v.HasCaptures(g, piece.Red) {
		t.Error("HasCaptures() should be false at game start")
	}

	// Create a capture situation
	g2 := newTestGameWithCapture()
	if !v.HasCaptures(g2, piece.Red) {
		t.Error("HasCaptures() should be true when capture is available")
	}
}

func TestExecuteMove(t *testing.T) {
	v := NewValidator()
	g := newTestGame()

	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 3, Col: 4}

	captured, err := v.ExecuteMove(g, from, to)
	if err != nil {
		t.Errorf("ExecuteMove() error: %v", err)
	}

	if len(captured) != 0 {
		t.Errorf("ExecuteMove() captured = %d, want 0 for simple move", len(captured))
	}

	// Verify piece moved
	if g.Board.GetPiece(from) != nil {
		t.Error("ExecuteMove() piece not removed from source")
	}
	if g.Board.GetPiece(to) == nil {
		t.Error("ExecuteMove() piece not placed at destination")
	}
}

func TestExecuteMove_Capture(t *testing.T) {
	v := NewValidator()
	g := newTestGameWithCapture()

	from := board.Position{Row: 2, Col: 3}
	to := board.Position{Row: 4, Col: 5}
	capturePos := board.Position{Row: 3, Col: 4}

	// Verify black piece exists before capture
	if g.Board.GetPiece(capturePos) == nil {
		t.Fatal("Test setup error: no piece to capture")
	}

	captured, err := v.ExecuteMove(g, from, to)
	if err != nil {
		t.Errorf("ExecuteMove() error: %v", err)
	}

	if len(captured) != 1 {
		t.Errorf("ExecuteMove() captured %d pieces, want 1", len(captured))
	}

	if captured[0] != capturePos {
		t.Errorf("ExecuteMove() captured position = %v, want %v", captured[0], capturePos)
	}

	// Verify captured piece removed
	if g.Board.GetPiece(capturePos) != nil {
		t.Error("ExecuteMove() captured piece not removed")
	}
}

func TestExecuteMove_Promotion(t *testing.T) {
	v := NewValidator()
	g := newTestGameWithPromotion()

	// Red piece at row 6 can promote by moving to row 7
	from := board.Position{Row: 6, Col: 3}
	to := board.Position{Row: 7, Col: 4}

	p := g.Board.GetPiece(from)
	if p == nil {
		t.Fatal("Test setup error: no piece at promotion position")
	}
	if p.IsKing {
		t.Fatal("Test setup error: piece is already a king")
	}

	_, err := v.ExecuteMove(g, from, to)
	if err != nil {
		t.Errorf("ExecuteMove() error: %v", err)
	}

	// Verify promotion
	movedPiece := g.Board.GetPiece(to)
	if movedPiece == nil {
		t.Fatal("Piece not at destination after move")
	}
	if !movedPiece.IsKing {
		t.Error("ExecuteMove() piece should be promoted to king")
	}
}

// Helper functions for test setup

func newTestGame() *game.Game {
	g := game.NewGame()
	g.AddPlayer(&game.Player{ID: "p1", Name: "Red", Type: "human"})
	g.AddPlayer(&game.Player{ID: "p2", Name: "Black", Type: "human"})
	return g
}

func newTestGameWithKing() *game.Game {
	g := game.NewGame()
	g.AddPlayer(&game.Player{ID: "p1", Name: "Red", Type: "human"})
	g.AddPlayer(&game.Player{ID: "p2", Name: "Black", Type: "human"})

	// Promote a red piece to king
	kingPiece := &piece.Piece{ID: "king", Color: piece.Red, IsKing: true}
	g.Board.SetPiece(board.Position{Row: 4, Col: 3}, kingPiece)

	return g
}

func newTestGameWithCapture() *game.Game {
	g := game.NewGame()
	g.AddPlayer(&game.Player{ID: "p1", Name: "Red", Type: "human"})
	g.AddPlayer(&game.Player{ID: "p2", Name: "Black", Type: "human"})

	// Set up a capture situation:
	// Red piece at (2,3), Black piece at (3,4), Empty at (4,5)
	b := board.NewEmpty()
	b.SetPiece(board.Position{Row: 2, Col: 3}, piece.New("r1", piece.Red))
	b.SetPiece(board.Position{Row: 3, Col: 4}, piece.New("b1", piece.Black))
	g.Board = b

	return g
}

func newTestGameWithPromotion() *game.Game {
	g := game.NewGame()
	g.AddPlayer(&game.Player{ID: "p1", Name: "Red", Type: "human"})
	g.AddPlayer(&game.Player{ID: "p2", Name: "Black", Type: "human"})

	// Set up promotion situation:
	// Red piece at row 6 (one move from promotion)
	b := board.NewEmpty()
	b.SetPiece(board.Position{Row: 6, Col: 3}, piece.New("r1", piece.Red))
	g.Board = b

	return g
}