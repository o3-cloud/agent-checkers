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
	apiws "github.com/stackable-specs/agent-checkers/src/api/websocket"
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

func TestHandlersServeOpenAPI(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	jsonResponse := httptest.NewRecorder()
	router.ServeHTTP(jsonResponse, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	if jsonResponse.Code != http.StatusOK {
		t.Fatalf("json status = %d, want %d, body %s", jsonResponse.Code, http.StatusOK, jsonResponse.Body.String())
	}
	if got := jsonResponse.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("json content type = %q, want application/json", got)
	}
	var spec map[string]any
	if err := json.Unmarshal(jsonResponse.Body.Bytes(), &spec); err != nil {
		t.Fatalf("decode OpenAPI JSON: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Fatalf("openapi = %v, want 3.1.0", spec["openapi"])
	}

	yamlResponse := httptest.NewRecorder()
	router.ServeHTTP(yamlResponse, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))

	if yamlResponse.Code != http.StatusOK {
		t.Fatalf("yaml status = %d, want %d, body %s", yamlResponse.Code, http.StatusOK, yamlResponse.Body.String())
	}
	if got := yamlResponse.Header().Get("Content-Type"); got != "application/yaml" {
		t.Fatalf("yaml content type = %q, want application/yaml", got)
	}
	if yamlResponse.Body.Len() == 0 {
		t.Fatal("yaml body is empty")
	}
}

