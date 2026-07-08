package websocket

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// WebSocketHandler upgrades authenticated game update requests.
type WebSocketHandler struct {
	hub      *Hub
	sessions *session.Manager
	store    store.GameStore
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a handler for game WebSocket connections.
func NewWebSocketHandler(hub *Hub, sessionManager *session.Manager, gameStore store.GameStore) *WebSocketHandler {
	return &WebSocketHandler{
		hub:      hub,
		sessions: sessionManager,
		store:    gameStore,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		},
	}
}

// ServeHTTP authenticates the session token, upgrades the request, and streams events.
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.hub == nil || h.sessions == nil || h.store == nil {
		http.Error(w, "websocket dependencies unavailable", http.StatusInternalServerError)
		return
	}

	gameID := chi.URLParam(r, "id")
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing session token", http.StatusUnauthorized)
		return
	}

	loadedSession, err := h.sessions.Load(token)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, session.ErrNotFound) {
			status = http.StatusInternalServerError
		}
		http.Error(w, "invalid session token", status)
		return
	}
	if loadedSession.GameID != gameID {
		http.Error(w, "session is not valid for this game", http.StatusForbidden)
		return
	}

	g, err := h.store.LoadGame(gameID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	if g.GetPlayer(loadedSession.PlayerID) == nil {
		http.Error(w, "session player is not in this game", http.StatusForbidden)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := newClient(h.hub, conn, gameID, loadedSession.PlayerID)
	h.hub.RegisterClient(client)
	if err := client.SendEvent(Event{
		Type: EventTypeGameState,
		Payload: GameStatePayload{
			GameState: dto.NewGameState(g),
		},
	}); err != nil {
		h.hub.UnregisterClient(client)
		_ = conn.Close()
		return
	}

	go client.writePump()
	client.readPump()
}
