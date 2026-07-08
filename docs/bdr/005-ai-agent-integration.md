# BDR-005: AI Agent Integration via MCP

## Status

Accepted

## Behavior

The system exposes game operations as MCP (Model Context Protocol) tools, enabling AI agents to play checkers through a standardized interface with the same capabilities as human players.

## Context

AI agents need a programmatic interface to interact with the game. The MCP server provides structured tool definitions that AI models can invoke. Each tool has defined parameters, return types, and error conditions. This enables Claude, GPT, and other AI agents to play checkers without visual processing.

## Acceptance Criteria

- AC-1: AI agents can register as players using `register_player` tool
- AC-2: AI agents can query game state using `get_game_state` tool
- AC-3: AI agents can make moves using `make_move` tool
- AC-4: AI agents can query valid moves using `get_valid_moves` tool
- AC-5: AI agents can resign using `resign` tool
- AC-6: AI agents can offer/accept draws using `offer_draw` and `accept_draw` tools
- AC-7: All tools return structured JSON responses
- AC-8: All tools return descriptive errors for invalid operations
- AC-9: The MCP server runs over stdio transport

## Verification

### Scenario 1: AI player registration

- **Given** the MCP server is running
- **When** an AI agent calls `register_player(name="Claude", type="ai")`
- **Then** the response contains `player_id` and `game_id`
- **And** the AI is registered as a player

### Scenario 2: AI queries valid moves

- **Given** an AI agent is registered and it's their turn
- **When** the AI calls `get_valid_moves(game_id)`
- **Then** the response contains a list of legal moves
- **And** each move includes `from` and `to` positions

### Scenario 3: AI makes a move

- **Given** an AI agent knows a valid move
- **When** the AI calls `make_move(game_id, from, to)`
- **Then** the move is applied if valid
- **Or** an error is returned with available valid moves if invalid

### Scenario 4: AI handles error

- **Given** an AI agent attempts an illegal move
- **When** the move is rejected
- **Then** the error response includes `valid_moves` field
- **And** the AI can use this to select a legal move

## MCP Tool Definitions

### `register_player`

```json
{
  "name": "register_player",
  "parameters": {
    "name": "string - Player display name",
    "type": "string - 'human' or 'ai'"
  },
  "returns": {
    "player_id": "string",
    "game_id": "string",
    "color": "red | black"
  }
}
```

### `get_game_state`

```json
{
  "name": "get_game_state",
  "parameters": {
    "game_id": "string"
  },
  "returns": {
    "board": "[[Piece | null]]",
    "current_turn": "red | black",
    "status": "waiting | active | red_wins | black_wins | draw",
    "players": { "red": Player, "black": Player },
    "move_history": "Move[]"
  }
}
```

### `make_move`

```json
{
  "name": "make_move",
  "parameters": {
    "game_id": "string",
    "from": { "row": "0-7", "col": "0-7" },
    "to": { "row": "0-7", "col": "0-7" }
  },
  "returns": {
    "success": "boolean",
    "game_state": "GameState",
    "captured": "Position[]"
  },
  "errors": {
    "invalid_move": "Move does not follow checkers rules",
    "not_your_turn": "It is not your turn",
    "game_over": "Game has ended"
  }
}
```

### `get_valid_moves`

```json
{
  "name": "get_valid_moves",
  "parameters": {
    "game_id": "string"
  },
  "returns": {
    "moves": [{ "from": Position, "to": Position[] }]
  }
}
```

### `resign`

```json
{
  "name": "resign",
  "parameters": {
    "game_id": "string",
    "player_id": "string"
  },
  "returns": {
    "winner": "red | black",
    "reason": "resignation"
  }
}
```

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| MCP Server | `src/mcp/server.go` |
| Protocol | JSON-RPC 2.0 over stdio |
| Transport | Standard input/output |

## Traceability

- Spec: `docs/spec.md` - Feature: MCP Server
- ADR: `docs/adr/007-interface-layers.md`