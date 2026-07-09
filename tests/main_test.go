// Package tests contains BDD (Behavior-Driven Development) test scenarios
// for the agent-checkers project, implemented with godog against the real
// REST API using net/http/httptest.
//
// Feature files live in tests/features/ as Gherkin .feature files.
// This file implements the step definitions and a TestMain-based runner
// that executes all BDD scenarios via godog.TestSuite.
package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	gorilla "github.com/gorilla/websocket"
	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/piece"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
	"github.com/stackable-specs/agent-checkers/src/api"
	apiws "github.com/stackable-specs/agent-checkers/src/api/websocket"
	"github.com/stackable-specs/agent-checkers/src/mcp"
)

// ---------------------------------------------------------------------------
// TestMain — runs the godog BDD suite then standard Go tests.
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	format := "pretty"
	opts := &godog.Options{
		Paths:  []string{"features"},
		Format: format,
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	status := suite.Run()
	if status != 0 {
		os.Exit(status)
	}

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// bddCtx — shared state for a single BDD scenario.
// ---------------------------------------------------------------------------

type bddCtx struct {
	gameStore *store.MemoryStore
	sessMgr   *session.Manager
	router    http.Handler

	// MCP server for JSON-RPC tool testing.
	mcpServer *mcp.Server
	mcpInput  *strings.Builder
	mcpOutput *bytes.Buffer
	mcpResult map[string]any

	// WebSocket test server and connection.
	wsServer *httptest.Server
	wsConn   *gorilla.Conn
	wsToken  string
	wsEvents []apiws.Event

	gameID         string
	redID          string
	blackID        string
	lastResp       *httptest.ResponseRecorder
	lastBody       map[string]any
	boardSetupDone bool
}

func newBDDCtx() *bddCtx {
	s := &bddCtx{}
	s.reset()
	return s
}

func (s *bddCtx) reset() {
	// Close any leftover WebSocket resources.
	if s.wsConn != nil {
		_ = s.wsConn.Close()
		s.wsConn = nil
	}
	if s.wsServer != nil {
		s.wsServer.Close()
		s.wsServer = nil
	}

	s.gameStore = store.NewMemoryStore()
	s.sessMgr = session.NewManager(time.Hour)
	s.router = api.NewRouter(s.gameStore, s.sessMgr)
	s.mcpServer = mcp.NewServer(s.gameStore)
	s.mcpInput = &strings.Builder{}
	s.mcpOutput = &bytes.Buffer{}
	s.mcpResult = nil
	s.wsToken = ""
	s.wsEvents = nil
	s.gameID = ""
	s.redID = ""
	s.blackID = ""
	s.lastResp = nil
	s.lastBody = nil
	s.boardSetupDone = false
}

// ---------------------------------------------------------------------------
// HTTP helper
// ---------------------------------------------------------------------------

func (s *bddCtx) doRequest(method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)
	return resp
}

func (s *bddCtx) parseBody() {
	s.lastBody = nil
	if s.lastResp == nil {
		return
	}
	var result map[string]any
	if err := json.Unmarshal(s.lastResp.Body.Bytes(), &result); err != nil {
		return
	}
	s.lastBody = result
}

// ---------------------------------------------------------------------------
// Step implementations — Background
// ---------------------------------------------------------------------------

func (s *bddCtx) theGameServerIsRunning() error {
	if s.router == nil {
		return fmt.Errorf("router not initialized")
	}
	return nil
}

func (s *bddCtx) theGameStateIsEmpty() error {
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — Registration
// ---------------------------------------------------------------------------

func (s *bddCtx) iRegisterPlayerWithNameAndType(name, playerType string) error {
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, name, playerType)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games", body)
	s.parseBody()
	if s.lastResp.Code == http.StatusCreated {
		s.gameID = s.lastBody["game_id"].(string)
		player := s.lastBody["player"].(map[string]any)
		s.redID = player["id"].(string)
	}
	return nil
}

func (s *bddCtx) theResponseStatusShouldBe(code int) error {
	if s.lastResp == nil {
		return fmt.Errorf("no response recorded")
	}
	if s.lastResp.Code != code {
		return fmt.Errorf("expected status %d, got %d (body: %s)", code, s.lastResp.Code, s.lastResp.Body.String())
	}
	return nil
}

func (s *bddCtx) theResponseShouldContainAGameID() error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	if _, ok := s.lastBody["game_id"]; !ok {
		return fmt.Errorf("response does not contain game_id: %v", s.lastBody)
	}
	return nil
}

func (s *bddCtx) theResponseShouldContainAPlayerID() error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	player, ok := s.lastBody["player"].(map[string]any)
	if !ok {
		return fmt.Errorf("response does not contain player object")
	}
	if _, ok := player["id"]; !ok {
		return fmt.Errorf("player object does not contain id")
	}
	return nil
}

func (s *bddCtx) thePlayerShouldBeAssignedColor(color string) error {
	player, ok := s.lastBody["player"].(map[string]any)
	if !ok {
		return fmt.Errorf("response does not contain player object")
	}
	if player["color"] != color {
		return fmt.Errorf("expected player color %s, got %v", color, player["color"])
	}
	return nil
}

func (s *bddCtx) theGameStatusShouldBe(status string) error {
	// If we're in an MCP scenario, we need to fetch game state via REST first.
	if s.mcpResult != nil && s.lastBody == nil {
		// Fetch game state via REST API to populate lastBody.
		s.lastResp = s.doRequest(http.MethodGet, "/api/v1/games/"+s.gameID, "")
		s.parseBody()
	}
	return s.assertGameStateField("status", status)
}

