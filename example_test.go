package straw_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/kozikowskik/straw"
)

// ExampleBind demonstrates connecting an application action to a key sequence.
func ExampleBind() {
	type action string

	binding := straw.Bind(action("go-home"),
		straw.TextSequence("gh"),
		straw.Description("go home"),
	)

	fmt.Println(binding.Action())
	fmt.Println(binding.Description())
	fmt.Println(len(binding.Sequence()))

	// Output:
	// go-home
	// go home
	// 2
}

// ExampleNew demonstrates creating a resolver from bindings.
func ExampleNew() {
	binding := straw.Bind("go-home", straw.TextSequence("gh"))
	resolver, err := straw.New([]straw.Binding[string]{binding})

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleNew_validation demonstrates checking resolver construction errors.
func ExampleNew_validation() {
	bindings := []straw.Binding[string]{
		straw.Bind("first", straw.TextSequence("gh")),
		straw.Bind("second", straw.TextSequence("gh")),
	}
	_, err := straw.New(bindings)

	fmt.Println(errors.Is(err, straw.ErrDuplicateSequence))

	// Output:
	// true
}

// ExampleWithCancelKeys demonstrates canceling pending input with configured keys.
func ExampleWithCancelKeys() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	}, straw.WithCancelKeys(straw.Code(straw.KeyEsc)))

	resolver.UpdateKey(straw.Text("g"))
	result, timeout := resolver.UpdateKey(straw.Code(straw.KeyEsc))

	fmt.Println(result.IsCanceled())
	fmt.Println(timeout.Scheduled())
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
	// false
}

// ExampleWithFailedPendingPassThrough demonstrates passing failed pending input back to the host.
func ExampleWithFailedPendingPassThrough() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	}, straw.WithFailedPendingPassThrough(true))

	resolver.UpdateKey(straw.Text("g"))
	result, _ := resolver.UpdateKey(straw.Text("x"))

	fmt.Println(result.IsUnmatched())
	fmt.Println(result.PassThrough())

	// Output:
	// true
	// true
}

// ExampleResolver_UpdateKey demonstrates matching a multi-key sequence.
func ExampleResolver_UpdateKey() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	result, timeout := resolver.UpdateKey(straw.Text("g"))
	fmt.Println(result.IsPending())
	fmt.Println(timeout.Scheduled())

	result, timeout = resolver.UpdateKey(straw.Text("h"))
	fmt.Println(result.Match("go-home"))
	fmt.Println(timeout.Scheduled())

	// Output:
	// true
	// true
	// true
	// false
}

// ExampleResolver_UpdateKey_unmatched demonstrates returning unmatched input to normal host handling.
func ExampleResolver_UpdateKey_unmatched() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	result, timeout := resolver.UpdateKey(straw.Text("j"))

	fmt.Println(result.IsUnmatched())
	fmt.Println(result.PassThrough())
	fmt.Println(timeout.Scheduled())

	// Output:
	// true
	// true
	// false
}

// ExampleResolver_UpdateTimeout demonstrates resolving an ambiguous short binding.
func ExampleResolver_UpdateTimeout() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("g")),
		straw.Bind("go-help", straw.TextSequence("gh")),
	})

	_, timeout := resolver.UpdateKey(straw.Text("g"))
	result := resolver.UpdateTimeout(timeout)

	fmt.Println(result.Match("go-home"))
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleTimeout demonstrates scheduling and resolving a pending-sequence timeout.
func ExampleTimeout() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("g")),
		straw.Bind("go-help", straw.TextSequence("gh")),
	}, straw.WithTimeout(250*time.Millisecond))

	_, timeout := resolver.UpdateKey(straw.Text("g"))

	fmt.Println(timeout.Scheduled())
	fmt.Println(timeout.Duration())
	fmt.Println(resolver.UpdateTimeout(timeout).Match("go-home"))

	// Output:
	// true
	// 250ms
	// true
}

// ExampleResolver_Reset demonstrates clearing pending input when context changes.
func ExampleResolver_Reset() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	resolver.UpdateKey(straw.Text("g"))
	fmt.Println(resolver.Pending())

	resolver.Reset()
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

// ExampleShouldPassThrough demonstrates deciding whether normal host key handling should run.
func ExampleShouldPassThrough() {
	resolver, _ := straw.New([]straw.Binding[string]{
		straw.Bind("go-home", straw.TextSequence("gh")),
	})

	result, _ := resolver.UpdateKey(straw.Text("x"))

	fmt.Println(straw.ShouldPassThrough(result))

	// Output:
	// true
}
