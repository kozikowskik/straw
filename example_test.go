package straw_test

import (
	"fmt"

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
