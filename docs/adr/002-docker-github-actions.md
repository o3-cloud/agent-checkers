# ADR-002: Docker and GitHub Actions for Delivery

## Status

Accepted

## Context and Problem Statement

We need a reliable way to build, package, and deploy the application. What delivery mechanism should we use for consistent and reproducible deployments?

## Decision Drivers

- **Reproducibility**: Same artifact runs identically in all environments
- **CI/CD Integration**: Seamless integration with GitHub hosting
- **Security**: Vulnerability scanning and supply chain integrity
- **Simplicity**: Minimal configuration overhead
- **AI-agent-friendly**: Clear, declarative configurations that are easy to generate and validate

## Considered Options

- **Option A**: Docker + GitHub Actions
- **Option B**: Kubernetes + GitLab CI
- **Option C**: Packer + Jenkins
- **Option D**: Nix + Hydra

## Decision Outcome

Chosen option: **Docker + GitHub Actions**, because:

1. **Single-binary + Docker**: Go produces static binaries that pair perfectly with minimal Docker images
2. **GitHub-native**: Repository is already on GitHub; GitHub Actions is built-in
3. **Marketplace ecosystem**: Rich ecosystem of pre-built actions for security scanning, linting, deployment
4. **SHA-pinnable**: Images and actions can be pinned to cryptographic hashes for supply chain security
5. **Simple compose files**: `compose.yaml` for local dev, production overrides optional

## Consequences

### Positive

- Consistent builds across all environments
- Built-in secret management in GitHub Actions
- Easy local development with `docker compose up`
- Rich security scanning ecosystem (Trivy, Hadolint)
- SHA pinning for supply chain integrity
- Matrix builds for multiple platforms

### Negative

- Docker layer caching requires configuration for optimal CI speed
- GitHub Actions minutes can be costly for large teams
- Vendor lock-in to GitHub ecosystem

## Compliance

This decision is enforced by:

| Spec | Enforcement |
|------|-------------|
| `delivery/docker.md` | Hadolint in pre-commit + CI |
| `delivery/docker.md` | Trivy IaC scanning in CI |
| `delivery/github-actions.md` | SHA-pinned actions via gh-pin |
| `delivery/docker-compose.md` | `docker compose config` validation |

## References

- **spec**: `docs/specs/delivery/docker.md`
- **spec**: `docs/specs/delivery/github-actions.md`
- **external**: https://docs.github.com/en/actions
- **external**: https://docs.docker.com/best-practices/