// Package openapi builds and serializes the REST API OpenAPI document.
package openapi

import (
	"encoding/json"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	version = "1.0.0"
)

var (
	jsonOnce sync.Once
	jsonSpec []byte
	jsonErr  error

	yamlOnce sync.Once
	yamlSpec []byte
	yamlErr  error
)

// Document is the OpenAPI 3.1 contract served by the API.
type Document struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       Info                `json:"info" yaml:"info"`
	Servers    []Server            `json:"servers" yaml:"servers"`
	Tags       []Tag               `json:"tags" yaml:"tags"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components Components          `json:"components" yaml:"components"`
}

// Info describes the API.
type Info struct {
	Title       string  `json:"title" yaml:"title"`
	Version     string  `json:"version" yaml:"version"`
	Description string  `json:"description" yaml:"description"`
	Contact     Contact `json:"contact" yaml:"contact"`
}

// Contact identifies the API owner.
type Contact struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
}

// Server describes a deployment target.
type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
}

// Tag groups related operations.
type Tag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// PathItem contains operations for one API path.
type PathItem struct {
	Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
}

// Operation documents a single endpoint.
type Operation struct {
	Tags        []string            `json:"tags" yaml:"tags"`
	OperationID string              `json:"operationId" yaml:"operationId"`
	Summary     string              `json:"summary" yaml:"summary"`
	Description string              `json:"description" yaml:"description"`
	Parameters  []Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses" yaml:"responses"`
}

// Parameter documents a path or query parameter.
type Parameter struct {
	Name        string         `json:"name" yaml:"name"`
	In          string         `json:"in" yaml:"in"`
	Required    bool           `json:"required" yaml:"required"`
	Description string         `json:"description" yaml:"description"`
	Schema      map[string]any `json:"schema" yaml:"schema"`
}

// RequestBody documents JSON request payloads.
type RequestBody struct {
	Required    bool                 `json:"required" yaml:"required"`
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
}

// Response documents one HTTP response.
type Response struct {
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// MediaType describes a response or request content schema.
type MediaType struct {
	Schema map[string]any `json:"schema" yaml:"schema"`
}

// Components contains reusable OpenAPI components.
type Components struct {
	Schemas map[string]map[string]any `json:"schemas" yaml:"schemas"`
}

// Build constructs the OpenAPI document.
func Build() Document {
	return Document{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:       "Agent Checkers API",
			Version:     version,
			Description: "REST API for creating checkers games, joining players, making moves, and inspecting game state.",
			Contact: Contact{
				Name: "Stackable Specs",
				URL:  "https://github.com/stackable-specs/agent-checkers",
			},
		},
		Servers: []Server{
			{URL: "http://localhost:8080", Description: "Default local server"},
		},
		Tags: []Tag{
			{Name: "games", Description: "Game lifecycle and player actions"},
			{Name: "moves", Description: "Move execution and discovery"},
			{Name: "health", Description: "Service health checks"},
			{Name: "meta", Description: "API contract discovery"},
		},
		Paths:      paths(),
		Components: Components{Schemas: Schemas()},
	}
}

// JSON returns a cached JSON serialization of the OpenAPI document.
func JSON() ([]byte, error) {
	jsonOnce.Do(func() {
		jsonSpec, jsonErr = json.Marshal(Build())
	})
	return jsonSpec, jsonErr
}

// YAML returns a cached YAML serialization of the OpenAPI document.
func YAML() ([]byte, error) {
	yamlOnce.Do(func() {
		yamlSpec, yamlErr = yaml.Marshal(Build())
	})
	return yamlSpec, yamlErr
}