func (s *bddCtx) theGameStateShouldHaveStatus(status string) error {
	return s.assertGameStateField("status", status)
}

func (s *bddCtx) theGameStateShouldHaveCurrentTurn(turn string) error {
	return s.assertGameStateField("current_turn", turn)
}

func (s *bddCtx) theRedPlayerShouldBe(name string) error {
	return s.assertPlayerName("red_player", name)
}

func (s *bddCtx) theBlackPlayerShouldBe(name string) error {
	return s.assertPlayerName("black_player", name)
}

func (s *bddCtx) theResponseShouldContainAnError() error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	errMsg, ok := s.lastBody["error"]
	if !ok {
		return fmt.Errorf("response does not contain error field: %v", s.lastBody)
	}
	if errMsg == "" || errMsg == nil {
		return fmt.Errorf("error field is empty")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — Join game
// ---------------------------------------------------------------------------

func (s *bddCtx) aGameIsWaitingForAnOpponentWithPlayer(name string) error {
	return s.createGameWithPlayer(name, "human")
}

func (s *bddCtx) aGameIsWaitingForAnOpponentWithPlayerOfType(name, playerType string) error {
	return s.createGameWithPlayer(name, playerType)
}

func (s *bddCtx) playerJoinsTheGameWithType(player, playerType string) error {
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, player, playerType)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/join", body)
	s.parseBody()
	if s.lastResp.Code == http.StatusOK {
		playerObj := s.lastBody["player"].(map[string]any)
		s.blackID = playerObj["id"].(string)
	}
	return nil
}

func (s *bddCtx) playerJoinsGameWithType(player, gameID, playerType string) error {
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, player, playerType)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+gameID+"/join", body)
	s.parseBody()
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — State management
// ---------------------------------------------------------------------------

func (s *bddCtx) aGameExistsWithTwoPlayers() error {
	if err := s.createGameWithPlayer("Alice", "human"); err != nil {
		return err
	}
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, "Claude", "ai")
	resp := s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/join", body)
	if resp.Code != http.StatusOK {
		return fmt.Errorf("failed to join game: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		return fmt.Errorf("failed to decode join response: %w", err)
	}
	player := result["player"].(map[string]any)
	s.blackID = player["id"].(string)
	return nil
}

func (s *bddCtx) iRequestTheGameState() error {
	s.lastResp = s.doRequest(http.MethodGet, "/api/v1/games/"+s.gameID, "")
	s.parseBody()
	return nil
}

func (s *bddCtx) iRequestTheGameStateForGame(gameID string) error {
	s.lastResp = s.doRequest(http.MethodGet, "/api/v1/games/"+gameID, "")
	s.parseBody()
	return nil
}

func (s *bddCtx) theBoardShouldBeAnXGrid() error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData, ok := gameState["board"].([]any)
	if !ok {
		return fmt.Errorf("no board in game state")
	}
	if len(boardData) != 8 {
		return fmt.Errorf("board has %d rows, want 8", len(boardData))
	}
	for i, row := range boardData {
		rowData, ok := row.([]any)
		if !ok {
			return fmt.Errorf("row %d is not an array", i)
		}
		if len(rowData) != 8 {
			return fmt.Errorf("row %d has %d cols, want 8", i, len(rowData))
		}
	}
	return nil
}

func (s *bddCtx) redPiecesShouldOccurRows(rng string) error {
	startRow, endRow, err := parseRange(rng)
	if err != nil {
		return err
	}
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	for row := startRow; row <= endRow; row++ {
		for col := 0; col < 8; col++ {
			cell := boardData[row].([]any)[col]
			if cell != nil {
				pieceMap, ok := cell.(map[string]any)
				if !ok {
					return fmt.Errorf("piece at row %d col %d is not a map", row, col)
				}
				if pieceMap["color"] != "red" {
					return fmt.Errorf("expected red piece at row %d col %d, got %v", row, col, pieceMap["color"])
				}
			}
		}
	}
	return nil
}

func (s *bddCtx) blackPiecesShouldOccurRows(rng string) error {
	startRow, endRow, err := parseRange(rng)
	if err != nil {
		return err
	}
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	for row := startRow; row <= endRow; row++ {
		for col := 0; col < 8; col++ {
			cell := boardData[row].([]any)[col]
			if cell != nil {
				pieceMap, ok := cell.(map[string]any)
				if !ok {
					return fmt.Errorf("piece at row %d col %d is not a map", row, col)
				}
				if pieceMap["color"] != "black" {
					return fmt.Errorf("expected black piece at row %d col %d, got %v", row, col, pieceMap["color"])
				}
			}
		}
	}
	return nil
}

func (s *bddCtx) eachPieceShouldHaveIdColorAndIsKingFields() error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			cell := boardData[row].([]any)[col]
			if cell == nil {
				continue
			}
			pieceMap, ok := cell.(map[string]any)
			if !ok {
				return fmt.Errorf("piece at row %d col %d is not a map", row, col)
			}
			for _, field := range []string{"id", "color", "is_king"} {
				if _, ok := pieceMap[field]; !ok {
					return fmt.Errorf("piece at row %d col %d missing %s", row, col, field)
				}
			}
		}
	}
	return nil
}

func (s *bddCtx) theBoardShouldHaveRedPieces(count int) error {
	return s.assertPieceCount("red", count)
}

func (s *bddCtx) theBoardShouldHaveBlackPieces(count int) error {
	return s.assertPieceCount("black", count)
}

