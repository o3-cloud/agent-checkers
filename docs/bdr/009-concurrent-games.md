# BDR-009: Concurrent Games

## Status

Accepted

## Behavior

The system supports multiple concurrent games, allowing players to participate in more than one game simultaneously without state conflicts.

## Context

Players may want to play multiple games at once, or the server may host games for many independent player pairs. The system must isolate game state so that moves in one game do not affect another, and provide discovery mechanisms for finding available games.

## Acceptance Criteria

- AC-1: The server can host multiple games simultaneously
- AC-2: Each game has a unique identifier
- AC-3: Players can query a list of available (waiting) games
- AC-4: Players can join a specific game by ID
- AC-5: Game IDs are URL-safe and human-readable
- AC-6: Game state is completely isolated between games
- AC-7: A player can be in multiple games simultaneously
- AC-8: Rate limiting prevents abuse (max games per player, max moves per second)

## Verification

### Scenario 1: Multiple concurrent games

- **Given** game A and game B are both active
- **When** player 1 makes a move in game A
- **Then** game A's state changes
- **And** game B's state is unaffected

### Scenario 2: List waiting games

- **Given** multiple games exist (some waiting, some active)
- **When** a player queries `GET /api/v1/games?status=waiting`
- **Then** the response lists only games with one player waiting for an opponent

### Scenario 3: Join specific game

- **Given** game `game-abc123` is waiting for a player
- **When** player 2 calls `POST /api/v1/games/game-abc123/join`
- **Then** the game transitions to active
- **And** both players can now make moves

### Scenario 4: Player in multiple games

- **Given** player "Alice" is in game A
- **When** Alice registers for game B
- **Then** Alice is now in both games
- **And** Alice can query game A or game B independently

### Scenario 5: Rate limiting

- **Given** a player attempts to create more than N games per hour
- **When** the limit is exceeded
- **Then** the request is rejected with HTTP 429 Too Many Requests

## Game ID Format

```
game-<nanoid>
Example: game-V1StGX8M5j3
```

- Uses nanoid for compact, URL-safe identifiers
- 12 characters, case-sensitive
- Collision-resistant for practical scales

## Limits

| Resource | Limit |
|----------|-------|
| Games per player | 10 active |
| Moves per second | 5 per game |
| Game lifetime | 7 days (configurable) |

## Interfaces

| Interface | Endpoint |
|-----------|----------|
| REST API | `GET /api/v1/games?status=waiting` |
| REST API | `POST /api/v1/games/{id}/join` |

## Traceability

- Spec: `docs/spec.md` - Feature: Concurrent Games
- Code: `internal/app/store/memory.go`