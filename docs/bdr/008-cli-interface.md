# BDR-008: Command-Line Interface

## Status

Accepted

## Behavior

The system provides a command-line interface for playing checkers from a terminal, useful for development, testing, and scripted gameplay.

## Context

Developers and AI agents may prefer a CLI interface for scripting, automation, or testing. The CLI provides the same capabilities as the REST API but through command invocations. Output is formatted for human readability (ASCII board) or machine parsing (JSON flag).

## Acceptance Criteria

- AC-1: CLI supports creating a new game with `agent-checkers new`
- AC-2: CLI supports joining an existing game with `agent-checkers join <game-id>`
- AC-3: CLI supports viewing the board with `agent-checkers board`
- AC-4: CLI supports making a move with `agent-checkers move <from> <to>`
- AC-5: CLI supports listing valid moves with `agent-checkers moves`
- AC-6: CLI supports resigning with `agent-checkers resign`
- AC-7: CLI supports playing against AI with `agent-checkers vs --ai`
- AC-8: CLI outputs human-readable ASCII board by default
- AC-9: CLI supports `--json` flag for machine-readable output
- AC-10: CLI stores session token locally for persistence across invocations

## Verification

### Scenario 1: Create new game

- **Given** the CLI is installed
- **When** user runs `agent-checkers new --name "Alice"`
- **Then** the output shows:
  ```
  Game created: game-abc123
  You are: Red
  Status: Waiting for opponent...
  Share: agent-checkers join game-abc123 --name "Bob"
  ```

### Scenario 2: Join existing game

- **Given** a game exists with ID `game-abc123`
- **When** user runs `agent-checkers join game-abc123 --name "Bob"`
- **Then** the output shows:
  ```
  Joined game: game-abc123
  You are: Black
  Status: Active
  Red player: Alice
  ```

### Scenario 3: View board

- **Given** an active game
- **When** user runs `agent-checkers board`
- **Then** the output shows an ASCII representation:
  ```
      0   1   2   3   4   5   6   7
    +---+---+---+---+---+---+---+---+
  0 |   | ● |   | ● |   | ● |   | ● |
    +---+---+---+---+---+---+---+---+
  1 | ● |   | ● |   | ● |   | ● |   |
    ...
  ```

### Scenario 4: Make a move

- **Given** it's the player's turn
- **When** user runs `agent-checkers move 2,3 3,4`
- **Then** the output shows:
  ```
  Move accepted: (2,3) → (3,4)
  Turn: Black
  ```

### Scenario 5: List valid moves

- **Given** it's the player's turn
- **When** user runs `agent-checkers moves`
- **Then** the output shows all legal moves for the current player

### Scenario 6: JSON output

- **Given** any CLI command
- **When** user adds `--json` flag
- **Then** the output is valid JSON suitable for scripting

## Commands

| Command | Description |
|---------|-------------|
| `agent-checkers new --name <name>` | Create new game |
| `agent-checkers join <id> --name <name>` | Join existing game |
| `agent-checkers board` | Display current board |
| `agent-checkers move <from> <to>` | Make a move |
| `agent-checkers moves` | List valid moves |
| `agent-checkers resign` | Resign game |
| `agent-checkers draw --offer` | Offer draw |
| `agent-checkers draw --accept` | Accept draw |
| `agent-checkers vs --ai <name>` | Play against AI |

## Session Storage

- Session token stored in `~/.agent-checkers/session.json`
- Game ID cached for convenience (last joined game)
- Token expires per server TTL (default 24 hours)

## Interfaces

| Interface | Implementation |
|-----------|----------------|
| CLI | `src/cli/main.go` |
| Config | `~/.agent-checkers/` |

## Traceability

- Spec: `docs/spec.md` - Feature: CLI
- ADR: `docs/adr/007-interface-layers.md`