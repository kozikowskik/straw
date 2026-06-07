<p align="center">
  <img src="logo.png" alt="straw logo" width="240">
</p>

---

# straw

Reusable key sequence and chord resolver for Bubble Tea applications.

Bubble Tea emits individual key press messages. `straw` turns those events into higher-level key sequence results such as pending prefix, matched binding, unmatched input, and canceled sequence.

## Installation

```sh
go get github.com/kozikowskik/straw
```

`straw` currently targets Bubble Tea v2 via `charm.land/bubbletea/v2`.

## Quick Start

Define application-owned actions, bind them to key sequences, and call the resolver from your Bubble Tea `Update` function.

```go
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/kozikowskik/straw"
)

type action string

const (
	goHome action = "go-home"
	goDashboard action = "go-dashboard"
)

type model struct {
	resolver *straw.Resolver[action]
}

func newModel() (model, error) {
	resolver, err := straw.New([]straw.Binding[action]{
		straw.Bind(goHome, straw.TextSequence("gh"), straw.Description("go home action")),
		straw.Bind(goDashboard, straw.Text("gd"), straw.Description("go dashboard action")),
	})
	if err != nil {
		return model{}, err
	}
	return model{resolver: resolver}, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.resolver.Update(msg)
	if cmd != nil {
		return m, cmd
	}

	switch {
	case result.Match(goHome):
		fmt.Println("go home action")
		return m, nil
	case result.Match(goDashboard):
		fmt.Println("go dashboard action")
		return m, nil
	}

	// Only unmatched pass-through keys should reach the host key switch.
	// Pending prefixes and matched-but-unhandled bindings stay consumed by straw.
	if straw.ShouldPassThrough(result) {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}
```

The resolver only reports key sequence outcomes. Your application owns the actions and decides which Bubble Tea commands to return.

## API Overview

- `Key` describes one key press.
- `Text`, `Code`, and `Modified` build keys for printable text, special keys, and modified keys.
- `Seq`, `Sequence`, and `TextSequence` build ordered key sequences.
- `Binding[A]` maps an application-owned action to a sequence.
- `Bind` creates bindings and accepts optional metadata such as `Description`.
- `New` validates bindings and builds a `Resolver[A]`.
- `Resolver.Update` accepts Bubble Tea messages and returns a `Result[A]` plus an optional `tea.Cmd`.
- `Result[A]` reports whether input is idle, pending, matched, unmatched, or canceled.
- `ShouldPassThrough` reports whether normal host key handling should run for a result.

## Resolver Behavior

`Resolver.Update` returns one of these result states:

- `Idle`: the message was not a key press or relevant timeout.
- `Pending`: the key sequence is a valid prefix and the resolver is waiting for another key.
- `Matched`: the sequence matched a binding.
- `Unmatched`: the sequence did not match any binding.
- `Canceled`: the pending sequence was canceled.

When a key is both a complete binding and a prefix for a longer binding, `straw` returns `Pending` and starts a timeout command. If no longer sequence arrives before the timeout, the pending exact match resolves.

By default, pending sequences time out after `500ms` and `esc` cancels pending input. Use `WithTimeout` and `WithCancelKeys` to customize that behavior.

## Host Key Handling

`straw` does not own every key in your application. When the resolver returns `Unmatched` with `PassThrough() == true`, the latest key can be handled by your normal host application key switch.

Failed keys after a pending prefix do not pass through by default. Use `WithFailedPendingPassThrough(true)` if your application wants those failed pending keys to be handled by the host app.

## Performance

The current v0 implementation uses a simple lookup path that is easy to understand and test. Benchmarks show linear scaling with binding count, which is expected for the current design. This is acceptable for typical terminal applications with modest binding sets.

Future versions may replace the current lookup with a trie or another prefix index so very large binding sets stay efficient.

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

## Roadmap

- Stabilize the v0 API through real Bubble Tea usage.
- Add a Bubble Tea v1 compatibility adapter.
- Improve lookup performance for large binding sets.
- Add public CI, release notes, and contribution workflow files before the first public release.

## Contributing

Contributions are welcome after the public workflow is in place. See `CONTRIBUTING.md` for local setup, tests, benchmarks, and pull request expectations.

## License

`straw` is available under the MIT License. See `LICENSE` for details.
