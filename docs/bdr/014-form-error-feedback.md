# BDR-014: Form Error Feedback

## Status

Proposed

## Behavior

The Web UI displays inline error messages inside modal forms when API calls fail, so the user understands what went wrong and can retry without losing context.

## Context

Currently when a Join Game or Create Game API call fails, the modal stays open but no error is shown. The user has no idea why the operation failed — whether the game is full, not found, or the input was invalid. This creates a confusing dead-end experience.

## Acceptance Criteria

- AC-1: When a form submission API call returns an error, an error message is displayed inside the modal form
- AC-2: Error messages are shown in red text below or above the form fields
- AC-3: Specific errors from the API are displayed (e.g., "Game is already full", "Game not found", "Invalid player name")
- AC-4: The modal stays open on error so the user can correct input and retry
- AC-5: The error message auto-dismisses when the user starts typing or resubmits
- AC-6: Network errors (server unreachable) show "Unable to connect to server"
- AC-7: Validation errors (empty name, missing game ID) are caught before API call

## Verification

### Scenario 1: Join full game

- **Given** a game already has 2 players
- **When** the user tries to join with that game ID
- **Then** the modal shows "Game is already full" in red text
- **And** the modal remains open for retry

### Scenario 2: Join nonexistent game

- **Given** the user enters a game ID that does not exist
- **When** the user clicks "Join"
- **Then** the modal shows "Game not found" in red text

### Scenario 3: Create game with empty name

- **Given** the user leaves the player name field empty
- **When** the user clicks "Create"
- **Then** the modal shows "Player name is required" in red text
- **And** no API call is made

### Scenario 4: Server unreachable

- **Given** the server is down
- **When** the user clicks "Create"
- **Then** the modal shows "Unable to connect to server" in red text

### Scenario 5: Error clears on retry

- **Given** an error message is displayed in the modal
- **When** the user starts typing in any field
- **Then** the error message disappears

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/app.js` — inline form error display |

## Traceability

- **BDR-006**: Web UI Board Visualization (extends with error feedback)
- **Issue #17**: UI: Join Game form does not show error feedback on failure