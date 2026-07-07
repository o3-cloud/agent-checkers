# Agent-Checkers: Implementation Contract

## Overview

**Project:** agent-checkers  
**Source Template:** stackable-specs/stack-base-go  
**Language:** Go  
**Interfaces:** Web UI, REST API, CLI, MCP Server  
**Database:** In-memory (Phase 1) → Redis (Phase 2)

---

## Phase 0: Project Scaffolding

**Goal:** Initialize project from stack-base-go template with project-specific configuration.

### Files Modified (from template)

| File | Changes |
|------|---------|
| `README.md` | Replace template README with agent-checkers description |
| `go.mod` | Update module name to `agent-checkers` |
| `.github/workflows/ci.yml` | Update project name references |
| `.cursorrules` | Add checkers-specific AI assistant rules |

### Files Added

| File | Purpose |
|------|---------|
| `docs/ADR-001-game-engine-architecture.md` | Decision: Core game engine structure |
| `docs/ADR-002-interface-layers.md` | Decision: Web/API/CLI/MCP interface design |
| `docs/ADR-003-state-management.md` | Decision: In-memory vs Redis persistence |

### Deliverables
- [ ] Project compiles (`go build ./...`)
- [ ] CI pipeline runs successfully
- [ ] README describes the project
- [ ] ADRs drafted for major decisions

---

## Phase 1: Core Game Engine

**Goal:** Implement checkers game logic with full rule validation.

### Files Added

| File | Purpose |
|------|---------|
| `internal/app/game/game.go` | Game struct, state management |
| `internal/app/game/game_test.go` | Unit tests for game logic |
| `internal/app/board/board.go` | Board representation (8x8 grid) |
| `internal/app/board/board_test.go` | Board tests |
| `internal/app/piece/piece.go` | Piece struct (color, king status, position) |
| `internal/app/rules/validator.go` | Move validation logic |
| `internal/app/rules/validator_test.go` | Validation tests |
| `internal/app/rules/captures.go` | Capture/jump logic |
| `internal/app/rules/captures_test.go` | Capture tests |
| `internal/app/player/player.go` | Player struct, registration |

### File Details

#### `internal/app/board/board.go`
```go
// Board represents an 8x8 checkers board
type Board struct {
    squares [8][8]*piece.Piece
}

// NewBoard creates a board with initial piece positions
func NewBoard() *Board

// GetPiece returns piece at position, nil if empty
func (b *Board) GetPiece(pos Position) *piece.Piece

// SetPiece places a piece at position
func (b *Board) SetPiece(pos Position, p *piece.Piece)

// RemovePiece removes piece from position
func (b *Board) RemovePiece(pos Position)
```

#### `internal/app/game/game.go`
```go
// Game represents a checkers game session
type Game struct {
    ID          string
    Board       *board.Board
    RedPlayer   *player.Player
    BlackPlayer *player.Player
    CurrentTurn piece.Color // red or black
    Status      Status      // waiting, active, completed, draw
    Moves       []Move
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// NewGame creates a new game with initial board state
func NewGame() *Game

// MakeMove attempts to execute a move, returns error if invalid
func (g *Game) MakeMove(playerID string, from, to board.Position) error

// GetValidMoves returns all valid moves for the current player
func (g *Game) GetValidMoves() []Move

// IsGameOver checks win/lose/draw conditions
func (g *Game) IsGameOver() (bool, GameResult)
```

#### `internal/app/rules/validator.go`
```go
// Validator checks move legality
type Validator struct{}

// ValidateMove returns nil if valid, error describing why if invalid
func (v *Validator) ValidateMove(g *game.Game, from, to board.Position) error

// GetValidMoves returns all legal moves for a piece
func (v *Validator) GetValidMoves(g *game.Game, pos board.Position) []board.Position
```

### Tests Added

