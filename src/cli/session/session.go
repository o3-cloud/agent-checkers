// Package session manages persisted CLI sessions.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const sessionFile = "session.json"

// Session is the persisted player session for CLI commands.
type Session struct {
	GameID       string `json:"game_id"`
	PlayerID     string `json:"player_id"`
	PlayerName   string `json:"player_name"`
	SessionToken string `json:"session_token"`
	ServerURL    string `json:"server_url"`
}

// DefaultPath returns ~/.agent-checkers/session.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".agent-checkers", sessionFile), nil
}

// Load reads a session from disk.
func Load(_ context.Context, path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no saved session; run agent-checkers-cli new or join first")
		}
		return nil, fmt.Errorf("read session: %w", err)
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	if session.GameID == "" || session.PlayerID == "" {
		return nil, errors.New("saved session is missing game_id or player_id")
	}
	return &session, nil
}

// Save writes a session to disk.
func Save(_ context.Context, path string, session Session) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}
