# ADR-008: State Management (In-Memory vs Redis)

## Status

Proposed

## Context

Agent-checkers needs to persist game state between moves. The system must:

1. Store game state (board, players, turn, status)
2. Support multiple concurrent games
3. Allow players to reconnect and resume
4. Enable game history and replay
5. Scale horizontally if needed

State access patterns:
- **Read-heavy:** Game state queried on every move validation
- **Write-medium:** State updated after each move
- **Latency-sensitive:** Players expect sub-100ms response

## Decision

We will implement a **two-phase persistence strategy**:

### Phase 1: In-Memory Store (Default)

```go
type MemoryStore struct {
    games   map[string]*game.Game
    players map[string]*player.Player
    mu      sync.RWMutex
}
```

**Characteristics:**
- Zero external dependencies
- Sub-millisecond latency
- Simple implementation
- Game state lost on server restart
- No horizontal scaling

**Use Case:** Development, testing, single-server deployments

### Phase 2: Redis Store (Production)

```go
type RedisStore struct {
    client *redis.Client
}
```

**Characteristics:**
- Persistent across restarts
- Supports reconnection
- Horizontal scaling possible
- Requires Redis infrastructure
- ~1-5ms latency overhead

**Use Case:** Production, multi-instance deployments, game history

## Store Interface

Both implementations will share a common interface:

```go
// Store defines the persistence contract
type Store interface {
    // Game operations
    SaveGame(g *game.Game) error
    LoadGame(id string) (*game.Game, error)
    DeleteGame(id string) error
    ListGames(filter GameFilter) ([]*game.Game, error)
    
    // Player operations
    SavePlayer(p *player.Player) error
    LoadPlayer(id string) (*player.Player, error)
    
    // Move history
    AppendMove(gameID string, move game.Move) error
    GetMoveHistory(gameID string) ([]game.Move, error)
}
```

## Configuration

```yaml
# config.yaml
store:
  type: memory  # or "redis"
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10
```

## Data Model

### Game State (Redis JSON)

```json
{
  "id": "game-123",
  "board": [
    [null, {"color": "black", "king": false}, null, ...],
    ...
  ],
  "red_player": {
    "id": "player-1",
    "name": "Alice",
    "type": "human"
  },
  "black_player": {
    "id": "player-2",
    "name": "Claude",
    "type": "ai"
  },
  "current_turn": "red",
  "status": "active",
  "moves": [
    {"from": {"row": 2, "col": 3}, "to": {"row": 3, "col": 4}, "timestamp": "..."}
  ],
  "created_at": "2026-07-07T12:00:00Z",
  "updated_at": "2026-07-07T12:30:00Z"
}
```

### Redis Key Schema

```
game:{id}              -> JSON game state
game:{id}:moves        -> List of move IDs
move:{id}              -> JSON move record
player:{id}            -> JSON player data
player:{id}:games      -> Set of game IDs
games:active           -> Set of active game IDs
```

## Alternatives Considered

### 1. PostgreSQL (Relational Database)
- **Pros:** ACID guarantees, rich querying, mature ecosystem
- **Cons:** Overkill for document-style game state, higher latency
- **Decision:** Rejected — Redis is better fit for key-value access pattern

### 2. SQLite (Embedded Database)
- **Pros:** Zero-config, file-based, good for single instance
- **Cons:** No horizontal scaling, file locking issues
- **Decision:** Rejected — Redis is more flexible for scaling

### 3. MongoDB (Document Database)
- **Pros:** Document model fits game state, flexible schema
- **Cons:** Heavier than needed, adds complexity
- **Decision:** Rejected — Redis is simpler and faster for this use case

### 4. In-Memory Only (No Redis Option)
- **Pros:** Simplest implementation
- **Cons:** No persistence, no scaling
- **Decision:** Rejected — Need path to production

## Consequences

### Positive
- Start simple (in-memory), add complexity when needed
- Same interface for both implementations
- Easy to test (swap in memory store)
- Clear upgrade path

### Negative
- Two implementations to maintain
- Need to ensure consistency between them
- Redis adds operational complexity

### Risks
- Memory store could cause data loss in production if misconfigured
- Need good monitoring to detect memory pressure

## Migration Path

When transitioning from memory to Redis:

```bash
# Export current games (if any)
agent-checkers export > games.json

# Start with Redis store
agent-checkers server --store=redis

# Import games (if needed)
agent-checkers import < games.json
```

## Implementation Notes

1. **Phase 1:** Implement `MemoryStore` with full test coverage
2. **Phase 2:** Implement `RedisStore` with same tests
3. **Phase 3:** Add configuration flag to select store
4. **Phase 4:** Add monitoring (memory usage, Redis connection pool)

## References

- [Redis Data Modeling](https://redis.com/redis-enterprise/technology/redis-data-modeling/)
- [Go Redis Client](https://github.com/redis/go-redis)