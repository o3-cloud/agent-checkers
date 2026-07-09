# BDR-015: King Piece Visualization

## Status

Proposed

## Behavior

The Web UI visually distinguishes king pieces from regular pieces using a crown symbol and distinct styling, so players can immediately identify kings on the board.

## Context

Currently all pieces render as identical colored circles regardless of whether they are kings. This makes it impossible to tell kings from regular pieces at a glance, which affects gameplay decisions. Kings can move in both directions and are significantly more powerful.

## Acceptance Criteria

- AC-1: King pieces display a crown symbol (♚ for red, ♛ for black) centered on the piece
- AC-2: King pieces have a distinct border style (e.g., gold/dark gold ring) different from regular pieces
- AC-3: Regular pieces remain as plain colored circles with no symbol
- AC-4: When a piece is promoted to king, the visual update is immediate (no page refresh needed)
- AC-5: King styling is visible in both selected and unselected states
- AC-6: King styling works on both light and dark board squares

## Verification

### Scenario 1: Initial board has no kings

- **Given** a new game has started
- **When** the board renders
- **Then** all 24 pieces are plain colored circles
- **And** no crown symbols are visible

### Scenario 2: Piece promotion to king

- **Given** a red piece moves to row 7 (opposite end)
- **When** the move is completed and the piece is promoted
- **Then** the piece at row 7 displays a crown symbol (♚)
- **And** the piece has a gold border
- **And** the visual update is immediate without page refresh

### Scenario 3: King vs regular distinction

- **Given** a board with both kings and regular pieces
- **When** the user views the board
- **Then** kings are clearly distinguishable by crown symbol and border
- **And** regular pieces have no symbol and standard styling

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/board.js` — piece rendering with king indicator |
| Web UI | `src/web/static/css/board.css` — king piece styling |

## Traceability

- **BDR-006**: Web UI Board Visualization (extends with king visualization)
- **Issue #18**: UI: Board pieces have no visual indicators (text/symbols)