package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/lobby"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

// ---------------------------------------------------------------------------
// Existing list_games tests
// ---------------------------------------------------------------------------

func TestListGamesToolFiltersByStatus(t *testing.T) {
	gameStore := store.NewMemoryStore()
	waiting := mcpGameForTest(t, gameStore, "waiting-game", game.StatusWaiting, time.Now(), "alice", "")
	mcpGameForTest(t, gameStore, "active-game", game.StatusActive, time.Now().Add(-1*time.Hour), "bob", "carol")

	response, err := NewServer(gameStore).ListGames("waiting", "")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(response.Games) != 1 {
		t.Fatalf("games length = %d, want 1", len(response.Games))
	}
	if response.Games[0].GameID != waiting.ID || response.Games[0].Status != "waiting" {
		t.Fatalf("game = %#v, want waiting game", response.Games[0])
	}
}

func TestListGamesToolFiltersByPlayerID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	aliceGame := mcpGameForTest(t, gameStore, "alice-game", game.StatusActive, time.Now(), "alice", "bob")
	mcpGameForTest(t, gameStore, "other-game", game.StatusActive, time.Now().Add(-1*time.Hour), "carol", "dave")

	response, err := NewServer(gameStore).ListGames("", "alice")
	if err != nil {
		t.Fatalf("ListGames() error = %v", err)
	}
	if len(response.Games) != 1 || response.Games[0].GameID != aliceGame.ID {
		t.Fatalf("games = %#v, want alice game", response.Games)
	}
}

func TestServerRunHandlesListGamesToolCall(t *testing.T) {
	gameStore := store.NewMemoryStore()
	mcpGameForTest(t, gameStore, "waiting-game", game.StatusWaiting, time.Now(), "alice", "")
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_games","arguments":{"status":"waiting"}}}` + "\n")
	var output bytes.Buffer

	if err := NewServer(gameStore).Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var response map[string]any
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := response["result"].(map[string]any)
	games := result["games"].([]any)
	if len(games) != 1 {
		t.Fatalf("games length = %d, want 1", len(games))
	}
}

// ---------------------------------------------------------------------------
// tools/list
// ---------------------------------------------------------------------------

func TestToolsListIncludesAllTools(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n")
	var output bytes.Buffer
	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := resp["result"].(map[string]any)
	tools := result["tools"].([]any)

	want := map[string]bool{
		"list_games":      false,
		"register_player": false,
		"get_game_state":  false,
		"make_move":       false,
		"get_valid_moves": false,
		"resign":          false,
		"offer_draw":      false,
		"accept_draw":     false,
	}
	for _, tool := range tools {
		name := tool.(map[string]any)["name"].(string)
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing tool %q in tools/list", name)
		}
	}
}

// ---------------------------------------------------------------------------
// register_player
// ---------------------------------------------------------------------------

func TestRegisterPlayerDirectly(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	resp, err := srv.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer error = %v", err)
	}
	if !resp.Success || resp.PlayerID == "" || resp.GameID == "" || resp.Color == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Color != "red" {
		t.Errorf("first player color = %q, want red", resp.Color)
	}

	g, err := gameStore.LoadGame(resp.GameID)
	if err != nil {
		t.Fatalf("LoadGame error = %v", err)
	}
	if g.Status != game.StatusWaiting {
		t.Errorf("game status = %v, want waiting", g.Status)
	}
}

func TestRegisterPlayerViaJSONRPC(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	input := strings.NewReader(`{"jsonrpc":"2.0","id":42,"method":"tools/call","params":{"name":"register_player","arguments":{"name":"Bob","type":"ai"}}}` + "\n")
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]any)
	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
	if result["color"] != "red" {
		t.Errorf("color = %v, want red", result["color"])
	}
}

func TestRegisterPlayerValidationErrors(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	if _, err := srv.RegisterPlayer("", "human"); err == nil {
		t.Error("expected error for empty name")
	}
	if _, err := srv.RegisterPlayer("Alice", ""); err == nil {
		t.Error("expected error for empty type")
	}
}

// ---------------------------------------------------------------------------
// get_game_state
// ---------------------------------------------------------------------------

func TestGetGameStateDirectly(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	create, err := srv.RegisterPlayer("Alice", "human")
	if err != nil {
		t.Fatalf("RegisterPlayer error = %v", err)
	}

	resp, err := srv.GetGameState(create.GameID)
	if err != nil {
		t.Fatalf("GetGameState error = %v", err)
	}
	if !resp.Success || resp.GameState == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.GameState.Status != "waiting" {
		t.Errorf("status = %q, want waiting", resp.GameState.Status)
	}
}

