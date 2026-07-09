package store

import (
	"context"
	"errors"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
)

func TestRedisStoreSaveGameLoadGameRoundTrip(t *testing.T) {
	store := newTestRedisStore(t)
	g := game.NewGame()
	g.ID = testRedisKey(t, "game-round-trip")
	if err := g.AddPlayer(&game.Player{ID: "p1", Name: "Alice", Type: "human"}); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}
	g.Board.RemovePiece(board.Position{Row: 0, Col: 1})
	g.Moves = []game.Move{{
		From:      board.Position{Row: 2, Col: 3},
		To:        board.Position{Row: 3, Col: 4},
		PlayerID:  "p1",
		Timestamp: time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC),
		Captured:  []board.Position{{Row: 4, Col: 5}},
	}}

	if err := store.SaveGame(g); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	loaded, err := store.LoadGame(g.ID)
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if loaded.ID != g.ID {
		t.Fatalf("LoadGame() ID = %q, want %q", loaded.ID, g.ID)
	}
	if loaded.RedPlayer == nil || loaded.RedPlayer.ID != "p1" {
		t.Fatalf("LoadGame() red player = %#v, want p1", loaded.RedPlayer)
	}
	redCount, _ := loaded.Board.CountPieces()
	if redCount != 11 {
		t.Fatalf("LoadGame() red piece count = %d, want 11", redCount)
	}
	if len(loaded.Moves) != 1 || len(loaded.Moves[0].Captured) != 1 {
		t.Fatalf("LoadGame() moves = %#v, want persisted captured move", loaded.Moves)
	}
}

func TestRedisStoreDeleteGame(t *testing.T) {
	store := newTestRedisStore(t)
	g := game.NewGame()
	g.ID = testRedisKey(t, "delete")
	if err := store.SaveGame(g); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	if err := store.DeleteGame(g.ID); err != nil {
		t.Fatalf("DeleteGame() error = %v", err)
	}
	_, err := store.LoadGame(g.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("LoadGame(deleted) error = %v, want ErrNotFound", err)
	}
}

func TestRedisStoreListGamesWithStatusFilter(t *testing.T) {
	store := newTestRedisStore(t)
	waiting := game.NewGame()
	waiting.ID = testRedisKey(t, "waiting")
	active := game.NewGame()
	active.ID = testRedisKey(t, "active")
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
		t.Fatalf("ListGames(active) error = %v", err)
	}
	if gameIDs(activeGames) != active.ID {
		t.Fatalf("ListGames(active) IDs = %q, want %q", gameIDs(activeGames), active.ID)
	}

	waitingGames, err := store.ListGames(GameFilter{Status: game.StatusWaiting, StatusSet: true})
	if err != nil {
		t.Fatalf("ListGames(waiting) error = %v", err)
	}
	if gameIDs(waitingGames) != waiting.ID {
		t.Fatalf("ListGames(waiting) IDs = %q, want %q", gameIDs(waitingGames), waiting.ID)
	}
}

func TestRedisStoreSavePlayerLoadPlayerRoundTrip(t *testing.T) {
	store := newTestRedisStore(t)
	player := &game.Player{ID: testRedisKey(t, "player"), Name: "Alice", Type: "human"}

	if err := store.SavePlayer(player); err != nil {
		t.Fatalf("SavePlayer() error = %v", err)
	}
	loaded, err := store.LoadPlayer(player.ID)
	if err != nil {
		t.Fatalf("LoadPlayer() error = %v", err)
	}
	if loaded.ID != player.ID || loaded.Name != "Alice" {
		t.Fatalf("LoadPlayer() = %#v, want %#v", loaded, player)
	}
}

func TestRedisStoreAppendMoveGetMoveHistory(t *testing.T) {
	store := newTestRedisStore(t)
	g := game.NewGame()
	g.ID = testRedisKey(t, "moves")
	if err := store.SaveGame(g); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}

	move := game.Move{
		From:      board.Position{Row: 2, Col: 1},
		To:        board.Position{Row: 3, Col: 2},
		PlayerID:  "p1",
		Timestamp: time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC),
		Captured:  []board.Position{{Row: 4, Col: 3}},
	}
	if err := store.AppendMove(g.ID, move); err != nil {
		t.Fatalf("AppendMove() error = %v", err)
	}

	history, err := store.GetMoveHistory(g.ID)
	if err != nil {
		t.Fatalf("GetMoveHistory() error = %v", err)
	}
	if len(history) != 1 || history[0].From != move.From || history[0].To != move.To {
		t.Fatalf("GetMoveHistory() = %#v, want appended move", history)
	}
	if len(history[0].Captured) != 1 || history[0].Captured[0] != move.Captured[0] {
		t.Fatalf("GetMoveHistory() captured = %#v, want %#v", history[0].Captured, move.Captured)
	}

	loaded, err := store.LoadGame(g.ID)
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if len(loaded.Moves) != 1 {
		t.Fatalf("LoadGame() move count = %d, want 1", len(loaded.Moves))
	}
}

func newTestRedisStore(t *testing.T) *RedisStore {
	t.Helper()

	rawURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if rawURL == "" {
		t.Skip("REDIS_URL is not set")
	}

	addr, password, db := parseTestRedisURL(t, rawURL)
	prefix := "agent-checkers:test:" + testRedisKey(t, "")
	store, err := NewRedisStore(addr, password, db, 2, prefix)
	if err != nil {
		t.Skipf("Redis is not available: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		iter := store.client.Scan(ctx, 0, prefix+":*", 0).Iterator()
		keys := make([]string, 0)
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if len(keys) > 0 {
			_ = store.client.Del(ctx, keys...).Err()
		}
		_ = store.client.Close()
	})

	return store
}

func parseTestRedisURL(t *testing.T, rawURL string) (string, string, int) {
	t.Helper()
	opts, err := redis.ParseURL(rawURL)
	if err == nil {
		return opts.Addr, opts.Password, opts.DB
	}

	parsed, parseErr := url.Parse(rawURL)
	if parseErr != nil || parsed.Host == "" {
		t.Fatalf("invalid REDIS_URL %q: %v", rawURL, err)
	}
	password, _ := parsed.User.Password()
	db := 0
	if parsed.Path != "" && parsed.Path != "/" {
		value, convErr := strconv.Atoi(strings.TrimPrefix(parsed.Path, "/"))
		if convErr != nil {
			t.Fatalf("invalid Redis DB in REDIS_URL %q: %v", rawURL, convErr)
		}
		db = value
	}
	return parsed.Host, password, db
}

func testRedisKey(t *testing.T, suffix string) string {
	t.Helper()
	key := strings.NewReplacer("/", "-", " ", "-", ":", "-").Replace(strings.ToLower(t.Name()))
	if suffix == "" {
		return key
	}
	return key + ":" + suffix
}

func gameIDs(games []*game.Game) string {
	ids := make([]string, len(games))
	for i, g := range games {
		ids[i] = g.ID
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}