func paths() map[string]PathItem {
	return map[string]PathItem{
		"/api/v1/games": {
			Post: operation("games", "createGame", "Create a game", "Create a new waiting game and register the first player as red. A player session is returned when session management is enabled.",
				nil, jsonBody("CreateGameRequest", "First player registration payload"), responses(
					success("201", "Game created", "PlayerGameResponse"),
					errorResponse("400", "Invalid request payload or player data"),
					errorResponse("500", "Session creation or persistence failed"),
				)),
		},
		"/api/v1/games/{id}/join": {
			Post: operation("games", "joinGame", "Join a game", "Register the second player for an existing waiting game and start play when the game becomes full.",
				[]Parameter{idPathParameter()}, jsonBody("JoinGameRequest", "Joining player registration payload"), responses(
					success("200", "Player joined game", "PlayerGameResponse"),
					errorResponse("400", "Invalid request payload or game state"),
					errorResponse("404", "Game not found"),
					errorResponse("500", "Session creation or persistence failed"),
				)),
		},
		"/api/v1/games/{id}": {
			Get: operation("games", "getGame", "Get game state", "Return the current board, players, turn, status, and result for a game.",
				[]Parameter{idPathParameter()}, nil, responses(
					success("200", "Current game state", "GameResponse"),
					errorResponse("404", "Game not found"),
				)),
			Delete: operation("games", "resignGame", "Resign a game", "End a game by resignation for the requesting player.",
				[]Parameter{idPathParameter()}, jsonBody("PlayerActionRequest", "Player resigning the game"), responses(
					success("200", "Game ended by resignation", "GameResponse"),
					errorResponse("400", "Invalid request payload or resignation"),
					errorResponse("404", "Game not found"),
					errorResponse("500", "Game persistence failed"),
				)),
		},
		"/api/v1/games/{id}/draw": {
			Post: operation("games", "offerOrAcceptDraw", "Offer or accept a draw", "Create a draw offer for the requesting player, or accept an existing draw offer from the opponent.",
				[]Parameter{idPathParameter()}, jsonBody("PlayerActionRequest", "Player offering or accepting a draw"), responses(
					success("200", "Draw offer recorded or accepted", "GameResponse"),
					errorResponse("400", "Invalid request payload or draw action"),
					errorResponse("404", "Game not found"),
					errorResponse("500", "Game persistence failed"),
				)),
		},
		"/api/v1/games/{id}/moves": {
			Post: operation("moves", "makeMove", "Make a move", "Validate and execute a move for a player, then return the move and updated game state.",
				[]Parameter{idPathParameter()}, jsonBody("MoveRequest", "Move execution payload"), responses(
					success("200", "Move accepted", "MoveResponse"),
					errorResponse("400", "Invalid request payload, move, turn, or piece ownership"),
					errorResponse("404", "Game not found"),
					errorResponse("500", "Game persistence failed"),
				)),
			Get: operation("moves", "getMoveHistory", "Get move history", "Return all recorded moves for a game in chronological order.",
				[]Parameter{idPathParameter()}, nil, responses(
					success("200", "Move history", "MoveHistoryResponse"),
					errorResponse("404", "Game not found"),
				)),
		},
		"/api/v1/games/{id}/valid-moves": {
			Get: operation("moves", "getValidMoves", "Get valid moves", "Return all valid destination squares for the piece at the requested row and column.",
				[]Parameter{idPathParameter(), boardQueryParameter("row"), boardQueryParameter("col")}, nil, responses(
					success("200", "Valid destination squares", "ValidMovesResponse"),
					errorResponse("400", "Invalid row, column, position, or game state"),
					errorResponse("404", "Game not found"),
				)),
		},
		"/health": {
			Get: operation("health", "healthCheck", "Health check", "Report whether the HTTP server is ready to handle requests.",
				nil, nil, responses(inlineSuccess("200", "Server is healthy", healthSchema()))),
		},
		"/openapi.json": {
			Get: operation("meta", "getOpenAPISpecJSON", "Get OpenAPI JSON", "Return this OpenAPI 3.1 contract encoded as JSON.",
				nil, nil, responses(inlineSuccess("200", "OpenAPI contract as JSON", map[string]any{"type": "object"}))),
		},
		"/openapi.yaml": {
			Get: operation("meta", "getOpenAPISpecYAML", "Get OpenAPI YAML", "Return this OpenAPI 3.1 contract encoded as YAML.",
				nil, nil, responses(inlineSuccess("200", "OpenAPI contract as YAML", map[string]any{"type": "object"}))),
		},
	}
}

func operation(tag, id, summary, description string, parameters []Parameter, requestBody *RequestBody, responses map[string]Response) *Operation {
	return &Operation{
		Tags:        []string{tag},
		OperationID: id,
		Summary:     summary,
		Description: description,
		Parameters:  parameters,
		RequestBody: requestBody,
		Responses:   responses,
	}
}

func jsonBody(schemaName, description string) *RequestBody {
	return &RequestBody{
		Required:    true,
		Description: description,
		Content:     jsonContent(ref(schemaName)),
	}
}

func responses(items ...namedResponse) map[string]Response {
	out := make(map[string]Response, len(items))
	for _, item := range items {
		out[item.status] = item.response
	}
	return out
}

type namedResponse struct {
	status   string
	response Response
}

func success(status, description, schemaName string) namedResponse {
	return inlineSuccess(status, description, ref(schemaName))
}

func inlineSuccess(status, description string, schema map[string]any) namedResponse {
	return namedResponse{
		status: status,
		response: Response{
			Description: description,
			Content:     jsonContent(schema),
		},
	}
}

func errorResponse(status, description string) namedResponse {
	return namedResponse{
		status: status,
		response: Response{
			Description: description,
			Content:     jsonContent(ref("ErrorResponse")),
		},
	}
}

func jsonContent(schema map[string]any) map[string]MediaType {
	return map[string]MediaType{"application/json": {Schema: schema}}
}

func ref(schemaName string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + schemaName}
}

func idPathParameter() Parameter {
	return Parameter{
		Name:        "id",
		In:          "path",
		Required:    true,
		Description: "Game identifier.",
		Schema:      map[string]any{"type": "string"},
	}
}

func boardQueryParameter(name string) Parameter {
	return Parameter{
		Name:        name,
		In:          "query",
		Required:    true,
		Description: "Zero-based board " + name + " from 0 through 7.",
		Schema: map[string]any{
			"type":    "integer",
			"minimum": 0,
			"maximum": 7,
		},
	}
}

func healthSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"required":   []string{"status"},
		"properties": map[string]any{"status": map[string]any{"type": "string", "example": "ok"}},
	}
}
