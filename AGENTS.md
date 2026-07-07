# Project: stack-base-go

Go base template following the stackable-specs methodology.

## Architecture

Layered specifications where each spec answers exactly one question about the system. See CLAUDE.md for full details.

## Key Commands

| Command | Description | Spec Reference |
|---------|-------------|----------------|
| `make test` | Run tests with race detection | `go.md:12` |
| `make lint` | Run golangci-lint | `go.md:16` |
| `make fmt` | Format with gofmt + goimports | `go.md:2-3` |
| `make vet` | Run go vet | `go.md:15` |
| `make check` | All quality gates | All specs |

## Quality Gates

- `gofmt` passes
- `go vet ./...` passes
- `golangci-lint run` passes
- `go test -race ./...` passes
- Coverage ≥80%

## Specs

All specs are in `docs/specs/`. Check relevant specs before making changes.

## Workflow

1. Check relevant specs in `docs/specs/`
2. Write tests first (TDD)
3. Run `make check` before committing
4. Reference specs in commit messages