# Bindings

A binding maps one application-owned action to one key sequence. `straw` does not define your actions or execute them. Your application owns both the action type and the behavior attached to each matched action.

## Action Types

Use any comparable type for actions. Small custom types usually make application code easier to read than raw strings.

```go
type action int

const (
	goHome action = iota + 1
	goDashboard
)
```

Strings are also valid for small examples or simple applications.

```go
straw.Bind("go-home", straw.TextSequence("gh"))
```

## Text Keys

Use `Text` for one printable text key and `TextSequence` for a string of printable text keys.

```go
straw.Text("g")
straw.TextSequence("gh")
```

`TextSequence` works with Unicode text by splitting the string into runes.

```go
sequence := straw.TextSequence("gé")
```

`Text` is for one key press. Use `Text("g")`, not `Text("gh")`. Use `TextSequence("gh")` when the user should press `g` and then `h`.

## Special Keys

Use `Code` for non-printable keys such as escape, enter, tab, and arrows.

```go
straw.Sequence(straw.Code(straw.KeyEsc))
```

Special key constants live in the root package and are re-exported by the adapter packages.

## Modified Keys

Use `Modified` for keys pressed with modifiers such as control or alt.

```go
straw.Sequence(straw.Modified('c', straw.ModCtrl))
straw.Sequence(straw.Modified('x', straw.ModAlt))
```

Modifier constants include values such as `ModCtrl` and `ModAlt`.

## Sequences And Chords

Use `Sequence` when you need to combine text, special, and modified keys in one binding.

```go
sequence := straw.Sequence(
	straw.Text("g"),
	straw.Code(straw.KeyEnter),
	straw.Modified('c', straw.ModCtrl),
)
```

Use `TextSequence` for the common case where every key is printable text.

```go
sequence := straw.TextSequence("gd")
```

In this documentation, a sequence means ordered key presses. A chord means one key press with modifiers, such as `ctrl+c`.

There is no separate chord builder. Use `Modified` for chord-like keys.

## Binding Metadata

Use `Description` to attach human-readable metadata to a binding. This is useful for help screens and command palettes.

```go
binding := straw.Bind(goHome,
	straw.TextSequence("gh"),
	straw.Description("go home"),
)
```

Read metadata back with `Description()`.

```go
fmt.Println(binding.Description())
```

## Validation

`New` validates bindings and returns an error for invalid keys, empty sequences, duplicate sequences, and invalid options.

These key shapes are invalid:

- `Text("")`, because it is empty.
- `Text("gh")`, because it contains more than one rune.
- `Text(" ")`, because whitespace is not a printable text binding key.
- `Code('g')`, because printable keys should use `Text("g")`.

Prefer the builders that match how the user presses the key. Use `TextSequence` for normal letters, `Code` for special keys, and `Modified` for keys such as `ctrl+c`.

```go
resolver, err := straw.New(bindings)
if err != nil {
	return err
}

_ = resolver
```

Check exported errors with `errors.Is` when your application wants to handle one category differently.

```go
if errors.Is(err, straw.ErrDuplicateSequence) {
	return fmt.Errorf("duplicate key sequence: %w", err)
}
```
