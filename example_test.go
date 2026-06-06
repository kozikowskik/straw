package straw_test

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/kozikowskik/straw"
)

// ExampleKey demonstrates using Key as the public key value type.
func ExampleKey() {
	var key straw.Key = straw.Text("g")
	sequence := straw.Sequence(key)

	fmt.Println(len(sequence))

	// Output:
	// 1
}

// ExampleSeq demonstrates using Seq as the public sequence value type.
func ExampleSeq() {
	var sequence straw.Seq = straw.TextSequence("gh")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

// ExampleText demonstrates building a printable text key.
func ExampleText() {
	sequence := straw.Sequence(straw.Text("g"))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

// ExampleTextSequence demonstrates building a text-only key sequence.
func ExampleTextSequence() {
	sequence := straw.TextSequence("gé")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

// ExampleCode demonstrates building a special Bubble Tea key.
func ExampleCode() {
	sequence := straw.Sequence(straw.Code(tea.KeyEsc))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

// ExampleModified demonstrates building a modified key.
func ExampleModified() {
	sequence := straw.Sequence(straw.Modified('c', tea.ModCtrl))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

// ExampleSequence demonstrates composing mixed key types into one sequence.
func ExampleSequence() {
	sequence := straw.Sequence(
		straw.Text("g"),
		straw.Code(tea.KeyEsc),
		straw.Modified('c', tea.ModCtrl),
	)

	fmt.Println(len(sequence))

	// Output:
	// 3
}

// ExampleBinding demonstrates using Binding as the public binding value type.
func ExampleBinding() {
	type action int

	const goHome action = 1

	var binding straw.Binding[action] = straw.Bind(goHome, straw.TextSequence("gh"))

	fmt.Println(binding.Action())

	// Output:
	// 1
}

// ExampleBindingOption demonstrates passing optional binding metadata.
func ExampleBindingOption() {
	var option straw.BindingOption = straw.Description("go home")
	binding := straw.Bind("go-home", straw.TextSequence("gh"), option)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

// ExampleDescription demonstrates attaching human-readable binding metadata.
func ExampleDescription() {
	binding := straw.Bind("go-home",
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

// ExampleBind demonstrates connecting an action to a key sequence.
func ExampleBind() {
	type action int

	const goHome action = 1

	binding := straw.Bind(goHome,
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Action())
	fmt.Println(binding.Description())
	fmt.Println(len(binding.Sequence()))

	// Output:
	// 1
	// go home
	// 2
}

// ExampleBinding_Action demonstrates reading the action stored in a binding.
func ExampleBinding_Action() {
	binding := straw.Bind("go-home", straw.TextSequence("gh"))

	fmt.Println(binding.Action())

	// Output:
	// go-home
}

// ExampleBinding_Sequence demonstrates reading the sequence stored in a binding.
func ExampleBinding_Sequence() {
	binding := straw.Bind("go-home", straw.TextSequence("gh"))

	fmt.Println(len(binding.Sequence()))

	// Output:
	// 2
}

// ExampleBinding_Description demonstrates reading binding metadata.
func ExampleBinding_Description() {
	binding := straw.Bind("go-home",
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

// ExampleErrInvalidBinding demonstrates checking for invalid binding errors.
func ExampleErrInvalidBinding() {
	fmt.Println(errors.Is(straw.ErrInvalidBinding, straw.ErrInvalidBinding))

	// Output:
	// true
}

// ExampleErrInvalidKey demonstrates checking for invalid key errors.
func ExampleErrInvalidKey() {
	fmt.Println(errors.Is(straw.ErrInvalidKey, straw.ErrInvalidKey))

	// Output:
	// true
}

// ExampleErrDuplicateSequence demonstrates checking for duplicate sequence errors.
func ExampleErrDuplicateSequence() {
	fmt.Println(errors.Is(straw.ErrDuplicateSequence, straw.ErrDuplicateSequence))

	// Output:
	// true
}

// ExampleErrInvalidOption demonstrates checking for invalid resolver options.
func ExampleErrInvalidOption() {
	_, err := straw.New[string](nil, straw.WithTimeout(0))

	fmt.Println(errors.Is(err, straw.ErrInvalidOption))

	// Output:
	// true
}

// ExampleOption demonstrates passing an option value to New.
func ExampleOption() {
	var option straw.Option = straw.WithTimeout(250 * time.Millisecond)
	resolver, err := straw.New[string](nil, option)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleWithTimeout demonstrates configuring pending-sequence timeout duration.
func ExampleWithTimeout() {
	resolver, err := straw.New[string](nil, straw.WithTimeout(250*time.Millisecond))

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleWithCancelKeys demonstrates configuring keys that cancel pending sequences.
func ExampleWithCancelKeys() {
	resolver, err := straw.New[string](nil,
		straw.WithCancelKeys(straw.Code(tea.KeyEsc), straw.Modified('c', tea.ModCtrl)),
	)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleWithFailedPendingPassThrough demonstrates configuring failed-pending pass-through behavior.
func ExampleWithFailedPendingPassThrough() {
	resolver, err := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	}, straw.WithFailedPendingPassThrough(true))
	if err != nil {
		fmt.Println(err)
		return
	}

	resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "x", Code: 'x'}))
	fmt.Println(result.IsUnmatched())
	fmt.Println(result.PassThrough())

	// Output:
	// true
	// true
}

// ExampleResolver demonstrates constructing a resolver.
func ExampleResolver() {
	resolver, err := straw.New[string](nil)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleNew demonstrates creating a resolver from bindings.
func ExampleNew() {
	bindings := []straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	}
	resolver, err := straw.New(bindings)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleResolver_Update demonstrates updating a resolver with key press messages.
func ExampleResolver_Update() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	fmt.Println(result.IsPending())
	fmt.Println(cmd != nil)

	result, cmd = resolver.Update(tea.KeyPressMsg(tea.Key{Text: "h", Code: 'h'}))
	fmt.Println(result.Match("go-home"))
	fmt.Println(cmd == nil)

	// Output:
	// true
	// true
	// true
	// true
}

// ExampleResolver_Update_unmatched demonstrates passing unmatched input back to the host.
func ExampleResolver_Update_unmatched() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "j", Code: 'j'}))
	fmt.Println(result.IsUnmatched())
	fmt.Println(result.PassThrough())
	fmt.Println(cmd == nil)

	// Output:
	// true
	// true
	// true
}

// ExampleResolver_Reset demonstrates clearing resolver pending state.
func ExampleResolver_Reset() {
	resolver, _ := straw.New[string](nil)
	resolver.Reset()

	fmt.Println(resolver.Pending())

	// Output:
	// false
}

// ExampleResolver_Pending demonstrates checking whether a resolver is pending.
func ExampleResolver_Pending() {
	resolver, _ := straw.New[string](nil)

	fmt.Println(resolver.Pending())

	// Output:
	// false
}

// ExampleState demonstrates using State as the public result-state type.
func ExampleState() {
	var state straw.State = straw.Idle

	fmt.Println(state)

	// Output:
	// 0
}

// ExampleIdle demonstrates the idle result state.
func ExampleIdle() {
	fmt.Println(straw.Idle)

	// Output:
	// 0
}

// ExamplePending demonstrates the pending result state.
func ExamplePending() {
	fmt.Println(straw.Pending)

	// Output:
	// 1
}

// ExampleMatched demonstrates the matched result state.
func ExampleMatched() {
	fmt.Println(straw.Matched)

	// Output:
	// 2
}

// ExampleUnmatched demonstrates the unmatched result state.
func ExampleUnmatched() {
	fmt.Println(straw.Unmatched)

	// Output:
	// 3
}

// ExampleCanceled demonstrates the canceled result state.
func ExampleCanceled() {
	fmt.Println(straw.Canceled)

	// Output:
	// 4
}

// ExampleResult demonstrates using Result as the public update outcome type.
func ExampleResult() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Printf("%T\n", result)

	// Output:
	// straw.Result[string]
}

// ExampleResult_Match demonstrates checking whether a result matched an action.
func ExampleResult_Match() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.Match("go-home"))

	// Output:
	// false
}

// ExampleResult_Binding demonstrates retrieving a matched binding when present.
func ExampleResult_Binding() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})
	_, ok := result.Binding()

	fmt.Println(ok)

	// Output:
	// false
}

