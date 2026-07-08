package game

import (
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusWaiting, "waiting"},
		{StatusActive, "active"},
		{StatusCompleted, "completed"},
		{StatusDraw, "draw"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("Status.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewGame(t *testing.T) {
	g := NewGame()

	if g.ID == "" {
		t.Error("NewGame() ID is empty")
	}
	if g.Board == nil {
		t.Error("NewGame() Board is nil")
	}
	if g.CurrentTurn != piece.Red {
		t.Errorf("NewGame() CurrentTurn = %v, want Red", g.CurrentTurn)
	}
	if g.Status != StatusWaiting {
		t.Errorf("NewGame() Status = %v, want Waiting", g.Status)
	}
	if len(g.Moves) != 0 {
		t.Errorf("NewGame() Moves = %v, want empty", g.Moves)
	}

	// Verify initial board has 12 pieces each
	redCount, blackCount := g.Board.CountPieces()
	if redCount != 12 {
		t.Errorf("NewGame() red pieces = %d, want 12", redCount)
	}
	if blackCount != 12 {
		t.Errorf("NewGame() black pieces = %d, want 12", blackCount)
	}
}

func TestAddPlayer(t *testing.T) {
	g := NewGame()

	// Add first player
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	err := g.AddPlayer(p1)
	if err != nil {
		t.Errorf("AddPlayer(p1) error: %v", err)
	}
	if p1.Color != piece.Red {
		t.Errorf("AddPlayer(p1) Color = %v, want Red", p1.Color)
	}
	if g.RedPlayer != p1 {
		t.Error("AddPlayer(p1) RedPlayer not set")
	}
	if g.Status != StatusWaiting {
		t.Errorf("AddPlayer(p1) Status = %v, want Waiting", g.Status)
	}

	// Add second player
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	err = g.AddPlayer(p2)
	if err != nil {
		t.Errorf("AddPlayer(p2) error: %v", err)
	}
	if p2.Color != piece.Black {
		t.Errorf("AddPlayer(p2) Color = %v, want Black", p2.Color)
	}
	if g.BlackPlayer != p2 {
		t.Error("AddPlayer(p2) BlackPlayer not set")
	}
	if g.Status != StatusActive {
		t.Errorf("AddPlayer(p2) Status = %v, want Active", g.Status)
	}

	// Try to add third player
	p3 := &Player{ID: "p3", Name: "Charlie", Type: "human"}
	err = g.AddPlayer(p3)
	if err == nil {
		t.Error("AddPlayer(p3) expected error, got nil")
	}
}

func TestGetPlayer(t *testing.T) {
	g := NewGame()
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	addPlayer(t, g, p1)
	addPlayer(t, g, p2)

	if got := g.GetPlayer("p1"); got != p1 {
		t.Error("GetPlayer(p1) did not return correct player")
	}
	if got := g.GetPlayer("p2"); got != p2 {
		t.Error("GetPlayer(p2) did not return correct player")
	}
	if got := g.GetPlayer("nonexistent"); got != nil {
		t.Error("GetPlayer(nonexistent) should return nil")
	}
}

func TestCurrentPlayer(t *testing.T) {
	g := NewGame()
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	addPlayer(t, g, p1)
	addPlayer(t, g, p2)

	// Red moves first
	if got := g.CurrentPlayer(); got != p1 {
		t.Error("CurrentPlayer() should return red player initially")
	}

	g.SwitchTurn()
	if got := g.CurrentPlayer(); got != p2 {
		t.Error("CurrentPlayer() should return black player after switch")
	}
}

func TestSwitchTurn(t *testing.T) {
	g := NewGame()

	if g.CurrentTurn != piece.Red {
		t.Error("Initial turn should be Red")
	}

	g.SwitchTurn()
	if g.CurrentTurn != piece.Black {
		t.Error("After SwitchTurn(), turn should be Black")
	}

	g.SwitchTurn()
	if g.CurrentTurn != piece.Red {
		t.Error("After second SwitchTurn(), turn should be Red")
	}
}

func TestRecordMove(t *testing.T) {
	g := NewGame()
	move := Move{
		From:     board.Position{Row: 2, Col: 3},
		To:       board.Position{Row: 3, Col: 4},
		PlayerID: "p1",
	}

	g.RecordMove(move)

	if len(g.Moves) != 1 {
		t.Errorf("RecordMove() Moves length = %d, want 1", len(g.Moves))
	}
	if g.Moves[0].Timestamp.IsZero() {
		t.Error("RecordMove() should set timestamp")
	}
}

func TestIsGameOver(t *testing.T) {
	// New game should not be over
	g := NewGame()
	over, _ := g.IsGameOver()
	if over {
		t.Error("NewGame().IsGameOver() = true, want false")
	}

	// Game with no red pieces - black wins
	g2 := board.NewEmpty()
	g2.SetPiece(board.Position{Row: 5, Col: 0}, piece.New("b1", piece.Black))
	game := &Game{Board: g2}
	over, result := game.IsGameOver()
	if !over {
		t.Error("IsGameOver() with no red pieces = false, want true")
	}
	if result.Winner != "black" {
		t.Errorf("IsGameOver() winner = %v, want black", result.Winner)
	}

	// Game with no black pieces - red wins
	g3 := board.NewEmpty()
	g3.SetPiece(board.Position{Row: 2, Col: 1}, piece.New("r1", piece.Red))
	game2 := &Game{Board: g3}
	over, result = game2.IsGameOver()
	if !over {
		t.Error("IsGameOver() with no black pieces = false, want true")
	}
	if result.Winner != "red" {
		t.Errorf("IsGameOver() winner = %v, want red", result.Winner)
	}
}

func TestEndGame(t *testing.T) {
	g := NewGame()
	result := &Result{Winner: "red", Reason: "resignation"}

	g.EndGame(result)

	if g.Status != StatusCompleted {
		t.Errorf("EndGame() Status = %v, want Completed", g.Status)
	}
	if g.Result != result {
		t.Error("EndGame() Result not set")
	}
}

func TestOfferAndAcceptDraw(t *testing.T) {
	g := NewGame()
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	addPlayer(t, g, p1)
	addPlayer(t, g, p2)

	// Player 1 offers draw
	err := g.OfferDraw("p1")
	if err != nil {
		t.Errorf("OfferDraw() error: %v", err)
	}
	if g.Result.DrawOffe != "p1" {
		t.Error("OfferDraw() did not set draw offer")
	}

	// Player 1 cannot accept their own offer
	err = g.AcceptDraw("p1")
	if err == nil {
		t.Error("AcceptDraw() player accepting own offer should error")
	}

	// Player 2 accepts draw
	err = g.AcceptDraw("p2")
	if err != nil {
		t.Errorf("AcceptDraw() error: %v", err)
	}
	if g.Status != StatusDraw {
		t.Errorf("AcceptDraw() Status = %v, want Draw", g.Status)
	}
	if g.Result.Reason != "draw_agreement" {
		t.Errorf("AcceptDraw() Reason = %v, want draw_agreement", g.Result.Reason)
	}
}

func TestResign(t *testing.T) {
	g := NewGame()
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	addPlayer(t, g, p1)
	addPlayer(t, g, p2)

	// Player 1 (red) resigns
	err := g.Resign("p1")
	if err != nil {
		t.Errorf("Resign() error: %v", err)
	}
	if g.Status != StatusCompleted {
		t.Errorf("Resign() Status = %v, want Completed", g.Status)
	}
	if g.Result.Winner != "black" {
		t.Errorf("Resign() Winner = %v, want black", g.Result.Winner)
	}
	if g.Result.Reason != "resignation" {
		t.Errorf("Resign() Reason = %v, want resignation", g.Result.Reason)
	}
}

func TestClone(t *testing.T) {
	g := NewGame()
	p1 := &Player{ID: "p1", Name: "Alice", Type: "human"}
	p2 := &Player{ID: "p2", Name: "Bob", Type: "human"}
	addPlayer(t, g, p1)
	addPlayer(t, g, p2)

	clone := g.Clone()

	// Verify clone has same values
	if clone.ID != g.ID {
		t.Error("Clone() ID mismatch")
	}
	if clone.CurrentTurn != g.CurrentTurn {
		t.Error("Clone() CurrentTurn mismatch")
	}
	if clone.Status != g.Status {
		t.Error("Clone() Status mismatch")
	}

	// Verify board is independent
	clone.Board.RemovePiece(board.Position{Row: 0, Col: 1})
	redOriginal, _ := g.Board.CountPieces()
	redClone, _ := clone.Board.CountPieces()

	if redOriginal != 12 {
		t.Error("Original board was modified by clone modification")
	}
	if redClone != 11 {
		t.Error("Clone board modification did not work")
	}
}

func TestMakeMoveErrors(t *testing.T) {
	tests := []struct {
		name      string
		setupGame func(t *testing.T) *Game
		playerID  string
		shouldErr bool
	}{
		{
			name:      "game not active",
			setupGame: func(t *testing.T) *Game { return NewGame() }, // StatusWaiting
			playerID:  "p1",
			shouldErr: true,
		},
		{
			name: "player not in game",
			setupGame: func(t *testing.T) *Game {
				g := NewGame()
				addPlayer(t, g, &Player{ID: "p1", Name: "Alice", Type: "human"})
				addPlayer(t, g, &Player{ID: "p2", Name: "Bob", Type: "human"})
				return g
			},
			playerID:  "p3",
			shouldErr: true,
		},
		{
			name: "wrong turn",
			setupGame: func(t *testing.T) *Game {
				g := NewGame()
				addPlayer(t, g, &Player{ID: "p1", Name: "Alice", Type: "human"})
				addPlayer(t, g, &Player{ID: "p2", Name: "Bob", Type: "human"})
				return g
			},
			playerID:  "p2", // Black, but it's red's turn
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setupGame(t)
			err := g.MakeMove(tt.playerID, board.Position{Row: 2, Col: 3}, board.Position{Row: 3, Col: 4})
			if tt.shouldErr && err == nil {
				t.Error("MakeMove() expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("MakeMove() unexpected error: %v", err)
			}
		})
	}
}

func addPlayer(t *testing.T, g *Game, p *Player) {
	t.Helper()
	if err := g.AddPlayer(p); err != nil {
		t.Fatalf("AddPlayer(%s) error = %v", p.ID, err)
	}
}
