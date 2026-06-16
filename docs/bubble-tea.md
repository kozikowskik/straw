# Bubble Tea Integration

Most applications should use a Bubble Tea adapter package instead of the root resolver directly. The adapters translate Bubble Tea key messages into `straw` keys and translate pending sequence timeouts into Bubble Tea commands.

## Import Paths

Use the adapter that matches your Bubble Tea major version.

| Bubble Tea version | Adapter import |
| --- | --- |
| v2 | `github.com/kozikowskik/straw/bubbletea/v2` |
| v1 | `github.com/kozikowskik/straw/bubbletea/v1` |

Alias the adapter import as `straw` in normal application code.

```go
import straw "github.com/kozikowskik/straw/bubbletea/v2"
```

If one file needs both the adapter and the root package, alias the root package as `strawcore`.

```go
import (
	strawcore "github.com/kozikowskik/straw"
	straw "github.com/kozikowskik/straw/bubbletea/v2"
)
```

## Create A Resolver

Create the resolver with adapter package bindings. Adapter packages re-export the root binding and key helpers, so normal application code can stay on one import.

```go
resolver, err := straw.New([]straw.Binding[action]{
	straw.Bind(goHome, straw.TextSequence("gh"), straw.Description("go home")),
	straw.Bind(goDashboard, straw.TextSequence("gd"), straw.Description("go dashboard")),
})
if err != nil {
	return model{}, err
}
```

Store the resolver on your Bubble Tea model.

```go
type model struct {
	resolver *straw.Resolver[action]
	message  string
}
```

## Update Loop Pattern

Call the resolver near the start of `Update`. Handle matched actions first, then gate normal host key handling with `ShouldPassThrough`.

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.resolver.Update(msg)

	switch {
	case result.Match(goHome):
		m.message = "matched: go home"
		return m, cmd
	case result.Match(goDashboard):
		m.message = "matched: go dashboard"
		return m, cmd
	case result.IsPending():
		m.message = "pending sequence..."
	case result.IsCanceled():
		m.message = "sequence canceled"
	}

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

Return the resolver command when the result is pending. The command schedules the timeout message used to resolve ambiguous prefixes.

## Modifier Behavior

Modifier matching is exact in v0. A binding for `Modified('c', ModCtrl)` does not match `ctrl+alt+c`; bind `Modified('c', ModCtrl|ModAlt)` when both modifiers are required.

Printable uppercase letters and shifted punctuation are treated as text when Bubble Tea reports printable text, even if Bubble Tea also reports an explicit Shift modifier. For example, bind uppercase `G` as `Text("G")`, not `Modified('g', ModShift)`.

Bubble Tea v1 and v2 expose different modifier models. The v1 adapter supports regular text, special keys, ctrl aliases, and Alt-modified text or special keys. The v2 adapter maps Bubble Tea's explicit modifier bits to straw modifiers, including Shift, Alt, Ctrl, Meta, Hyper, and Super.

## Bubble Tea v2 Notes

The v2 adapter accepts `charm.land/bubbletea/v2` messages and handles `tea.KeyPressMsg` key events.

```go
import (
	tea "charm.land/bubbletea/v2"
	straw "github.com/kozikowskik/straw/bubbletea/v2"
)
```

Runnable code is available in [examples/bubbletea-v2](../examples/bubbletea-v2).

## Bubble Tea v1 Notes

The v1 adapter accepts `github.com/charmbracelet/bubbletea` messages and handles `tea.KeyMsg` key events.

```go
import (
	tea "github.com/charmbracelet/bubbletea"
	straw "github.com/kozikowskik/straw/bubbletea/v1"
)
```

Runnable code is available in [examples/bubbletea-v1](../examples/bubbletea-v1).

## Timeout And Cancel Options

Adapter packages expose the same resolver options as the root package.

```go
resolver, err := straw.New(bindings,
	straw.WithTimeout(250*time.Millisecond),
	straw.WithCancelKeys(straw.Code(straw.KeyEsc)),
)
```

See [examples/timeout-cancel](../examples/timeout-cancel) for a runnable example.

## Root Core Versus Adapters

Use the root package for:

- Adapter implementations.
- Tests that exercise resolver behavior without Bubble Tea.
- Applications that already translate input into `straw.Key` values.

Use an adapter package for:

- Bubble Tea applications.
- Automatic translation from Bubble Tea key messages.
- Timeout commands that integrate with the Bubble Tea update loop.
