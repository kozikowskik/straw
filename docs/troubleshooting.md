# Troubleshooting

This guide lists common symptoms and the usual cause in `straw` integrations.

## A Binding Does Not Match

Check that the binding sequence uses the same key shape the adapter receives.

For printable text, prefer `TextSequence`.

```go
straw.Bind(goHome, straw.TextSequence("gh"))
```

For special keys, use `Code`.

```go
straw.Bind(cancel, straw.Sequence(straw.Code(straw.KeyEsc)))
```

For modified keys, use `Modified`.

```go
straw.Bind(quit, straw.Sequence(straw.Modified('c', straw.ModCtrl)))
```

If you construct keys manually, make sure each key is valid. `Text("g")` is valid. `Text("gh")`, `Text("")`, `Text(" ")`, and `Code('g')` are not valid binding keys. Use `TextSequence("gh")` for multiple printable key presses.

## A Short Binding Waits Instead Of Matching Immediately

This happens when the short binding is also a prefix of a longer binding.

For example, `g` and `gh` are ambiguous after the first `g`. The resolver returns `Pending` and waits for another key until the timeout expires.

If you want the short binding to resolve faster, reduce the timeout.

```go
resolver, err := straw.New(bindings, straw.WithTimeout(200*time.Millisecond))
```

## Host Keys Stop Working

Make sure your Bubble Tea update loop checks `ShouldPassThrough` before normal host key handling.

```go
result, cmd := m.resolver.Update(msg)

if !straw.ShouldPassThrough(result) {
	return m, cmd
}

switch msg := msg.(type) {
case tea.KeyPressMsg:
	// host key handling
}
```

Without this guard, a key consumed by `straw` can also trigger host behavior. With the guard in the wrong direction, host behavior may never run for unmatched keys.

## A Failed Pending Key Does Not Reach Host Handling

By default, if a sequence starts with a valid prefix and then fails, the failed pending key does not pass through to the host application.

Enable failed pending pass-through if your application wants that key to be handled normally.

```go
resolver, err := straw.New(bindings, straw.WithFailedPendingPassThrough(true))
```

## Escape Does Not Cancel A Sequence

The default cancel key is `esc`, but it only cancels pending input. If no sequence is pending, escape can pass through to host handling.

If you override cancel keys, include escape explicitly if you still want escape to cancel.

```go
resolver, err := straw.New(bindings,
	straw.WithCancelKeys(straw.Code(straw.KeyEsc)),
)
```

## Timeout Matching Does Not Happen

In Bubble Tea applications, return the command from `resolver.Update` when a result is pending or otherwise consumed by `straw`.

```go
if !straw.ShouldPassThrough(result) {
	return m, cmd
}
```

If you drop the command, Bubble Tea will not send the timeout message back through `Update`, and ambiguous short bindings will remain unresolved until another key arrives.

With the root resolver, you need to schedule the timeout yourself.

```go
result, timeout := resolver.UpdateKey(straw.Text("g"))
if result.IsPending() && timeout.Scheduled() {
	// Wait for timeout.Duration(), then call UpdateTimeout(timeout).
}
```

If `UpdateTimeout` returns `Idle`, the timeout token was stale or belonged to a different resolver. This can happen after another key, a cancel key, or `Reset`. Ignore that result and keep handling the current input.

## Old Keys Affect A New Screen Or Mode

Call `Reset` when pending input should not carry into the next screen, mode, popup, or keymap.

```go
resolver.Reset()
```

This clears only pending sequence state. It does not remove bindings.

For applications with separate Bubble Tea child models, another option is to give each active screen its own resolver and route messages only to that screen.

## Two Resolvers React To The Same Keys

Avoid broadcasting the same key message to several resolvers unless you intentionally want each resolver to keep independent pending state.

For most multi-screen applications, route the message to the active child model and let that child use its own resolver. If you switch screens while a resolver is pending, call `Reset` on the resolver that should stop tracking the old sequence.

## Non-Key Messages Appear To Do Nothing

This is expected. Adapter resolvers return `Idle` for non-key messages and for timeout messages that do not belong to the current pending sequence.

Handle non-key Bubble Tea messages as usual after the pass-through check or before calling the resolver if they should always run.

The adapters also ignore key input that does not map to one supported key press. Pasted text, multi-rune key messages, and key release messages are ignored by the current adapters.

## Duplicate Sequence Errors

Each sequence can map to only one action. If `New` returns `ErrDuplicateSequence`, remove or change one of the duplicate bindings.

```go
if errors.Is(err, straw.ErrDuplicateSequence) {
	return fmt.Errorf("duplicate key sequence: %w", err)
}
```

The sentinel errors live in the root package. Adapter packages re-export binding builders and key constants, but when you want explicit error checks, import the root package too.

```go
import strawcore "github.com/kozikowskik/straw"

if errors.Is(err, strawcore.ErrDuplicateSequence) {
	return err
}
```

## Import Confusion

Use the adapter import for Bubble Tea applications.

```go
import straw "github.com/kozikowskik/straw/bubbletea/v2"
```

Use the root package only for resolver-core code or adapter code.

```go
import strawcore "github.com/kozikowskik/straw"
```
