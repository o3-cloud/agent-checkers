package store

import (
	"errors"
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
)

func TestMemoryStoreSaveAndLoadGame(t *testing.T) {
	store := NewMemoryStore()
	g := game.NewGame()
	if err := g.AddPlayer(&game.Player{ID: "p1", Name: "Alice", Type: "human"}); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}

	if err := store.SaveGame(g); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	loaded, err := store.LoadGame(g.ID)
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}

	if loaded.ID != g.ID {
		t.Errorf("LoadGame() ID = %q, want %q", loaded.ID, g.ID)
	}
	if loaded.RedPlayer == nil || loaded.RedPlayer.ID != "p1" {
		t.Fatalf("LoadGame() red player = %#v, want p1", loaded.RedPlayer)
	}

	loaded.RedPlayer.Name = "changed"
	loaded.Board.RemovePiece(board.Position{Row: 0, Col: 1})

	reloaded, err := store.LoadGame(g.ID)
	if err != nil {
		t.Fatalf("LoadGame() after mutation error = %v", err)
	}
	if reloaded.RedPlayer.Name != "Alice" {
		t.Errorf("LoadGame() returned mutable player reference")
	}
	redCount, _ := reloaded.Board.CountPieces()
	if redCount != 12 {
		t.Errorf("LoadGame() returned mutable board reference; red count = %d, want 12", redCount)
	}
}

func TestMemoryStoreLoadMissingGame(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.LoadGame("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("LoadGame() error = %v, want ErrNotFound", err)
	}
}

func TestMemoryStorePlayers(t *testing.T) {
	store := NewMemoryStore()
	player := &game.Player{ID: "p1", Name: "Alice", Type: "human"}

	if err := store.SavePlayer(player); err != nil {
		t.Fatalf("SavePlayer() error = %v", err)
	}

	loaded, err := store.LoadPlayer("p1")
	if err != nil {
		t.Fatalf("LoadPlayer() error = %v", err)
	}
	if loaded.Name != "Alice" {
		t.Errorf("LoadPlayer() Name = %q, want Alice", loaded.Name)
	}

	loaded.Name = "changed"
	reloaded, err := store.LoadPlayer("p1")
	if err != nil {
		t.Fatalf("LoadPlayer() after mutation error = %v", err)
	}
	if reloaded.Name != "Alice" {
		t.Errorf("LoadPlayer() returned mutable player reference")
	}
}

func TestMemoryStoreAppendAndGetMoveHistory(t *testing.T) {
	store := NewMemoryStore()
	g := game.NewGame()
	if err := store.SaveGame(g); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	move := game.Move{
		From:     board.Position{Row: 2, Col: 3},
		To:       board.Position{Row: 3, Col: 4},
		PlayerID: "p1",
	}
	if err := store.AppendMove(g.ID, move); err != nil {
		t.Fatalf("AppendMove() error = %v", err)
	}

	history, err := store.GetMoveHistory(g.ID)
	if err != nil {
		t.Fatalf("GetMoveHistory() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("GetMoveHistory() length = %d, want 1", len(history))
	}
	if history[0].From != move.From || history[0].To != move.To {
		t.Errorf("GetMoveHistory() move = %#v, want %#v", history[0], move)
	}

	history[0].PlayerID = "changed"
	reloaded, err := store.GetMoveHistory(g.ID)
	if err != nil {
		t.Fatalf("GetMoveHistory() after mutation error = %v", err)
	}
	if reloaded[0].PlayerID != "p1" {
		t.Errorf("GetMoveHistory() returned mutable history")
	}
}

func TestMemoryStoreListAndDeleteGames(t *testing.T) {
	store := NewMemoryStore()
	waiting := game.NewGame()
	active := game.NewGame()
	if err := active.AddPlayer(&game.Player{ID: "p1", Name: "Alice", Type: "human"}); err != nil {
		t.Fatalf("AddPlayer(p1) error = %v", err)
	}
	if err := active.AddPlayer(&game.Player{ID: "p2", Name: "Bob", Type: "human"}); err != nil {
		t.Fatalf("AddPlayer(p2) error = %v", err)
	}

	if err := store.SaveGame(waiting); err != nil {
		t.Fatalf("SaveGame(waiting) error = %v", err)
	}
	if err := store.SaveGame(active); err != nil {
		t.Fatalf("SaveGame(active) error = %v", err)
	}

	activeGames, err := store.ListGames(GameFilter{Status: game.StatusActive})
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(activeGames) != 1 || activeGames[0].ID != active.ID {
		t.Fatalf("ListGames(active) = %#v, want active game", activeGames)
	}

	waitingGames, err := store.ListGames(GameFilter{Status: game.StatusWaiting, StatusSet: true})
	if err != nil {
		t.Fatalf("ListGames(waiting) error = %v", err)
	}
	if len(waitingGames) != 1 || waitingGames[0].ID != waiting.ID {
		t.Fatalf("ListGames(waiting) = %#v, want waiting game", waitingGames)
	}

	if err := store.DeleteGame(waiting.ID); err != nil {
		t.Fatalf("DeleteGame() error = %v", err)
	}
	_, err = store.LoadGame(waiting.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("LoadGame(deleted) error = %v, want ErrNotFound", err)
	}
}