func (s *bddCtx) emptySquaresShouldBeNull() error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			pos := board.Position{Row: row, Col: col}
			if !pos.IsPlayable() {
				cell := boardData[row].([]any)[col]
				if cell != nil {
					return fmt.Errorf("non-playable square at row %d col %d is not null", row, col)
				}
			}
		}
	}
	return nil
}

func (s *bddCtx) allPiecesShouldHaveIsKingFalseAtStart() error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			cell := boardData[row].([]any)[col]
			if cell == nil {
				continue
			}
			pieceMap, ok := cell.(map[string]any)
			if !ok {
				return fmt.Errorf("piece at row %d col %d is not a map", row, col)
			}
			isKing, ok := pieceMap["is_king"].(bool)
			if !ok {
				return fmt.Errorf("is_king at row %d col %d is not bool", row, col)
			}
			if isKing {
				return fmt.Errorf("piece at row %d col %d is a king at start", row, col)
			}
		}
	}
	return nil
}

func (s *bddCtx) iRequestTheHealthEndpoint() error {
	s.lastResp = s.doRequest(http.MethodGet, "/health", "")
	s.parseBody()
	return nil
}

func (s *bddCtx) theResponseShouldContainStatus(status string) error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	if s.lastBody["status"] != status {
		return fmt.Errorf("expected status %s, got %v", status, s.lastBody["status"])
	}
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — Moves
// ---------------------------------------------------------------------------

func (s *bddCtx) aRedPieceIsAtPositionRowCol(row, col int) error {
	return s.setupPiece(row, col, piece.Red)
}

func (s *bddCtx) itIsRedsTurn() error {
	return nil
}

func (s *bddCtx) theRedPlayerMovesFromRowColToRowCol(fromRow, fromCol, toRow, toCol int) error {
	body := fmt.Sprintf(`{"player_id":"%s","from":{"row":%d,"col":%d},"to":{"row":%d,"col":%d}}`,
		s.redID, fromRow, fromCol, toRow, toCol)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/moves", body)
	s.parseBody()
	return nil
}

func (s *bddCtx) theBlackPlayerMovesFromRowColToRowCol(fromRow, fromCol, toRow, toCol int) error {
	body := fmt.Sprintf(`{"player_id":"%s","from":{"row":%d,"col":%d},"to":{"row":%d,"col":%d}}`,
		s.blackID, fromRow, fromCol, toRow, toCol)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/moves", body)
	s.parseBody()
	return nil
}

func (s *bddCtx) theMoveShouldBeAccepted() error {
	if s.lastResp == nil {
		return fmt.Errorf("no response recorded")
	}
	if s.lastResp.Code != http.StatusOK {
		return fmt.Errorf("move was rejected: status %d, body: %s", s.lastResp.Code, s.lastResp.Body.String())
	}
	return nil
}

func (s *bddCtx) theMoveShouldBeRejected() error {
	if s.lastResp == nil {
		return fmt.Errorf("no response recorded")
	}
	if s.lastResp.Code == http.StatusOK {
		return fmt.Errorf("move was accepted but should have been rejected")
	}
	return nil
}

func (s *bddCtx) thePieceShouldNowBeAtRowCol(row, col int) error {
	return s.verifyPieceAt(row, col, true)
}

func (s *bddCtx) theTurnShouldPassTo(color string) error {
	return s.assertGameStateField("current_turn", color)
}

func (s *bddCtx) aBlackPieceIsAtPositionRowCol(row, col int) error {
	return s.setupPiece(row, col, piece.Black)
}

func (s *bddCtx) positionRowColIsEmpty(row, col int) error {
	return nil
}

func (s *bddCtx) setupPiece(row, col int, color piece.Color) error {
	g, err := s.gameStore.LoadGame(s.gameID)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}
	if !s.boardSetupDone {
		g.Board = board.NewEmpty()
		s.boardSetupDone = true
	}
	pos := board.Position{Row: row, Col: col}
	id := fmt.Sprintf("%s%d%d", color.String()[:1], row, col)
	g.Board.SetPiece(pos, piece.New(id, color))
	g.CurrentTurn = piece.Red
	g.Status = game.StatusActive
	if err := s.gameStore.SaveGame(g); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}
	return nil
}

func (s *bddCtx) theBlackPieceAtRowColShouldBeRemoved(row, col int) error {
	return s.verifyPieceAt(row, col, false)
}

func (s *bddCtx) thePieceAtRowColShouldBeAKing(row, col int) error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	cell := boardData[row].([]any)[col]
	if cell == nil {
		return fmt.Errorf("no piece at row %d col %d", row, col)
	}
	pieceMap, ok := cell.(map[string]any)
	if !ok {
		return fmt.Errorf("piece at row %d col %d is not a map", row, col)
	}
	isKing, ok := pieceMap["is_king"].(bool)
	if !ok {
		return fmt.Errorf("is_king is not a boolean")
	}
	if !isKing {
		return fmt.Errorf("piece at row %d col %d is not a king", row, col)
	}
	return nil
}

func (s *bddCtx) iRequestValidMovesForRowCol(row, col int) error {
	url := fmt.Sprintf("/api/v1/games/%s/valid-moves?row=%d&col=%d", s.gameID, row, col)
	s.lastResp = s.doRequest(http.MethodGet, url, "")
	s.parseBody()
	return nil
}

func (s *bddCtx) theResponseShouldContainAListOfValidMoves() error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	moves, ok := s.lastBody["moves"].([]any)
	if !ok {
		return fmt.Errorf("no moves array in response")
	}
	if len(moves) == 0 {
		return fmt.Errorf("expected at least one valid move, got empty list")
	}
	return nil
}

