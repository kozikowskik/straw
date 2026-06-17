# Contributing

Thank you for taking the time to improve `straw`.

`straw` is currently pre-release v0 software. Small, focused changes are easiest to review while the API is still settling.

## Local Setup

Install Go, then download module dependencies:

```sh
go mod download
```

Run the test suite:

```sh
go test ./...
```

## Code Style

- Keep changes small and focused.
- Prefer simple, readable code over clever abstractions.
- Run `gofmt` on changed Go files.
- Add or update tests for behavior changes.
- Add runnable examples for public API additions.
- Do not commit private notes, local experiments, editor settings, or generated files that are not part of the project.

## Benchmarks

Run benchmarks locally when changing resolver lookup, allocation behavior, or timeout handling:

```sh
go test -run '^$' -bench=. -benchmem ./...
```

Focused benchmark groups can be useful while working on resolver lookup behavior:

```sh
go test -run '^$' -bench '^BenchmarkNew$' -benchmem .
go test -run '^$' -bench '^BenchmarkUpdateExact' -benchmem .
go test -run '^$' -bench '^BenchmarkTimeout' -benchmem .
```

For before/after comparisons, capture repeated runs and compare them with `benchstat`:

```sh
go test -run '^$' -bench=. -benchmem -count=5 ./... > /tmp/straw-before.txt
go test -run '^$' -bench=. -benchmem -count=5 ./... > /tmp/straw-after.txt
benchstat /tmp/straw-before.txt /tmp/straw-after.txt
```

Benchmark results are not currently enforced in CI. Treat them as local evidence for performance-sensitive changes.

## Pull Request Expectations

Before opening a pull request, please check:

- `go test ./...` passes for the root package and both Bubble Tea adapter packages.
- Changed Go files are formatted with `gofmt`.
- Public API changes include tests and examples.
- Documentation is updated when behavior or user-facing commands change.
- The PR description explains the problem, solution, and verification performed.

## Issues And Feature Requests

Please include enough detail for maintainers to reproduce or evaluate the request:

- What you expected to happen.
- What actually happened.
- Relevant key bindings or resolver options.
- Go version and Bubble Tea version.
- A small reproduction when possible.

## Maintainer Notes

This project is currently maintained as a small open-source library. Review and response times may vary.
