// Package handlers adapts HTTP requests to the application layer.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/lobby"
	"github.com/stackable-specs/agent-checkers/internal/app/rules"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
	apiws "github.com/stackable-specs/agent-checkers/src/api/websocket"
)

// Lobby registers players for new and existing games.
type Lobby interface {
	RegisterPlayer(name string, playerType string) (*game.Game, *game.Player, error)
	JoinGame(gameID string, name string, playerType string) (*game.Player, error)
}

// SessionCreator creates player sessions for API clients.
type SessionCreator interface {
	Create(playerID, gameID string) (*session.Session, error)
}

// Broadcaster publishes real-time game events.
type Broadcaster interface {
	BroadcastEvent(gameID string, event apiws.Event)
}

// Handlers owns REST dependencies.
type Handlers struct {
	store       store.GameStore
	lobby       Lobby
	sessions    SessionCreator
	validator   *rules.Validator
	broadcaster Broadcaster
}

// New creates REST handlers backed by a game store.
func New(gameStore store.GameStore, sessions SessionCreator) *Handlers {
	return NewWithLobby(gameStore, lobby.New(gameStore), sessions)
}

// NewWithLobby creates REST handlers with an injected lobby.
func NewWithLobby(gameStore store.GameStore, gameLobby Lobby, sessions SessionCreator) *Handlers {
	return &Handlers{
		store:     gameStore,
		lobby:     gameLobby,
		sessions:  sessions,
		validator: rules.NewValidator(),
	}
}

// NewWithBroadcaster creates REST handlers with real-time event broadcasting.
func NewWithBroadcaster(gameStore store.GameStore, sessions SessionCreator, broadcaster Broadcaster) *Handlers {
	h := New(gameStore, sessions)
	h.broadcaster = broadcaster
	return h
}

// RegisterRoutes attaches all REST routes to a chi router.
func (h *Handlers) RegisterRoutes(router chi.Router) {
	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/games", h.CreateGame)
		r.Post("/games/{id}/join", h.JoinGame)
		r.Get("/games/{id}", h.GetGame)
		r.Delete("/games/{id}", h.ResignGame)
		r.Post("/games/{id}/draw", h.OfferOrAcceptDraw)
		r.Post("/games/{id}/moves", h.MakeMove)
		r.Get("/games/{id}/moves", h.GetMoveHistory)
		r.Get("/games/{id}/valid-moves", h.GetValidMoves)
	})
	router.Get("/health", Health)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, dto.ErrorResponse{
		Error:      message,
		StatusCode: status,
	})
}

func errorStatus(err error) int {
	if errors.Is(err, store.ErrNotFound) {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func gameID(r *http.Request) string {
	return chi.URLParam(r, "id")
}
