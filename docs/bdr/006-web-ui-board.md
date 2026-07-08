# BDR-006: Web UI Board Visualization

## Status

Accepted

## Behavior

The system provides a browser-based visual interface displaying the checkers board with interactive piece selection, valid move highlighting, and real-time state updates.

## Context

Human players need a visual interface to play checkers. The Web UI must render an 8x8 board with pieces, allow players to select pieces and make moves via click/tap, display valid moves visually, and update the board in real-time when moves occur (via WebSocket). The UI must work on desktop and mobile browsers.

## Acceptance Criteria

- AC-1: The board displays an 8x8 grid with alternating dark/light squares
- AC-2: Red and black pieces are visually distinct with king indication
- AC-3: Clicking a piece highlights it and shows valid destination squares
- AC-4: Clicking a valid destination square executes the move
- AC-5: The board updates immediately after the current player's move
- AC-6: The board updates in real-time when the opponent moves (WebSocket)
- AC-7: Game status (waiting, active, winner) is displayed prominently
- AC-8: Move history is visible in a sidebar
- AC-9: The UI is responsive and works on mobile devices

## Verification

### Scenario 1: Initial board display

- **Given** a player opens the Web UI for an active game
- **Then** the board shows 12 red pieces on rows 0-2 and 12 black pieces on rows 5-7
- **And** pieces appear only on dark squares

### Scenario 2: Piece selection

- **Given** it is red's turn
- **When** red clicks on a red piece
- **Then** the piece is highlighted
- **And** valid move destinations are shown (highlighted squares)

### Scenario 3: Valid move execution

- **Given** a piece is selected and valid moves are highlighted
- **When** the player clicks a highlighted destination square
- **Then** the piece moves to that square
- **And** the board updates immediately
- **And** the turn indicator changes to the opponent

### Scenario 4: Real-time opponent move

- **Given** the opponent makes a move
- **When** the move is processed by the server
- **Then** the board updates via WebSocket without polling
- **And** the current player sees the opponent's move within 100ms

### Scenario 5: King promotion animation

- **Given** a piece reaches the opposite end of the board
- **Then** the piece is visually marked as a king
- **And** a subtle animation indicates the promotion

### Scenario 6: Capture animation

- **Given** a piece captures an opponent's piece
- **Then** the captured piece is removed with a visual animation
- **And** the capturing piece lands on the destination square

## Board Rendering

```
Visual representation:

    0   1   2   3   4   5   6   7
  +---+---+---+---+---+---+---+---+
0 |   | ● |   | ● |   | ● |   | ● |  ← Black pieces
  +---+---+---+---+---+---+---+---+
1 | ● |   | ● |   | ● |   | ● |   |
  +---+---+---+---+---+---+---+---+
...
  +---+---+---+---+---+---+---+---+
5 | ○ |   | ○ |   | ○ |   | ○ |   |  ← Red pieces
  +---+---+---+---+---+---+---+---+
6 |   | ○ |   | ○ |   | ○ |   | ○ |
  +---+---+---+---+---+---+---+---+
7 | ○ |   | ○ |   | ○ |   | ○ |   |
  +---+---+---+---+---+---+---+---+

○ = red piece    ● = black piece
♚ = red king      ♛ = black king
```

## Technologies

| Component | Technology |
|-----------|------------|
| Frontend | HTML5, CSS3, JavaScript |
| State Sync | WebSocket (Phase 4) |
| REST Client | Fetch API |
| Mobile | Responsive design |

## Interfaces

| Interface | Endpoint |
|-----------|----------|
| Web UI | `GET /` (static HTML) |
| WebSocket | `GET /api/v1/games/{id}/ws` |

## Traceability

- Spec: `docs/spec.md` - Feature: Web UI
- ADR: `docs/adr/007-interface-layers.md`