| Test File | Coverage |
|-----------|----------|
| `game_test.go` | Game creation, turn management, win conditions |
| `board_test.go` | Piece placement, removal, position conversion |
| `validator_test.go` | Valid moves, invalid moves, edge cases |
| `captures_test.go` | Single jumps, multi-jumps, mandatory captures |

### Deliverables
- [ ] All tests pass (`go test ./internal/app/...`)
- [ ] 80%+ code coverage on game engine
- [ ] Game correctly validates all checkers rules:
  - [ ] Diagonal movement only
  - [ ] Forward movement for non-kings
  - [ ] Mandatory captures
  - [ ] Multi-jump sequences
  - [ ] King promotion
  - [ ] Win/lose/draw detection

---

## Phase 2: Player Management & Registration

**Goal:** Handle player registration, session management, and game matchmaking.

### Files Added

| File | Purpose |
|------|---------|
| `internal/app/lobby/lobby.go` | Lobby manages waiting players and matchmaking |
| `internal/app/lobby/lobby_test.go` | Lobby tests |
| `internal/app/session/session.go` | Player session management |
| `internal/app/session/session_test.go` | Session tests |
| `internal/app/store/memory.go` | In-memory game store |
| `internal/app/store/memory_test.go` | Store tests |

### File Details

#### `internal/app/lobby/lobby.go`
```go
// Lobby manages player matching
type Lobby struct {
    waitingPlayer *player.Player
    store         store.GameStore
    mu            sync.Mutex
}

// RegisterPlayer adds player to lobby, creates game if matched
func (l *Lobby) RegisterPlayer(name string, playerType player.Type) (*game.Game, *player.Player, error)

// JoinGame registers second player and starts game
func (l *Lobby) JoinGame(gameID string, name string, playerType player.Type) (*player.Player, error)
```

#### `internal/app/store/memory.go`
```go
// MemoryStore implements GameStore interface
type MemoryStore struct {
    games   map[string]*game.Game
    players map[string]*player.Player
    mu      sync.RWMutex
}

// SaveGame persists game state
func (m *MemoryStore) SaveGame(g *game.Game) error

// LoadGame retrieves game by ID
func (m *MemoryStore) LoadGame(id string) (*game.Game, error)

// SavePlayer persists player
func (m *MemoryStore) SavePlayer(p *player.Player) error
```

### Deliverables
- [ ] Players can register and receive unique IDs
- [ ] Lobby creates games when two players join
- [ ] Game state persists in memory between moves
- [ ] Sessions track player identities

---

## Phase 3: REST API Layer

**Goal:** Expose game operations via RESTful HTTP API.

### Files Added

| File | Purpose |
|------|---------|
| `src/api/server.go` | HTTP server setup and routing |
| `src/api/handlers/games.go` | Game CRUD handlers |
| `src/api/handlers/games_test.go` | Handler tests |
| `src/api/handlers/moves.go` | Move execution handlers |
| `src/api/handlers/moves_test.go` | Move handler tests |
| `src/api/middleware/auth.go` | Player authentication middleware |
| `src/api/middleware/cors.go` | CORS configuration |
| `src/api/responses/responses.go` | JSON response helpers |
| `src/api/routes/routes.go` | Route definitions |

### API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/api/v1/games` | Create new game |
| `POST` | `/api/v1/games/{id}/join` | Join existing game |
| `GET` | `/api/v1/games/{id}` | Get game state |
| `POST` | `/api/v1/games/{id}/moves` | Make a move |
| `GET` | `/api/v1/games/{id}/moves` | Get move history |
| `DELETE` | `/api/v1/games/{id}` | Resign game |
| `POST` | `/api/v1/games/{id}/draw` | Offer/accept draw |

### Deliverables
- [ ] API server starts on configured port
- [ ] All endpoints return correct HTTP status codes
- [ ] Error responses include helpful messages
- [ ] Integration tests cover all endpoints

---

## Phase 4: WebSocket Real-Time Updates

