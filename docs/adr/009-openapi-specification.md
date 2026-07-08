# ADR-009: OpenAPI Specification and Discovery

## Status

Proposed

## Context and Problem Statement

The agent-checkers REST API currently has no machine-readable contract. Consumers (AI agents, frontend developers, integrators, testing tools) must read source code or ad-hoc documentation to discover available endpoints, request/response schemas, and error formats. This creates friction for automated clients (MCP servers, code generators, contract testers) and allows the API surface to drift silently from documentation. An OpenAPI 3.1 specification published at a well-known endpoint solves this by providing a single source of truth that tools can fetch, validate against, and generate from.

## Decision Drivers

- AI agents (BDR-005) need a machine-readable API contract to discover available operations
- Frontend developers (BDR-006) need request/response schemas for type-safe client generation
- Contract testing tools need a spec to validate API responses against
- The API surface should be self-describing — clients should not need to read source code
- OpenAPI 3.1 is the industry standard and aligns with JSON Schema 2020-12

## Considered Options

- **Option A: Hand-written OpenAPI YAML** — Manually author and maintain a `openapi.yaml` file served at `/openapi.yaml`
- **Option B: Code-generated from annotations** — Use `swaggo/swag` or `swaggo/gin-swagger` annotations on handlers to generate OpenAPI at build time
- **Option C: Runtime introspection from chi routes** — Build the OpenAPI document programmatically from the chi router at startup, serve as JSON

## Decision Outcome

Chosen option: **Option C (runtime introspection)**, because it keeps the spec in sync with the actual routes at all times, requires no separate annotation language or build step, and naturally produces JSON that can be served at both `/openapi.json` and `/api/v1/openapi.json`. The server constructs an OpenAPI 3.1 document at startup by walking the chi router and consulting handler-level metadata structs. Schemas for request/response DTOs are generated from existing `dto` package struct tags.

A static `openapi.yaml` is also committed to the repo under `docs/openapi.yaml` for repository-based tools (linters, CI checks) that need to validate the spec without running a server. The server's runtime document and the static file are kept in sync by a CI step that fetches `/openapi.json` and diffs against the committed file.

## Consequences

- Positive: Single source of truth — the running server IS the spec
- Positive: AI agents can discover the API by fetching `/openapi.json`
- Positive: Code generators (oapi-codegen, openapi-typescript) can target the endpoint or the file
- Positive: Contract tests can validate responses against schemas
- Negative: Runtime construction adds complexity to server startup
- Negative: Keeping the static file in sync requires a CI check
- Negative: Schemas are only as accurate as the struct tags — missing tags mean missing fields

## References

- **BDR-010**: OpenAPI Specification and Discovery
- **ADR-007**: Interface Layers (Web, API, CLI, MCP)
- **external** `https://spec.openapis.org/oas/v3.1.0` — OpenAPI Specification 3.1.0
- **external** `https://github.com/OAI/OpenAPI-Specification` — OpenAPI Specification repository
- **external** `https://github.com/deepmap/oapi-codegen` — OpenAPI code generator for Go