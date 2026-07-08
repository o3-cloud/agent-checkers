# BDR-011: List Active Games

## Status

Proposed

## Behavior

The system allows clients to list all active (in-progress and waiting) games across all interfaces — REST API, CLI, and MCP server — returning each game's ID, status, players, and current turn.

## Context

Players and AI agents need to discover existing games they can join or observe. Without a list endpoint, a client must know a game ID in advance, which blocks matchmaking, spectating, and multi-game orchestration. The store already supports `ListGames(filter)` with status and player filtering; this BDR exposes that capability through every interface layer so humans, scripts, and AI agents can discover available games.

## Acceptance Criteria

### Shared (all interfaces)

- AC-1: The list operation returns games with status `waiting` or `active` by default; completed games are excluded
- AC-2: Each game entry includes: `game_id`, `status`, `current_turn`, `red_player` (name, color), `black_player` (name, color, or null if waiting)
- AC-3: The list can be filtered by status (`waiting`, `active`, `completed`, or `all`)
- AC-4: The list can be filtered by `player_id` to show only games a specific player is in
- AC-5: An empty result returns an empty list, not an error
- AC-6: The list is sorted by `created_at` descending (newest first)

### REST API

- AC-7: `GET /api/v1/games` returns a JSON array of game summaries with HTTP 200
- AC-8: `GET /api/v1/games?status=waiting` filters by status
- AC-9: `GET /api/v1/games?player_id={id}` filters by player
- AC-10: Response `Content-Type` is `application/json`

### CLI

- AC-11: `agent-checkers games` lists active and waiting games in a table format
- AC-12: `agent-checkers games --status waiting` filters by status
- AC-13: `agent-checkers games --player {id}` filters by player
- AC-14: `agent-checkers games --json` outputs machine-readable JSON
- AC-15: Empty list prints "No games found" (or `[]` with `--json`)

### MCP Server

- AC-16: `list_games` tool returns an array of game summaries
- AC-17: `list_games(status="waiting")` filters by status
- AC-18: `list_games(player_id="...")` filters by player
- AC-19: Tool response includes `games` array with `game_id`, `status`, `current_turn`, and `players` fields

## Verification

### Scenario 1: List games via REST API (no filter)

- **Given** the game server is running
- **And** there are 2 active games and 1 completed game in the store
- **When** a client sends `GET /api/v1/games`
- **Then** the response status is 200
- **And** the response body is a JSON array with 2 entries (active and waiting only)
- **And** each entry has `game_id`, `status`, `current_turn`, `red_player`, and `black_player`
- **And** the completed game is NOT included

### Scenario 2: Filter by status via REST API

- **Given** there are 1 waiting game and 2 active games
- **When** a client sends `GET /api/v1/games?status=waiting`
- **Then** the response contains exactly 1 entry
- **And** that entry has `status: "waiting"`

### Scenario 3: Filter by player via REST API

- **Given** player "Alice" is in 2 games and player "Bob" is in 1 game
- **When** a client sends `GET /api/v1/games?player_id={alice_id}`
- **Then** the response contains 2 entries
- **And** each entry has Alice listed as a player

### Scenario 4: Empty list via REST API

- **Given** no games exist in the store
- **When** a client sends `GET /api/v1/games`
- **Then** the response status is 200
- **And** the response body is `[]`

### Scenario 5: List games via CLI

- **Given** there are 2 active games and 1 waiting game
- **When** the user runs `agent-checkers games`
- **Then** the output is a table with 3 rows
- **And** each row shows game ID (truncated), status, red player, black player, and turn

### Scenario 6: Filter by status via CLI

- **Given** there are 1 waiting game and 2 active games
- **When** the user runs `agent-checkers games --status waiting`
- **Then** the output shows 1 row
- **And** that row has status "waiting"

### Scenario 7: CLI JSON output

- **Given** there are 2 active games
- **When** the user runs `agent-checkers games --json`
- **Then** the output is valid JSON array with 2 entries
- **And** each entry has `game_id`, `status`, `current_turn`, and `players` fields

### Scenario 8: CLI empty list

- **Given** no games exist
- **When** the user runs `agent-checkers games`
- **Then** the output prints "No games found"

### Scenario 9: List games via MCP

- **Given** the MCP server is running
- **And** there are 2 active games and 1 waiting game
- **When** an AI agent calls `list_games()`
- **Then** the response contains a `games` array with 3 entries
- **And** each entry has `game_id`, `status`, `current_turn`, and `players`

### Scenario 10: Filter via MCP

- **Given** there are 1 waiting game and 2 active games
- **When** an AI agent calls `list_games(status="waiting")`
- **Then** the response `games` array contains exactly 1 entry
- **And** that entry has `status: "waiting"`

### Scenario 11: AI agent finds a game to join

- **Given** an AI agent wants to play checkers
- **And** there is 1 waiting game with no opponent
- **When** the AI calls `list_games(status="waiting")`
- **Then** the AI receives the waiting game's `game_id`
- **And** the AI can call `join_game(game_id)` to start playing

## Interfaces

### REST API

| Method | Path | Query Params | Description |
|--------|------|--------------|-------------|
| `GET` | `/api/v1/games` | `status`, `player_id` | List games with optional filtering |

### CLI

| Command | Flags | Description |
|---------|-------|-------------|
| `agent-checkers games` | `--status`, `--player`, `--json` | List games in table or JSON format |

### MCP Server

| Tool | Parameters | Description |
|------|------------|-------------|
| `list_games` | `status` (optional), `player_id` (optional) | Returns array of game summaries |

## Response Schema (shared)

```json
{
  "games": [
    {
      "game_id": "uuid",
      "status": "waiting | active | completed",
      "current_turn": "red | black",
      "red_player": { "name": "Alice", "color": "red", "type": "human" },
      "black_player": { "name": "Bob", "color": "black", "type": "human" },
      "created_at": "2026-07-08T12:00:00Z"
    }
  ]
}
```

## Traceability

- **BDR-001**: Player Registration (list shows registered players)
- **BDR-003**: Game State Management (list queries game state)
- **BDR-005**: AI Agent Integration via MCP (MCP `list_games` tool)
- **BDR-008**: Command-Line Interface (CLI `games` command)
- **BDR-009**: Concurrent Games (list spans all concurrent games)
- **BDR-010**: OpenAPI Specification (endpoint documented in spec)
- **ADR-007**: Interface Layers (all interfaces share same domain layer)