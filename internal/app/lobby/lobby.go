// Package lobby provides player registration and matchmaking.
package lobby

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

type waitingPlayer struct {
	gameID string
}

// Lobby manages waiting players and game matchmaking.
type Lobby struct {
	waiting *waitingPlayer
	store   store.GameStore
	mu      sync.Mutex
}

// New creates a lobby backed by the provided store.
func New(gameStore store.GameStore) *Lobby {
	return &Lobby{store: gameStore}
}

// RegisterPlayer adds a player to the lobby and matches them when possible.
func (l *Lobby) RegisterPlayer(name string, playerType string) (*game.Game, *game.Player, error) {
	if err := validateRegistration(name, playerType); err != nil {
		return nil, nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	player := newPlayer(name, playerType)
	if l.waiting == nil {
		g := game.NewGame()
		if err := g.AddPlayer(player); err != nil {
			return nil, nil, fmt.Errorf("add first player: %w", err)
		}
		if err := l.store.SaveGame(g); err != nil {
			return nil, nil, fmt.Errorf("save waiting game: %w", err)
		}
		if err := l.store.SavePlayer(player); err != nil {
			return nil, nil, fmt.Errorf("save assigned player: %w", err)
		}
		l.waiting = &waitingPlayer{gameID: g.ID}
		return g, player, nil
	}

	g, err := l.store.LoadGame(l.waiting.gameID)
	if err != nil {
		l.waiting = nil
		return nil, nil, fmt.Errorf("load waiting game: %w", err)
	}
	if err := g.AddPlayer(player); err != nil {
		l.waiting = nil
		return nil, nil, fmt.Errorf("match waiting game: %w", err)
	}
	if err := l.store.SaveGame(g); err != nil {
		return nil, nil, fmt.Errorf("save matched game: %w", err)
	}
	if err := l.store.SavePlayer(player); err != nil {
		return nil, nil, fmt.Errorf("save assigned player: %w", err)
	}

	l.waiting = nil
	return g, player, nil
}

// JoinGame adds a player to an existing waiting game.
func (l *Lobby) JoinGame(gameID string, name string, playerType string) (*game.Player, error) {
	if gameID == "" {
		return nil, errors.New("game ID is empty")
	}
	if err := validateRegistration(name, playerType); err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	g, err := l.store.LoadGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("load game: %w", err)
	}

	player := newPlayer(name, playerType)
	if err := g.AddPlayer(player); err != nil {
		return nil, fmt.Errorf("join game: %w", err)
	}
	if err := l.store.SavePlayer(player); err != nil {
		return nil, fmt.Errorf("save player: %w", err)
	}
	if err := l.store.SaveGame(g); err != nil {
		return nil, fmt.Errorf("save game: %w", err)
	}

	if l.waiting != nil && l.waiting.gameID == gameID {
		l.waiting = nil
	}

	return player, nil
}

func validateRegistration(name string, playerType string) error {
	if name == "" {
		return errors.New("player name is empty")
	}
	if playerType == "" {
		return errors.New("player type is empty")
	}
	return nil
}

func newPlayer(name string, playerType string) *game.Player {
	return &game.Player{
		ID:   uuid.New().String(),
		Name: name,
		Type: playerType,
	}
}
