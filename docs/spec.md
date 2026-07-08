# Agent-Checkers: Behavioral Specification

A checkers game designed for human-AI interaction with multiple interface layers.

---

## Feature: Game Registration

### Scenario: Two players register for a new game
```gherkin
Feature: Game Registration
  As a player (human or AI agent)
  I want to register for a checkers game
  So that I can play against another player

  Background:
    Given the game server is running
    And the game state is empty

  Scenario: Human registers via Web UI
    Given the Web UI is available at "/"
    When I navigate to the Web UI
    And I enter player name "Alice"
    And I select color preference "red"
    And I click "Join Game"
    Then I should see a waiting message "Waiting for opponent..."
    And my player ID should be stored in session

  Scenario: AI agent registers via MCP server
    Given the MCP server is available
    When the agent calls "register_player" with:
      | name   | "Claude"      |
      | type   | "ai"          |
    Then the response should contain a player_id
    And the response status should be "waiting_for_opponent"

  Scenario: Second player joins and game starts
    Given player "Alice" is waiting for an opponent
    When player "Claude" registers via MCP
    Then a new game should be created
    And both players should receive a game_id
    And the game state should show:
      | turn    | "red"    |
      | status  | "active" |

  Scenario: AI vs AI game registration
    Given two AI agents are available
    When agent1 calls "register_player" with name "Agent1"
    And agent2 calls "register_player" with name "Agent2"
    Then a new game should be created
    And both agents should receive the game_id
```

---

## Feature: Game State Management

### Scenario: Game state is persisted and queryable
```gherkin
Feature: Game State Management
  As a player or game observer
  I want to query the current game state
  So that I can understand the board position and whose turn it is

  Background:
    Given a game exists with id "game-123"
    And players "Alice" (red) and "Claude" (black) are registered

  Scenario: Query game state via API
    When I request GET /api/v1/games/game-123/state
    Then the response should contain:
      | game_id       | "game-123"           |
      | current_turn  | "red"                |
      | board         | <initial_board_json> |
      | red_player    | "Alice"              |
      | black_player  | "Claude"             |
      | status        | "active"             |

  Scenario: Query game state via MCP
    Given the MCP server is available
    When the agent calls "get_game_state" with game_id "game-123"
    Then the response should contain:
      | board        | <8x8 array>        |
      | current_turn | "red"              |
      | valid_moves  | <list of moves>    |

  Scenario: Board state representation
    Given a standard checkers starting position
    Then the board should be an 8x8 grid
    And red pieces should occupy rows 0-2
    And black pieces should occupy rows 5-7
    And empty squares should be null or empty
    And each piece should have:
      | id       | unique identifier |
      | color    | "red" or "black"  |
      | is_king  | boolean           |
      | position | {row, col}        |
```

---

## Feature: Move Validation and Execution

### Scenario: Valid moves are accepted and executed
```gherkin
Feature: Move Validation and Execution
  As a player
  I want to make valid moves and have them rejected if invalid
  So that the game follows standard checkers rules

  Background:
    Given a game exists with id "game-123"
    And it is red's turn

  Scenario: Valid simple move
    Given a red piece is at position {2, 3}
    And the destination {3, 4} is empty
    When the player makes a move:
      | from | {2, 3} |
      | to   | {3, 4} |
    Then the move should be accepted
    And the piece should now be at {3, 4}
    And the turn should pass to black

  Scenario: Invalid move - wrong turn
    Given it is red's turn
    And a black piece is at position {5, 4}
    When player "Claude" attempts to move from {5, 4} to {4, 3}
    Then the move should be rejected
    And the error message should be "It is not your turn"
    And the game state should remain unchanged

  Scenario: Invalid move - wrong direction
    Given a red non-king piece is at position {2, 3}
    When player "Alice" attempts to move from {2, 3} to {1, 4}
    Then the move should be rejected
    And the error message should be "Invalid move direction"

  Scenario: Valid capture move
    Given a red piece is at position {2, 3}
    And a black piece is at position {3, 4}
    And position {4, 5} is empty
    When player "Alice" moves from {2, 3} to {4, 5}
    Then the move should be accepted
    And the black piece at {3, 4} should be removed
    And the turn should pass to black

  Scenario: Mandatory capture - must jump if available
    Given a red piece is at position {2, 3}
    And a black piece is at position {3, 4}
    And position {4, 5} is empty
    And a red piece is at position {1, 2}
    And position {2, 1} is empty
    When player "Alice" attempts a simple move from {1, 2} to {2, 1}
    Then the move should be rejected
    And the error message should be "A capture is available, you must capture"

  Scenario: Multi-jump sequence
    Given a red piece is at position {2, 1}
    And black pieces are at positions {3, 2} and {5, 4}
    And positions {4, 3} and {6, 5} are empty
    When player "Alice" captures first piece {3, 2}
    Then the game should detect another capture is available
    And the player must continue capturing
    And only after all captures is the turn passed

  Scenario: King promotion
    Given a red piece is at position {6, 3}
    And position {7, 4} is empty
    When player "Alice" moves from {6, 3} to {7, 4}
    Then the piece should be promoted to king
    And the piece should have is_king = true
    And the piece should now be able to move in both directions
```

