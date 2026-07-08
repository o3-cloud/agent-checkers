# BDR-004: Win/Lose/Draw Conditions

## Status

Accepted

## Behavior

The system detects and declares game-ending conditions: win by capture, win by blocking, draw by agreement, and resignation, transitioning the game state accordingly.

## Context

A checkers game must end when a player loses all pieces, cannot make a legal move, agrees to a draw, or resigns. The system must detect these conditions automatically and update the game status atomically so that subsequent queries reflect the correct terminal state.

## Acceptance Criteria

- AC-1: A player wins when the opponent has no remaining pieces
- AC-2: A player wins when the opponent has no legal moves available
- AC-3: A draw occurs when both players agree to a draw
- AC-4: A player can resign at any time during their turn
- AC-5: When a terminal condition is reached, the game status transitions to `red_wins`, `black_wins`, or `draw`
- AC-6: No further moves are allowed after a terminal status is reached
- AC-7: All connected interfaces are notified of the game-ending event

## Verification

### Scenario 1: Win by capture all pieces

- **Given** red captures black's last piece
- **Then** the game status becomes `red_wins`
- **And** no further moves are accepted

### Scenario 2: Win by blocking

- **Given** black has pieces remaining but no legal moves
- **Then** the game status becomes `red_wins` (it's black's turn but cannot move)
- **And** the win reason is recorded as "blocked"

### Scenario 3: Draw by agreement

- **Given** it is red's turn
- **When** red offers a draw and black accepts
- **Then** the game status becomes `draw`
- **And** both players' consent is recorded

### Scenario 4: Resignation

- **Given** it is red's turn
- **When** red resigns
- **Then** the game status becomes `black_wins`
- **And** the win reason is recorded as "resignation"

### Scenario 5: Move rejected after game ends

- **Given** the game status is `red_wins`
- **When** black attempts to make a move
- **Then** the move is rejected with error "game is over"

## Terminal States

| Status | Description | Next Action |
|--------|-------------|-------------|
| `red_wins` | Red player has won | No moves accepted |
| `black_wins` | Black player has won | No moves accepted |
| `draw` | Players agreed to draw | No moves accepted |

## Interfaces

| Interface | Endpoint/Tool |
|-----------|---------------|
| REST API | `POST /api/v1/games/{id}/draw` |
| REST API | `DELETE /api/v1/games/{id}` (resign) |
| CLI | `agent-checkers resign` |
| CLI | `agent-checkers draw --offer` |
| MCP | `resign(game_id)` |
| MCP | `offer_draw(game_id)` |
| MCP | `accept_draw(game_id)` |

## Traceability

- Spec: `docs/spec.md` - Feature: Win/Lose/Draw
- Code: `internal/app/game/game.go`