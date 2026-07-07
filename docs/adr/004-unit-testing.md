# ADR-004: Unit Testing as Quality Gate

## Status

Accepted

## Context and Problem Statement

We need a baseline quality gate that ensures code correctness before merging. What testing discipline should we adopt for the stack-base-go template?

## Decision Drivers

- **Coverage**: Measurable test coverage with clear thresholds
- **Speed**: Fast feedback loop for developers
- **Concurrency**: Go's concurrency model requires race detection
- **Simplicity**: Minimal test infrastructure overhead
- **AI-agent-friendly**: Tests should be easy for AI agents to generate and understand

## Considered Options

- **Option A**: Unit testing with race detection + coverage threshold
- **Option B**: Integration testing only
- **Option C**: Property-based testing only
- **Option D**: No enforced testing

## Decision Outcome

Chosen option: **Unit testing with race detection + coverage threshold**, because:

1. **Fast feedback**: Unit tests run in milliseconds, enabling rapid iteration
2. **Race detection**: `go test -race` catches concurrency bugs early
3. **Coverage visibility**: `go test -coverprofile` generates measurable metrics
4. **TDD-aligned**: Unit tests are the foundation of test-driven development
5. **Table-driven tests**: Go idiom for clear, maintainable test cases

## Consequences

### Positive

- Fast CI pipeline (unit tests complete in seconds)
- Race detector catches concurrency bugs early
- Coverage metrics are actionable and trackable over time
- Table-driven tests are easy for AI agents to generate
- Tests document expected behavior

### Negative

- Unit tests don't catch integration issues
- Coverage threshold may need adjustment per project
- Race detector adds overhead to test execution

## Compliance

This decision is enforced by:

| Spec | Enforcement |
|------|-------------|
| `quality/unit-testing.md` | `go test -race -coverprofile` in CI |
| `quality/unit-testing.md` | Coverage threshold (default ≥80%) |
| `language/go.md:13-14` | Tests in `_test.go` files, table-driven with `t.Run` |

## References

- **spec**: `docs/specs/quality/unit-testing.md`
- **spec**: `docs/specs/language/go.md` (rules 12-14)
- **external**: https://go.dev/doc/tutorial/add-a-test