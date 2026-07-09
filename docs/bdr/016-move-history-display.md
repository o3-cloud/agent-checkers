# BDR-016: Move History Display

## Status

Proposed

## Behavior

The Web UI displays a chronological list of all moves made in the current game, showing each move's player color, origin square, destination square, and any captures or promotions, in a readable notation.

## Context

The Move History sidebar currently shows empty list entries — the count is correct but the content is blank. Players cannot review what moves were made, making it impossible to analyze the game or verify opponents' moves. A clear move history is essential for gameplay comprehension.

## Acceptance Criteria

- AC-1: Each move entry displays the player color (Red/Black) and move number
- AC-2: Each move entry shows the from and to positions in algebraic notation (e.g., "c3 → d4")
- AC-3: Capture moves are marked with an "×" or "capture" indicator
- AC-4: King promotions are marked with a crown symbol or "K" indicator
- AC-5: The move list updates in real-time after each move (no manual refresh)
- AC-6: The list auto-scrolls to the latest move after a new move is added
- AC-7: The move list is ordered oldest-first (move 1 at top, latest at bottom)
- AC-8: Empty move list (new game) shows no entries without error

## Verification

### Scenario 1: Move appears in history

- **Given** a game is active and Red moves from c3 to d4
- **When** the move is completed
- **Then** the Move History sidebar shows: "1. Red: c3 → d4"

### Scenario 2: Capture marked in history

- **Given** Red jumps from c3 to e5 capturing a black piece
- **When** the move is completed
- **Then** the history shows: "3. Red: c3 × e5" (or "c3 → e5 capture")

### Scenario 3: Promotion marked in history

- **Given** Red moves to row 7 and the piece is promoted to king
- **When** the move is completed
- **Then** the history shows: "5. Red: d6 → d8 ♚" (or "d6 → d8 K")

### Scenario 4: Auto-scroll to latest

- **Given** 10 moves have been made and the list is scrollable
- **When** move 11 is completed
- **Then** the list scrolls to show move 11 at the bottom

### Scenario 5: Real-time update

- **Given** the opponent makes a move via API
- **When** the WebSocket event is received
- **Then** the move appears in the history without page refresh

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/app.js` — move history rendering |
| Web UI | `src/web/static/js/board.js` — algebraic notation conversion |

## Traceability

- **BDR-003**: Game State Management (move history is part of game state)
- **BDR-006**: Web UI Board Visualization (extends with move history display)
- **BDR-007**: Real-Time Game Updates (WebSocket triggers history update)
- **Issue #19**: UI: Move history entries are empty/not rendering content