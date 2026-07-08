package mcp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/internal/app/discovery"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/rules"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

type toolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type listGamesArgs struct {
	Status   string `json:"status,omitempty"`
	PlayerID string `json:"player_id,omitempty"`
}

// toolDefinitions returns the full set of MCP tool definitions.
func toolDefinitions() []toolDefinition {
	return []toolDefinition{
		listGamesTool(),
		registerPlayerTool(),
		getGameStateTool(),
		makeMoveTool(),
		getValidMovesTool(),
		resignTool(),
		offerDrawTool(),
		acceptDrawTool(),
	}
}

// ----------------------------------------------------------------------------
// Tool definitions
// ----------------------------------------------------------------------------

func listGamesTool() toolDefinition {
	return toolDefinition{
		Name:        "list_games",
		Description: "List games filtered by status and player. Defaults to active and waiting games.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"status": map[string]any{
					"type": "string",
					"enum": []string{"waiting", "active", "completed", "draw", "all"},
				},
				"player_id": map[string]any{"type": "string"},
			},
		},
	}
}

func registerPlayerTool() toolDefinition {
	return toolDefinition{
		Name:        "register_player",
		Description: "Register a new player and create a waiting game. Returns the player ID, game ID, and assigned color.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"type": map[string]any{
					"type": "string",
					"enum": []string{"human", "ai"},
				},
			},
			"required": []string{"name", "type"},
		},
	}
}

func getGameStateTool() toolDefinition {
	return toolDefinition{
		Name:        "get_game_state",
		Description: "Get the full state of a game by ID: board, current turn, status, players, and result.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id": map[string]any{"type": "string"},
			},
			"required": []string{"game_id"},
		},
	}
}

func makeMoveTool() toolDefinition {
	return toolDefinition{
		Name:        "make_move",
		Description: "Execute a move on the board. Positions use {row, col} with 0-7 indices. On invalid move, the error data includes valid_moves for AI recovery.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id": map[string]any{"type": "string"},
				"from": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"row": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
						"col": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
					},
					"required": []string{"row", "col"},
				},
				"to": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"row": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
						"col": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
					},
					"required": []string{"row", "col"},
				},
			},
			"required": []string{"game_id", "from", "to"},
		},
	}
}

func getValidMovesTool() toolDefinition {
	return toolDefinition{
		Name:        "get_valid_moves",
		Description: "List all legal moves for the current player. Each entry has a 'from' position and an array of 'to' destination positions.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id": map[string]any{"type": "string"},
			},
			"required": []string{"game_id"},
		},
	}
}

func resignTool() toolDefinition {
	return toolDefinition{
		Name:        "resign",
		Description: "Resign the game. The opposing player wins by resignation.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id":   map[string]any{"type": "string"},
				"player_id": map[string]any{"type": "string"},
			},
			"required": []string{"game_id", "player_id"},
		},
	}
}

func offerDrawTool() toolDefinition {
	return toolDefinition{
		Name:        "offer_draw",
		Description: "Offer a draw in the game. The opponent must accept_draw to end the game.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id":   map[string]any{"type": "string"},
				"player_id": map[string]any{"type": "string"},
			},
			"required": []string{"game_id", "player_id"},
		},
	}
}

func acceptDrawTool() toolDefinition {
	return toolDefinition{
		Name:        "accept_draw",
		Description: "Accept a draw offer from the opponent, ending the game in a draw.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"game_id":   map[string]any{"type": "string"},
				"player_id": map[string]any{"type": "string"},
			},
			"required": []string{"game_id", "player_id"},
		},
	}
}

// ----------------------------------------------------------------------------
// Dispatch
// ----------------------------------------------------------------------------

