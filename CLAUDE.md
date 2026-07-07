# Project: stack-base-go

Go base template following the stackable-specs methodology.

## Architecture

Layered specifications where each spec answers exactly one question about the system:
- **Language** (go.md): Go idioms, formatting, module hygiene
- **Practices** (madr.md, bdr.md, tdd.md, git.md, conventional-commits.md): Workflow and decision records
- **Quality** (unit-testing.md): Test scope and coverage
- **Security** (dependency-management.md): Dependency policy
- **Delivery** (docker.md, github-actions.md): Container and CI/CD conventions

## Key Commands

| Command | Description | Spec Reference |
|---------|-------------|----------------|
| `make test` | Run tests with race detection | `go.md:12`, `unit-testing.md` |
| `make lint` | Run golangci-lint | `go.md:16` |
| `make fmt` | Format with gofmt + goimports | `go.md:2-3` |
| `make vet` | Run go vet | `go.md:15` |
| `make build` | Build binary | `docker.md` |
| `make check` | All quality gates | All specs |

## Quality Gates

- **Format**: `gofmt -d .` returns empty (spec: go.md:2)
- **Imports**: `goimports` groups correctly (spec: go.md:3)
- **Vet**: `go vet ./...` passes (spec: go.md:15)
- **Lint**: `golangci-lint run` passes (spec: go.md:16)
- **Race**: `go test -race ./...` passes (spec: go.md:12)
- **Coverage**: ≥80% line coverage (spec: unit-testing.md)

## Specs

All specs are in `docs/specs/`. Each file defines rules for a specific layer:

```
docs/specs/
├── delivery/
│   ├── docker.md
│   └── github-actions.md
├── language/
│   └── go.md
├── practices/
│   ├── bdr.md
│   ├── conventional-commits.md
│   ├── git.md
│   ├── madr.md
│   └── tdd.md
├── quality/
│   └── unit-testing.md
└── security/
    └── dependency-management.md
```

## Workflow

1. Check relevant specs before making changes
2. Write tests first (TDD: `practices/tdd.md`)
3. Run quality gates before committing (`make check`)
4. Reference specs in commit messages and PR descriptions

## Spec Traceability

Map spec rules to enforcement:

| Spec Rule | Enforcement |
|-----------|-------------|
| `go.md:2` | `gofmt` in pre-commit + CI |
| `go.md:3` | `goimports` in pre-commit + CI |
| `go.md:12` | `go test -race` in CI |
| `go.md:15` | `go vet ./...` in pre-commit + CI |
| `go.md:16` | `staticcheck` via golangci-lint |
| `go.md:17` | Git track `go.mod` + `go.sum` |
| `unit-testing.md` | `go test -coverprofile` with threshold |
| `conventional-commits.md` | commitlint in pre-commit + CI |

## Adding a Spec

When adding a new spec:
1. Copy from [stackable-specs/specs](https://github.com/stackable-specs/specs)
2. Add references in CLAUDE.md/AGENTS.md
3. Add quality gates that enforce the spec's rules
4. Create ADR if it's a significant decision