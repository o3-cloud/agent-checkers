package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSession(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "session.json")
	want := Session{
		GameID:       "game-1",
		PlayerID:     "player-1",
		PlayerName:   "Alice",
		SessionToken: "token-1",
		ServerURL:    "http://example.test",
	}

	if err := Save(context.Background(), path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("session mode = %v, want 0600", mode)
	}

	got, err := Load(context.Background(), path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if *got != want {
		t.Fatalf("Load() = %+v, want %+v", *got, want)
	}
}

func TestDefaultPathUsesHomeDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	want := filepath.Join(home, ".agent-checkers", "session.json")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestLoadMissingSessionReturnsHelpfulError(t *testing.T) {
	t.Parallel()

	_, err := Load(context.Background(), filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