func (s *bddCtx) iRequestTheMoveHistory() error {
	s.lastResp = s.doRequest(http.MethodGet, "/api/v1/games/"+s.gameID+"/moves", "")
	s.parseBody()
	return nil
}

func (s *bddCtx) theMoveHistoryShouldContainMove(count int) error {
	if s.lastBody == nil {
		return fmt.Errorf("no response body")
	}
	moves, ok := s.lastBody["moves"].([]any)
	if !ok {
		return fmt.Errorf("no moves array in response")
	}
	if len(moves) != count {
		return fmt.Errorf("expected %d moves, got %d", count, len(moves))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — Game end
// ---------------------------------------------------------------------------

func (s *bddCtx) redHasPieceAtRowCol(count int, row, col int) error {
	g, err := s.gameStore.LoadGame(s.gameID)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}
	g.Board = board.NewEmpty()
	g.Board.SetPiece(board.Position{Row: row, Col: col}, piece.New("r1", piece.Red))
	g.Status = game.StatusActive
	if over, result := g.IsGameOver(); over {
		g.EndGame(result)
	}
	if err := s.gameStore.SaveGame(g); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}
	return nil
}

func (s *bddCtx) blackHasPiecesRemaining(count int) error {
	if count == 0 {
		g, err := s.gameStore.LoadGame(s.gameID)
		if err != nil {
			return fmt.Errorf("failed to load game: %w", err)
		}
		for row := 0; row < 8; row++ {
			for col := 0; col < 8; col++ {
				p := g.Board.GetPiece(board.Position{Row: row, Col: col})
				if p != nil && p.Color == piece.Black {
					g.Board.RemovePiece(board.Position{Row: row, Col: col})
				}
			}
		}
		if over, result := g.IsGameOver(); over {
			g.EndGame(result)
		}
		if err := s.gameStore.SaveGame(g); err != nil {
			return fmt.Errorf("failed to save game: %w", err)
		}
	}
	return nil
}

func (s *bddCtx) blackHasPiecesButNoValidMoves() error {
	g, err := s.gameStore.LoadGame(s.gameID)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}
	g.Board = board.NewEmpty()
	// Black piece at (0,1) — row 0 is the edge.  Non-king black pieces
	// move "forward" = decreasing row, so a piece at row 0 cannot move.
	g.Board.SetPiece(board.Position{Row: 0, Col: 1}, piece.New("b1", piece.Black))
	// Red piece elsewhere on the board that has valid forward moves.
	g.Board.SetPiece(board.Position{Row: 3, Col: 2}, piece.New("r1", piece.Red))
	g.CurrentTurn = piece.Black
	g.Status = game.StatusActive
	if over, result := g.IsGameOver(); over {
		g.EndGame(result)
	}
	if err := s.gameStore.SaveGame(g); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}
	return nil
}

func (s *bddCtx) redHasValidMovesAvailable() error {
	return nil
}

func (s *bddCtx) theWinnerShouldBe(winner string) error {
	// If we're in an MCP scenario, fetch game state via REST API.
	if s.mcpResult != nil && s.lastBody == nil {
		s.lastResp = s.doRequest(http.MethodGet, "/api/v1/games/"+s.gameID, "")
		s.parseBody()
	}
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	result, ok := gameState["result"].(map[string]any)
	if !ok {
		return fmt.Errorf("no result in game state: %v", gameState)
	}
	if result["winner"] != winner {
		return fmt.Errorf("expected winner %s, got %v", winner, result["winner"])
	}
	return nil
}

func (s *bddCtx) theResultReasonShouldBe(reason string) error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	result, ok := gameState["result"].(map[string]any)
	if !ok {
		return fmt.Errorf("no result in game state")
	}
	if result["reason"] != reason {
		return fmt.Errorf("expected reason %s, got %v", reason, result["reason"])
	}
	return nil
}

func (s *bddCtx) theGameIsInProgress() error {
	return nil
}

func (s *bddCtx) playerOffersADraw(player string) error {
	playerID := s.playerIDByName(player)
	body := fmt.Sprintf(`{"player_id":"%s"}`, playerID)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/draw", body)
	s.parseBody()
	return nil
}

func (s *bddCtx) playerAcceptsTheDraw(player string) error {
	playerID := s.playerIDByName(player)
	body := fmt.Sprintf(`{"player_id":"%s"}`, playerID)
	s.lastResp = s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/draw", body)
	s.parseBody()
	return nil
}

func (s *bddCtx) theBlackPlayerResigns() error {
	body := fmt.Sprintf(`{"player_id":"%s"}`, s.blackID)
	s.lastResp = s.doRequest(http.MethodDelete, "/api/v1/games/"+s.gameID, body)
	s.parseBody()
	return nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (s *bddCtx) createGameWithPlayer(name, playerType string) error {
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, name, playerType)
	resp := s.doRequest(http.MethodPost, "/api/v1/games", body)
	if resp.Code != http.StatusCreated {
		return fmt.Errorf("failed to create game: %d %s", resp.Code, resp.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		return fmt.Errorf("failed to decode create response: %w", err)
	}
	s.gameID = result["game_id"].(string)
	player := result["player"].(map[string]any)
	s.redID = player["id"].(string)
	return nil
}

func (s *bddCtx) assertGameStateField(field, expected string) error {
	// If we're in an MCP scenario, check the MCP response.
	if s.mcpResult != nil {
		resultMap, ok := s.mcpResult["result"].(map[string]any)
		if !ok {
			return fmt.Errorf("no result in MCP response")
		}
		gameState, ok := resultMap["game_state"].(map[string]any)
		if !ok {
			return fmt.Errorf("no game_state in MCP response: %v", resultMap)
		}
		if gameState[field] != expected {
			return fmt.Errorf("expected %s = %s, got %v", field, expected, gameState[field])
		}
		return nil
	}
	// REST API response.
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	if gameState[field] != expected {
		return fmt.Errorf("expected %s = %s, got %v", field, expected, gameState[field])
	}
	return nil
}

func (s *bddCtx) assertPlayerName(playerField, name string) error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	player, ok := gameState[playerField].(map[string]any)
	if !ok {
		return fmt.Errorf("no %s in game state", playerField)
	}
	if player["name"] != name {
		return fmt.Errorf("expected %s = %s, got %v", playerField, name, player["name"])
	}
	return nil
}

