package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestHandlersCreateJoinAndGetGame(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	createBody := bytes.NewBufferString(`{"player_name":"Alice","player_type":"human"}`)
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games", createBody))

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body %s", createResponse.Code, http.StatusCreated, createResponse.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	gameID := created["game_id"].(string)
	player := created["player"].(map[string]any)
	if player["color"] != "red" {
		t.Fatalf("created player color = %v, want red", player["color"])
	}

	joinBody := bytes.NewBufferString(`{"player_name":"Bob","player_type":"human"}`)
	joinResponse := httptest.NewRecorder()
	router.ServeHTTP(joinResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/join", joinBody))

	if joinResponse.Code != http.StatusOK {
		t.Fatalf("join status = %d, want %d, body %s", joinResponse.Code, http.StatusOK, joinResponse.Body.String())
	}

	getResponse := httptest.NewRecorder()
	router.ServeHTTP(getResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+gameID, nil))

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body %s", getResponse.Code, http.StatusOK, getResponse.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(getResponse.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	state := got["game_state"].(map[string]any)
	if state["status"] != "active" {
		t.Fatalf("game status = %v, want active", state["status"])
	}
}

func TestHandlersMakeMoveRecordsHistory(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	g := game.NewGame()
	red := &game.Player{ID: "red-player", Name: "Alice", Type: "human"}
	black := &game.Player{ID: "black-player", Name: "Bob", Type: "human"}
	if err := g.AddPlayer(red); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer(black); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveGame(g); err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(`{"player_id":"red-player","from":{"row":2,"col":1},"to":{"row":3,"col":0}}`)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+g.ID+"/moves", body))

	if response.Code != http.StatusOK {
		t.Fatalf("move status = %d, want %d, body %s", response.Code, http.StatusOK, response.Body.String())
	}

	historyResponse := httptest.NewRecorder()
	router.ServeHTTP(historyResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+g.ID+"/moves", nil))

	if historyResponse.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d, body %s", historyResponse.Code, http.StatusOK, historyResponse.Body.String())
	}

	var history map[string]any
	if err := json.Unmarshal(historyResponse.Body.Bytes(), &history); err != nil {
		t.Fatalf("decode move history: %v", err)
	}
	moves := history["moves"].([]any)
	if len(moves) != 1 {
		t.Fatalf("move history length = %d, want 1", len(moves))
	}
}

func TestHandlersReturnJSONErrors(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/games/missing", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload["error"] == "" {
		t.Fatal("error response did not include an error message")
	}
}

type mockStore struct {
	games   map[string]*game.Game
	players map[string]*game.Player
}

func newMockStore() *mockStore {
	return &mockStore{
		games:   make(map[string]*game.Game),
		players: make(map[string]*game.Player),
	}
}

func (m *mockStore) SaveGame(g *game.Game) error {
	m.games[g.ID] = g.Clone()
	return nil
}

func (m *mockStore) LoadGame(id string) (*game.Game, error) {
	g, ok := m.games[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return g.Clone(), nil
}

func (m *mockStore) DeleteGame(id string) error {
	if _, ok := m.games[id]; !ok {
		return store.ErrNotFound
	}
	delete(m.games, id)
	return nil
}

func (m *mockStore) ListGames(filter store.GameFilter) ([]*game.Game, error) {
	return nil, nil
}

func (m *mockStore) SavePlayer(p *game.Player) error {
	m.players[p.ID] = p
	return nil
}

func (m *mockStore) LoadPlayer(id string) (*game.Player, error) {
	p, ok := m.players[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	cloned := *p
	return &cloned, nil
}

func (m *mockStore) AppendMove(gameID string, move game.Move) error {
	return errors.New("AppendMove should not be called by handlers")
}

func (m *mockStore) GetMoveHistory(gameID string) ([]game.Move, error) {
	g, ok := m.games[gameID]
	if !ok {
		return nil, store.ErrNotFound
	}
	return append([]game.Move(nil), g.Moves...), nil
}
