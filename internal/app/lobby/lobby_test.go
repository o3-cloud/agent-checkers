package lobby

import (
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestRegisterPlayerCreatesWaitingGameForFirstPlayer(t *testing.T) {
	gameStore := store.NewMemoryStore()
	lobby := New(gameStore)

	g, player, err := lobby.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer() error = %v", err)
	}

	if player.ID == "" {
		t.Fatal("RegisterPlayer() player ID is empty")
	}
	if player.Name != "Alice" {
		t.Errorf("RegisterPlayer() player name = %q, want Alice", player.Name)
	}
	if player.Color != piece.Red {
		t.Errorf("RegisterPlayer() player color = %v, want red", player.Color)
	}
	if g.Status != game.StatusWaiting {
		t.Errorf("RegisterPlayer() game status = %v, want waiting", g.Status)
	}

	loaded, err := gameStore.LoadGame(g.ID)
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if loaded.RedPlayer == nil || loaded.RedPlayer.ID != player.ID {
		t.Fatalf("LoadGame() red player = %#v, want registered player", loaded.RedPlayer)
	}
}

func TestRegisterPlayerMatchesSecondPlayer(t *testing.T) {
	gameStore := store.NewMemoryStore()
	lobby := New(gameStore)

	waitingGame, redPlayer, err := lobby.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer(Alice) error = %v", err)
	}

	matchedGame, blackPlayer, err := lobby.RegisterPlayer("Bot", "ai")
	if err != nil {
		t.Fatalf("RegisterPlayer(Bot) error = %v", err)
	}

	if matchedGame.ID != waitingGame.ID {
		t.Errorf("RegisterPlayer() matched game ID = %q, want %q", matchedGame.ID, waitingGame.ID)
	}
	if matchedGame.Status != game.StatusActive {
		t.Errorf("RegisterPlayer() matched game status = %v, want active", matchedGame.Status)
	}
	if matchedGame.RedPlayer.ID != redPlayer.ID {
		t.Errorf("RegisterPlayer() red player changed")
	}
	if matchedGame.BlackPlayer.ID != blackPlayer.ID {
		t.Errorf("RegisterPlayer() black player ID = %q, want %q", matchedGame.BlackPlayer.ID, blackPlayer.ID)
	}
	if blackPlayer.Color != piece.Black {
		t.Errorf("RegisterPlayer() second player color = %v, want black", blackPlayer.Color)
	}
	if blackPlayer.Type != "ai" {
		t.Errorf("RegisterPlayer() second player type = %q, want ai", blackPlayer.Type)
	}
}

func TestJoinGameStartsWaitingGame(t *testing.T) {
	gameStore := store.NewMemoryStore()
	lobby := New(gameStore)

	waitingGame, _, err := lobby.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer() error = %v", err)
	}

	player, err := lobby.JoinGame(waitingGame.ID, "Bob", "human")
	if err != nil {
		t.Fatalf("JoinGame() error = %v", err)
	}
	if player.Color != piece.Black {
		t.Errorf("JoinGame() player color = %v, want black", player.Color)
	}

	loaded, err := gameStore.LoadGame(waitingGame.ID)
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if loaded.Status != game.StatusActive {
		t.Errorf("JoinGame() game status = %v, want active", loaded.Status)
	}
	if loaded.BlackPlayer == nil || loaded.BlackPlayer.ID != player.ID {
		t.Fatalf("JoinGame() black player = %#v, want joined player", loaded.BlackPlayer)
	}
}

func TestJoinGameRejectsInvalidGame(t *testing.T) {
	lobby := New(store.NewMemoryStore())

	if _, err := lobby.JoinGame("missing", "Bob", "human"); err == nil {
		t.Fatal("JoinGame() missing game error = nil, want error")
	}
}