**Goal:** Enable real-time board updates for Web UI.

### Files Added

| File | Purpose |
|------|---------|
| `src/api/websocket/hub.go` | WebSocket connection hub |
| `src/api/websocket/client.go` | Individual client connection |
| `src/api/websocket/protocol.go` | Message types and serialization |
| `src/api/websocket/websocket_test.go` | WebSocket tests |

### WebSocket Events

| Event | Direction | Payload |
|-------|-----------|---------|
| `game_started` | Server → Client | `{game_id, players}` |
| `move_made` | Server → Client | `{from, to, captured}` |
| `turn_changed` | Server → Client | `{player_id}` |
| `game_ended` | Server → Client | `{winner, reason}` |

### Deliverables
- [ ] WebSocket endpoint at `/api/v1/games/{id}/ws`
- [ ] Clients receive real-time move notifications
- [ ] Reconnection preserves game state
- [ ] Multiple clients can observe same game

---

## Phase 5: Web UI

**Goal:** Visual checkers board for human players.

### Files Added

| File | Purpose |
|------|---------|
| `src/web/static/css/board.css` | Board styling |
| `src/web/static/js/game.js` | Game logic, move handling |
| `src/web/static/js/websocket.js` | WebSocket client |
| `src/web/templates/board.html` | Main game board template |
| `src/web/templates/lobby.html` | Registration/waiting room |
| `src/web/handlers/pages.go` | Page serving handlers |

### UI Features

| Feature | Description |
|---------|-------------|
| Board rendering | 8x8 grid with pieces |
| Piece selection | Click to select, show valid moves |
| Move execution | Click destination to move |
| Turn indicator | Visual cue for whose turn |
| Game status | Win/lose/draw messages |
| Play again | Restart with same opponent |

### Deliverables
- [ ] Board displays correctly in browser
- [ ] Pieces can be selected and moved
- [ ] Valid moves are highlighted
- [ ] Moves update board in real-time
- [ ] Mobile responsive design

---

## Phase 6: CLI Interface

**Goal:** Command-line interface for testing and development.

### Files Added

| File | Purpose |
|------|---------|
| `src/cli/main.go` | CLI entry point |
| `src/cli/commands/root.go` | Root command |
| `src/cli/commands/new.go` | `new` - create game |
| `src/cli/commands/join.go` | `join` - join game |
| `src/cli/commands/board.go` | `board` - display board |
| `src/cli/commands/move.go` | `move` - make a move |
| `src/cli/commands/moves.go` | `moves` - list valid moves |
| `src/cli/commands/watch.go` | `watch` - observe game |

### CLI Commands

```
agent-checkers new --name "Alice"              # Create new game
agent-checkers join <game-id> --name "Bob"     # Join existing game
agent-checkers board                            # Display current board
agent-checkers move <from> <to>                # Make move (e.g., e3 f4)
agent-checkers moves                            # List valid moves
agent-checkers watch                            # Watch game with live updates
agent-checkers vs --ai "Claude"                 # Play against AI
```

### Deliverables
- [ ] All commands work with local or remote server
- [ ] Board displays as ASCII art
- [ ] Moves validated before sending to server
- [ ] Exit codes indicate success/failure

---

## Phase 7: MCP Server

**Goal:** Model Context Protocol server for AI agent integration.

### Files Added

| File | Purpose |
|------|---------|
| `src/mcp/server.go` | MCP server implementation |
| `src/mcp/tools/register.go` | `register_player` tool |
| `src/mcp/tools/game_state.go` | `get_game_state` tool |
| `src/mcp/tools/make_move.go` | `make_move` tool |
| `src/mcp/tools/valid_moves.go` | `get_valid_moves` tool |
| `src/mcp/tools/resign.go` | `resign` tool |
| `src/mcp/tools/draw.go` | `offer_draw`, `accept_draw` tools |
| `src/mcp/protocol/handler.go` | JSON-RPC request handling |
| `docs/skills/checkers-playing.md` | Skill file for agents |

