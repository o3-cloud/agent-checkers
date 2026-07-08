# BDR-001: Player Registration

## Status

Accepted

## Behavior

The system allows human players and AI agents to register for a checkers game, assigning them a unique identifier, color (red or black), and establishing their place in the game.

## Context

Checkers is a two-player game. Before gameplay begins, players must identify themselves to the system. The registration process determines which player moves first (red) and creates a waiting state for matchmaking. AI agents register the same way as humans, enabling human-vs-AI and AI-vs-AI games.

## Acceptance Criteria

- AC-1: A human player can register with a name and receive a unique player ID
- AC-2: An AI agent can register with a name and type="ai" and receive a unique player ID
- AC-3: The first player to register is assigned the color red and the waiting state
- AC-4: The second player to register is assigned the color black and the game becomes active
- AC-5: Registration returns a session token for subsequent authenticated operations
- AC-6: Registration fails with an error if the game already has two players

## Verification

### Scenario 1: First player registration (Human)

- **Given** no game exists
- **When** a human registers with name "Alice"
- **Then** the system creates a game in "waiting" status
- **And** assigns the player color "red"
- **And** returns a player ID and session token

### Scenario 2: Second player registration (AI)

- **Given** a game exists with one player (red, waiting)
- **When** an AI agent registers with name "Claude" and type="ai"
- **Then** the system sets the game status to "active"
- **And** assigns the player color "black"
- **And** returns a player ID and session token

### Scenario 3: Registration rejected for full game

- **Given** a game exists with two players (active status)
- **When** a third player attempts to register
- **Then** the system returns an error "game is full"

### Scenario 4: Session token validation

- **Given** a player has registered and received a session token
- **When** the player makes a move using the session token
- **Then** the system validates the token and accepts the operation

## Interfaces

| Interface | Endpoint/Tool |
|-----------|---------------|
| REST API | `POST /api/v1/games` |
| REST API | `POST /api/v1/games/{id}/join` |
| CLI | `agent-checkers join <game-id> --name "Bob"` |
| MCP | `register_player(name, type)` |

## Traceability

- Spec: `docs/spec.md` - Feature: Game Registration
- ADR: `docs/adr/007-interface-layers.md`