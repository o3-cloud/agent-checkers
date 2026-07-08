# BDR-002: Move Validation

## Status

Accepted

## Behavior

The system validates all move requests according to standard American checkers rules before applying them to the game state, rejecting any illegal moves with a descriptive error.

## Context

Checkers has strict rules governing piece movement. Non-king pieces can only move forward diagonally. Kings can move in all four diagonal directions. Captures are mandatory—if a capture is available, the player must take it. Multi-jump sequences must be completed in a single turn. Invalid moves must be rejected to maintain game integrity, whether the move comes from a human player or an AI agent.

## Acceptance Criteria

- AC-1: Moves must be diagonal (non-orthogonal, non-knight moves)
- AC-2: Moves must be to an adjacent empty square (simple move) or jump over an opponent piece (capture)
- AC-3: Non-king pieces can only move forward (red moves toward row 7, black toward row 0)
- AC-4: King pieces can move in all four diagonal directions
- AC-5: If a capture is available, the player must make a capture (mandatory jump rule)
- AC-6: Multi-jump sequences must be completed; the turn does not end until no more captures are available
- AC-7: Moves on the wrong turn are rejected
- AC-8: The system returns the list of valid moves when an invalid move is attempted

## Verification

### Scenario 1: Valid simple move

- **Given** it is red's turn and a red piece is at position (2, 3)
- **When** red attempts to move from (2, 3) to (3, 4)
- **Then** the move is accepted
- **And** the piece is now at (3, 4)

### Scenario 2: Invalid non-diagonal move

- **Given** it is red's turn and a red piece is at position (2, 3)
- **When** red attempts to move from (2, 3) to (3, 3) (same column)
- **Then** the move is rejected
- **And** the error message includes "diagonal"

### Scenario 3: Wrong turn

- **Given** it is red's turn
- **When** black attempts to move
- **Then** the move is rejected with error "not your turn"

### Scenario 4: Mandatory capture

- **Given** it is red's turn and a capture is available
- **When** red attempts a simple (non-capture) move
- **Then** the move is rejected with error "a capture is available, you must capture"

### Scenario 5: King promotion

- **Given** a red piece reaches row 7 (the far end of the board)
- **Then** the piece is promoted to a king
- **And** can now move backward on subsequent turns

### Scenario 6: Multi-jump sequence

- **Given** a red piece can capture two black pieces in sequence
- **When** red makes the first capture
- **Then** the turn does not end
- **And** red must continue capturing until no more captures are available from the landing square

## Interfaces

| Interface | Endpoint/Tool |
|-----------|---------------|
| REST API | `POST /api/v1/games/{id}/moves` |
| CLI | `agent-checkers move <from> <to>` |
| MCP | `make_move(game_id, from, to)` |

## Traceability

- Spec: `docs/spec.md` - Feature: Move Validation
- Code: `internal/app/rules/validator.go`
- ADR: `docs/adr/006-game-engine-architecture.md`