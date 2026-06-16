package v2_test

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	strawtea "github.com/kozikowskik/straw/bubbletea/v2"
)

func ExampleKey() {
	var key strawtea.Key = strawtea.Text("g")
	sequence := strawtea.Sequence(key)

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleSeq() {
	var sequence strawtea.Seq = strawtea.TextSequence("gh")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

func ExampleBinding() {
	binding := strawtea.Bind("go-home", strawtea.TextSequence("gh"))

	fmt.Println(binding.Action())

	// Output:
	// go-home
}

func ExampleBindingOption() {
	var option strawtea.BindingOption = strawtea.Description("go home")
	binding := strawtea.Bind("go-home", strawtea.TextSequence("gh"), option)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

func ExampleOption() {
	var option strawtea.Option = strawtea.WithTimeout(250 * time.Millisecond)
	resolver, err := strawtea.New[string](nil, option)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

func ExampleResult() {
	var result strawtea.Result[string]

	fmt.Printf("%T\n", result)

	// Output:
	// straw.Result[string]
}

func ExampleState() {
	var state strawtea.State = strawtea.Idle

	fmt.Println(state)

	// Output:
	// 0
}

func ExampleMod() {
	var mod strawtea.Mod = strawtea.ModCtrl | strawtea.ModAlt

	fmt.Println(mod != 0)

	// Output:
	// true
}

func ExampleText() {
	sequence := strawtea.Sequence(strawtea.Text("g"))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleTextSequence() {
	sequence := strawtea.TextSequence("gh")

	fmt.Println(len(sequence))

	// Output:
	// 2
}

func ExampleCode() {
	sequence := strawtea.Sequence(strawtea.Code(strawtea.KeyEsc))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleModified() {
	sequence := strawtea.Sequence(strawtea.Modified('c', strawtea.ModCtrl))

	fmt.Println(len(sequence))

	// Output:
	// 1
}

func ExampleSequence() {
	sequence := strawtea.Sequence(
		strawtea.Text("g"),
		strawtea.Code(strawtea.KeyEsc),
		strawtea.Modified('c', strawtea.ModCtrl),
	)

	fmt.Println(len(sequence))

	// Output:
	// 3
}

func ExampleBind() {
	binding := strawtea.Bind("go-home",
		strawtea.TextSequence("gh"),
		strawtea.Description("go home"),
	)

	fmt.Println(binding.Action())
	fmt.Println(binding.Description())

	// Output:
	// go-home
	// go home
}

func ExampleDescription() {
	binding := strawtea.Bind("go-home",
		strawtea.TextSequence("gh"),
		strawtea.Description("go home"),
	)

	fmt.Println(binding.Description())

	// Output:
	// go home
}

func ExampleWithTimeout() {
	resolver, err := strawtea.New[string](nil, strawtea.WithTimeout(250*time.Millisecond))

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

func ExampleWithCancelKeys() {
	resolver, err := strawtea.New[string](nil,
		strawtea.WithCancelKeys(strawtea.Code(strawtea.KeyEsc), strawtea.Modified('c', strawtea.ModCtrl)),
	)

	fmt.Println(err == nil)
	fmt.Println(resolver.Pending())

	// Output:
	// true
	// false
}

func ExampleWithFailedPendingPassThrough() {
	resolver, _ := strawtea.New([]strawtea.Binding[string]{
		strawtea.Bind("go-home", strawtea.TextSequence("gh")),
	}, strawtea.WithFailedPendingPassThrough(true))

	resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "x", Code: 'x'}))

	fmt.Println(result.IsUnmatched())
	fmt.Println(result.PassThrough())

	// Output:
	// true
	// true
}

func ExampleShouldPassThrough() {
	resolver, _ := strawtea.New[string](nil)
	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "x", Code: 'x'}))

	fmt.Println(strawtea.ShouldPassThrough(result))

	// Output:
	// true
}

func ExampleNew() {
	resolver, err := strawtea.New([]strawtea.Binding[string]{
		strawtea.Bind("go-home", strawtea.TextSequence("gh")),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

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

func ExampleResolver_Update() {
	resolver, _ := strawtea.New([]strawtea.Binding[string]{
		strawtea.Bind("go-home", strawtea.TextSequence("gh")),
	})

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	fmt.Println(result.IsPending())
	fmt.Println(cmd != nil)

	// Output:
	// true
	// true
}

func ExampleResolver_Reset() {
	resolver, _ := strawtea.New([]strawtea.Binding[string]{
		strawtea.Bind("go-home", strawtea.TextSequence("gh")),
	})
	resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	resolver.Reset()

	fmt.Println(resolver.Pending())

	// Output:
	// false
}

func ExampleResolver_Pending() {
	resolver, _ := strawtea.New[string](nil)

	fmt.Println(resolver.Pending())

	// Output:
	// false
}