func (s *bddCtx) assertPieceCount(color string, expected int) error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	count := 0
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			cell := boardData[row].([]any)[col]
			if cell == nil {
				continue
			}
			pieceMap, ok := cell.(map[string]any)
			if !ok {
				continue
			}
			if pieceMap["color"] == color {
				count++
			}
		}
	}
	if count != expected {
		return fmt.Errorf("expected %d %s pieces, got %d", expected, color, count)
	}
	return nil
}

func (s *bddCtx) verifyPieceAt(row, col int, shouldExist bool) error {
	gameState, ok := s.lastBody["game_state"].(map[string]any)
	if !ok {
		return fmt.Errorf("no game_state in response")
	}
	boardData := gameState["board"].([]any)
	cell := boardData[row].([]any)[col]
	if shouldExist && cell == nil {
		return fmt.Errorf("expected piece at row %d col %d, got nil", row, col)
	}
	if !shouldExist && cell != nil {
		return fmt.Errorf("expected no piece at row %d col %d, got %v", row, col, cell)
	}
	return nil
}

func (s *bddCtx) playerIDByName(name string) string {
	switch name {
	case "Alice":
		return s.redID
	case "Claude":
		return s.blackID
	}
	return ""
}

func parseRange(rng string) (int, int, error) {
	parts := strings.Split(rng, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range: %s", rng)
	}
	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid range numbers: %s", rng)
	}
	return start, end, nil
}

