package straw

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func ExampleBind() {
	type action int

	const goHome action = 1

	binding := Bind(goHome,
		TextSequence("gh"),
		Description("go home"),
	)

	fmt.Println(binding.Action())
	fmt.Println(binding.Description())
	fmt.Println(len(binding.Sequence()))

	// Output:
	// 1
	// go home
	// 2
}

func ExampleSequence() {
	sequence := Sequence(
		Text("g"),
		Code(tea.KeyEsc),
		Modified('c', tea.ModCtrl),
	)

	fmt.Println(len(sequence))

	// Output:
	// 3
}
