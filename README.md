# straw

Reusable key sequence/chord resolver for Bubble Tea applications.

Bubble Tea emits individual key press messages. This project will turn those events into higher-level key sequence results such as pending prefix, matched binding, no match, and cancelled sequence.

## Benchmarks

Run the full local benchmark suite with allocation reporting:

```sh
go test -run '^$' -bench=. -benchmem ./...
```

Run a focused benchmark group:

```sh
go test -run '^$' -bench '^BenchmarkNew$' -benchmem .
go test -run '^$' -bench '^BenchmarkUpdateExact' -benchmem .
go test -run '^$' -bench '^BenchmarkTimeout' -benchmem .
```

Capture a local baseline for later comparison:

```sh
go test -run '^$' -bench=. -benchmem ./... > /tmp/straw-benchmark-baseline.txt
```

Each benchmark result reports average time, bytes allocated, and allocation count per operation. For example, `1331402 ns/op`, `1280017 B/op`, and `10002 allocs/op` mean about `1.33ms`, `1.28MB`, and `10002` allocations per measured operation.

Current benchmarks are a local baseline, not a CI threshold or public performance guarantee. Use them to compare future optimization work, especially changes to resolver lookup behavior.