// ---------------------------------------------------------------------------
// InitializeScenario — registers all step definitions with godog.
// ---------------------------------------------------------------------------

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := newBDDCtx()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.reset()
		return ctx, nil
	})

	// Background
	ctx.Step(`^the game server is running$`, s.theGameServerIsRunning)
	ctx.Step(`^the game state is empty$`, s.theGameStateIsEmpty)

	// Registration
	ctx.Step(`^I register a player with name "([^"]*)" and type "([^"]*)"$`, s.iRegisterPlayerWithNameAndType)
	ctx.Step(`^the response status should be (\d+)$`, s.theResponseStatusShouldBe)
	ctx.Step(`^the response should contain a game_id$`, s.theResponseShouldContainAGameID)
	ctx.Step(`^the response should contain a player_id$`, s.theResponseShouldContainAPlayerID)
	ctx.Step(`^the player should be assigned color "([^"]*)"$`, s.thePlayerShouldBeAssignedColor)
	ctx.Step(`^the game status should be "([^"]*)"$`, s.theGameStatusShouldBe)
	ctx.Step(`^the game state should have status "([^"]*)"$`, s.theGameStateShouldHaveStatus)
	ctx.Step(`^the game state should have current_turn "([^"]*)"$`, s.theGameStateShouldHaveCurrentTurn)
	ctx.Step(`^the red player should be "([^"]*)"$`, s.theRedPlayerShouldBe)
	ctx.Step(`^the black player should be "([^"]*)"$`, s.theBlackPlayerShouldBe)
	ctx.Step(`^the response should contain an error$`, s.theResponseShouldContainAnError)

	// Join game
	ctx.Step(`^a game is waiting for an opponent with player "([^"]*)"$`, s.aGameIsWaitingForAnOpponentWithPlayer)
	ctx.Step(`^a game is waiting for an opponent with player "([^"]*)" of type "([^"]*)"$`, s.aGameIsWaitingForAnOpponentWithPlayerOfType)
	ctx.Step(`^player "([^"]*)" joins the game with type "([^"]*)"$`, s.playerJoinsTheGameWithType)
	ctx.Step(`^player "([^"]*)" joins game "([^"]*)" with type "([^"]*)"$`, s.playerJoinsGameWithType)

	// State management
	ctx.Step(`^a game exists with two players$`, s.aGameExistsWithTwoPlayers)
	ctx.Step(`^I request the game state$`, s.iRequestTheGameState)
	ctx.Step(`^I request the game state for game "([^"]*)"$`, s.iRequestTheGameStateForGame)
	ctx.Step(`^the board should be an 8x8 grid$`, s.theBoardShouldBeAnXGrid)
	ctx.Step(`^red pieces should occupy rows ([^"]*)$`, s.redPiecesShouldOccurRows)
	ctx.Step(`^black pieces should occupy rows ([^"]*)$`, s.blackPiecesShouldOccurRows)
	ctx.Step(`^each piece should have id, color, and is_king fields$`, s.eachPieceShouldHaveIdColorAndIsKingFields)
	ctx.Step(`^the board should have (\d+) red pieces$`, s.theBoardShouldHaveRedPieces)
	ctx.Step(`^the board should have (\d+) black pieces$`, s.theBoardShouldHaveBlackPieces)
	ctx.Step(`^empty squares should be null$`, s.emptySquaresShouldBeNull)
	ctx.Step(`^all pieces should have is_king false at start$`, s.allPiecesShouldHaveIsKingFalseAtStart)
	ctx.Step(`^I request the health endpoint$`, s.iRequestTheHealthEndpoint)
	ctx.Step(`^the response should contain status "([^"]*)"$`, s.theResponseShouldContainStatus)

	// Moves
	ctx.Step(`^a red piece is at position row (\d+) col (\d+)$`, s.aRedPieceIsAtPositionRowCol)
	ctx.Step(`^it is red's turn$`, s.itIsRedsTurn)
	ctx.Step(`^the red player moves from row (\d+) col (\d+) to row (\d+) col (\d+)$`, s.theRedPlayerMovesFromRowColToRowCol)
	ctx.Step(`^the black player moves from row (\d+) col (\d+) to row (\d+) col (\d+)$`, s.theBlackPlayerMovesFromRowColToRowCol)
	ctx.Step(`^the move should be accepted$`, s.theMoveShouldBeAccepted)
	ctx.Step(`^the move should be rejected$`, s.theMoveShouldBeRejected)
	ctx.Step(`^the piece should now be at row (\d+) col (\d+)$`, s.thePieceShouldNowBeAtRowCol)
	ctx.Step(`^the turn should pass to "([^"]*)"$`, s.theTurnShouldPassTo)
	ctx.Step(`^a black piece is at position row (\d+) col (\d+)$`, s.aBlackPieceIsAtPositionRowCol)
	ctx.Step(`^position row (\d+) col (\d+) is empty$`, s.positionRowColIsEmpty)
	ctx.Step(`^the black piece at row (\d+) col (\d+) should be removed$`, s.theBlackPieceAtRowColShouldBeRemoved)
	ctx.Step(`^the piece at row (\d+) col (\d+) should be a king$`, s.thePieceAtRowColShouldBeAKing)
	ctx.Step(`^I request valid moves for row (\d+) col (\d+)$`, s.iRequestValidMovesForRowCol)
	ctx.Step(`^the response should contain a list of valid moves$`, s.theResponseShouldContainAListOfValidMoves)
	ctx.Step(`^I request the move history$`, s.iRequestTheMoveHistory)
	ctx.Step(`^the move history should contain (\d+) move$`, s.theMoveHistoryShouldContainMove)

	// Game end
	ctx.Step(`^red has (\d+) piece at row (\d+) col (\d+)$`, s.redHasPieceAtRowCol)
	ctx.Step(`^black has (\d+) pieces remaining$`, s.blackHasPiecesRemaining)
	ctx.Step(`^black has pieces but no valid moves$`, s.blackHasPiecesButNoValidMoves)
	ctx.Step(`^red has valid moves available$`, s.redHasValidMovesAvailable)
	ctx.Step(`^the winner should be "([^"]*)"$`, s.theWinnerShouldBe)
	ctx.Step(`^the result reason should be "([^"]*)"$`, s.theResultReasonShouldBe)
	ctx.Step(`^the game is in progress$`, s.theGameIsInProgress)
	ctx.Step(`^player "([^"]*)" offers a draw$`, s.playerOffersADraw)
	ctx.Step(`^player "([^"]*)" accepts the draw$`, s.playerAcceptsTheDraw)
	ctx.Step(`^the black player resigns$`, s.theBlackPlayerResigns)

	// WebSocket
	ctx.Step(`^I connect a WebSocket to the game$`, s.iConnectAWebSocketToTheGame)
	ctx.Step(`^I connect a WebSocket with an invalid token$`, s.iConnectAWebSocketWithAnInvalidToken)
	ctx.Step(`^I have connected a WebSocket to the game$`, s.iConnectAWebSocketToTheGame)
	ctx.Step(`^I should receive a "([^"]*)" event$`, s.iShouldReceiveAEvent)
	ctx.Step(`^the event should contain the game_id$`, s.theEventShouldContainTheGameID)
	ctx.Step(`^the WebSocket connection should be rejected with status (\d+)$`, s.theWebSocketConnectionShouldBeRejectedWithStatus)

	// MCP
	ctx.Step(`^the MCP server is available$`, s.theMCPServerIsAvailable)
	ctx.Step(`^the agent calls "([^"]*)" with name "([^"]*)" and type "([^"]*)"$`, s.agentCallsRegisterPlayer)
	ctx.Step(`^the agent calls "([^"]*)" with the game_id$`, s.agentCallsWithGameID)
	ctx.Step(`^the agent calls "([^"]*)" with the red player's id$`, s.agentCallsResignWithRedPlayerID)
	ctx.Step(`^the agent calls "([^"]*)" with from row (\d+) col (\d+) to row (\d+) col (\d+)$`, s.agentCallsMakeMove)
	ctx.Step(`^the agent calls "([^"]*)"$`, s.agentCallsMethod)
	ctx.Step(`^the MCP response should contain a player_id$`, s.mcpResponseShouldContainPlayerID)
	ctx.Step(`^the MCP response should contain a game_id$`, s.mcpResponseShouldContainGameID)
	ctx.Step(`^the MCP response should contain a game_state$`, s.mcpResponseShouldContainGameState)
	ctx.Step(`^the MCP response should contain a list of moves$`, s.mcpResponseShouldContainListOfMoves)
	ctx.Step(`^the MCP response should contain a list of tools$`, s.mcpResponseShouldContainListOfTools)
	ctx.Step(`^the MCP response should have color "([^"]*)"$`, s.mcpResponseShouldHaveColor)
	ctx.Step(`^the MCP response should have success (true|false)$`, s.mcpResponseShouldHaveSuccess)
	ctx.Step(`^the MCP response should contain an error$`, s.mcpResponseShouldContainAnError)
	ctx.Step(`^the tools should include "([^"]*)"$`, s.toolsShouldInclude)
	ctx.Step(`^a game exists with two players via MCP$`, s.aGameExistsWithTwoPlayersViaMCP)
}

