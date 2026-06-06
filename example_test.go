package straw_test

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/kozikowskik/straw"
)

func ExampleKey() {
	var key straw.Key = straw.Text("g")
	sequence := straw.Sequence(key)

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleSeq() {
	var sequence straw.Seq = straw.TextSequence("gh")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

func ExampleText() {
	sequence := straw.Sequence(straw.Text("g"))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleTextSequence() {
	sequence := straw.TextSequence("gé")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

func ExampleCode() {
	sequence := straw.Sequence(straw.Code(tea.KeyEsc))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleModified() {
	sequence := straw.Sequence(straw.Modified('c', tea.ModCtrl))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

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

func ExampleBinding() {
	type action int

	const goHome action = 1

	var binding straw.Binding[action] = straw.Bind(goHome, straw.TextSequence("gh"))

	fmt.Println(binding.Action())

	// Output:
	// 1
}

func ExampleBindingOption() {
	var option straw.BindingOption = straw.Description("go home")
	binding := straw.Bind("go-home", straw.TextSequence("gh"), option)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

func ExampleDescription() {
	binding := straw.Bind("go-home",
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

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

func ExampleBinding_Action() {
	binding := straw.Bind("go-home", straw.TextSequence("gh"))

	fmt.Println(binding.Action())

	// Output:
	// go-home
}

func ExampleBinding_Sequence() {
	binding := straw.Bind("go-home", straw.TextSequence("gh"))

	fmt.Println(len(binding.Sequence()))

	// Output:
	// 2
}

func ExampleBinding_Description() {
	binding := straw.Bind("go-home",
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

func ExampleErrInvalidBinding() {
	fmt.Println(errors.Is(straw.ErrInvalidBinding, straw.ErrInvalidBinding))

	// Output:
	// true
}

func ExampleErrInvalidKey() {
	fmt.Println(errors.Is(straw.ErrInvalidKey, straw.ErrInvalidKey))

	// Output:
	// true
}

func ExampleErrDuplicateSequence() {
	fmt.Println(errors.Is(straw.ErrDuplicateSequence, straw.ErrDuplicateSequence))

	// Output:
	// true
}

func ExampleErrInvalidOption() {
	_, err := straw.New[string](nil, straw.WithTimeout(0))

	fmt.Println(errors.Is(err, straw.ErrInvalidOption))

	// Output:
	// true
}

func ExampleOption() {
	var option straw.Option = straw.WithTimeout(250 * time.Millisecond)
	resolver, err := straw.New[string](nil, option)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

func ExampleWithTimeout() {
	resolver, err := straw.New[string](nil, straw.WithTimeout(250*time.Millisecond))

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

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

func ExampleWithFailedPendingPassThrough() {
	resolver, err := straw.New[string](nil, straw.WithFailedPendingPassThrough(true))

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

func ExampleResolver() {
	resolver, err := straw.New[string](nil)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

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

func ExampleResolver_Update() {
	resolver, _ := straw.New[string](nil)
	result, cmd := resolver.Update(struct{}{})

	fmt.Printf("%T\n", result)
	fmt.Println(cmd == nil)

	// Output:
	// straw.Result[string]
	// true
}

func ExampleResolver_Reset() {
	resolver, _ := straw.New[string](nil)
	resolver.Reset()

	fmt.Println(resolver.Pending())

	// Output:
	// false
}

func ExampleResolver_Pending() {
	resolver, _ := straw.New[string](nil)

	fmt.Println(resolver.Pending())

	// Output:
	// false
}

func ExampleState() {
	var state straw.State = straw.Idle

	fmt.Println(state)

	// Output:
	// 0
}

func ExampleIdle() {
	fmt.Println(straw.Idle)

	// Output:
	// 0
}

func ExamplePending() {
	fmt.Println(straw.Pending)

	// Output:
	// 1
}

func ExampleMatched() {
	fmt.Println(straw.Matched)

	// Output:
	// 2
}

func ExampleUnmatched() {
	fmt.Println(straw.Unmatched)

	// Output:
	// 3
}

func ExampleCanceled() {
	fmt.Println(straw.Canceled)

	// Output:
	// 4
}

func ExampleResult() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Printf("%T\n", result)

	// Output:
	// straw.Result[string]
}

func ExampleResult_Match() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.Match("go-home"))

	// Output:
	// false
}

func ExampleResult_Binding() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})
	_, ok := result.Binding()

	fmt.Println(ok)

	// Output:
	// false
}

func ExampleResult_State() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.State())

	// Output:
	// 0
}

func ExampleResult_IsIdle() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsIdle())

	// Output:
	// true
}

func ExampleResult_IsPending() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsPending())

	// Output:
	// false
}

func ExampleResult_IsMatched() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsMatched())

	// Output:
	// false
}

func ExampleResult_IsUnmatched() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsUnmatched())

	// Output:
	// false
}

func ExampleResult_IsCanceled() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.IsCanceled())

	// Output:
	// false
}

func ExampleResult_PassThrough() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(result.PassThrough())

	// Output:
	// false
}

func ExampleResult_Key() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Printf("%T\n", result.Key())

	// Output:
	// straw.Key
}

func ExampleResult_Sequence() {
	resolver, _ := straw.New[string](nil)
	result, _ := resolver.Update(struct{}{})

	fmt.Println(len(result.Sequence()))

	// Output:
	// 0
}
