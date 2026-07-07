# ADR-005: Dependency Management Policy

## Status

Accepted

## Context and Problem Statement

Go modules provide dependency management, but we need policies for pinning, updating, and auditing third-party dependencies to ensure supply chain security and reproducibility.

## Decision Drivers

- **Reproducibility**: Builds should be deterministic
- **Security**: Vulnerabilities in dependencies must be detectable
- **Auditability**: Clear record of dependency provenance
- **Update cadence**: Controlled, reviewed dependency updates
- **AI-agent-friendly**: Machine-readable dependency declarations

## Considered Options

- **Option A**: Pin to semantic versions + automated vulnerability scanning + Renovate
- **Option B**: Vendored dependencies + manual updates
- **Option C**: Latest versions always (no pinning)

## Decision Outcome

Chosen option: **Pin to semantic versions + automated vulnerability scanning + Renovate**, because:

1. **Semantic versioning**: Go modules already enforce semver compatibility
2. **go.mod + go.sum**: Committed lockfile ensures reproducibility
3. **Vulnerability scanning**: `govulncheck` detects known CVEs in dependencies
4. **Renovate**: Automated, configurable PRs for dependency updates
5. **Grouped updates**: Reduce noise by grouping minor/patch updates

## Consequences

### Positive

- Reproducible builds via committed `go.mod` and `go.sum`
- Known vulnerabilities detected automatically
- Renovate reduces manual dependency maintenance burden
- Clear audit trail of dependency changes via PR history
- No `replace` directives pointing to local paths in release branches

### Negative

- Renovate requires initial configuration
- Vulnerability alerts may create urgent update PRs
- Large dependency trees can slow `go mod download`

## Compliance

This decision is enforced by:

| Spec | Enforcement |
|------|-------------|
| `security/dependency-management.md` | `go mod tidy` in pre-commit |
| `language/go.md:17` | Git track `go.mod` + `go.sum` |
| `language/go.md:18` | No `replace` in release branches |
| `security/vulnerability-scanning.md` | `govulncheck` in CI |
| `security/renovate.md` | Renovate bot for automated updates |

## References

- **spec**: `docs/specs/security/dependency-management.md`
- **spec**: `docs/specs/language/go.md` (rules 17-18)
- **external**: https://go.dev/ref/mod
- **external**: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- **external**: https://docs.renovatebot.com/