### MCP Tools

| Tool | Parameters | Returns |
|------|-----------|---------|
| `register_player` | `name`, `type` | `player_id`, `game_id` |
| `get_game_state` | `game_id` | board, turn, status |
| `make_move` | `game_id`, `from`, `to` | success, new_state |
| `get_valid_moves` | `game_id` | list of valid moves |
| `resign` | `game_id` | winner, reason |
| `offer_draw` | `game_id` | status |

### Deliverables
- [ ] MCP server runs via stdio (JSON-RPC)
- [ ] All tools return properly formatted responses
- [ ] Error handling with descriptive messages
- [ ] Skill file teaches checkers rules to agents

---

## Phase 8: BDD Test Suite

**Goal:** Automated acceptance tests from Gherkin specs.

### Files Added

| File | Purpose |
|------|---------|
| `tests/features/game_registration.feature` | Registration scenarios |
| `tests/features/move_validation.feature` | Move validation scenarios |
| `tests/features/win_conditions.feature` | Win/lose/draw scenarios |
| `tests/features/mcp_interface.feature` | MCP tool scenarios |
| `tests/features/api_interface.feature` | REST API scenarios |
| `tests/step_definitions/game_steps.go` | Godog step implementations |
| `tests/step_definitions/move_steps.go` | Move step definitions |
| `tests/step_definitions/api_steps.go` | API step definitions |

### Example Step Definition

```go
// tests/step_definitions/game_steps.go
func (s *GameSteps) iRegisterAsAPlayer(name string) error {
    resp, err := s.api.RegisterPlayer(name, "human")
    if err != nil {
        return err
    }
    s.playerID = resp.PlayerID
    return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    steps := &GameSteps{}
    ctx.Step(`^I register as a player "([^"]*)"$`, steps.iRegisterAsAPlayer)
    ctx.Step(`^the response should contain a player_id$`, steps.responseContainsPlayerID)
}
```

### Deliverables
- [ ] All Gherkin scenarios have step definitions
- [ ] `make test-bdd` runs all acceptance tests
- [ ] CI pipeline includes BDD tests
- [ ] Coverage report generated

---

## Phase 9: Redis Persistence (Optional)

**Goal:** Replace in-memory store with Redis for production use.

### Files Added

| File | Purpose |
|------|---------|
| `internal/app/store/redis.go` | Redis-backed game store |
| `internal/app/store/redis_test.go` | Redis store tests |
| `deploy/docker-compose.yml` | Local Redis container |
| `deploy/redis.conf` | Redis configuration |

### Files Modified

| File | Changes |
|------|---------|
| `internal/app/store/store.go` | Add interface definition |
| `src/api/server.go` | Add store selection logic |
| `docs/ADR-003-state-management.md` | Update with Redis decision |

### Deliverables
- [ ] Redis store implements same interface as memory store
- [ ] Configuration flag to select store type
- [ ] Docker compose file for local Redis
- [ ] Game state survives server restart

---

## Phase 10: AI Integration & Skill Publishing

**Goal:** Package skill file for easy agent adoption.

### Files Added

| File | Purpose |
|------|---------|
| `docs/skills/checkers-playing.md` | Agent skill definition |
| `docs/skills/examples/basic-game.md` | Example: Play a basic game |
| `docs/skills/examples/strategy.md` | Strategy hints for agents |
| `scripts/publish-skill.sh` | Publish skill to skill registry |

### Skill Content

```markdown
# Checkers Playing Skill

## Tools Available
- `register_player` - Join a game
- `get_game_state` - View current board
- `get_valid_moves` - List legal moves
- `make_move` - Execute a move

