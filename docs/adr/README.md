# Architectural Decision Records

This directory holds MADR-format records for this stack.
Format and lifecycle rules: [`docs/specs/practices/madr.md`](../specs/practices/madr.md).

## Index

|| ADR | Title | Status |
|| --- | ----- | ------ |
|| [001](001-go-language.md) | Go as Primary Language | Accepted |
|| [002](002-docker-github-actions.md) | Docker and GitHub Actions for Delivery | Accepted |
|| [003](003-development-practices.md) | Development Practices (MADR, BDR, TDD, Git, Conventional Commits) | Accepted |
|| [004](004-unit-testing.md) | Unit Testing as Quality Gate | Accepted |
|| [005](005-dependency-management.md) | Dependency Management Policy | Accepted |
|| [006](006-game-engine-architecture.md) | Game Engine Architecture | Proposed |
|| [007](007-interface-layers.md) | Interface Layers (Web, API, CLI, MCP) | Proposed |
|| [008](008-state-management.md) | State Management (In-Memory vs Redis) | Proposed |

## Authoring

Copy [`000-template.md`](000-template.md), assign the next monotonic number, and open a PR.