# BDR-013: Modal Dialogs for Destructive Actions

## Status

Proposed

## Behavior

The Web UI uses custom HTML modal dialogs for all destructive or confirmable actions (resign, leave game) instead of native browser `confirm()` calls, keeping the page responsive and testable.

## Context

The resign button currently calls `confirm()` which blocks the page thread, making the UI unresponsive in automated testing and creating a jarring UX. Custom modals are already used for New Game and Join Game — the resign action should follow the same pattern.

## Acceptance Criteria

- AC-1: Clicking "Resign" opens a custom HTML modal (not native `confirm()`)
- AC-2: The modal displays "Are you sure you want to resign?" with the current game result shown
- AC-3: The modal has "Confirm Resign" and "Cancel" buttons
- AC-4: The page remains responsive while the modal is open
- AC-5: Clicking "Confirm Resign" calls the API to resign the game
- AC-6: Clicking "Cancel" closes the modal without action
- AC-7: Pressing Escape closes the modal without action
- AC-8: The modal works in automated testing (Playwright, Selenium)

## Verification

### Scenario 1: Resign confirmation

- **Given** a user is in an active game
- **When** the user clicks "Resign"
- **Then** a custom modal appears with "Are you sure you want to resign?"
- **And** the page remains responsive
- **And** "Confirm Resign" and "Cancel" buttons are visible

### Scenario 2: Cancel resign

- **Given** the resign modal is open
- **When** the user clicks "Cancel"
- **Then** the modal closes
- **And** the game continues unchanged

### Scenario 3: Confirm resign

- **Given** the resign modal is open
- **When** the user clicks "Confirm Resign"
- **Then** the API resign endpoint is called
- **And** the game status updates to "completed"
- **And** the result shows the winner

### Scenario 4: Escape key

- **Given** the resign modal is open
- **When** the user presses the Escape key
- **Then** the modal closes without action

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| Web UI | `src/web/static/js/app.js` — custom modal dialog |

## Traceability

- **BDR-006**: Web UI Board Visualization (extends with modal dialogs)
- **Issue #16**: UI: Resign button triggers confirm() dialog that blocks page