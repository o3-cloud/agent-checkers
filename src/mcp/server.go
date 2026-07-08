// Package mcp exposes a minimal JSON-RPC tool server for agent integrations.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stackable-specs/agent-checkers/internal/app/lobby"
	"github.com/stackable-specs/agent-checkers/internal/app/rules"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

// Server handles MCP-style JSON-RPC requests.
type Server struct {
	store     store.GameStore
	lobby     *lobby.Lobby
	validator *rules.Validator
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewServer creates an MCP server backed by a game store.
func NewServer(gameStore store.GameStore) *Server {
	return &Server{
		store:     gameStore,
		lobby:     lobby.New(gameStore),
		validator: rules.NewValidator(),
	}
}

// Run processes newline-delimited JSON-RPC requests from in and writes responses to out.
func (s *Server) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var request rpcRequest
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			if encodeErr := encoder.Encode(errorResponse(nil, -32700, "parse error", nil)); encodeErr != nil {
				return encodeErr
			}
			continue
		}
		if err := encoder.Encode(s.handle(request)); err != nil {
			return fmt.Errorf("encode response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	return nil
}

func (s *Server) handle(request rpcRequest) rpcResponse {
	switch request.Method {
	case "initialize":
		return successResponse(request.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    "agent-checkers",
				"version": "1.0.0",
			},
			"capabilities": map[string]any{"tools": map[string]any{}},
		})
	case "tools/list":
		return successResponse(request.ID, map[string]any{"tools": toolDefinitions()})
	case "tools/call":
		result, err := s.handleToolCall(request.Params)
		if err != nil {
			return errorResponse(request.ID, -32602, err.Error(), s.errData(err))
		}
		return successResponse(request.ID, result)
	case "list_games":
		var params listGamesArgs
		if len(request.Params) > 0 {
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return errorResponse(request.ID, -32602, "invalid list_games parameters", nil)
			}
		}
		result, err := s.ListGames(params.Status, params.PlayerID)
		if err != nil {
			return errorResponse(request.ID, -32602, err.Error(), s.errData(err))
		}
		return successResponse(request.ID, result)
	default:
		return errorResponse(request.ID, -32601, "method not found", nil)
	}
}

func successResponse(id any, result any) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func errorResponse(id any, code int, message string, data json.RawMessage) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}