## Game Rules
1. Red moves first, pieces move diagonally forward
2. Non-king pieces can only move forward
3. Captures are mandatory - if you can jump, you must
4. Multi-jumps continue until no more captures
5. Reaching opposite end promotes piece to king
6. Kings can move diagonally in any direction
7. Game ends when a player loses all pieces or cannot move

## Strategy Hints
- Control the center early
- Protect your king row
- Trade pieces when ahead
- Force opponent into bad positions

## Move Format
{
  "from": {"row": 2, "col": 3},
  "to": {"row": 3, "col": 4}
}

Coordinates are 0-indexed, row 0 is black's home row.
```

### Deliverables
- [ ] Skill file documents all MCP tools
- [ ] Example prompts show how to play
- [ ] Agent can successfully join and play a game
- [ ] Documentation published to project wiki

---

## Summary: File Count by Phase

| Phase | Added | Modified | Total |
|------|-------|----------|-------|
| 0: Scaffolding | 3 | 4 | 7 |
| 1: Game Engine | 10 | 0 | 10 |
| 2: Player Mgmt | 6 | 0 | 6 |
| 3: REST API | 10 | 0 | 10 |
| 4: WebSocket | 4 | 1 | 5 |
| 5: Web UI | 6 | 0 | 6 |
| 6: CLI | 8 | 0 | 8 |
| 7: MCP Server | 10 | 0 | 10 |
| 8: BDD Tests | 8 | 0 | 8 |
| 9: Redis | 3 | 3 | 6 |
| 10: AI Integration | 4 | 0 | 4 |
| **Total** | **72** | **8** | **80** |

---

## Testing Strategy by Phase

| Phase | Unit Tests | Integration Tests | BDD Tests |
|-------|------------|-------------------|-----------|
| 1 | ✅ Required | ❌ | ❌ |
| 2 | ✅ Required | ✅ Required | ❌ |
| 3 | ✅ Required | ✅ Required | ❌ |
| 4 | ✅ Required | ✅ Required | ❌ |
| 5 | ❌ | ✅ Required | ✅ Required |
| 6 | ✅ Required | ❌ | ✅ Required |
| 7 | ✅ Required | ✅ Required | ✅ Required |
| 8 | ❌ | ❌ | ✅ Required |
| 9 | ✅ Required | ✅ Required | ✅ Existing |

---

## Dependency Order

```
Phase 0 ───> Phase 1 ───> Phase 2 ───> Phase 3 ───> Phase 4
                │                          │
                │                          └──> Phase 5
                │
                └──> Phase 6
                │
                └──> Phase 7 ───> Phase 10
                │
                └──> Phase 8
                │
                └──> Phase 9 (optional)
```

---

## Acceptance Criteria (Per Phase)

Each phase is considered **complete** when:
1. All listed files are created/modified
2. All tests pass (`go test ./...`)
3. Deliverables checklist is satisfied
4. Code review approved
5. Documentation updated
6. Demo to stakeholder

---

## Estimated Timeline

| Phase | Estimated Time | Dependencies |
|-------|----------------|--------------|
| 0 | 1 hour | None |
| 1 | 8 hours | Phase 0 |
| 2 | 4 hours | Phase 1 |
| 3 | 6 hours | Phase 2 |
| 4 | 4 hours | Phase 3 |
| 5 | 8 hours | Phase 4 |
| 6 | 4 hours | Phase 2 |
| 7 | 6 hours | Phase 2 |
| 8 | 6 hours | Phases 1-7 |
| 9 | 4 hours | Phase 2 |
| 10 | 2 hours | Phase 7 |
| **Total** | **~49 hours** | |

---

## Next Action

**Ready to begin Phase 0?**

Confirm:
1. Project location: `~/projects/agent-checkers` (or specify)
2. GitHub repo: `stackable-specs/agent-checkers` (or specify)
3. Start implementation now?

Once confirmed, I'll:
1. Clone `stack-base-go` as template
2. Update project configuration
3. Create initial ADRs
4. Set up CI pipeline