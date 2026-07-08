---
id: openapi
layer: interface
extends: []
---

# OpenAPI Specification

## Purpose

A machine-readable API contract is the single source of truth for what the REST API does. Without one, AI agents, frontend developers, and contract-testing tools must read source code to discover endpoints, schemas, and error formats — which is brittle, drifts from reality, and blocks automation. OpenAPI 3.1 (aligned with JSON Schema 2020-12) is the industry standard; serving it at a well-known endpoint lets clients auto-discover the API, generate typed clients, validate responses, and detect breaking changes. This spec pins where the document lives, how it is generated, what it must contain, and how it stays in sync with the running server.

## References

- **external** `https://spec.openapis.org/oas/v3.1.0` — OpenAPI Specification 3.1.0
- **external** `https://github.com/OAI/OpenAPI-Specification` — OpenAPI Specification repository
- **external** `https://github.com/deepmap/oapi-codegen` — OpenAPI code generator for Go
- **external** `https://www.spectral-lint.io/` — Spectral linter for OpenAPI specs
- **adr** `009-openapi-specification` — ADR for the OpenAPI decision
- **bdr** `010-openapi-specification` — BDR for the OpenAPI behavior

## Rules

1. Serve the OpenAPI 3.1 document at `GET /openapi.json` with `Content-Type: application/json` and at `GET /openapi.yaml` with `Content-Type: application/yaml`; both endpoints must return semantically identical content.
2. Set `openapi` field to `3.1.0`; do not use 3.0.x — 3.1 aligns with JSON Schema 2020-12 and supports full `null` types.
3. Include an `info` object with `title`, `version` (matching the application version from `go.mod`), `description`, and `contact` fields.
4. Include a `servers` array with at least one entry for the default URL (`http://localhost:8080`); document production and staging URLs when known.
5. List every registered REST endpoint in the `paths` object, including path parameters (`{id}`), query parameters, request bodies, and all possible response codes.
6. Give every operation a unique `operationId` in camelCase (e.g., `createGame`, `joinGame`, `makeMove`); the `operationId` is the primary key for code generators and client SDKs.
7. Give every operation a `summary` (short, one line) and `description` (longer, may include examples); do not leave either empty.
8. Define reusable schemas in `components.schemas` and reference them via `$ref: "#/components/schemas/<Name>"`; do not inline the same schema in multiple operations.
9. Name schema components after the Go struct they represent (e.g., `CreateGameRequest` for `dto.CreateGameRequest`); use PascalCase for component names.
10. Document error responses (400, 404, 409, 500) for every endpoint that can return them; each error response must reference the `ErrorResponse` component schema.
11. Include `tags` on each operation to group endpoints by domain (e.g., `games`, `moves`, `health`, `meta`); use the tag name as the group label in generated documentation.
12. Commit a static `docs/openapi.yaml` file to the repository for offline tooling; update it whenever the API surface changes.
13. Add a CI step that fetches `/openapi.json` from the running server and diffs it against the committed `docs/openapi.yaml`; fail CI if they diverge.
14. Validate the committed `docs/openapi.yaml` with Spectral (or equivalent OpenAPI linter) in CI; do not merge a spec with validation errors.
15. Do not include authentication or security schemes in the initial spec; add `securitySchemes` and `security` when authentication is implemented and document in a new ADR.
16. Version the API path prefix (`/api/v1/...`); increment the prefix in the OpenAPI `servers` and `paths` when a new major version is introduced.