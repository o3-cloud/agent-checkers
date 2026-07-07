# ADR-001: Go as Primary Language

## Status

Accepted

## Context and Problem Statement

We need a programming language for building reliable, maintainable services. The project should be AI-agent-friendly, with clear idioms and strong tooling that can be enforced automatically. What language should we choose for the stack-base-go template?

## Decision Drivers

- **Readability**: Code should be easy for humans and AI agents to understand
- **Tooling**: Strong built-in tooling for formatting, linting, and testing
- **Concurrency**: First-class support for concurrent programming
- **Compilation**: Fast compile times and single-binary deployment
- **Ecosystem**: Mature standard library and tooling ecosystem
- **Simplicity**: Minimal language complexity, opinionated defaults

## Considered Options

- **Option A**: Go
- **Option B**: Python
- **Option C**: TypeScript
- **Option D**: Rust

## Decision Outcome

Chosen option: **Go**, because:

1. **Opinionated tooling baked in**: `gofmt`, `go vet`, `go test` eliminate bikeshedding about style and basic correctness
2. **AI-agent-friendly**: Uniform style across all codebases means AI agents can read and write idiomatic Go without guessing conventions
3. **Fast compilation**: Enables rapid iteration and CI feedback loops
4. **Single binary deployment**: Simplifies containerization and delivery
5. **Strong concurrency primitives**: Goroutines and channels with race detector built in
6. **Minimal language complexity**: Fewer footguns, easier for AI agents to generate correct code
7. **Explicit error handling**: Forces explicit handling at every call site, reducing hidden bugs

## Consequences

### Positive

- Consistent code style across all contributors (human and AI)
- Built-in race detection in `go test -race`
- Fast CI pipelines due to quick compilation
- Single-binary output simplifies Docker images
- Strong static analysis ecosystem (golangci-lint, staticcheck)
- Explicit error handling reduces silent failures

### Negative

- More verbose than Python for simple scripts
- No generics until Go 1.18 (now available)
- Manual error handling can feel repetitive
- Stricter type system than Python, less flexible than TypeScript

## Compliance

This decision is enforced by:

| Spec | Enforcement |
|------|-------------|
| `language/go.md` | `gofmt`, `goimports` in pre-commit + CI |
| `language/go.md:12` | `go test -race` in CI |
| `language/go.md:15` | `go vet ./...` in pre-commit + CI |
| `language/go.md:16` | `golangci-lint` with staticcheck enabled |

## References

- **spec**: `docs/specs/language/go.md`
- **external**: https://go.dev/doc/effective_go
- **external**: https://google.github.io/styleguide/go/