---

## Feature: Win/Lose/Draw Conditions

### Scenario: Game ends when a player loses all pieces
```gherkin
Feature: Win/Lose/Draw Conditions
  As a player
  I want the game to correctly detect and announce game endings
  So that there is a clear winner

  Background:
    Given a game exists with id "game-123"

  Scenario: Win by capturing all opponent pieces
    Given red has 1 piece remaining
    And black has 0 pieces remaining
    When the game state is evaluated
    Then the status should be "completed"
    And the winner should be "red"
    And both players should receive a game_ended event

  Scenario: Win by blocking all opponent moves
    Given black has pieces but no valid moves
    And red has valid moves available
    When the game state is evaluated
    Then the status should be "completed"
    And the winner should be "red"

  Scenario: Draw by mutual agreement
    Given the game is in progress
    When player "Alice" requests a draw
    And player "Claude" accepts the draw
    Then the game status should be "draw"
    And both players should receive a game_ended event

  Scenario: AI agent resignation
    Given an AI agent is playing
    When the agent calls "resign" with game_id
    Then the game status should be "completed"
    And the opponent should be declared winner
    And both players should receive a game_ended event
```

---

## Feature: Human Plays via Web UI

### Scenario: Human interacts with checkers board through web interface
```gherkin
Feature: Web UI Interface
  As a human player
  I want to play checkers through a visual web interface
  So that I can see the board and make moves naturally

  Background:
    Given I am a registered player in game "game-123"
    And I am viewing the game board

  Scenario: View initial board
    When I load the game board
    Then I should see an 8x8 checkerboard
    And I should see my pieces highlighted
    And I should see opponent pieces
    And I should see whose turn it is
    And valid moves should be visually indicated

  Scenario: Select and move a piece
    Given it is my turn
    And a piece at {2, 3} has valid moves to {3, 4} and {3, 2}
    When I click on the piece at {2, 3}
    Then valid destination squares should be highlighted
    When I click on {3, 4}
    Then the move should be submitted
    And the board should update
    And I should see "Opponent's turn" message

  Scenario: Capture animation
    Given a capture is available
    When I execute the capture
    Then I should see the jumping animation
    And I should see the captured piece removed
    And if another capture is available, I should be prompted to continue

  Scenario: King promotion visual
    Given my piece reaches the opposite end
    Then the piece should visually change to indicate king status
    And a crown icon should appear on the piece

  Scenario: Real-time opponent moves
    Given it is the opponent's turn
    And the opponent makes a move via MCP
    Then my board should update in real-time
    And I should see the move animation
    And I should see "Your turn" message
```

---

## Feature: AI Agent Plays via MCP Server