func TestGetGameStateInvalidID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	_, err := srv.GetGameState("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent game")
	}
}

// ---------------------------------------------------------------------------
// make_move
// ---------------------------------------------------------------------------

func mcpActiveGame(t *testing.T, srv *Server) (gameID, redID, blackID string) {
	t.Helper()
	red, err := srv.RegisterPlayer("Red", "human")
	if err != nil {
		t.Fatalf("register red: %v", err)
	}
	// Join as second player via lobby directly
	black, err := srv.lobby.JoinGame(red.GameID, "Black", "ai")
	if err != nil {
		t.Fatalf("join game: %v", err)
	}
	return red.GameID, red.PlayerID, black.ID
}

func TestMakeMoveValid(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	gameID, redID, _ := mcpActiveGame(t, srv)
	_ = redID

	// Red moves first: from (2,1) to (3,0) is a legal simple move.
	from := board.Position{Row: 2, Col: 1}
	to := board.Position{Row: 3, Col: 0}

	resp, err := srv.MakeMove(gameID, from, to)
	if err != nil {
		t.Fatalf("MakeMove error = %v", err)
	}
	if !resp.Success {
		t.Fatalf("success = false")
	}
	if resp.GameState == nil {
		t.Fatal("game_state is nil")
	}
	// After red move, it should be black's turn
	if resp.GameState.CurrentTurn != "black" {
		t.Errorf("current_turn = %q, want black", resp.GameState.CurrentTurn)
	}
}

func TestMakeMoveViaJSONRPC(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, _ := mcpActiveGame(t, srv)

	reqJSON := `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"make_move","arguments":{"game_id":"` + gameID + `","from":{"row":2,"col":1},"to":{"row":3,"col":0}}}}` + "\n"
	input := strings.NewReader(reqJSON)
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]any)
	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
}

func TestMakeMoveInvalidReturnsValidMovesInErrorData(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, _ := mcpActiveGame(t, srv)

	// Try an invalid move: from (5,0) which is a black piece, but it's red's turn.
	from := board.Position{Row: 5, Col: 0}
	to := board.Position{Row: 4, Col: 1}

	_, err := srv.MakeMove(gameID, from, to)
	if err == nil {
		t.Fatal("expected error for wrong turn")
	}

	// Now verify via JSON-RPC that error.data has valid_moves.
	reqJSON := `{"jsonrpc":"2.0","id":99,"method":"tools/call","params":{"name":"make_move","arguments":{"game_id":"` + gameID + `","from":{"row":5,"col":0},"to":{"row":4,"col":1}}}}` + "\n"
	input := strings.NewReader(reqJSON)
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error in response")
	}
	data, hasData := errObj["data"].(map[string]any)
	if !hasData {
		t.Fatal("expected data field in error")
	}
	if _, ok := data["valid_moves"]; !ok {
		t.Error("expected valid_moves in error data")
	}
}

func TestMakeMoveInvalidGameID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	_, err := srv.MakeMove("nonexistent", board.Position{Row: 2, Col: 1}, board.Position{Row: 3, Col: 0})
	if err == nil {
		t.Fatal("expected error for nonexistent game")
	}
}

// ---------------------------------------------------------------------------
// get_valid_moves
// ---------------------------------------------------------------------------

func TestGetValidMovesDirectly(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, _ := mcpActiveGame(t, srv)

	resp, err := srv.GetValidMoves(gameID)
	if err != nil {
		t.Fatalf("GetValidMoves error = %v", err)
	}
	if !resp.Success {
		t.Fatal("success = false")
	}
	if len(resp.Moves) == 0 {
		t.Fatal("expected at least one valid move group")
	}
	// Each move group should have a from and non-empty to list.
	for _, mg := range resp.Moves {
		if !mg.From.IsValid() {
			t.Errorf("from = %v is invalid", mg.From)
		}
		if len(mg.To) == 0 {
			t.Errorf("no destinations for from %v", mg.From)
		}
	}
}

func TestGetValidMovesViaJSONRPC(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, _ := mcpActiveGame(t, srv)

	reqJSON := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_valid_moves","arguments":{"game_id":"` + gameID + `"}}}` + "\n"
	input := strings.NewReader(reqJSON)
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := resp["result"].(map[string]any)
	moves := result["moves"].([]any)
	if len(moves) == 0 {
		t.Fatal("expected at least one valid move group")
	}
}

