package session

import (
	"errors"
	"testing"
	"time"
)

func TestManagerCreateAndLoadSession(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	manager := NewManagerWithClock(time.Hour, func() time.Time { return now })

	session, err := manager.Create("player-1", "game-1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if session.ID == "" {
		t.Error("Create() ID is empty")
	}
	if session.Token == "" {
		t.Error("Create() Token is empty")
	}
	if session.PlayerID != "player-1" {
		t.Errorf("Create() PlayerID = %q, want player-1", session.PlayerID)
	}
	if session.GameID != "game-1" {
		t.Errorf("Create() GameID = %q, want game-1", session.GameID)
	}

	loaded, err := manager.Load(session.Token)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.ID != session.ID {
		t.Errorf("Load() ID = %q, want %q", loaded.ID, session.ID)
	}

	loaded.PlayerID = "changed"
	reloaded, err := manager.Load(session.Token)
	if err != nil {
		t.Fatalf("Load() after mutation error = %v", err)
	}
	if reloaded.PlayerID != "player-1" {
		t.Errorf("Load() returned mutable session reference")
	}
}

func TestManagerRejectsEmptyIdentity(t *testing.T) {
	manager := NewManager(time.Hour)

	if _, err := manager.Create("", "game-1"); err == nil {
		t.Fatal("Create() with empty player ID error = nil, want error")
	}
	if _, err := manager.Create("player-1", ""); err == nil {
		t.Fatal("Create() with empty game ID error = nil, want error")
	}
}

func TestManagerExpiresSessions(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	manager := NewManagerWithClock(time.Minute, func() time.Time { return now })

	session, err := manager.Create("player-1", "game-1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	now = now.Add(2 * time.Minute)
	_, err = manager.Load(session.Token)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() expired error = %v, want ErrNotFound", err)
	}
}

func TestManagerDeleteSession(t *testing.T) {
	manager := NewManager(time.Hour)
	session, err := manager.Create("player-1", "game-1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := manager.Delete(session.Token); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err = manager.Load(session.Token)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() deleted error = %v, want ErrNotFound", err)
	}
}