### Scenario: AI agent uses MCP tools to play
```gherkin
Feature: MCP Server Interface
  As an AI agent
  I want to interact with the game via MCP tools
  So that I can make moves programmatically

  Background:
    Given the MCP server is running
    And the agent has registered as a player

  Scenario: MCP tool - get_game_state
    Given a game exists with id "game-123"
    When the agent calls the MCP tool "get_game_state":
      | game_id | "game-123" |
    Then the tool should return:
      | board        | 8x8 array of piece objects |
      | current_turn | "red" or "black"            |
      | valid_moves  | list of {from, to} objects  |
      | game_status  | "active"                    |

  Scenario: MCP tool - make_move
    Given it is the agent's turn
    And the agent controls black pieces
    When the agent calls the MCP tool "make_move":
      | game_id | "game-123"      |
      | from    | {row: 5, col: 2} |
      | to      | {row: 4, col: 3} |
    Then the tool should return:
      | success    | true              |
      | new_state  | updated board     |
      | next_turn  | "red"             |

  Scenario: MCP tool - get_valid_moves
    Given it is the agent's turn
    When the agent calls the MCP tool "get_valid_moves":
      | game_id | "game-123" |
    Then the tool should return a list of valid moves:
      | piece_id | from_position | valid_destinations |
      | "r1"     | {2, 3}        | [{3, 4}, {3, 2}]   |

  Scenario: MCP tool - offer_draw
    Given the game is in progress
    When the agent calls the MCP tool "offer_draw":
      | game_id | "game-123" |
    Then the tool should return:
      | status    | "draw_offered" |
      | opponent_notified | true |

  Scenario: MCP tool - resign
    Given the agent wants to end the game
    When the agent calls the MCP tool "resign":
      | game_id | "game-123" |
    Then the tool should return:
      | status   | "completed" |
      | winner   | opponent_id |
      | reason   | "resignation" |
```

---

## Feature: AI Skill for Playing Checkers

### Scenario: Agent receives skill with game rules
```gherkin
Feature: Checkers Playing Skill
  As an AI agent
  I want to receive a skill that teaches me checkers rules
  So that I can play legally and strategically

  Background:
    Given an MCP server is available for agent-checkers
    And the agent has the "checkers-playing" skill loaded

  Scenario: Skill provides game rules
    Given the agent loads the checkers skill
    Then the skill should contain:
      | piece_movement  | How pieces move diagonally     |
      | king_movement   | How kings move in both directions |
      | capture_rules   | Mandatory jumps, multi-jumps   |
      | promotion       | King promotion at opposite end |
      | win_conditions  | No pieces or no moves = loss   |

  Scenario: Skill provides strategy hints
    Given the agent is analyzing a position
    When the agent consults the skill for strategy
    Then the skill should suggest:
      | control_center   | Occupy center squares early      |
      | protect_kings_row | Keep back row for king safety   |
      | force_trades     | Trade when ahead in pieces       |
      | create_forks     | Set up double-capture threats    |

  Scenario: Skill provides move format specification
    Given the agent needs to make a move
    Then the skill should specify:
      | move_format | {from: {row, col}, to: {row, col}} |
      | coordinate_system | 0-indexed, top-left origin |
      | row_direction | red moves +1, black moves -1 |
```

---

## Feature: REST API Interface

### Scenario: Third-party clients use REST API
```gherkin
Feature: REST API Interface
  As a third-party application developer
  I want to interact with the game via REST API
  So that I can build custom clients

  Background:
    Given the API server is running on port 8080
    And the API version is v1

  Scenario: Create a new game
    When I POST /api/v1/games with:
      | player1_name | "Alice" |
      | player1_type | "human" |
    Then the response should be 201 Created
    And the body should contain:
      | game_id      | <uuid>              |
      | player1_id   | <uuid>              |
      | status       | "waiting_for_player" |

  Scenario: Join an existing game
    Given a game exists with status "waiting_for_player"
    When I POST /api/v1/games/{game_id}/join with:
      | player2_name | "Claude" |
      | player2_type | "ai"     |
    Then the response should be 200 OK
    And the game status should be "active"

  Scenario: Get game state
    Given a game exists with id "game-123"
    When I GET /api/v1/games/game-123/state
    Then the response should be 200 OK
    And the body should contain the full game state

  Scenario: Make a move
    Given a game exists and it is player1's turn
    When I POST /api/v1/games/game-123/moves with:
      | from | {row: 2, col: 3} |
      | to   | {row: 3, col: 4} |
    Then the response should be 200 OK
    And the body should contain:
      | success   | true        |
      | new_state | <state>     |

  Scenario: Invalid move returns error
    Given it is player1's turn
    When I POST /api/v1/games/game-123/moves with an invalid move
    Then the response should be 400 Bad Request
    And the body should contain:
      | error     | "Invalid move: <reason>" |
      | valid_moves | <list of valid moves> |

  Scenario: WebSocket for real-time updates
    Given I am connected to WebSocket /api/v1/games/game-123/ws
    When a move is made by either player
    Then I should receive a message:
      | event     | "move_made" |
      | move      | {from, to}  |
      | new_state | <state>     |
      | next_turn | <player>    |
```

