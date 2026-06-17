<p align="center">
  <img src="logo.png" alt="straw logo" width="400">
</p>

---

# straw

Reusable key sequence resolver for Bubble Tea applications with modifier-aware matching.

Bubble Tea emits individual key press messages. `straw` turns those events into higher-level key sequence results such as pending prefix, matched binding, unmatched input, and canceled sequence.

## Installation

```sh
go get github.com/kozikowskik/straw
```

Bubble Tea v2 users should import the v2 adapter:

```go
import straw "github.com/kozikowskik/straw/bubbletea/v2"
```

Bubble Tea v1 users should import the v1 adapter:

```go
import straw "github.com/kozikowskik/straw/bubbletea/v1"
```

The root `github.com/kozikowskik/straw` package contains the version-neutral resolver core for advanced use and adapter authors. If you need both packages in one file, import the root package as `strawcore`.

## Quick Start

Define application-owned actions, bind them to key sequences, and call the resolver from your Bubble Tea `Update` function.

```go
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	straw "github.com/kozikowskik/straw/bubbletea/v2"
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
		straw.Bind(goDashboard, straw.TextSequence("gd"), straw.Description("go dashboard action")),
	})
	if err != nil {
		return model{}, err
	}
	return model{resolver: resolver}, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.resolver.Update(msg)

	switch {
	case result.Match(goHome):
		fmt.Println("go home action")
		return m, cmd
	case result.Match(goDashboard):
		fmt.Println("go dashboard action")
		return m, cmd
	}

	// Only unmatched pass-through keys should reach the host key switch.
	// Pending prefixes and matched-but-unhandled bindings stay consumed by straw.
	if !straw.ShouldPassThrough(result) {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, cmd
}
```

The resolver only reports key sequence outcomes. Your application owns the actions and decides which Bubble Tea commands to return.

Runnable examples are available in [`examples/bubbletea-v2`](examples/bubbletea-v2), [`examples/bubbletea-v1`](examples/bubbletea-v1), and [`examples/timeout-cancel`](examples/timeout-cancel):

```sh
go run ./examples/bubbletea-v2
go run ./examples/bubbletea-v1
go run ./examples/timeout-cancel
```

Detailed guides are available in [`docs/`](docs/): [core concepts](docs/concepts.md), [bindings](docs/bindings.md), [Bubble Tea integration](docs/bubble-tea.md), and [troubleshooting](docs/troubleshooting.md).

## API Overview

- `Bind` maps one application-owned action to one key sequence.
- `TextSequence`, `Sequence`, `Text`, `Code`, and `Modified` build the key sequences users press.
- `New` validates bindings and creates a resolver.
- Adapter `Resolver.Update` methods accept Bubble Tea messages and return a `Result[A]` plus an optional `tea.Cmd`.
- `Result[A]` reports matched, pending, unmatched, canceled, and idle input.
- `ShouldPassThrough` tells the host application when normal key handling should run.

## Resolver Behavior

Adapter `Resolver.Update`, root `Resolver.UpdateKey`, and root `Resolver.UpdateTimeout` all return `Result[A]`, but they reach `Idle` in different situations:

- `Idle`: adapter input was not a key press, or a timeout token was stale or unrelated to the current pending sequence.
- `Pending`: the key sequence is a valid prefix and the resolver is waiting for another key.
- `Matched`: the sequence matched a binding.
- `Unmatched`: the sequence did not match any binding.
- `Canceled`: the pending sequence was canceled.

When a key is both a complete binding and a prefix for a longer binding, `straw` returns `Pending` and starts a timeout command. If no longer sequence arrives before the timeout, the pending exact match resolves.

By default, pending sequences time out after `500ms` and `esc` cancels pending input. Use `WithTimeout` and `WithCancelKeys` to customize that behavior.

Call `Reset` when the old pending keys should no longer matter. Common cases are switching screens, closing a palette, changing modes, or replacing the active keymap.

## Host Key Handling

`straw` does not own every key in your application. When the resolver returns `Unmatched` with `PassThrough() == true`, the latest key can be handled by your normal host application key switch.

Failed keys after a pending prefix do not pass through by default. Use `WithFailedPendingPassThrough(true)` if your application wants those failed pending keys to be handled by the host app.

## Performance

The current implementation uses a simple lookup path that is easy to understand and test. It is intended for typical terminal applications with modest binding sets.

Future versions may replace the current lookup with a trie or another prefix index so very large binding sets stay efficient.

## Current Limitations

`straw` is intentionally small. These limits are part of the current design:

- Matching is exact. `ctrl+c`, `c`, `esc`, and `enter` are different key shapes.
- Use `TextSequence("gh")` for multi-key text sequences. `Text("gh")` is not a valid single key.
- Modified keys such as `ctrl+c` and `alt+enter` are supported. Simultaneous non-modifier chords, such as pressing `g+h+d` at the same time, are not supported.
- Adapter packages ignore Bubble Tea input that cannot be represented as one supported key press, such as pasted text or key release events.
- Timeout tokens are tied to one resolver and one pending generation. A stale timeout returns `Idle` and should be ignored.
- The root resolver is mutable. Use it from one update flow at a time, as you normally would inside a Bubble Tea model.
- Binding analysis, file-based binding configuration, modes, contexts, enabled or disabled bindings, and continuation inspection are deferred for now.

## Roadmap

`straw` is pre-release v0 software. The public API is intended to be small and stable enough for early use, but breaking changes may still happen before v1 as real Bubble Tea integrations shape the resolver model.

- Stabilize the v0 API through real Bubble Tea usage.
- Consider binding analysis and reporting for larger keymaps.
- Consider continuation inspection so applications can show possible next keys for pending prefixes.
- Improve lookup performance if real applications need larger binding sets.

## Contributing

Contributions are welcome. See `CONTRIBUTING.md` for local setup, tests, benchmarks, and pull request expectations.

## License

`straw` is available under the MIT License. See `LICENSE` for details.