func (s *Server) handleToolCall(raw json.RawMessage) (any, error) {
	var params toolCallParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, errors.New("invalid tool call parameters")
	}

	switch params.Name {
	case "list_games":
		args := listGamesArgs{}
		if value, ok := params.Arguments["status"].(string); ok {
			args.Status = value
		}
		if value, ok := params.Arguments["player_id"].(string); ok {
			args.PlayerID = value
		}
		return s.ListGames(args.Status, args.PlayerID)
	case "register_player":
		name, _ := params.Arguments["name"].(string)
		playerType, _ := params.Arguments["type"].(string)
		return s.RegisterPlayer(name, playerType)
	case "get_game_state":
		gameID, _ := params.Arguments["game_id"].(string)
		return s.GetGameState(gameID)
	case "make_move":
		gameID, _ := params.Arguments["game_id"].(string)
		from, err := parsePosition(params.Arguments["from"])
		if err != nil {
			return nil, fmt.Errorf("invalid 'from' position: %w", err)
		}
		to, err := parsePosition(params.Arguments["to"])
		if err != nil {
			return nil, fmt.Errorf("invalid 'to' position: %w", err)
		}
		return s.MakeMove(gameID, from, to)
	case "get_valid_moves":
		gameID, _ := params.Arguments["game_id"].(string)
		return s.GetValidMoves(gameID)
	case "resign":
		gameID, _ := params.Arguments["game_id"].(string)
		playerID, _ := params.Arguments["player_id"].(string)
		return s.Resign(gameID, playerID)
	case "offer_draw":
		gameID, _ := params.Arguments["game_id"].(string)
		playerID, _ := params.Arguments["player_id"].(string)
		return s.OfferDraw(gameID, playerID)
	case "accept_draw":
		gameID, _ := params.Arguments["game_id"].(string)
		playerID, _ := params.Arguments["player_id"].(string)
		return s.AcceptDraw(gameID, playerID)
	default:
		return nil, fmt.Errorf("unknown tool %q", params.Name)
	}
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func parsePosition(raw any) (board.Position, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return board.Position{}, errors.New("expected object with row and col")
	}
	row, ok := toInt(m["row"])
	if !ok {
		return board.Position{}, errors.New("missing or invalid row")
	}
	col, ok := toInt(m["col"])
	if !ok {
		return board.Position{}, errors.New("missing or invalid col")
	}
	pos := board.Position{Row: row, Col: col}
	if !pos.IsValid() {
		return board.Position{}, fmt.Errorf("position out of bounds: %v", pos)
	}
	return pos, nil
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	}
	return 0, false
}

// moveError wraps a validation/game error together with the game it occurred
// in, so that the JSON-RPC error response can include valid_moves for AI
// recovery. The wrapper is not returned to the caller directly; instead it
// flows through errData() into the JSON-RPC error.data field.
type moveError struct {
	err    error
	gameID string
}

func (m *moveError) Error() string { return m.err.Error() }
func (m *moveError) Unwrap() error { return m.err }

// errData produces the optional `data` field for JSON-RPC error responses.
// For move errors it embeds the current legal moves so the AI agent can
// recover without an extra round-trip. A ValidationError with pre-populated
// ValidMoves is also honored.
func (s *Server) errData(err error) json.RawMessage {
	var me *moveError
	if errors.As(err, &me) {
		data := s.buildMoveErrorData(me.gameID)
		if data != nil {
			return data
		}
	}

	var ve *rules.ValidationError
	if errors.As(err, &ve) && len(ve.ValidMoves) > 0 {
		data, marshalErr := json.Marshal(map[string]any{
			"valid_moves": ve.ValidMoves,
		})
		if marshalErr == nil {
			return data
		}
	}
	return nil
}

// buildMoveErrorData loads the game and computes all legal moves, returning
// a JSON payload suitable for embedding in an error response. Returns nil if
// the game cannot be loaded or no moves are available.
func (s *Server) buildMoveErrorData(gameID string) json.RawMessage {
	if gameID == "" {
		return nil
	}
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil
	}
	allMoves, err := s.validator.GetAllValidMoves(g)
	if err != nil || len(allMoves) == 0 {
		return nil
	}
	groups := make([]ValidMovesGroup, 0, len(allMoves))
	for from, tos := range allMoves {
		groups = append(groups, ValidMovesGroup{From: from, To: tos})
	}
	data, marshalErr := json.Marshal(map[string]any{
		"valid_moves": groups,
	})
	if marshalErr != nil {
		return nil
	}
	return data
}

// ----------------------------------------------------------------------------
// Tool handlers
// ----------------------------------------------------------------------------

// ListGames runs the list_games tool directly.
func (s *Server) ListGames(status, playerID string) (*dto.ListGamesResponse, error) {
	games, err := discovery.ListGames(s.store, status, playerID)
	if err != nil {
		if errors.Is(err, discovery.ErrInvalidStatus) {
			return nil, errors.New("invalid status")
		}
		return nil, err
	}
	return &dto.ListGamesResponse{
		Success: true,
		Games:   dto.NewGameSummaries(games),
	}, nil
}