// ---------------------------------------------------------------------------
// Step implementations — WebSocket
// ---------------------------------------------------------------------------

func (s *bddCtx) iConnectAWebSocketToTheGame() error {
	// Start a real HTTP test server with the router.
	s.wsServer = httptest.NewServer(s.router)

	// Create a valid session for the red player.
	sess, err := s.sessMgr.Create(s.redID, s.gameID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	s.wsToken = sess.Token

	wsURL := "ws" + strings.TrimPrefix(s.wsServer.URL, "http") + "/api/v1/games/" + s.gameID + "/ws?token=" + sess.Token
	conn, _, err := gorilla.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect websocket: %w", err)
	}
	s.wsConn = conn

	// Read the initial game_state event.
	var event apiws.Event
	if err := conn.ReadJSON(&event); err != nil {
		return fmt.Errorf("failed to read initial event: %w", err)
	}
	s.wsEvents = append(s.wsEvents, event)
	return nil
}

func (s *bddCtx) iConnectAWebSocketWithAnInvalidToken() error {
	s.wsServer = httptest.NewServer(s.router)
	wsURL := "ws" + strings.TrimPrefix(s.wsServer.URL, "http") + "/api/v1/games/" + s.gameID + "/ws?token=invalid-token"
	conn, resp, err := gorilla.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		_ = conn.Close()
		return fmt.Errorf("websocket connection should have failed")
	}
	if resp == nil || resp.StatusCode != 401 {
		return fmt.Errorf("expected status 401, got %v", resp)
	}
	s.lastResp = httptest.NewRecorder()
	s.lastResp.Code = resp.StatusCode
	return nil
}

func (s *bddCtx) iShouldReceiveAEvent(eventType string) error {
	// First check if we already received this event (e.g. the initial game_state
	// event that is read during connection).
	for _, e := range s.wsEvents {
		if string(e.Type) == eventType {
			return nil
		}
	}
	if s.wsConn == nil {
		return fmt.Errorf("no websocket connection")
	}
	// Read events with a timeout, looking for the requested type.
	deadline := time.Now().Add(3 * time.Second)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("timed out waiting for %q event; received %d events", eventType, len(s.wsEvents))
		}
		_ = s.wsConn.SetReadDeadline(time.Now().Add(remaining))
		var event apiws.Event
		if err := s.wsConn.ReadJSON(&event); err != nil {
			return fmt.Errorf("failed to read event: %w", err)
		}
		s.wsEvents = append(s.wsEvents, event)
		if string(event.Type) == eventType {
			return nil
		}
	}
}

