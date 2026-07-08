package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestListGamesToolFiltersByStatus(t *testing.T) {
	gameStore := store.NewMemoryStore()
	waiting := mcpGameForTest(t, gameStore, "waiting-game", game.StatusWaiting, time.Now(), "alice", "")
	mcpGameForTest(t, gameStore, "active-game", game.StatusActive, time.Now().Add(-1*time.Hour), "bob", "carol")

	response, err := NewServer(gameStore).ListGames("waiting", "")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(response.Games) != 1 {
		t.Fatalf("games length = %d, want 1", len(response.Games))
	}
	if response.Games[0].GameID != waiting.ID || response.Games[0].Status != "waiting" {
		t.Fatalf("game = %#v, want waiting game", response.Games[0])
	}
}

func TestListGamesToolFiltersByPlayerID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	aliceGame := mcpGameForTest(t, gameStore, "alice-game", game.StatusActive, time.Now(), "alice", "bob")
	mcpGameForTest(t, gameStore, "other-game", game.StatusActive, time.Now().Add(-1*time.Hour), "carol", "dave")

	response, err := NewServer(gameStore).ListGames("", "alice")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(response.Games) != 1 || response.Games[0].GameID != aliceGame.ID {
		t.Fatalf("games = %#v, want alice game", response.Games)
	}
}

func TestServerRunHandlesListGamesToolCall(t *testing.T) {
	gameStore := store.NewMemoryStore()
	mcpGameForTest(t, gameStore, "waiting-game", game.StatusWaiting, time.Now(), "alice", "")
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_games","arguments":{"status":"waiting"}}}` + "\n")
	var output bytes.Buffer

	if err := NewServer(gameStore).Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var response map[string]any
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := response["result"].(map[string]any)
	games := result["games"].([]any)
	if len(games) != 1 {
		t.Fatalf("games length = %d, want 1", len(games))
	}
}

func mcpGameForTest(t *testing.T, gameStore store.GameStore, id string, status game.Status, createdAt time.Time, redID, blackID string) *game.Game {
	t.Helper()

	g := game.NewGame()
	g.ID = id
	g.CreatedAt = createdAt
	g.UpdatedAt = createdAt
	if redID != "" {
		if err := g.AddPlayer(&game.Player{ID: redID, Name: redID, Type: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	if blackID != "" {
		if err := g.AddPlayer(&game.Player{ID: blackID, Name: blackID, Type: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	g.Status = status
	if err := gameStore.SaveGame(g); err != nil {
		t.Fatal(err)
	}
	return g
}
