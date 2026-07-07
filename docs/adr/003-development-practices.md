# ADR-003: Development Practices (MADR, BDR, TDD, Git, Conventional Commits)

## Status

Accepted

## Context and Problem Statement

A project needs consistent development practices to ensure maintainability, traceability, and quality over time. What practices should we adopt for decision documentation, testing discipline, and version control workflow?

## Decision Drivers

- **Traceability**: Every decision should be documented and discoverable
- **Quality**: Tests should drive development, not follow it
- **Collaboration**: Clear workflow for branching, merging, and reviewing
- **Automation**: Practices should be enforceable via CI/CD
- **AI-agent-friendly**: Structured formats that AI agents can read, write, and validate

## Considered Options

- **Option A**: MADR + BDR + TDD + Conventional Commits + Trunk-based development
- **Option B**: RFCs + Behavior-driven development + Gitflow
- **Option C**: ADR (Y-structure) + No formal BDR + Manual testing + Free-form commits

## Decision Outcome

Chosen option: **MADR + BDR + TDD + Conventional Commits + Git workflow**, because:

1. **MADR (Micha Koller ADR format)**: Lightweight, structured, widely adopted for documenting architectural decisions
2. **BDR (Behavior Decision Records)**: Captures runtime/business behavior decisions separately from architecture
3. **TDD (Test-Driven Development)**: Red-green-refactor cycle ensures tests exist before code
4. **Conventional Commits**: Machine-readable commit messages enable automated changelogs and versioning
5. **Git workflow**: Clear branch/merge rules with spec-referenced commits

## Consequences

### Positive

- Every decision is discoverable in `docs/adr/` or `docs/bdr/`
- Tests document intent; code matches tests
- Automated changelog generation from commit messages
- Clear git history with spec references
- AI agents can understand and generate compliant commits

### Negative

- Overhead of writing ADRs for every decision
- TDD requires discipline, especially for beginners
- Conventional commits need commitlint enforcement

## Compliance

This decision is enforced by:

| Spec | Enforcement |
|------|-------------|
| `practices/madr.md` | ADR template in `docs/adr/000-template.md` |
| `practices/bdr.md` | BDR template in `docs/bdr/000-template.md` |
| `practices/tdd.md` | Tests must pass before merge (CI gate) |
| `practices/conventional-commits.md` | commitlint in pre-commit + CI |
| `practices/git.md` | Branch protection rules, PR workflow |

## References

- **spec**: `docs/specs/practices/madr.md`
- **spec**: `docs/specs/practices/bdr.md`
- **spec**: `docs/specs/practices/tdd.md`
- **spec**: `docs/specs/practices/conventional-commits.md`
- **spec**: `docs/specs/practices/git.md`
- **external**: https://adr.github.io/madr/
- **external**: https://www.conventionalcommits.org/