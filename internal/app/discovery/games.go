// Package discovery provides read-side game queries for interface adapters.
package discovery

import (
	"errors"
	"sort"
	"strings"

	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

// ErrInvalidStatus is returned when a list query uses an unsupported status.
var ErrInvalidStatus = errors.New("invalid status")

// ListGames returns games matching the status and player filters.
func ListGames(gameStore store.GameStore, statusValue, playerID string) ([]*game.Game, error) {
	statusValue = strings.ToLower(strings.TrimSpace(statusValue))

	filter := store.GameFilter{PlayerID: playerID}
	defaultActiveOnly := statusValue == ""
	if statusValue != "" && statusValue != "all" {
		status, ok := game.ParseStatus(statusValue)
		if !ok {
			return nil, ErrInvalidStatus
		}
		filter.Status = status
		filter.StatusSet = true
	}

	games, err := gameStore.ListGames(filter)
	if err != nil {
		return nil, err
	}

	if defaultActiveOnly {
		games = activeAndWaitingOnly(games)
	}

	sort.Slice(games, func(i, j int) bool {
		return games[i].CreatedAt.After(games[j].CreatedAt)
	})
	return games, nil
}

func activeAndWaitingOnly(games []*game.Game) []*game.Game {
	filtered := make([]*game.Game, 0, len(games))
	for _, g := range games {
		if g.Status == game.StatusWaiting || g.Status == game.StatusActive {
			filtered = append(filtered, g)
		}
	}
	return filtered
}
