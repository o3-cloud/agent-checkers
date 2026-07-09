# BDR-017: Game Discovery in Web UI

## Status

Proposed

## Behavior

The Web UI displays a browsable list of waiting games when the user clicks "Join Game", allowing them to select and join a game by clicking it rather than manually typing a game UUID.

## Context

Currently the Join Game form requires the user to manually type a full game UUID. There is no way to discover available games from the UI, creating a major UX barrier. BDR-011 defines the list games capability across all interfaces — this BDR specifies how it appears in the Web UI specifically.

## Acceptance Criteria

- AC-1: When "Join Game" is clicked, the modal shows a list of waiting games fetched from `GET /api/v1/games?status=waiting`
- AC-2: Each game entry displays: red player name, game ID (truncated), and time created
- AC-3: Clicking a game from the list populates the game ID field and focuses the player name input
- AC-4: The user can still manually enter a game UUID as a fallback
- AC-5: If no waiting games exist, the list shows "No games waiting — create a new one!"
- AC-6: The list refreshes when the modal is opened (stale data is not shown)
- AC-7: After selecting a game and entering a name, clicking "Join" joins the selected game

## Verification

### Scenario 1: Browse waiting games

- **Given** 2 waiting games exist on the server
- **When** the user clicks "Join Game"
- **Then** the modal shows 2 game entries
- **And** each entry shows the red player's name and truncated game ID

### Scenario 2: Select and join from list

- **Given** the waiting games list is displayed
- **When** the user clicks a game entry
- **Then** the game ID field is populated with the selected game's UUID
- **And** the player name input is focused
- **When** the user enters their name and clicks "Join"
- **Then** the game is joined and the board renders

### Scenario 3: No waiting games

- **Given** no waiting games exist
- **When** the user clicks "Join Game"
- **Then** the list shows "No games waiting — create a new one!"
- **And** the game ID input is still available for manual entry

### Scenario 4: Manual UUID entry still works

- **Given** the waiting games list is displayed
- **When** the user ignores the list and types a full UUID manually
- **Then** clicking "Join" attempts to join that specific game

### Scenario 5: List refreshes on open

- **Given** a game was waiting, then was joined by another player
- **When** the user closes and reopens the Join Game modal
- **Then** that game is no longer in the waiting list

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/app.js` — game list fetch and display |
| REST API | `GET /api/v1/games?status=waiting` (from BDR-011) |

## Traceability

- **BDR-006**: Web UI Board Visualization (extends with game discovery)
- **BDR-011**: List Active Games (provides the API endpoint)
- **Issue #20**: UI: No way to list/join waiting games from the UI