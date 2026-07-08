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

func TestRegisterPlayerCreatesSeparateWaitingGames(t *testing.T) {
	gameStore := store.NewMemoryStore()
	lobby := New(gameStore)

	firstGame, firstPlayer, err := lobby.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer(Alice) error = %v", err)
	}

	secondGame, secondPlayer, err := lobby.RegisterPlayer("Bot", "ai")
	if err != nil {
		t.Fatalf("RegisterPlayer(Bot) error = %v", err)
	}

	if secondGame.ID == firstGame.ID {
		t.Fatalf("RegisterPlayer() reused game ID %q, want separate waiting games", secondGame.ID)
	}
	if firstGame.Status != game.StatusWaiting {
		t.Errorf("RegisterPlayer() first game status = %v, want waiting", firstGame.Status)
	}
	if secondGame.Status != game.StatusWaiting {
		t.Errorf("RegisterPlayer() second game status = %v, want waiting", secondGame.Status)
	}
	if firstGame.RedPlayer.ID != firstPlayer.ID {
		t.Errorf("RegisterPlayer() first game red player changed")
	}
	if secondGame.RedPlayer.ID != secondPlayer.ID {
		t.Errorf("RegisterPlayer() second game red player ID = %q, want %q", secondGame.RedPlayer.ID, secondPlayer.ID)
	}
	if secondPlayer.Color != piece.Red {
		t.Errorf("RegisterPlayer() second player color = %v, want red", secondPlayer.Color)
	}
	if secondPlayer.Type != "ai" {
		t.Errorf("RegisterPlayer() second player type = %q, want ai", secondPlayer.Type)
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

func TestJoinGameDoesNotAffectOtherWaitingGames(t *testing.T) {
	gameStore := store.NewMemoryStore()
	lobby := New(gameStore)

	gameOne, playerOne, err := lobby.RegisterPlayer("P1", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer(P1) error = %v", err)
	}
	gameTwo, playerThree, err := lobby.RegisterPlayer("P3", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer(P3) error = %v", err)
	}

	playerTwo, err := lobby.JoinGame(gameOne.ID, "P2", "human")
	if err != nil {
		t.Fatalf("JoinGame(gameOne) error = %v", err)
	}

	loadedOne, err := gameStore.LoadGame(gameOne.ID)
	if err != nil {
		t.Fatalf("LoadGame(gameOne) error = %v", err)
	}
	if loadedOne.Status != game.StatusActive {
		t.Errorf("gameOne status = %v, want active", loadedOne.Status)
	}
	if loadedOne.RedPlayer == nil || loadedOne.RedPlayer.ID != playerOne.ID {
		t.Fatalf("gameOne red player = %#v, want P1", loadedOne.RedPlayer)
	}
	if loadedOne.BlackPlayer == nil || loadedOne.BlackPlayer.ID != playerTwo.ID {
		t.Fatalf("gameOne black player = %#v, want P2", loadedOne.BlackPlayer)
	}

	loadedTwo, err := gameStore.LoadGame(gameTwo.ID)
	if err != nil {
		t.Fatalf("LoadGame(gameTwo) error = %v", err)
	}
	if loadedTwo.Status != game.StatusWaiting {
		t.Errorf("gameTwo status = %v, want waiting", loadedTwo.Status)
	}
	if loadedTwo.RedPlayer == nil || loadedTwo.RedPlayer.ID != playerThree.ID {
		t.Fatalf("gameTwo red player = %#v, want P3", loadedTwo.RedPlayer)
	}
	if loadedTwo.BlackPlayer != nil {
		t.Fatalf("gameTwo black player = %#v, want nil", loadedTwo.BlackPlayer)
	}
}