---

## Feature: CLI Interface

### Scenario: Developer plays via command line
```gherkin
Feature: CLI Interface
  As a developer or tester
  I want to interact with the game via CLI
  So that I can quickly test or debug

  Background:
    Given the CLI binary is installed as "agent-checkers"

  Scenario: Start a new game
    When I run `agent-checkers new --name "Alice"`
    Then the output should contain:
      | Game created: game-123 |
      | Player registered: Alice (red) |
      | Waiting for opponent... |
    And a session should be started

  Scenario: Join a game
    Given a game exists with id "game-123"
    When I run `agent-checkers join game-123 --name "Bob"`
    Then the output should contain:
      | Joined game: game-123 |
      | You are playing as black |
      | Game started! |

  Scenario: View board
    Given I am in a game
    When I run `agent-checkers board`
    Then the output should display an ASCII board:
      ```
        0   1   2   3   4   5   6   7
      +---+---+---+---+---+---+---+---+
    0 |   | ● |   | ● |   | ● |   | ● |
      +---+---+---+---+---+---+---+---+
    1 | ● |   | ● |   | ● |   | ● |   |
      +---+---+---+---+---+---+---+---+
    ...
      ```
    And current turn should be indicated

  Scenario: Make a move
    Given it is my turn
    When I run `agent-checkers move e3 f4`
    Then the output should contain:
      | Move accepted |
    And the updated board should be displayed

  Scenario: List valid moves
    Given it is my turn
    When I run `agent-checkers moves`
    Then the output should list all valid moves:
      | Piece at a3: can move to b4 |
      | Piece at c3: can move to b4 or d4 |

  Scenario: Play against AI
    When I run `agent-checkers new --name "Alice" --ai "Claude"`
    Then a game should start with AI opponent
    And the AI should make moves automatically
```

---

## Feature: Game History and Replay

### Scenario: Games can be reviewed
```gherkin
Feature: Game History and Replay
  As a player or analyst
  I want to review game history
  So that I can learn from past games

  Background:
    Given a completed game exists with id "game-123"

  Scenario: Get game history
    When I request GET /api/v1/games/game-123/history
    Then the response should contain:
      | game_id   | "game-123" |
      | moves     | <list of all moves> |
      | players   | ["Alice", "Claude"] |
      | winner    | "Alice" |
      | duration  | "4m 32s" |

  Scenario: Replay a game move by move
    Given the game history is loaded
    When I request GET /api/v1/games/game-123/replay
    Then I should be able to step through moves:
      | step  | query move number |
      | back  | previous position  |
      | forward | next position    |
      | reset | return to start    |

  Scenario: Export game in standard notation
    When I request GET /api/v1/games/game-123/export?format=pgn
    Then the response should contain a PGN-like notation:
      | [White "Alice"] |
      | [Black "Claude"] |
      | [Result "1-0"] |
      | 1. a3-b4 c5-d4 ... |
```

---

## Feature: Multiple Concurrent Games

### Scenario: Server handles multiple games simultaneously
```gherkin
Feature: Concurrent Games
  As the game server
  I want to support multiple concurrent games
  So that many players can use the system

  Scenario: Multiple independent games
    Given 100 games are in progress simultaneously
    When I query any individual game state
    Then each game should return its correct state
    And moves in one game should not affect others

  Scenario: Player in multiple games
    Given player "Alice" is in game-123 and game-456
    When "Alice" makes a move in game-123
    Then game-456 should be unaffected
    And "Alice" should be able to switch between games

  Scenario: Rate limiting
    Given an AI agent is making moves rapidly
    When the agent exceeds 10 moves per second
    Then the server should return 429 Too Many Requests
    And the game state should remain consistent
```

---

## Feature: Persistence and Recovery

