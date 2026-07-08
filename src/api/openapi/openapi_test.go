package openapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildOpenAPIDocument(t *testing.T) {
	doc := Build()

	if doc.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version = %q, want 3.1.0", doc.OpenAPI)
	}
	if doc.Info.Title != "Agent Checkers API" {
		t.Fatalf("info title = %q, want Agent Checkers API", doc.Info.Title)
	}
	if doc.Info.Version != version {
		t.Fatalf("info version = %q, want %q", doc.Info.Version, version)
	}
	if len(doc.Servers) == 0 || doc.Servers[0].URL != "http://localhost:8080" {
		t.Fatalf("default server = %#v, want http://localhost:8080", doc.Servers)
	}

	assertOperation(t, doc, "/api/v1/games", "post", "createGame")
	assertOperation(t, doc, "/api/v1/games", "get", "listGames")
	assertOperation(t, doc, "/api/v1/games/{id}/join", "post", "joinGame")
	assertOperation(t, doc, "/api/v1/games/{id}", "get", "getGame")
	assertOperation(t, doc, "/api/v1/games/{id}", "delete", "resignGame")
	assertOperation(t, doc, "/api/v1/games/{id}/draw", "post", "offerOrAcceptDraw")
	assertOperation(t, doc, "/api/v1/games/{id}/moves", "post", "makeMove")
	assertOperation(t, doc, "/api/v1/games/{id}/moves", "get", "getMoveHistory")
	assertOperation(t, doc, "/api/v1/games/{id}/valid-moves", "get", "getValidMoves")
	assertOperation(t, doc, "/health", "get", "healthCheck")
	assertOperation(t, doc, "/openapi.json", "get", "getOpenAPISpecJSON")
	assertOperation(t, doc, "/openapi.yaml", "get", "getOpenAPISpecYAML")

	for _, schema := range []string{"CreateGameRequest", "JoinGameRequest", "MoveRequest", "GameState", "GameSummary", "ListGamesResponse", "PlayerResponse", "ErrorResponse", "MoveResponse"} {
		if _, ok := doc.Components.Schemas[schema]; !ok {
			t.Fatalf("components.schemas missing %s", schema)
		}
	}
}

func TestSerialization(t *testing.T) {
	jsonBytes, err := JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	var jsonDoc map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonDoc); err != nil {
		t.Fatalf("JSON() returned invalid JSON: %v", err)
	}

	yamlBytes, err := YAML()
	if err != nil {
		t.Fatalf("YAML() error: %v", err)
	}
	var yamlDoc map[string]any
	if err := yaml.Unmarshal(yamlBytes, &yamlDoc); err != nil {
		t.Fatalf("YAML() returned invalid YAML: %v", err)
	}

	if jsonDoc["openapi"] != yamlDoc["openapi"] {
		t.Fatalf("JSON and YAML OpenAPI versions differ: %v != %v", jsonDoc["openapi"], yamlDoc["openapi"])
	}
}

func TestStaticOpenAPIYAMLIsCurrent(t *testing.T) {
	yamlBytes, err := YAML()
	if err != nil {
		t.Fatalf("YAML() error: %v", err)
	}
	path := filepath.Join("..", "..", "..", "docs", "openapi.yaml")
	if os.Getenv("UPDATE_OPENAPI") == "1" {
		if err := os.WriteFile(path, yamlBytes, 0o644); err != nil {
			t.Fatalf("write static OpenAPI YAML: %v", err)
		}
	}

	staticBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read static OpenAPI YAML: %v", err)
	}
	if string(staticBytes) != string(yamlBytes) {
		t.Fatal("docs/openapi.yaml is out of sync with the runtime OpenAPI document")
	}
}

func assertOperation(t *testing.T, doc Document, path, method, operationID string) {
	t.Helper()
	item, ok := doc.Paths[path]
	if !ok {
		t.Fatalf("paths missing %s", path)
	}

	var op *Operation
	switch method {
	case "get":
		op = item.Get
	case "post":
		op = item.Post
	case "delete":
		op = item.Delete
	default:
		t.Fatalf("unsupported test method %s", method)
	}

	if op == nil {
		t.Fatalf("%s %s missing operation", method, path)
	}
	if op.OperationID != operationID {
		t.Fatalf("%s %s operationId = %q, want %q", method, path, op.OperationID, operationID)
	}
	if op.Summary == "" {
		t.Fatalf("%s %s missing summary", method, path)
	}
	if op.Description == "" {
		t.Fatalf("%s %s missing description", method, path)
	}
	if len(op.Tags) == 0 {
		t.Fatalf("%s %s missing tags", method, path)
	}
	if len(op.Responses) == 0 {
		t.Fatalf("%s %s missing responses", method, path)
	}
}
