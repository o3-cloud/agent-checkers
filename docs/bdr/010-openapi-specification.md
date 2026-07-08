# BDR-010: OpenAPI Specification and Discovery

## Status

Proposed

## Behavior

The system exposes a machine-readable OpenAPI 3.1 specification at a well-known endpoint, enabling clients to discover available operations, request/response schemas, and error formats without reading source code.

## Context

AI agents (BDR-005), frontend developers (BDR-006), and contract testing tools all need to understand the REST API surface. Currently, the API has no machine-readable contract — consumers must read source code or ad-hoc documentation. An OpenAPI spec at a predictable endpoint allows tools to auto-discover and generate typed clients, validate responses, and detect API drift.

## Acceptance Criteria

- AC-1: `GET /openapi.json` returns a valid OpenAPI 3.1 document with HTTP 200 and `Content-Type: application/json`
- AC-2: `GET /openapi.yaml` returns the same spec in YAML format with `Content-Type: application/yaml`
- AC-3: The OpenAPI document lists all REST endpoints registered on the chi router, including path parameters, request bodies, and response schemas
- AC-4: Each endpoint entry includes a summary, description, and operationId
- AC-5: Request and response schemas reference reusable components for DTOs (CreateGameRequest, JoinGameRequest, MoveRequest, GameState, ErrorResponse, etc.)
- AC-6: Error responses (400, 404, 500) are documented for every endpoint that can return them
- AC-7: The `/health` endpoint is included in the spec with its response schema
- AC-8: A static `docs/openapi.yaml` file is committed to the repository for offline tooling
- AC-9: The spec includes a `servers` field with the default URL (`http://localhost:8080`)
- AC-10: The spec includes `info` with title, version (from go.mod), and description

## Verification

### Scenario 1: Fetch OpenAPI JSON

- **Given** the game server is running on port 8080
- **When** a client sends `GET /openapi.json`
- **Then** the response status is 200
- **And** the response `Content-Type` is `application/json`
- **And** the response body is valid OpenAPI 3.1 JSON
- **And** the `info.title` field is "Agent Checkers API"
- **And** the `info.version` field matches the application version

### Scenario 2: Fetch OpenAPI YAML

- **Given** the game server is running on port 8080
- **When** a client sends `GET /openapi.yaml`
- **Then** the response status is 200
- **And** the response `Content-Type` is `application/yaml`
- **And** the response body is valid OpenAPI 3.1 YAML

### Scenario 3: All endpoints are documented

- **Given** the OpenAPI spec is fetched from `/openapi.json`
- **When** the `paths` object is inspected
- **Then** it includes entries for:
  - `POST /api/v1/games`
  - `POST /api/v1/games/{id}/join`
  - `GET /api/v1/games/{id}`
  - `DELETE /api/v1/games/{id}`
  - `POST /api/v1/games/{id}/draw`
  - `POST /api/v1/games/{id}/moves`
  - `GET /api/v1/games/{id}/moves`
  - `GET /api/v1/games/{id}/valid-moves`
  - `GET /health`
  - `GET /openapi.json`
  - `GET /openapi.yaml`

### Scenario 4: Reusable component schemas

- **Given** the OpenAPI spec is fetched from `/openapi.json`
- **When** the `components.schemas` object is inspected
- **Then** it includes definitions for:
  - `CreateGameRequest`
  - `JoinGameRequest`
  - `MoveRequest`
  - `GameState`
  - `PlayerResponse`
  - `ErrorResponse`
  - `MoveResponse`
- **And** each endpoint's request body and response reference these components via `$ref`

### Scenario 5: Error responses documented

- **Given** the OpenAPI spec is fetched from `/openapi.json`
- **When** the path `POST /api/v1/games/{id}/moves` is inspected
- **Then** it documents responses for:
  - `200` (successful move)
  - `400` (invalid move, wrong turn, not your piece)
  - `404` (game not found)
  - `500` (internal server error)

### Scenario 6: AI agent discovers API

- **Given** an AI agent connects to the server for the first time
- **When** the agent fetches `GET /openapi.json`
- **Then** the agent can determine all available operations
- **And** the agent can construct valid requests from the schemas
- **And** the agent can interpret responses using the documented schemas

## Interfaces

### REST API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/openapi.json` | OpenAPI 3.1 spec in JSON format |
| `GET` | `/openapi.yaml` | OpenAPI 3.1 spec in YAML format |

## Traceability

- **ADR-009**: OpenAPI Specification and Discovery
- **ADR-007**: Interface Layers (Web, API, CLI, MCP)
- **BDR-005**: AI Agent Integration via MCP
- **BDR-006**: Web UI Board Visualization