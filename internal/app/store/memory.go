// Package store provides persistence contracts and in-memory storage.
package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
)

// ErrNotFound is returned when a game or player cannot be found.
var ErrNotFound = errors.New("not found")

// GameFilter filters game list results.
type GameFilter struct {
	Status   game.Status
	PlayerID string
}

// GameStore defines the game persistence contract.
type GameStore interface {
	SaveGame(g *game.Game) error
	LoadGame(id string) (*game.Game, error)
	DeleteGame(id string) error
	ListGames(filter GameFilter) ([]*game.Game, error)
	SavePlayer(p *game.Player) error
	LoadPlayer(id string) (*game.Player, error)
	AppendMove(gameID string, move game.Move) error
	GetMoveHistory(gameID string) ([]game.Move, error)
}

// MemoryStore stores games and players in process memory.
type MemoryStore struct {
	games   map[string]*game.Game
	players map[string]*game.Player
	mu      sync.RWMutex
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		games:   make(map[string]*game.Game),
		players: make(map[string]*game.Player),
	}
}

// SaveGame persists a game state.
func (m *MemoryStore) SaveGame(g *game.Game) error {
	if g == nil {
		return errors.New("game is nil")
	}
	if g.ID == "" {
		return errors.New("game ID is empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.games[g.ID] = g.Clone()
	return nil
}

// LoadGame retrieves a game by ID.
func (m *MemoryStore) LoadGame(id string) (*game.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.games[id]
	if !ok {
		return nil, fmt.Errorf("load game %q: %w", id, ErrNotFound)
	}
	return g.Clone(), nil
}

// DeleteGame removes a game by ID.
func (m *MemoryStore) DeleteGame(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.games[id]; !ok {
		return fmt.Errorf("delete game %q: %w", id, ErrNotFound)
	}
	delete(m.games, id)
	return nil
}

// ListGames returns games matching the provided filter.
func (m *MemoryStore) ListGames(filter GameFilter) ([]*game.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]*game.Game, 0, len(m.games))
	for _, g := range m.games {
		if filter.Status != 0 && g.Status != filter.Status {
			continue
		}
		if filter.PlayerID != "" && !gameHasPlayer(g, filter.PlayerID) {
			continue
		}
		games = append(games, g.Clone())
	}
	return games, nil
}

// SavePlayer persists a player.
func (m *MemoryStore) SavePlayer(p *game.Player) error {
	if p == nil {
		return errors.New("player is nil")
	}
	if p.ID == "" {
		return errors.New("player ID is empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.players[p.ID] = clonePlayer(p)
	return nil
}

// LoadPlayer retrieves a player by ID.
func (m *MemoryStore) LoadPlayer(id string) (*game.Player, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.players[id]
	if !ok {
		return nil, fmt.Errorf("load player %q: %w", id, ErrNotFound)
	}
	return clonePlayer(p), nil
}

// AppendMove appends a move to a stored game's history.
func (m *MemoryStore) AppendMove(gameID string, move game.Move) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, ok := m.games[gameID]
	if !ok {
		return fmt.Errorf("append move for game %q: %w", gameID, ErrNotFound)
	}
	cloned := g.Clone()
	cloned.Moves = append(cloned.Moves, cloneMove(move))
	m.games[gameID] = cloned
	return nil
}

// GetMoveHistory retrieves the move history for a game.
func (m *MemoryStore) GetMoveHistory(gameID string) ([]game.Move, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.games[gameID]
	if !ok {
		return nil, fmt.Errorf("get move history for game %q: %w", gameID, ErrNotFound)
	}
	return cloneMoves(g.Moves), nil
}

func gameHasPlayer(g *game.Game, playerID string) bool {
	return (g.RedPlayer != nil && g.RedPlayer.ID == playerID) ||
		(g.BlackPlayer != nil && g.BlackPlayer.ID == playerID)
}

func clonePlayer(p *game.Player) *game.Player {
	if p == nil {
		return nil
	}
	return &game.Player{
		ID:    p.ID,
		Name:  p.Name,
		Color: p.Color,
		Type:  p.Type,
	}
}

func cloneMove(move game.Move) game.Move {
	cloned := move
	if move.Captured != nil {
		cloned.Captured = append([]board.Position(nil), move.Captured...)
	}
	return cloned
}

func cloneMoves(moves []game.Move) []game.Move {
	cloned := make([]game.Move, len(moves))
	for i, move := range moves {
		cloned[i] = cloneMove(move)
	}
	return cloned
}