func (s *bddCtx) theEventShouldContainTheGameID() error {
	if len(s.wsEvents) == 0 {
		return fmt.Errorf("no events received")
	}
	event := s.wsEvents[0]
	if event.Type != apiws.EventTypeGameState {
		return fmt.Errorf("expected game_state event, got %s", event.Type)
	}
	// The payload is a GameStatePayload; we need to check game_id inside game_state.
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}
	var payload struct {
		GameState struct {
			ID string `json:"id"`
		} `json:"game_state"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal game state payload: %w", err)
	}
	if payload.GameState.ID != s.gameID {
		return fmt.Errorf("expected game_id %s, got %s", s.gameID, payload.GameState.ID)
	}
	return nil
}

func (s *bddCtx) theWebSocketConnectionShouldBeRejectedWithStatus(expectedStatus int) error {
	if s.lastResp == nil {
		return fmt.Errorf("no response recorded")
	}
	if s.lastResp.Code != expectedStatus {
		return fmt.Errorf("expected status %d, got %d", expectedStatus, s.lastResp.Code)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Step implementations — MCP
// ---------------------------------------------------------------------------

func (s *bddCtx) theMCPServerIsAvailable() error {
	if s.mcpServer == nil {
		return fmt.Errorf("MCP server not initialized")
	}
	return nil
}

func (s *bddCtx) callMCP(method string, params map[string]any) error {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	reqJSON, _ := json.Marshal(req)
	s.mcpInput = &strings.Builder{}
	s.mcpInput.Write(reqJSON)
	s.mcpInput.WriteString("\n")
	s.mcpOutput = &bytes.Buffer{}
	if err := s.mcpServer.Run(context.Background(), strings.NewReader(s.mcpInput.String()), s.mcpOutput); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	s.mcpResult = nil
	var result map[string]any
	if err := json.Unmarshal(s.mcpOutput.Bytes(), &result); err != nil {
		return fmt.Errorf("failed to decode MCP response: %w (body: %s)", err, s.mcpOutput.String())
	}
	s.mcpResult = result
	return nil
}

func (s *bddCtx) callMCPTool(toolName string, args map[string]any) error {
	return s.callMCP("tools/call", map[string]any{
		"name":      toolName,
		"arguments": args,
	})
}

func (s *bddCtx) agentCallsRegisterPlayer(method, name, playerType string) error {
	if method != "register_player" {
		return fmt.Errorf("expected register_player, got %s", method)
	}
	return s.callMCPTool("register_player", map[string]any{
		"name": name,
		"type": playerType,
	})
}

func (s *bddCtx) agentCallsWithGameID(method string) error {
	return s.callMCPTool(method, map[string]any{
		"game_id": s.gameID,
	})
}

func (s *bddCtx) agentCallsResignWithRedPlayerID(method string) error {
	return s.callMCPTool(method, map[string]any{
		"game_id":   s.gameID,
		"player_id": s.redID,
	})
}

func (s *bddCtx) agentCallsMakeMove(method string, fromRow, fromCol, toRow, toCol int) error {
	return s.callMCPTool(method, map[string]any{
		"game_id": s.gameID,
		"from":    map[string]any{"row": fromRow, "col": fromCol},
		"to":      map[string]any{"row": toRow, "col": toCol},
	})
}

func (s *bddCtx) agentCallsMethod(method string) error {
	if method == "tools/list" {
		return s.callMCP("tools/list", nil)
	}
	// For unknown tools or direct method calls
	return s.callMCPTool(method, map[string]any{})
}

func (s *bddCtx) mcpResultMap() (map[string]any, error) {
	if s.mcpResult == nil {
		return nil, fmt.Errorf("no MCP response")
	}
	// JSON-RPC wraps the actual result in a "result" field.
	// For tool calls and direct methods, the data we want is inside "result".
	// For error responses, "error" is at the top level.
	if resultRaw, ok := s.mcpResult["result"]; ok {
		if resultMap, ok := resultRaw.(map[string]any); ok {
			return resultMap, nil
		}
	}
	return s.mcpResult, nil
}

func (s *bddCtx) mcpResultField(key string) (any, error) {
	result, err := s.mcpResultMap()
	if err != nil {
		return nil, err
	}
	val, ok := result[key]
	if !ok {
		return nil, fmt.Errorf("MCP response does not contain %q: %v", key, result)
	}
	return val, nil
}

func (s *bddCtx) mcpResponseShouldContainPlayerID() error {
	_, err := s.mcpResultField("player_id")
	return err
}

func (s *bddCtx) mcpResponseShouldContainGameID() error {
	_, err := s.mcpResultField("game_id")
	return err
}

func (s *bddCtx) mcpResponseShouldContainGameState() error {
	_, err := s.mcpResultField("game_state")
	return err
}

func (s *bddCtx) mcpResponseShouldContainListOfMoves() error {
	result, err := s.mcpResultMap()
	if err != nil {
		return err
	}
	if result["error"] != nil {
		return fmt.Errorf("unexpected error: %v", result["error"])
	}
	moves, ok := result["moves"].([]any)
	if !ok {
		return fmt.Errorf("no moves array in MCP response: %v", result)
	}
	if len(moves) == 0 {
		return fmt.Errorf("expected at least one move, got empty list")
	}
	return nil
}

func (s *bddCtx) mcpResponseShouldContainListOfTools() error {
	result, err := s.mcpResultMap()
	if err != nil {
		return err
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		return fmt.Errorf("no tools array in MCP response: %v", result)
	}
	if len(tools) == 0 {
		return fmt.Errorf("expected at least one tool, got empty list")
	}
	return nil
}

func (s *bddCtx) mcpResponseShouldHaveColor(color string) error {
	val, err := s.mcpResultField("color")
	if err != nil {
		return err
	}
	if val != color {
		return fmt.Errorf("expected color %s, got %v", color, val)
	}
	return nil
}

func (s *bddCtx) mcpResponseShouldHaveSuccess(expected string) error {
	val, err := s.mcpResultField("success")
	if err != nil {
		return err
	}
	success, ok := val.(bool)
	if !ok {
		return fmt.Errorf("success is not a boolean: %v", val)
	}
	want := expected == "true"
	if success != want {
		return fmt.Errorf("expected success %s, got %v", expected, success)
	}
	return nil
}

func (s *bddCtx) mcpResponseShouldContainAnError() error {
	if s.mcpResult == nil {
		return fmt.Errorf("no MCP response")
	}
	if s.mcpResult["error"] == nil {
		return fmt.Errorf("MCP response does not contain an error: %v", s.mcpResult)
	}
	return nil
}

func (s *bddCtx) toolsShouldInclude(toolName string) error {
	result, err := s.mcpResultMap()
	if err != nil {
		return err
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		return fmt.Errorf("no tools array in MCP response")
	}
	for _, t := range tools {
		tool, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] == toolName {
			return nil
		}
	}
	return fmt.Errorf("tools list does not include %q", toolName)
}

func (s *bddCtx) aGameExistsWithTwoPlayersViaMCP() error {
	// Register first player via MCP.
	if err := s.callMCPTool("register_player", map[string]any{
		"name": "Alice",
		"type": "human",
	}); err != nil {
		return err
	}
	result, err := s.mcpResultMap()
	if err != nil {
		return err
	}
	gameID, ok := result["game_id"].(string)
	if !ok {
		return fmt.Errorf("MCP register_player response missing game_id")
	}
	s.gameID = gameID
	playerID, ok := result["player_id"].(string)
	if !ok {
		return fmt.Errorf("MCP register_player response missing player_id")
	}
	s.redID = playerID

	// Register second player via REST API join endpoint (to start the game).
	body := fmt.Sprintf(`{"player_name":"%s","player_type":"%s"}`, "Claude", "ai")
	resp := s.doRequest(http.MethodPost, "/api/v1/games/"+s.gameID+"/join", body)
	if resp.Code != http.StatusOK {
		return fmt.Errorf("failed to join game: %d %s", resp.Code, resp.Body.String())
	}
	var joinResult map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &joinResult); err != nil {
		return fmt.Errorf("failed to decode join response: %w", err)
	}
	player := joinResult["player"].(map[string]any)
	s.blackID = player["id"].(string)
	return nil
}

// Suppress unused import warnings for bufio (used by mcp.Run internally).
var _ = bufio.ScanLines