### Scenario: Game state survives server restart
```gherkin
Feature: Game Persistence
  As a player
  I want my game to be saved and recoverable
  So that I don't lose progress if interrupted

  Scenario: Auto-save after each move
    Given a game is in progress
    When a move is made
    Then the game state should be persisted to storage
    And the move should be appended to the move log

  Scenario: Reconnect to in-progress game
    Given a player loses connection during their game
    When the player reconnects
    Then they should see the current game state
    And they should be able to continue playing

  Scenario: Resume after server restart
    Given the server restarts
    And games were persisted before restart
    When the server comes back online
    Then all in-progress games should be restored
    And players should be able to reconnect

  Scenario: Game expiration
    Given a game has been inactive for 30 days
    When the cleanup job runs
    Then the game should be archived
    And the archive should be queryable by game_id
```

---

## Non-Functional Requirements

### Performance
```gherkin
Feature: Performance Requirements
  As the game server
  I want to respond quickly to all requests
  So that gameplay feels responsive

  Scenario: Move latency
    Given a player makes a move
    Then the response should be received within 100ms
    And the updated state should be broadcast within 50ms

  Scenario: Concurrent connections
    Given 1000 WebSocket connections
    When a move is broadcast
    Then all clients should receive the update within 200ms

  Scenario: MCP tool response
    Given an AI agent calls an MCP tool
    Then the tool should respond within 50ms for read operations
    And within 100ms for write operations
```

### Security
```gherkin
Feature: Security
  As a game administrator
  I want to prevent cheating and abuse
  So that games are fair

  Scenario: Player authentication
    Given a player wants to make a move
    When they submit a move without a valid session
    Then the request should be rejected with 401 Unauthorized

  Scenario: Move authorization
    Given it is player2's turn
    When player1 attempts to make a move
    Then the request should be rejected with 403 Forbidden

  Scenario: Rate limiting per player
    Given a player exceeds rate limits
    When they attempt another request
    Then they should receive 429 Too Many Requests
    And a retry-after header should be provided
```

---

## Implementation Notes

### Technology Stack (from stack-base-go)
- **Language**: Go (Golang)
- **Architecture**: Clean architecture (internal/app)
- **Interfaces**: 
  - Web UI (HTML/JS/WebSocket)
  - REST API (HTTP)
  - CLI (Cobra)
  - MCP Server (JSON-RPC over stdio)
- **State Management**: In-memory with optional Redis persistence
- **Testing**: Gherkin specs → Go tests via godog or similar

### Directory Structure
```
agent-checkers/
├── .github/           # CI/CD workflows
├── docs/              # ADRs and documentation
├── internal/
│   └── app/
│       ├── game/      # Game engine core
│       ├── board/     # Board representation
│       ├── rules/     # Checkers rules validation
│       └── player/    # Player management
├── src/
│   ├── api/           # REST API handlers
│   ├── mcp/           # MCP server implementation
│   ├── web/           # Web UI templates and static
│   └── cli/           # CLI commands
├── tests/
│   ├── features/      # Gherkin feature files
│   └── integration/   # Integration tests
└── verify/            # Verification scripts
```

---

## Feature: OpenAPI Specification

