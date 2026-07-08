// Package session tracks authenticated player sessions.
package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a session cannot be found or has expired.
var ErrNotFound = errors.New("session not found")

// Session represents a player identity for a game.
type Session struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"player_id"`
	GameID     string    `json:"game_id"`
	Token      string    `json:"token"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// Manager manages player sessions in memory.
type Manager struct {
	sessions map[string]*Session
	ttl      time.Duration
	now      func() time.Time
	mu       sync.RWMutex
}

// NewManager creates a session manager with the given time to live.
func NewManager(ttl time.Duration) *Manager {
	return NewManagerWithClock(ttl, time.Now)
}

// NewManagerWithClock creates a session manager with an injected clock.
func NewManagerWithClock(ttl time.Duration, now func() time.Time) *Manager {
	if now == nil {
		now = time.Now
	}
	return &Manager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
		now:      now,
	}
}

// Create creates a session for a player in a game.
func (m *Manager) Create(playerID, gameID string) (*Session, error) {
	if playerID == "" {
		return nil, errors.New("player ID is empty")
	}
	if gameID == "" {
		return nil, errors.New("game ID is empty")
	}

	now := m.now()
	session := &Session{
		ID:         uuid.New().String(),
		PlayerID:   playerID,
		GameID:     gameID,
		Token:      uuid.New().String(),
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(m.ttl),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.Token] = cloneSession(session)
	return cloneSession(session), nil
}

// Load retrieves a non-expired session by token and refreshes LastSeenAt.
func (m *Manager) Load(token string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[token]
	if !ok {
		return nil, fmt.Errorf("load session: %w", ErrNotFound)
	}

	now := m.now()
	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		delete(m.sessions, token)
		return nil, fmt.Errorf("load session: %w", ErrNotFound)
	}

	session.LastSeenAt = now
	return cloneSession(session), nil
}

// Delete removes a session by token.
func (m *Manager) Delete(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[token]; !ok {
		return fmt.Errorf("delete session: %w", ErrNotFound)
	}
	delete(m.sessions, token)
	return nil
}

func cloneSession(session *Session) *Session {
	if session == nil {
		return nil
	}
	cloned := *session
	return &cloned
}