func TestHandlersCreateGamesAndJoinSpecificGameInIsolation(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	gameOneID := createGameForTest(t, router, "P1")
	gameTwoID := createGameForTest(t, router, "P3")

	joinBody := bytes.NewBufferString(`{"player_name":"P2","player_type":"human"}`)
	joinResponse := httptest.NewRecorder()
	router.ServeHTTP(joinResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameOneID+"/join", joinBody))

	if joinResponse.Code != http.StatusOK {
		t.Fatalf("join status = %d, want %d, body %s", joinResponse.Code, http.StatusOK, joinResponse.Body.String())
	}

	gameOne := decodeGameStateForTest(t, joinResponse.Body.Bytes())
	if gameOne["status"] != "active" {
		t.Fatalf("game one status = %v, want active", gameOne["status"])
	}
	gameOneRed := gameOne["red_player"].(map[string]any)
	if gameOneRed["name"] != "P1" {
		t.Fatalf("game one red player = %v, want P1", gameOneRed["name"])
	}
	gameOneBlack := gameOne["black_player"].(map[string]any)
	if gameOneBlack["name"] != "P2" {
		t.Fatalf("game one black player = %v, want P2", gameOneBlack["name"])
	}

	getGameTwoResponse := httptest.NewRecorder()
	router.ServeHTTP(getGameTwoResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+gameTwoID, nil))

	if getGameTwoResponse.Code != http.StatusOK {
		t.Fatalf("get game two status = %d, want %d, body %s", getGameTwoResponse.Code, http.StatusOK, getGameTwoResponse.Body.String())
	}

	gameTwo := decodeGameStateForTest(t, getGameTwoResponse.Body.Bytes())
	if gameTwo["status"] != "waiting" {
		t.Fatalf("game two status = %v, want waiting", gameTwo["status"])
	}
	gameTwoRed := gameTwo["red_player"].(map[string]any)
	if gameTwoRed["name"] != "P3" {
		t.Fatalf("game two red player = %v, want P3", gameTwoRed["name"])
	}
	if gameTwo["black_player"] != nil {
		t.Fatalf("game two black player = %v, want nil", gameTwo["black_player"])
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

func TestHandlersBroadcastGameStartedWhenSecondPlayerJoins(t *testing.T) {
	store := newMockStore()
	broadcaster := &recordingBroadcaster{}
	h := NewWithBroadcaster(store, nil, broadcaster)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	gameID := createGameForTest(t, router, "Alice")
	joinBody := bytes.NewBufferString(`{"player_name":"Bob","player_type":"human"}`)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/join", joinBody))

	if response.Code != http.StatusOK {
		t.Fatalf("join status = %d, want %d, body %s", response.Code, http.StatusOK, response.Body.String())
	}
	event := broadcaster.singleEvent(t)
	if event.gameID != gameID {
		t.Fatalf("broadcast game ID = %q, want %q", event.gameID, gameID)
	}
	if event.event.Type != apiws.EventTypeGameStarted {
		t.Fatalf("event type = %q, want %q", event.event.Type, apiws.EventTypeGameStarted)
	}
}

func TestHandlersBroadcastMoveAndTurnChangeWhenMoveSucceeds(t *testing.T) {
	store := newMockStore()
	broadcaster := &recordingBroadcaster{}
	h := NewWithBroadcaster(store, nil, broadcaster)
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
	events := broadcaster.events
	if len(events) != 2 {
		t.Fatalf("broadcast count = %d, want 2", len(events))
	}
	if events[0].event.Type != apiws.EventTypeMoveMade {
		t.Fatalf("first event type = %q, want %q", events[0].event.Type, apiws.EventTypeMoveMade)
	}
	if events[1].event.Type != apiws.EventTypeTurnChanged {
		t.Fatalf("second event type = %q, want %q", events[1].event.Type, apiws.EventTypeTurnChanged)
	}
}

func TestHandlersDrawOfferAndAccept(t *testing.T) {
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
	redPlayer := created["player"].(map[string]any)
	redPlayerID := redPlayer["id"].(string)

	joinBody := bytes.NewBufferString(`{"player_name":"Bob","player_type":"human"}`)
	joinResponse := httptest.NewRecorder()
	router.ServeHTTP(joinResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/join", joinBody))

	if joinResponse.Code != http.StatusOK {
		t.Fatalf("join status = %d, want %d, body %s", joinResponse.Code, http.StatusOK, joinResponse.Body.String())
	}

	var joined map[string]any
	if err := json.Unmarshal(joinResponse.Body.Bytes(), &joined); err != nil {
		t.Fatalf("decode join response: %v", err)
	}
	blackPlayer := joined["player"].(map[string]any)
	blackPlayerID := blackPlayer["id"].(string)

	offerBody := bytes.NewBufferString(`{"player_id":"` + redPlayerID + `"}`)
	offerResponse := httptest.NewRecorder()
	router.ServeHTTP(offerResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/draw", offerBody))

	if offerResponse.Code != http.StatusOK {
		t.Fatalf("offer status = %d, want %d, body %s", offerResponse.Code, http.StatusOK, offerResponse.Body.String())
	}

	var offered map[string]any
	if err := json.Unmarshal(offerResponse.Body.Bytes(), &offered); err != nil {
		t.Fatalf("decode offer response: %v", err)
	}
	if offered["success"] != true {
		t.Fatalf("offer success = %v, want true", offered["success"])
	}
	offeredState := offered["game_state"].(map[string]any)
	offeredResult := offeredState["result"].(map[string]any)
	if offeredResult["draw_offer"] != redPlayerID {
		t.Fatalf("draw offer = %v, want %v", offeredResult["draw_offer"], redPlayerID)
	}

	acceptBody := bytes.NewBufferString(`{"player_id":"` + blackPlayerID + `"}`)
	acceptResponse := httptest.NewRecorder()
	router.ServeHTTP(acceptResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/draw", acceptBody))

	if acceptResponse.Code != http.StatusOK {
		t.Fatalf("accept status = %d, want %d, body %s", acceptResponse.Code, http.StatusOK, acceptResponse.Body.String())
	}

	var accepted map[string]any
	if err := json.Unmarshal(acceptResponse.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode accept response: %v", err)
	}
	if accepted["success"] != true {
		t.Fatalf("accept success = %v, want true", accepted["success"])
	}
	acceptedState := accepted["game_state"].(map[string]any)
	if acceptedState["status"] != "draw" {
		t.Fatalf("accepted game status = %v, want draw", acceptedState["status"])
	}
	acceptedResult := acceptedState["result"].(map[string]any)
	if acceptedResult["reason"] != "draw_agreement" {
		t.Fatalf("draw reason = %v, want draw_agreement", acceptedResult["reason"])
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

func createGameForTest(t *testing.T, router http.Handler, playerName string) string {
	t.Helper()

	body := bytes.NewBufferString(`{"player_name":"` + playerName + `","player_type":"human"}`)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/games", body))

	if response.Code != http.StatusCreated {
		t.Fatalf("create game for %s status = %d, want %d, body %s", playerName, response.Code, http.StatusCreated, response.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	return created["game_id"].(string)
}

func decodeGameStateForTest(t *testing.T, payload []byte) map[string]any {
	t.Helper()

	var response map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("decode game response: %v", err)
	}
	return response["game_state"].(map[string]any)
}

type mockStore struct {
	games   map[string]*game.Game
	players map[string]*game.Player
}

type recordedEvent struct {
	gameID string
	event  apiws.Event
}

type recordingBroadcaster struct {
	events []recordedEvent
}

func (r *recordingBroadcaster) BroadcastEvent(gameID string, event apiws.Event) {
	r.events = append(r.events, recordedEvent{gameID: gameID, event: event})
}

func (r *recordingBroadcaster) singleEvent(t *testing.T) recordedEvent {
	t.Helper()
	if len(r.events) != 1 {
		t.Fatalf("broadcast count = %d, want 1", len(r.events))
	}
	return r.events[0]
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

func TestHandlersGetValidMoves(t *testing.T) {
	store := newMockStore()
	h := New(store, nil)
	router := chi.NewRouter()
	h.RegisterRoutes(router)

	// Create + join a game so it's active.
	gameID := createGameForTest(t, router, "Alice")

	joinBody := bytes.NewBufferString(`{"player_name":"Bob","player_type":"human"}`)
	joinResponse := httptest.NewRecorder()
	router.ServeHTTP(joinResponse, httptest.NewRequest(http.MethodPost, "/api/v1/games/"+gameID+"/join", joinBody))
	if joinResponse.Code != http.StatusOK {
		t.Fatalf("join status = %d, want %d, body %s", joinResponse.Code, http.StatusOK, joinResponse.Body.String())
	}

	// Red's first piece is at row 0, col 1 (playable square in row 0).
	// Request valid moves for that square.
	validMovesResponse := httptest.NewRecorder()
	router.ServeHTTP(validMovesResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+gameID+"/valid-moves?row=2&col=1", nil))

	if validMovesResponse.Code != http.StatusOK {
		t.Fatalf("valid-moves status = %d, want %d, body %s", validMovesResponse.Code, http.StatusOK, validMovesResponse.Body.String())
	}

	var validMovesResult map[string]any
	if err := json.Unmarshal(validMovesResponse.Body.Bytes(), &validMovesResult); err != nil {
		t.Fatalf("decode valid-moves response: %v", err)
	}
	if validMovesResult["success"] != true {
		t.Fatalf("success = %v, want true", validMovesResult["success"])
	}
	moves := validMovesResult["moves"].([]any)
	if len(moves) == 0 {
		t.Fatalf("expected non-empty valid moves for red piece at (2,1), got empty")
	}

	// Assert 404 for unknown game.
	unknownResponse := httptest.NewRecorder()
	router.ServeHTTP(unknownResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/nonexistent-id/valid-moves?row=0&col=1", nil))
	if unknownResponse.Code != http.StatusNotFound {
		t.Fatalf("unknown game status = %d, want %d", unknownResponse.Code, http.StatusNotFound)
	}

	// Assert 400 for bad query params.
	badRowResponse := httptest.NewRecorder()
	router.ServeHTTP(badRowResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+gameID+"/valid-moves?row=abc&col=1", nil))
	if badRowResponse.Code != http.StatusBadRequest {
		t.Fatalf("bad row param status = %d, want %d", badRowResponse.Code, http.StatusBadRequest)
	}

	badColResponse := httptest.NewRecorder()
	router.ServeHTTP(badColResponse, httptest.NewRequest(http.MethodGet, "/api/v1/games/"+gameID+"/valid-moves?row=2&col=xyz", nil))
	if badColResponse.Code != http.StatusBadRequest {
		t.Fatalf("bad col param status = %d, want %d", badColResponse.Code, http.StatusBadRequest)
	}
}
