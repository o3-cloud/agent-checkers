# BDR-003: Game State Management

## Status

Accepted

## Behavior

The system maintains the complete state of each game including the board configuration, current turn, player information, move history, and game status, making it queryable through all interfaces.

## Context

All players (human and AI) need to understand the current game state to make informed decisions. The game state must be consistent across all interface layers (REST API, WebSocket, CLI, MCP). The state persists between moves and survives server restarts when persistence is configured.

## Acceptance Criteria

- AC-1: The game state includes the 8x8 board configuration with all pieces and their positions
- AC-2: The game state identifies whose turn it is (red or black)
- AC-3: The game state identifies both players (names, types, colors)
- AC-4: The game state includes the current status (waiting, active, red_wins, black_wins, draw)
- AC-5: The game state includes the complete move history
- AC-6: State queries return a complete, serializable snapshot (JSON)
- AC-7: State is updated atomically after each move

## Verification

### Scenario 1: Query game state

- **Given** a game with two players and 3 moves completed
- **When** a player queries the game state
- **Then** the response includes:
  - The board configuration
  - Current turn (red or black)
  - Both player names and colors
  - Status "active"
  - Move history with 3 entries

### Scenario 2: Board representation

- **Given** a new game has started
- **When** the board is queried
- **Then** pieces are arranged in the standard checkers starting position:
  - Red pieces on rows 0, 1, 2 (12 pieces)
  - Black pieces on rows 5, 6, 7 (12 pieces)
  - Pieces only on dark squares

### Scenario 3: Move history

- **Given** a game with move history
- **When** a new move is made
- **Then** the move is appended to the history
- **And** each history entry includes: from position, to position, captured pieces, player ID

### Scenario 4: Waiting game state

- **Given** only one player has registered
- **When** the game state is queried
- **Then** status is "waiting"
- **And** only one player is listed

## Board Encoding

```
Position encoding: { row: 0-7, col: 0-7 }

Board cell values:
  null           - empty square
  { color: "red", king: false }   - red piece
  { color: "red", king: true }    - red king
  { color: "black", king: false } - black piece
  { color: "black", king: true }  - black king
```

## Interfaces

| Interface | Endpoint/Tool |
|-----------|---------------|
| REST API | `GET /api/v1/games/{id}` |
| REST API | `GET /api/v1/games/{id}/moves` |
| CLI | `agent-checkers board` |
| MCP | `get_game_state(game_id)` |

## Traceability

- Spec: `docs/spec.md` - Feature: Game State Management
- Code: `internal/app/game/game.go`
- Code: `internal/app/board/board.go`