// RegisterPlayerResponse is the structured result of register_player.
type RegisterPlayerResponse struct {
	Success  bool   `json:"success"`
	PlayerID string `json:"player_id"`
	GameID   string `json:"game_id"`
	Color    string `json:"color"`
}

// RegisterPlayer runs the register_player tool directly.
func (s *Server) RegisterPlayer(name, playerType string) (*RegisterPlayerResponse, error) {
	g, player, err := s.lobby.RegisterPlayer(name, playerType)
	if err != nil {
		return nil, err
	}
	return &RegisterPlayerResponse{
		Success:  true,
		PlayerID: player.ID,
		GameID:   g.ID,
		Color:    player.Color.String(),
	}, nil
}

// GetGameState runs the get_game_state tool directly.
func (s *Server) GetGameState(gameID string) (*dto.GameResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}
	return &dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	}, nil
}

// MoveResponse is the structured result of make_move.
type MoveResponse struct {
	Success   bool             `json:"success"`
	Captured  []board.Position `json:"captured,omitempty"`
	Move      game.Move        `json:"move"`
	GameState *dto.GameState   `json:"game_state"`
}

// MakeMove runs the make_move tool directly.
func (s *Server) MakeMove(gameID string, from, to board.Position) (*MoveResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}

	if err := s.validator.ValidateMove(g, from, to); err != nil {
		return nil, &moveError{err: err, gameID: gameID}
	}

	captured, err := s.validator.ExecuteMove(g, from, to)
	if err != nil {
		return nil, &moveError{err: err, gameID: gameID}
	}

	// If no capture or no further captures available, switch turn.
	if len(captured) == 0 || !s.validator.HasCapturesForPiece(g, to) {
		g.SwitchTurn()
	}

	// Check game over.
	if over, result := g.IsGameOver(); over {
		g.EndGame(result)
	}

	if err := s.store.SaveGame(g); err != nil {
		return nil, err
	}

	move := g.Moves[len(g.Moves)-1]
	return &MoveResponse{
		Success:   true,
		Captured:  captured,
		Move:      move,
		GameState: dto.NewGameState(g),
	}, nil
}

// ValidMovesGroup groups a from position with its legal destinations.
type ValidMovesGroup struct {
	From board.Position   `json:"from"`
	To   []board.Position `json:"to"`
}

// ValidMovesResponse is the structured result of get_valid_moves.
type ValidMovesResponse struct {
	Success   bool              `json:"success"`
	Moves     []ValidMovesGroup `json:"moves"`
	GameState *dto.GameState    `json:"game_state"`
}

// GetValidMoves runs the get_valid_moves tool directly.
func (s *Server) GetValidMoves(gameID string) (*ValidMovesResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}

	allMoves, err := s.validator.GetAllValidMoves(g)
	if err != nil {
		return nil, err
	}

	groups := make([]ValidMovesGroup, 0, len(allMoves))
	for from, tos := range allMoves {
		groups = append(groups, ValidMovesGroup{From: from, To: tos})
	}

	return &ValidMovesResponse{
		Success:   true,
		Moves:     groups,
		GameState: dto.NewGameState(g),
	}, nil
}

// Resign runs the resign tool directly.
func (s *Server) Resign(gameID, playerID string) (*dto.GameResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}
	if err := g.Resign(playerID); err != nil {
		return nil, err
	}
	if err := s.store.SaveGame(g); err != nil {
		return nil, err
	}
	return &dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	}, nil
}

// OfferDraw runs the offer_draw tool directly.
func (s *Server) OfferDraw(gameID, playerID string) (*dto.GameResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}
	if err := g.OfferDraw(playerID); err != nil {
		return nil, err
	}
	if err := s.store.SaveGame(g); err != nil {
		return nil, err
	}
	return &dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	}, nil
}

// AcceptDraw runs the accept_draw tool directly.
func (s *Server) AcceptDraw(gameID, playerID string) (*dto.GameResponse, error) {
	g, err := s.store.LoadGame(gameID)
	if err != nil {
		return nil, err
	}
	if err := g.AcceptDraw(playerID); err != nil {
		return nil, err
	}
	if err := s.store.SaveGame(g); err != nil {
		return nil, err
	}
	return &dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	}, nil
}