// ---------------------------------------------------------------------------
// resign
// ---------------------------------------------------------------------------

func TestResignDirectly(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, redID, blackID := mcpActiveGame(t, srv)

	resp, err := srv.Resign(gameID, redID)
	if err != nil {
		t.Fatalf("Resign error = %v", err)
	}
	if !resp.Success {
		t.Fatal("success = false")
	}
	if resp.GameState.Status != "completed" {
		t.Errorf("status = %q, want completed", resp.GameState.Status)
	}
	// Winner should be black
	if resp.GameState.Result == nil {
		t.Fatal("result is nil")
	}
	if resp.GameState.Result.Winner != "black" {
		t.Errorf("winner = %q, want black", resp.GameState.Result.Winner)
	}
	_ = blackID
}

func TestResignInvalidPlayer(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, _ := mcpActiveGame(t, srv)

	_, err := srv.Resign(gameID, "nonexistent-player")
	if err == nil {
		t.Fatal("expected error for nonexistent player")
	}
}

// ---------------------------------------------------------------------------
// offer_draw and accept_draw
// ---------------------------------------------------------------------------

func TestOfferAndAcceptDraw(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, redID, blackID := mcpActiveGame(t, srv)

	// Red offers draw.
	offerResp, err := srv.OfferDraw(gameID, redID)
	if err != nil {
		t.Fatalf("OfferDraw error = %v", err)
	}
	if !offerResp.Success {
		t.Fatal("offer success = false")
	}
	if offerResp.GameState.Status != "active" {
		t.Errorf("status after offer = %q, want active", offerResp.GameState.Status)
	}

	// Black accepts draw.
	acceptResp, err := srv.AcceptDraw(gameID, blackID)
	if err != nil {
		t.Fatalf("AcceptDraw error = %v", err)
	}
	if !acceptResp.Success {
		t.Fatal("accept success = false")
	}
	if acceptResp.GameState.Status != "draw" {
		t.Errorf("status after accept = %q, want draw", acceptResp.GameState.Status)
	}
}

func TestAcceptDrawWithoutOffer(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, _, blackID := mcpActiveGame(t, srv)

	_, err := srv.AcceptDraw(gameID, blackID)
	if err == nil {
		t.Fatal("expected error for accept without offer")
	}
}

func TestOfferDrawViaJSONRPC(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)
	gameID, redID, _ := mcpActiveGame(t, srv)

	reqJSON := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"offer_draw","arguments":{"game_id":"` + gameID + `","player_id":"` + redID + `"}}}` + "\n"
	input := strings.NewReader(reqJSON)
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result := resp["result"].(map[string]any)
	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
}

// ---------------------------------------------------------------------------
// Error handling: invalid game_id for various tools
// ---------------------------------------------------------------------------

func TestGetValidMovesInvalidGameID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	_, err := srv.GetValidMoves("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent game")
	}
}

func TestResignInvalidGameID(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	_, err := srv.Resign("nonexistent", "p1")
	if err == nil {
		t.Fatal("expected error for nonexistent game")
	}
}

func TestUnknownToolReturnsError(t *testing.T) {
	gameStore := store.NewMemoryStore()
	srv := NewServer(gameStore)

	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"bogus_tool","arguments":{}}}` + "\n")
	var output bytes.Buffer

	if err := srv.Run(context.Background(), input, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] == nil {
		t.Fatal("expected error for unknown tool")
	}
}

// ---------------------------------------------------------------------------
// ensure lobby.Lobby import is used (sanity for build)
// ---------------------------------------------------------------------------

var _ lobby.Lobby

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func mcpGameForTest(t *testing.T, gameStore store.GameStore, id string, status game.Status, createdAt time.Time, redID, blackID string) *game.Game {
	t.Helper()

	g := game.NewGame()
	g.ID = id
	g.CreatedAt = createdAt
	g.UpdatedAt = createdAt
	if redID != "" {
		if err := g.AddPlayer(&game.Player{ID: redID, Name: redID, Type: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	if blackID != "" {
		if err := g.AddPlayer(&game.Player{ID: blackID, Name: blackID, Type: "human"}); err != nil {
			t.Fatal(err)
		}
	}
	g.Status = status
	if err := gameStore.SaveGame(g); err != nil {
		t.Fatal(err)
	}
	return g
}

// silence unused import warnings for piece package when test build evolves.
var _ piece.Color
