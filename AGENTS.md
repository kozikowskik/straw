# Agent Guide

## Project and boundaries

`straw` is a Go library that resolves key sequences for Bubble Tea applications. See [README.md](README.md) for usage and [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations.

- The root package owns version-neutral matching, pending state, and timeout validation.
- `bubbletea/v1` and `bubbletea/v2` translate their respective Bubble Tea messages and schedule timeout commands. Keep Bubble Tea dependencies and message handling in these adapters.
- Applications own and execute actions; the resolver returns action values and metadata.
- Shared API changes should account for both adapter surfaces and relevant examples. Preserve the documented version-specific input and modifier behavior; the adapters are not interchangeable. See [Bubble Tea integration](docs/bubble-tea.md).

## Setup and checks

Run commands from the repository root. Use a Go toolchain meeting the requirement in [go.mod](go.mod), then download dependencies:

```sh
go mod download
```

### Go changes

Run `gofmt` on changed Go files and run:

```sh
go test ./...
```

This runs existing tests across the core and both adapters and compiles the example applications. It does not exercise interactive application behavior.

### Full CI checks

The [CI workflow](.github/workflows/ci.yml) runs the following checks. This is a reference for reproducing CI, not a requirement to run every command after each edit:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go test -race ./...
go test -run '^$' -bench=. -benchtime=100ms ./...
go mod tidy -diff && git diff --exit-code -- go.mod go.sum
```

### Performance-sensitive changes

For changes to resolver lookup, allocations, or timeout handling, also measure benchmarks:

```sh
go test -run '^$' -bench=. -benchmem ./...
```

Benchmark smoke checks do not establish performance equivalence. See [CONTRIBUTING.md](CONTRIBUTING.md) for repeated before/after comparisons.

### Documentation-only changes

Check affected text, links, and the diff. Pure prose changes do not require the Go suite. Changes to executable Go examples need relevant tests or compilation checks; report any verification of interactive behavior separately.

## Making changes

- Read the relevant definitions and existing tests before changing behavior; verify signatures and data structures rather than assuming them.
- Keep changes focused and consistent with surrounding code. Avoid unrelated refactors, dependencies, and features.
- Inspect existing working-tree changes and preserve unrelated contributor work.
- For resolver changes, preserve these contracts unless the requested change explicitly alters them:
  - Timeout tokens belong to one resolver and pending generation; stale or foreign tokens must not resolve current state.
  - Host key handling runs only when `ShouldPassThrough` permits it; pending or matched keys stay consumed.
  - APIs documented to return copies must not expose mutable internal sequences or slices. Action values retain their existing value semantics.
  - Pending-state inspection does not execute actions or change timeout state.

## Tests, examples, and documentation

- Add or update tests for changed behavior, including regression coverage for fixes.
- Public API additions need runnable examples. Check the core and both adapter surfaces when the API is shared.
- Update affected public documentation when behavior or user-facing commands change. Use README for entry-point guidance, `docs/` for detailed behavior, and the changelog for release-visible changes.
- In application examples, alias the adapter import as `straw`; use `strawcore` for the root package when both appear in one file.
- Keep instructions and documented commands aligned with the implementation and CI.

## Handoff

Summarize the change, the checks actually executed and their results, and any verification gaps or remaining questions. Distinguish source inspection, compilation, automated tests, and manual checks. Do not report recommended or unexecuted checks as passing.