### Scenario: Client fetches OpenAPI JSON spec
```gherkin
Feature: OpenAPI Specification
  As an API client (human developer or AI agent)
  I want to fetch a machine-readable OpenAPI specification
  So that I can discover available operations and schemas without reading source code

  Background:
    Given the game server is running on port 8080

  Scenario: Fetch OpenAPI JSON
    When I send a GET request to "/openapi.json"
    Then the response status should be 200
    And the response Content-Type should be "application/json"
    And the response body should be valid OpenAPI 3.1 JSON
    And the info.title should be "Agent Checkers API"
    And the info.version should match the application version

  Scenario: Fetch OpenAPI YAML
    When I send a GET request to "/openapi.yaml"
    Then the response status should be 200
    And the response Content-Type should be "application/yaml"
    And the response body should be valid OpenAPI 3.1 YAML

  Scenario: All endpoints are documented
    Given the OpenAPI spec is fetched from "/openapi.json"
    When the paths object is inspected
    Then it should include entries for:
      | method | path                          |
      | POST   | /api/v1/games                 |
      | POST   | /api/v1/games/{id}/join       |
      | GET    | /api/v1/games/{id}            |
      | DELETE | /api/v1/games/{id}            |
      | POST   | /api/v1/games/{id}/draw       |
      | POST   | /api/v1/games/{id}/moves      |
      | GET    | /api/v1/games/{id}/moves      |
      | GET    | /api/v1/games/{id}/valid-moves|
      | GET    | /health                      |

  Scenario: Reusable component schemas are defined
    Given the OpenAPI spec is fetched from "/openapi.json"
    When the components.schemas object is inspected
    Then it should include definitions for:
      | schema              |
      | CreateGameRequest   |
      | JoinGameRequest     |
      | MoveRequest         |
      | GameState           |
      | PlayerResponse      |
      | ErrorResponse       |
      | MoveResponse        |

  Scenario: Error responses are documented
    Given the OpenAPI spec is fetched from "/openapi.json"
    When the path "POST /api/v1/games/{id}/moves" is inspected
    Then it should document responses for:
      | status | description                     |
      | 200    | successful move                 |
      | 400    | invalid move, wrong turn        |
      | 404    | game not found                   |
      | 500    | internal server error           |

  Scenario: AI agent discovers API via OpenAPI
    Given an AI agent connects to the server for the first time
    When the agent fetches "GET /openapi.json"
    Then the agent can determine all available operations
    And the agent can construct valid requests from the schemas
    And the agent can interpret responses using the documented schemas

  Scenario: Static spec file exists in repository
    Given the repository is cloned
    When the docs/openapi.yaml file is inspected
    Then it should be valid OpenAPI 3.1 YAML
    And it should match the runtime spec served at /openapi.json
```

---

## Feature: List Active Games

### Scenario: List games via REST API
```gherkin
Feature: List Active Games
  As a player, AI agent, or observer
  I want to list all active and waiting games
  So that I can find a game to join, spectate, or manage

  Background:
    Given the game server is running on port 8080
    And there are 2 active games and 1 completed game in the store

  Scenario: List all active games via API
    When I send a GET request to "/api/v1/games"
    Then the response status should be 200
    And the response Content-Type should be "application/json"
    And the response body should be a JSON array with 2 entries
    And each entry should have game_id, status, current_turn, and players
    And the completed game should NOT be included

  Scenario: Filter by status via API
    Given there is 1 waiting game and 2 active games
    When I send a GET request to "/api/v1/games?status=waiting"
    Then the response should contain exactly 1 entry
    And that entry should have status "waiting"

  Scenario: Filter by player via API
    Given player "Alice" is in 2 games
    When I send a GET request to "/api/v1/games?player_id=alice_id"
    Then the response should contain 2 entries
    And each entry should list Alice as a player

  Scenario: Empty list via API
    Given no games exist in the store
    When I send a GET request to "/api/v1/games"
    Then the response status should be 200
    And the response body should be "[]"

  Scenario: List games via CLI
    Given there are 2 active games and 1 waiting game
    When I run "agent-checkers games"
    Then the output should be a table with 3 rows
    And each row should show game ID, status, red player, black player, and turn

  Scenario: Filter by status via CLI
    Given there is 1 waiting game and 2 active games
    When I run "agent-checkers games --status waiting"
    Then the output should show 1 row
    And that row should have status "waiting"

  Scenario: CLI JSON output
    Given there are 2 active games
    When I run "agent-checkers games --json"
    Then the output should be valid JSON array with 2 entries
    And each entry should have game_id, status, current_turn, and players

  Scenario: CLI empty list
    Given no games exist
    When I run "agent-checkers games"
    Then the output should print "No games found"

  Scenario: List games via MCP
    Given the MCP server is running
    And there are 2 active games and 1 waiting game
    When an AI agent calls "list_games"
    Then the response should contain a games array with 3 entries
    And each entry should have game_id, status, current_turn, and players

  Scenario: Filter via MCP
    Given there is 1 waiting game and 2 active games
    When an AI agent calls "list_games" with status="waiting"
    Then the games array should contain exactly 1 entry
    And that entry should have status "waiting"

  Scenario: AI agent finds a game to join
    Given an AI agent wants to play checkers
    And there is 1 waiting game with no opponent
    When the AI calls "list_games" with status="waiting"
    Then the AI should receive the waiting game's game_id
    And the AI can call "join_game" with that game_id to start playing
```