// ExampleResult_State demonstrates reading the result state.
func ExampleResult_State() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.State())

	// Output:
	// 0
}

// ExampleResult_IsIdle demonstrates checking for idle results.
func ExampleResult_IsIdle() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsIdle())

	// Output:
	// true
}

// ExampleResult_IsPending demonstrates checking for pending results.
func ExampleResult_IsPending() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsPending())

	// Output:
	// false
}

// ExampleResult_IsMatched demonstrates checking for matched results.
func ExampleResult_IsMatched() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsMatched())

	// Output:
	// false
}

// ExampleResult_IsUnmatched demonstrates checking for unmatched results.
func ExampleResult_IsUnmatched() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsUnmatched())

	// Output:
	// false
}

// ExampleResult_IsCanceled demonstrates checking for canceled results.
func ExampleResult_IsCanceled() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsCanceled())

	// Output:
	// false
}

// ExampleResult_PassThrough demonstrates checking host pass-through behavior.
func ExampleResult_PassThrough() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.PassThrough())

	// Output:
	// false
}

// ExampleResult_Key demonstrates retrieving the key associated with a result.
func ExampleResult_Key() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Printf("%T\n", result.Key())

	// Output:
	// straw.Key
}

// ExampleResult_Sequence demonstrates retrieving the sequence associated with a result.
func ExampleResult_Sequence() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(len(result.Sequence()))

	// Output:
	// 0
}
