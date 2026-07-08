package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	gorilla "github.com/gorilla/websocket"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestHubBroadcastsEventsToGameClients(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 1),
		gameID: "game-1",
	}
	hub.RegisterClient(client)

	hub.BroadcastEvent("game-1", Event{
		Type: EventTypeTurnChanged,
		Payload: TurnChangedPayload{
			CurrentPlayer: "player-2",
		},
	})

	select {
	case message := <-client.send:
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			t.Fatalf("decode broadcast event: %v", err)
		}
		if event.Type != EventTypeTurnChanged {
			t.Fatalf("event type = %q, want %q", event.Type, EventTypeTurnChanged)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast event")
	}
}

func TestEventMarshalUsesDocumentedEnvelope(t *testing.T) {
	message, err := json.Marshal(Event{
		Type: EventTypeGameEnded,
		Payload: GameEndedPayload{
			Winner: "red",
			Reason: "all_pieces_captured",
		},
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(message, &got); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if got["type"] != "game_ended" {
		t.Fatalf("type = %v, want game_ended", got["type"])
	}
	payload := got["payload"].(map[string]any)
	if payload["winner"] != "red" {
		t.Fatalf("winner = %v, want red", payload["winner"])
	}
}

func TestWebSocketConnectSendsCurrentGameState(t *testing.T) {
	gameStore := store.NewMemoryStore()
	sessionManager := session.NewManager(time.Hour)
	g := game.NewGame()
	red := &game.Player{ID: "red-player", Name: "Alice", Type: "human"}
	if err := g.AddPlayer(red); err != nil {
		t.Fatal(err)
	}
	if err := gameStore.SaveGame(g); err != nil {
		t.Fatal(err)
	}
	session, err := sessionManager.Create(red.ID, g.ID)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(webSocketRouter(gameStore, sessionManager))
	defer server.Close()

	conn, _, err := gorilla.DefaultDialer.Dial(wsURL(server.URL, "/api/v1/games/"+g.ID+"/ws?token="+session.Token), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	var event struct {
		Type    EventType `json:"type"`
		Payload struct {
			GameState struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"game_state"`
		} `json:"payload"`
	}
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatalf("read current state event: %v", err)
	}
	if event.Type != EventTypeGameState {
		t.Fatalf("event type = %q, want %q", event.Type, EventTypeGameState)
	}
	if event.Payload.GameState.ID != g.ID {
		t.Fatalf("game ID = %q, want %q", event.Payload.GameState.ID, g.ID)
	}
}

func TestWebSocketRejectsInvalidSessionToken(t *testing.T) {
	gameStore := store.NewMemoryStore()
	sessionManager := session.NewManager(time.Hour)
	g := game.NewGame()
	if err := gameStore.SaveGame(g); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(webSocketRouter(gameStore, sessionManager))
	defer server.Close()

	_, response, err := gorilla.DefaultDialer.Dial(wsURL(server.URL, "/api/v1/games/"+g.ID+"/ws?token=invalid"), nil)
	if err == nil {
		t.Fatal("dial succeeded, want authentication failure")
	}
	if response == nil || response.StatusCode != 401 {
		t.Fatalf("status = %v, want 401", responseStatus(response))
	}
}

func webSocketRouter(gameStore store.GameStore, sessionManager *session.Manager) *chi.Mux {
	router := chi.NewRouter()
	handler := NewWebSocketHandler(NewHub(), sessionManager, gameStore)
	router.Handle("/api/v1/games/{id}/ws", handler)
	return router
}

func wsURL(baseURL, path string) string {
	return "ws" + strings.TrimPrefix(baseURL, "http") + path
}

func responseStatus(response *http.Response) int {
	if response == nil {
		return 0
	}
	return response.StatusCode
}
