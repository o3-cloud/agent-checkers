package discovery

import (
	"errors"
	"testing"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestListGamesDefaultsToActiveAndWaitingSortedNewestFirst(t *testing.T) {
	gameStore := store.NewMemoryStore()
	waiting := discoveryGame(t, gameStore, "waiting", game.StatusWaiting, time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC), "")
	active := discoveryGame(t, gameStore, "active", game.StatusActive, time.Date(2026, 7, 8, 13, 0, 0, 0, time.UTC), "p1")
	discoveryGame(t, gameStore, "completed", game.StatusCompleted, time.Date(2026, 7, 8, 14, 0, 0, 0, time.UTC), "")

	games, err := ListGames(gameStore, "", "")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if got, want := discoveryIDs(games), []string{active.ID, waiting.ID}; !sameStrings(got, want) {
		t.Fatalf("ListGames() IDs = %#v, want %#v", got, want)
	}
}

func TestListGamesFiltersByStatusAndPlayer(t *testing.T) {
	gameStore := store.NewMemoryStore()
	active := discoveryGame(t, gameStore, "active", game.StatusActive, time.Date(2026, 7, 8, 13, 0, 0, 0, time.UTC), "p1")
	discoveryGame(t, gameStore, "waiting", game.StatusWaiting, time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC), "p2")

	games, err := ListGames(gameStore, " active ", "p1")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if got, want := discoveryIDs(games), []string{active.ID}; !sameStrings(got, want) {
		t.Fatalf("ListGames() IDs = %#v, want %#v", got, want)
	}
}

func TestListGamesAllIncludesCompleted(t *testing.T) {
	gameStore := store.NewMemoryStore()
	completed := discoveryGame(t, gameStore, "completed", game.StatusCompleted, time.Date(2026, 7, 8, 14, 0, 0, 0, time.UTC), "")
	waiting := discoveryGame(t, gameStore, "waiting", game.StatusWaiting, time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC), "")

	games, err := ListGames(gameStore, "all", "")
	if err != nil {
		t.Fatalf("ListGames(all) error = %v", err)
	}
	if got, want := discoveryIDs(games), []string{completed.ID, waiting.ID}; !sameStrings(got, want) {
		t.Fatalf("ListGames(all) IDs = %#v, want %#v", got, want)
	}
}

func TestListGamesRejectsInvalidStatus(t *testing.T) {
	_, err := ListGames(store.NewMemoryStore(), "paused", "")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("ListGames() error = %v, want ErrInvalidStatus", err)
	}
}

func discoveryGame(t *testing.T, gameStore store.GameStore, id string, status game.Status, createdAt time.Time, playerID string) *game.Game {
	t.Helper()
	g := game.NewGame()
	g.ID = id
	g.Status = status
	g.CreatedAt = createdAt
	g.UpdatedAt = createdAt
	if playerID != "" {
		g.RedPlayer = &game.Player{ID: playerID, Name: playerID, Type: "human"}
	}
	if err := gameStore.SaveGame(g); err != nil {
		t.Fatalf("SaveGame(%s) error = %v", id, err)
	}
	return g
}

func discoveryIDs(games []*game.Game) []string {
	ids := make([]string, len(games))
	for i, g := range games {
		ids[i] = g.ID
	}
	return ids
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
