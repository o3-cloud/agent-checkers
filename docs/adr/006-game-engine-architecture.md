# ADR-006: Game Engine Architecture

## Status

Proposed

## Context

Agent-checkers needs a game engine that can:
1. Validate moves according to standard checkers rules
2. Track game state (board, pieces, turns)
3. Detect win/lose/draw conditions
4. Support multiple concurrent games
5. Be queried by multiple interfaces (Web, API, CLI, MCP)

The engine must be:
- Fast (sub-100ms move validation)
- Testable (100% rule coverage)
- Interface-agnostic (no HTTP/WebSocket dependencies)

## Decision

We will implement a **layered game engine** with clean architecture:

```
┌─────────────────────────────────────┐
│         Interface Layer             │
│  (HTTP, WebSocket, CLI, MCP)        │
├─────────────────────────────────────┤
│         Application Layer           │
│  (Lobby, Session, Game Manager)     │
├─────────────────────────────────────┤
│         Domain Layer                │
│  (Game, Board, Piece, Player)       │
│  (Rules, Validator, Captures)        │
├─────────────────────────────────────┤
│         Infrastructure Layer        │
│  (Memory Store, Redis Store)        │
└─────────────────────────────────────┘
```

### Domain Layer (Core)

**`internal/app/board/board.go`**
- 8x8 grid representation
- Position abstraction (row, col)
- Piece placement and removal

**`internal/app/piece/piece.go`**
- Color enum (Red, Black)
- King status
- Position tracking

**`internal/app/game/game.go`**
- Game struct (ID, board, players, turn, status)
- Move execution
- Turn management
- Win condition detection

**`internal/app/rules/validator.go`**
- Move validation (diagonal, forward for non-kings)
- Capture detection
- Mandatory jump enforcement
- Multi-jump sequence handling

### Application Layer

**`internal/app/lobby/lobby.go`**
- Player registration
- Game matchmaking
- Waiting queue management

**`internal/app/session/session.go`**
- Player session tracking
- Authentication tokens
- Game-to-player mapping

### Infrastructure Layer

**`internal/app/store/memory.go`**
- In-memory game storage (Phase 1)
- Concurrent-safe map with mutex

**`internal/app/store/redis.go`** (Phase 9)
- Redis-backed persistence
- JSON serialization

## Board Representation

We will use a **simple 2D array** rather than bitboard optimization:

```go
type Board struct {
    squares [8][8]*Piece  // nil = empty
}

type Position struct {
    Row int  // 0-7, row 0 is black's home row
    Col int  // 0-7, col 0 is leftmost from player's view
}
```

**Rationale:**
- Clarity over performance for game logic
- 8x8 = 64 squares, trivial overhead
- Easy to debug and visualize
- Bitboards would be premature optimization

## Alternatives Considered

### 1. Bitboard Representation
- **Pros:** Faster move generation, memory efficient
- **Cons:** Harder to debug, overkill for checkers (vs chess)
- **Decision:** Rejected - premature optimization

### 2. Entity Component System (ECS)
- **Pros:** Flexible, composable
- **Cons:** Over-engineered for this domain
- **Decision:** Rejected - doesn't match problem complexity

### 3. External Game Engine Library
- **Pros:** Battle-tested rules
- **Cons:** Learning curve, less control, may not support our interface needs
- **Decision:** Rejected - core domain should be owned

## Consequences

### Positive
- Clean separation of concerns
- Highly testable (domain layer has no external dependencies)
- Easy to add new interfaces without touching game logic
- Performance is sufficient (64 squares, simple rules)

### Negative
- More boilerplate than monolithic approach
- Requires careful interface design to avoid leaking abstractions

### Risks
- Over-abstraction if not disciplined
- Need to ensure domain layer stays pure

## Implementation Notes

1. Start with domain layer (Phase 1)
2. Add application layer (Phase 2)
3. Interfaces build on application layer (Phases 3-7)
4. Infrastructure can be swapped (Phase 9 for Redis)

## References

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/08/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)