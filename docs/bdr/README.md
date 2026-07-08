# Behavior Decision Records

Behavior contracts for this stack.
Each BDR captures a single capability the system agrees to uphold,
with acceptance criteria a black-box observer can confirm.

## Index

| BDR | Title | Status |
| --- | ----- | ------ |
| [001](001-player-registration.md) | Player Registration | Accepted |
| [002](002-move-validation.md) | Move Validation | Accepted |
| [003](003-game-state-management.md) | Game State Management | Accepted |
| [004](004-win-lose-draw.md) | Win/Lose/Draw Conditions | Accepted |
| [005](005-ai-agent-integration.md) | AI Agent Integration via MCP | Accepted |
| [006](006-web-ui-board.md) | Web UI Board Visualization | Accepted |
| [007](007-real-time-updates.md) | Real-Time Game Updates | Accepted |
| [008](008-cli-interface.md) | Command-Line Interface | Accepted |
| [009](009-concurrent-games.md) | Concurrent Games | Accepted |

## Summary

### Core Game Behaviors

- **BDR-001**: Player registration for humans and AI agents
- **BDR-002**: Move validation per American checkers rules
- **BDR-003**: Game state persistence and querying
- **BDR-004**: Win/lose/draw condition detection

### Interface Behaviors

- **BDR-005**: AI agent tools via MCP (Model Context Protocol)
- **BDR-006**: Web UI board visualization and interaction
- **BDR-007**: Real-time updates via WebSocket
- **BDR-008**: Command-line interface for terminals

### System Behaviors

- **BDR-009**: Concurrent game support and isolation

## Authoring

Copy [`000-template.md`](000-template.md), assign the next monotonic number, and open a PR.