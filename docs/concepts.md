# Core Concepts

`straw` resolves individual key presses into application-owned actions. It is useful when a Bubble Tea application wants Vim-style key sequences such as `gh` or `gd` without making every screen reimplement sequence state.

## Resolver Core

The root package, `github.com/kozikowskik/straw`, contains the version-neutral resolver core. It does not depend on Bubble Tea.

Use the root package directly when you are writing adapter code or when your application already has a non-Bubble Tea key event type.

```go
resolver, err := straw.New([]straw.Binding[string]{
	straw.Bind("go-home", straw.TextSequence("g")),
	straw.Bind("go-help", straw.TextSequence("gh")),
})
if err != nil {
	return err
}

result, timeout := resolver.UpdateKey(straw.Text("g"))
if result.IsPending() && timeout.Scheduled() {
	// In a real event loop, wait for timeout.Duration(), then call UpdateTimeout.
	result = resolver.UpdateTimeout(timeout)
}
```

The root resolver does not start timers for you. It returns a `Timeout[A]` token when the application should schedule timeout work. Check `timeout.Scheduled()` before scheduling anything. Use `timeout.Duration()` as the wait time. When the timer fires, pass the same token to `UpdateTimeout`.

Timeout tokens are safe to ignore when they are no longer relevant. If another key, cancel key, or reset changed the pending sequence, the old token is stale. Passing a stale token to `UpdateTimeout` returns `Idle` and leaves the current pending sequence alone.

Most Bubble Tea applications should use one of the adapter packages instead of calling `UpdateKey` directly.

## Result States

Every resolver update returns a `Result[A]`, where `A` is your action type.

| State | Meaning |
| --- | --- |
| `Idle` | The input did not affect the resolver. Adapter packages return this for non-key messages and irrelevant timeout messages. |
| `Pending` | The current keys are a valid prefix and the resolver is waiting for another key. |
| `Matched` | The current sequence matched a binding. |
| `Unmatched` | The input did not match any binding. |
| `Canceled` | Pending input was canceled by a configured cancel key. |

Check result state with methods such as `IsPending`, `IsMatched`, `IsUnmatched`, and `IsCanceled`. Use `Match(action)` when you want to handle one specific action.

```go
switch {
case result.Match(goHome):
	return openHome()
case result.IsPending():
	return waitForMoreKeys()
case result.IsCanceled():
	return clearStatus()
}
```

## Sequence Matching

A sequence is an ordered list of keys. `straw.TextSequence("gh")` creates the two-key sequence `g`, then `h`.

If the current input is both a complete binding and a prefix for a longer binding, the resolver returns `Pending` first. If no longer sequence arrives before the timeout, the pending exact match resolves.

For example, with bindings for `g` and `gh`:

1. Pressing `g` returns `Pending` because `gh` might still arrive.
2. Pressing `h` before the timeout returns `Matched` for `gh`.
3. Letting the timeout expire returns `Matched` for `g`.

## Pending Sequence Inspection

Use `PendingSequence()` to display the active prefix.

```go
resolver.UpdateKey(straw.Text("g"))
pending := resolver.PendingSequence()
```

The returned sequence is a copy. Mutating it does not change resolver state. `PendingSequence()` returns an empty sequence when the resolver is idle, reset, canceled, or resolved by timeout.

Use `NextChoices()` to list the immediate keys that can follow the current prefix.

```go
for _, choice := range resolver.NextChoices() {
	label := choice.Key.Label()
	if choice.HasBinding {
		fmt.Println(label, choice.Binding.Description())
	}
}
```

When the resolver is idle, `NextChoices()` returns root-level choices. When a prefix is pending, it returns choices under that prefix. Each row is one immediate next key, not every descendant binding. A choice can complete a binding, continue to longer bindings, or both.

## Timeout Behavior

Pending sequences use a timeout so a short binding can coexist with a longer binding that shares its prefix.

By default, the timeout is `500ms`. Customize it with `WithTimeout`.

```go
resolver, err := straw.New(bindings, straw.WithTimeout(250*time.Millisecond))
```

The root resolver returns timeout values from `UpdateKey`. Adapter resolvers return Bubble Tea commands that send the timeout message back into the application update loop.

## Reset Behavior

`Reset` clears pending input. It does not change bindings. Use it when a pending sequence belongs to an old context and should not affect the next key.

Common cases include switching screens, leaving a mode, closing a popup, or replacing the active keymap.

```go
resolver.Reset()
```

Any timeout token created before `Reset` becomes stale. If it is delivered later, `UpdateTimeout` returns `Idle`.

## Cancel Keys

By default, `esc` cancels a pending sequence. Configure cancel keys with `WithCancelKeys`.

```go
resolver, err := straw.New(bindings,
	straw.WithCancelKeys(straw.Code(straw.KeyEsc), straw.Modified('c', straw.ModCtrl)),
)
```

A cancel key only cancels pending input. If there is no pending sequence, normal host key handling can still decide what to do with that key.

## Pass-Through Behavior

`straw` only consumes keys that are part of resolver behavior. Host application key handling should run when the result can pass through.

With adapter packages, use `ShouldPassThrough(result)` before your normal Bubble Tea key switch.

```go
result, cmd := m.resolver.Update(msg)

if !straw.ShouldPassThrough(result) {
	return m, cmd
}

switch msg := msg.(type) {
case tea.KeyPressMsg:
	// normal host key handling
}
```

Failed keys after a pending prefix do not pass through by default. Use `WithFailedPendingPassThrough(true)` if your application wants those failed pending keys to be handled by the host application.

## Multiple Resolvers

Separate screens or child models can each own their own resolver. The root model can route messages to the active child instead of making one resolver handle every screen in the application.

This pattern keeps pending sequence state local to the screen that owns it. If you reuse one resolver while changing screens, call `Reset` when old pending input should not affect the next screen.

## Current Limitations

The resolver is intentionally small and predictable. These limits are part of the current design:

- Matching is exact. A text key, a special key, and a modified key are different shapes.
- Use `TextSequence("gh")` for multiple printable key presses. `Text("gh")` is invalid because `Text` represents one key press.
- Bubble Tea adapters ignore input that is not one supported key press, such as pasted text, multi-rune key messages, and key release messages.
- Lookup is simple and scales linearly with the number of bindings. This is fine for typical terminal keymaps.
- The resolver is mutable and is intended to be used from one update loop at a time.
- Binding analysis, file-based binding configuration, modes, contexts, enabled or disabled bindings, and continuation inspection are deferred.
