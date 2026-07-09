# BDR-012: Session Persistence Across Page Refreshes

## Status

Proposed

## Behavior

The Web UI persists the player's game session (game ID, player ID, session token) in the browser so that refreshing or navigating away and back restores the active game without requiring the user to rejoin.

## Context

Currently all game state lives in JavaScript memory. A page refresh, accidental navigation, or browser crash loses the game entirely — the user sees "No game loaded" with no way to return to an active game. This is a critical UX gap for a web application. Browsers provide `localStorage` for exactly this purpose.

## Acceptance Criteria

- AC-1: On game creation, the UI stores `game_id`, `player_id`, `session_token`, and `player_color` in `localStorage`
- AC-2: On game join, the UI stores the same fields in `localStorage`
- AC-3: On page load, if `localStorage` contains a valid session, the UI automatically loads the game and renders the board
- AC-4: The status bar shows "Reconnecting..." while the game state is being loaded on refresh
- AC-5: If the stored game no longer exists (server restarted with in-memory store), the UI clears `localStorage` and shows "No game loaded"
- AC-6: When a game ends (completed, draw, or resigned), the UI clears `localStorage` so the next visit starts fresh
- AC-7: WebSocket reconnects automatically after page refresh using the stored session

## Verification

### Scenario 1: Game persists across refresh

- **Given** a user has created a game and the board is displayed
- **When** the user refreshes the page
- **Then** the board re-renders with the current game state
- **And** the status bar shows the correct game status
- **And** the player info shows the same player color

### Scenario 2: Game cleared after completion

- **Given** a game has ended (resignation or win)
- **When** the user refreshes the page
- **Then** the UI shows "No game loaded"
- **And** `localStorage` does not contain game session data

### Scenario 3: Stale session cleaned up

- **Given** the server was restarted (in-memory store lost)
- **When** the user refreshes the page with stale `localStorage`
- **Then** the UI detects the game no longer exists
- **And** clears `localStorage`
- **And** shows "No game loaded"

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/app.js` — localStorage read/write on load/exit |

## Traceability

- **BDR-006**: Web UI Board Visualization (extends with session persistence)
- **BDR-007**: Real-Time Game Updates (WebSocket reconnection)
- **Issue #15**: UI: Game state not persisted on page refresh