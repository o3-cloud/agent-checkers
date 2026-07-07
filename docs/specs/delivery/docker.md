---
id: docker
layer: delivery
extends: []
---

# Docker

## Purpose

A Docker image is the unit of delivery — the file system, the user, the entrypoint, and every byte of installed software the runtime will execute. Sloppy image construction silently undoes the controls every other layer worked to enforce: a `FROM ubuntu:latest` makes builds non-reproducible across days, a missing `USER` directive ships every workload as root, an `apt-get install <pkg>` without a version drifts dependencies between rebuilds, secrets baked into `ENV` or `ARG` leak into every layer of the image's history, and a single-stage Dockerfile ships compilers and test frameworks into production. This spec pins how Dockerfiles are written, how base images and packages are pinned, how secrets are injected, how images are tagged and signed, and how vulnerability scans and SBOMs gate publication so the artifact a registry hands to a runtime is exactly the artifact CI built — minimal, reproducible, attested, and free of credentials.

## References

- **spec** `sbom` — every image must ship with a Software Bill of Materials
- **spec** `github-actions` — sibling delivery-layer spec for the CI workflow that builds and publishes images
- **spec** `terraform` — sibling delivery-layer spec; image references in IaC must be pinned by digest
- **external** `https://www.docker.com/` — Docker home
- **external** `https://docs.docker.com/build/buildkit/` — BuildKit
- **external** `https://docs.docker.com/develop/develop-images/dockerfile_best-practices/` — Dockerfile best practices
- **external** `https://github.com/sigstore/cosign` — cosign image signing
- **external** `https://github.com/aquasecurity/trivy` — Trivy image scanner

## Rules

1. Build images from a `Dockerfile` checked into the repo; do not build images via ad-hoc shell scripts that wrap `docker commit`.
2. Reference every base image by immutable digest (`FROM <image>@sha256:<digest>`) with the human-readable tag in a trailing comment; do not reference base images by tag alone.
3. Use multi-stage builds — separate `builder` stages from the final runtime stage — and copy only the runtime artifacts into the final image; do not ship compilers, test frameworks, package managers, or shell history in the runtime image.
4. Base production images on a minimal distribution (distroless, Alpine, Chainguard, or `*-slim`/`*-chiseled` variants); do not use full-distribution images for runtime stages.
5. Run the container as a non-root user via an explicit `USER <uid>:<gid>` directive in the final stage; do not leave the runtime user as root.
6. Pin every package installed in the image (`apt-get install <pkg>=<version>`, `apk add <pkg>=<version>`, `pip install <pkg>==<version>`); do not install packages without an explicit version.
7. Combine `apt-get update` with the `apt-get install` and clean `/var/lib/apt/lists/*` in the same `RUN` instruction; do not leave package indexes or caches in the final image.
8. Use `COPY` for local files; reserve `ADD` for the cases it uniquely handles (remote-URL fetch, tarball auto-extract) and document the reason inline.
9. Set `WORKDIR` explicitly in every stage that runs commands; do not rely on the inherited working directory of a base image.
10. Commit a `.dockerignore` that excludes `.git/`, `node_modules/`, build outputs, tests, documentation, and any secret material; do not rely on `COPY` patterns alone to keep secrets out of the build context.
11. Inject build-time secrets via BuildKit `--mount=type=secret`; do not pass credentials through `ARG`, `ENV`, or files baked into image layers.
12. Specify `ENTRYPOINT` and `CMD` in exec form (`["bin", "arg"]`); do not use shell form for the primary entrypoint.
13. Declare a `HEALTHCHECK` for any long-running service image; do not ship a service image with no health probe.
14. Build with BuildKit (`docker buildx build` or `DOCKER_BUILDKIT=1`); do not use the legacy builder for new pipelines.
15. Tag every published image with the immutable build identifier (commit SHA or release version) and pull/run by digest in production manifests; do not deploy production from a mutable tag like `:latest` or `:main`.
16. Sign every published image with cosign (or an equivalent) and verify the signature on pull in production; do not deploy unsigned images.
17. Generate an SBOM for every published image and attach it as an OCI referrer or registry-side artifact (refs: sbom).
18. Run a vulnerability scanner (Trivy, Grype, or equivalent) against every published image in CI; treat findings at the configured severity as build failures.
19. Run containers with a read-only root filesystem and an explicit, minimal capability set; do not use `--privileged` in production and do not grant capabilities the workload does not require.
20. Reference images in Kubernetes manifests, Compose files, and Terraform by digest; do not reference production images by tag alone in deployment manifests.
