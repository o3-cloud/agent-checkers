# agent-checkers

A checkers game designed for human-AI interaction with multiple interface layers.

## Overview

Agent-checkers is a checkers game where humans and AI agents can play against each other through multiple interfaces:

- **Web UI** — Visual board for human players
- **REST API** — HTTP endpoints for programmatic access
- **CLI** — Command-line interface for testing and development
- **MCP Server** — Model Context Protocol for AI agent integration (Claude Code, etc.)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     agent-checkers                           │
│                                                              │
│  ┌─────────┐    ┌─────────┐    ┌──────────┐    ┌────────┐ │
│  │ Web UI  │    │   CLI   │    │ REST API │    │   MCP  │ │
│  │ (Human) │    │ (Dev)   │    │  (HTTP)  │    │ Server │ │
│  └────┬────┘    └────┬────┘    └────┬─────┘    └───┬────┘ │
│       │              │              │              │       │
│       └──────────────┴──────────────┴──────────────┘       │
│                              │                              │
│                      ┌───────▼───────┐                      │
│                      │  Game Engine  │                      │
│                      │   (Go core)   │                      │
│                      └───────┬───────┘                      │
│                              │                              │
│                      ┌───────▼───────┐                      │
│                      │  Game State   │                      │
│                      │  (in-memory   │                      │
│                      │   or Redis)   │                      │
│                      └───────────────┘                      │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.24+
- Make

### Run Locally

```bash
# Clone the repository
git clone https://github.com/stackable-specs/agent-checkers.git
cd agent-checkers

# Run tests
make test

# Run all quality gates
make check

# Build
make build

# Run the server
./bin/agent-checkers server
```

### Play via CLI

```bash
# Create a new game
./bin/agent-checkers new --name "Alice"

# Join existing game
./bin/agent-checkers join <game-id> --name "Bob"

# View board
./bin/agent-checkers board

# Make a move
./bin/agent-checkers move e3 f4

# List valid moves
./bin/agent-checkers moves
```

## Development

### Project Structure

```
agent-checkers/
├── .github/           # CI/CD workflows
├── docs/              # Architecture Decision Records (ADRs)
├── internal/
│   └── app/
│       ├── game/      # Game engine core
│       ├── board/     # Board representation
│       ├── rules/     # Move validation
│       └── player/    # Player management
├── src/
│   ├── api/           # REST API handlers
│   ├── mcp/           # MCP server implementation
│   ├── web/           # Web UI templates
│   └── cli/           # CLI commands
├── tests/
│   ├── features/      # Gherkin feature files
│   └── integration/   # Integration tests
└── verify/            # Verification scripts
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# BDD tests
make test-bdd

# All tests with coverage
make test-coverage
```

### Quality Gates

- `gofmt` passes
- `go vet ./...` passes
- `golangci-lint run` passes
- `go test -race ./...` passes
- Coverage ≥80%

## Documentation

- [BDD Specification](../agent-checkers-spec.md) — Behavioral specs in Gherkin
- [Implementation Contract](../agent-checkers-contract.md) — Phased delivery plan
- [ADRs](./docs/) — Architecture Decision Records

## Stackable-Specs Methodology

This project follows the [stackable-specs](https://github.com/stackable-specs/specs) methodology with layered specifications:

| Layer | Question |
|-------|----------|
| `language` | What is it written in? (Go 1.24) |
| `platform` | Where does it run? (Linux/macOS/Windows) |
| `interface` | How is it invoked? (Web, API, CLI, MCP) |
| `presentation` | What does it output? (Board, JSON, ASCII) |
| `delivery` | How is it shipped? (Binary, Docker) |
| `observability` | How is it inspected? (Logs, metrics) |
| `security` | What must never break trust? (Input validation) |
| `practices` | How do we work? (TDD, code review) |
| `quality` | How do we enforce? (CI, linting, coverage) |

## License

MIT License - see [LICENSE](LICENSE) for details.