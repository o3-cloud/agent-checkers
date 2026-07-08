package handlers

import (
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
	apiws "github.com/stackable-specs/agent-checkers/src/api/websocket"
)

func (h *Handlers) broadcastGameStarted(g *game.Game) {
	if h.broadcaster == nil || g == nil {
		return
	}
	players := []*dto.PlayerResponse{
		dto.NewPlayerResponse(g.RedPlayer),
		dto.NewPlayerResponse(g.BlackPlayer),
	}
	h.broadcaster.BroadcastEvent(g.ID, apiws.Event{
		Type: apiws.EventTypeGameStarted,
		Payload: apiws.GameStartedPayload{
			GameID:  g.ID,
			Players: players,
		},
	})
}

func (h *Handlers) broadcastMoveMade(g *game.Game, move game.Move) {
	if h.broadcaster == nil || g == nil {
		return
	}
	h.broadcaster.BroadcastEvent(g.ID, apiws.Event{
		Type: apiws.EventTypeMoveMade,
		Payload: apiws.MoveMadePayload{
			From:     move.From,
			To:       move.To,
			Captured: move.Captured,
			Board:    g.Board.ToJSON(),
			Turn:     g.CurrentTurn.String(),
		},
	})
}

func (h *Handlers) broadcastTurnChanged(g *game.Game) {
	if h.broadcaster == nil || g == nil {
		return
	}
	h.broadcaster.BroadcastEvent(g.ID, apiws.Event{
		Type: apiws.EventTypeTurnChanged,
		Payload: apiws.TurnChangedPayload{
			CurrentPlayer: currentPlayerID(g),
		},
	})
}

func (h *Handlers) broadcastGameEnded(g *game.Game) {
	if h.broadcaster == nil || g == nil || g.Result == nil {
		return
	}
	h.broadcaster.BroadcastEvent(g.ID, apiws.Event{
		Type: apiws.EventTypeGameEnded,
		Payload: apiws.GameEndedPayload{
			Winner: g.Result.Winner,
			Reason: g.Result.Reason,
		},
	})
}

func currentPlayerID(g *game.Game) string {
	current := g.CurrentPlayer()
	if current == nil {
		return g.CurrentTurn.String()
	}
	return current.ID
}
