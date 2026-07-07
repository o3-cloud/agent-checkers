---
id: go
layer: language
extends: []
---

# Go

## Purpose

Go optimizes for readability and uniform style across authors. Its toolchain bakes in opinionated defaults (`gofmt`, `go vet`, `go mod`) that eliminate bikeshedding when they are actually used. This spec pins the language version, formatting, module hygiene, and idiomatic patterns so code written by any contributor looks and behaves like code written by the team, and so the built-in guarantees (vetting, race detection, module integrity) are not silently bypassed.

## References

- **external** `https://go.dev/ref/spec` — The Go Programming Language Specification
- **external** `https://go.dev/doc/effective_go` — Effective Go
- **external** `https://google.github.io/styleguide/go/` — Google Go Style Guide
- **external** `https://github.com/golang/go/wiki/CodeReviewComments` — Go Code Review Comments
- **external** `https://pkg.go.dev/cmd/gofmt` — gofmt reference
- **external** `https://staticcheck.dev/` — staticcheck analyzer
- **external** `https://golangci-lint.run/` — golangci-lint aggregator

## Rules

1. Declare the minimum supported Go version in `go.mod` with a `go` directive, and keep it within the two most recent stable minor releases supported by the Go team. (refs: https://go.dev/doc/devel/release)
2. Format every `.go` file with `gofmt` (or `goimports`) before commit; CI must fail on unformatted files.
3. Order imports in two groups separated by a single blank line: standard library, then everything else. Use `goimports` to enforce grouping automatically.
4. Name packages as short, lowercase, single-word identifiers without underscores or mixedCaps; the package name must match its directory. (refs: https://go.dev/blog/package-names)
5. Do not use the blank identifier to silence errors returned from function calls; handle the error or document why it is discarded with a comment.
6. Return errors as the last return value and check them at every call site before using any other return values.
7. Wrap errors with `fmt.Errorf("...: %w", err)` when adding context, and compare with `errors.Is` / `errors.As` rather than string matching or `==`.
8. Do not `panic` in library code for expected failure modes; reserve `panic` for truly unrecoverable programmer errors.
9. Accept a `context.Context` as the first parameter of any function that performs I/O, blocks, or crosses an API boundary, and propagate it rather than calling `context.Background()` downstream.
10. Do not store a `context.Context` in a struct field; pass it explicitly through call chains.
11. Define interfaces in the package that consumes them, not the package that implements them, and keep them small (prefer single-method interfaces named with the `-er` suffix when natural).
12. Guard concurrent access to shared state with the `sync` package or channels; run `go test -race` in CI on every package that exercises goroutines.
13. Place tests in `_test.go` files in the same package (or an `_test` sibling package for black-box tests) and name test functions `TestXxx(t *testing.T)`.
14. Use table-driven subtests with `t.Run` for cases that vary only by input/expected output.
15. Vet every package with `go vet ./...` in CI; treat any diagnostic as a build failure.
16. Run `staticcheck` (or `golangci-lint` with staticcheck enabled) on every package in CI; do not disable checks inline without a comment explaining why.
17. Commit `go.mod` and `go.sum`; run `go mod tidy` before committing dependency changes so both files remain consistent.
18. Pin third-party dependencies to tagged semantic versions in `go.mod`; do not use `replace` directives pointing to local paths or forks in released branches.
19. Do not use `init()` for work that can be done lazily or explicitly; reserve it for registration with runtime-resolved registries.
20. Export only identifiers that must be part of the package's public API; every exported identifier requires a doc comment beginning with the identifier's name. (refs: https://go.dev/doc/effective_go#commentary)
