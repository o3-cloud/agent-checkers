# ADR-007: Interface Layers (Web, API, CLI, MCP)

## Status

Proposed

## Context

Agent-checkers must support four distinct interface layers:

1. **Web UI** — Visual board for human players (browser-based)
2. **REST API** — HTTP endpoints for programmatic access
3. **CLI** — Command-line interface for development/testing
4. **MCP Server** — Model Context Protocol for AI agent integration

All four interfaces must:
- Share the same game engine
- Support the same operations (create, join, move, state, resign)
- Handle authentication consistently
- Report errors uniformly

## Decision

We will implement **interface adapters** that translate between external protocols and the application layer:

```
┌─────────────────────────────────────────────────────────────┐
│                    External World                            │
│  Browser     HTTP Client     Terminal      AI Agent         │
└─────┬───────────┬──────────────┬──────────────┬─────────────┘
      │           │              │              │
      ▼           ▼              ▼              ▼
┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────┐
│ Web UI  │ │ REST API │ │   CLI    │ │ MCP Server │
│ (HTML)  │ │ (JSON)   │ │ (Text)   │ │ (JSON-RPC) │
└────┬────┘ └────┬─────┘ └────┬─────┘ └──────┬──────┘
     │           │            │             │
     └───────────┴────────────┴─────────────┘
                       │
                       ▼
           ┌─────────────────────┐
           │  Application Layer  │
           │  (Lobby, Session,   │
           │   Game Manager)     │
           └──────────┬──────────┘
                      │
                      ▼
           ┌─────────────────────┐
           │    Domain Layer     │
           │  (Game, Board,      │
           │   Rules, Player)    │
           └─────────────────────┘
```

### Interface Definitions

#### REST API (`src/api/`)

**Endpoints:**

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/games` | Create new game |
| `POST` | `/api/v1/games/{id}/join` | Join existing game |
| `GET` | `/api/v1/games/{id}` | Get game state |
| `POST` | `/api/v1/games/{id}/moves` | Make a move |
| `GET` | `/api/v1/games/{id}/moves` | Get move history |
| `DELETE` | `/api/v1/games/{id}` | Resign game |
| `POST` | `/api/v1/games/{id}/draw` | Offer/accept draw |
| `GET` | `/api/v1/games/{id}/ws` | WebSocket endpoint |

**Request/Response Format:**

```json
// POST /api/v1/games/{id}/moves
{
  "from": { "row": 2, "col": 3 },
  "to": { "row": 3, "col": 4 }
}

// Response
{
  "success": true,
  "game_state": {
    "board": [...],
    "current_turn": "black",
    "status": "active"
  }
}

// Error Response
{
  "error": "Invalid move: destination square is not empty",
  "valid_moves": [{ "from": {...}, "to": [...] }]
}
```

#### WebSocket Protocol (`src/api/websocket/`)

**Events:**

| Event | Direction | Payload |
|-------|-----------|---------|
| `game_started` | Server → Client | `{ "game_id": "...", "players": [...] }` |
| `move_made` | Server → Client | `{ "from": {...}, "to": {...}, "captured": [...] }` |
| `turn_changed` | Server → Client | `{ "current_player": "red" }` |
| `game_ended` | Server → Client | `{ "winner": "red", "reason": "capture_all" }` |
| `move` | Client → Server | `{ "from": {...}, "to": {...} }` |

#### CLI (`src/cli/`)

**Commands:**

```
agent-checkers new --name "Alice"              # Create game
agent-checkers join <game-id> --name "Bob"     # Join game
agent-checkers board                            # Display board
agent-checkers move <from> <to>                # Make move
agent-checkers moves                            # List valid moves
agent-checkers watch                            # Watch game live
agent-checkers vs --ai "Claude"                # Play vs AI
```

**Board Display:**

```
    0   1   2   3   4   5   6   7
  +---+---+---+---+---+---+---+---+
0 |   | ● |   | ● |   | ● |   | ● |
  +---+---+---+---+---+---+---+---+
1 | ● |   | ● |   | ● |   | ● |   |
  +---+---+---+---+---+---+---+---+
2 |   | ○ |   |   |   | ● |   | ● |
  +---+---+---+---+---+---+---+---+
...

● = black piece   ○ = red piece
♛ = black king    ♚ = red king
```

#### MCP Server (`src/mcp/`)

**Tools:**

| Tool | Parameters | Returns |
|------|-----------|---------|
| `register_player` | `name`, `type` | `player_id`, `game_id` |
| `get_game_state` | `game_id` | board, turn, status |
| `make_move` | `game_id`, `from`, `to` | success, new_state |
| `get_valid_moves` | `game_id` | list of valid moves |
| `resign` | `game_id` | winner, reason |
| `offer_draw` | `game_id` | status |
| `accept_draw` | `game_id` | game status |

**Protocol:** JSON-RPC 2.0 over stdio

**Example:**

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "make_move",
  "params": {
    "game_id": "game-123",
    "from": { "row": 2, "col": 3 },
    "to": { "row": 3, "col": 4 }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "success": true,
    "board": [...],
    "current_turn": "black"
  }
}
```

## Alternatives Considered

### 1. GraphQL instead of REST
- **Pros:** Flexible queries, single endpoint
- **Cons:** Overkill for simple CRUD, learning curve
- **Decision:** Rejected — REST is simpler for this domain

### 2. gRPC for API
- **Pros:** Type-safe, efficient binary protocol
- **Cons:** Requires protobuf definitions, harder to debug
- **Decision:** Rejected — JSON is easier for web integration

### 3. TUI instead of CLI commands
- **Pros:** Interactive terminal UI
- **Cons:** Harder to script, harder to test
- **Decision:** Rejected — CLI commands are composable

## Consequences

### Positive
- Each interface can be developed independently
- Domain logic has zero HTTP/WebSocket/CLI dependencies
- Easy to add new interfaces (e.g., gRPC, GraphQL)
- Testing is simplified (mock application layer)

### Negative
- Duplicate DTOs (Data Transfer Objects) across interfaces
- Need to keep error messages consistent

### Risks
- Interface layer could become "fat" if not careful
- Need good integration tests to catch protocol mismatches

## Implementation Order

1. **Phase 3:** REST API (foundation for testing)
2. **Phase 4:** WebSocket (real-time updates)
3. **Phase 5:** Web UI (uses REST + WebSocket)
4. **Phase 6:** CLI (uses REST client)
5. **Phase 7:** MCP Server (uses application layer directly)

## References

- [Ports and Adapters Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [MCP Specification](https://modelcontextprotocol.io/)