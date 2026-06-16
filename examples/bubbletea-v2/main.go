package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	straw "github.com/kozikowskik/straw/bubbletea/v2"
)

type action int

const (
	goHome action = iota + 1
	goDashboard
)

type model struct {
	resolver *straw.Resolver[action]
	message  string
}

func newModel() (model, error) {
	resolver, err := straw.New([]straw.Binding[action]{
		straw.Bind(goHome, straw.TextSequence("gh"), straw.Description("go home")),
		straw.Bind(goDashboard, straw.TextSequence("gd"), straw.Description("go dashboard")),
	})
	if err != nil {
		return model{}, err
	}

	return model{resolver: resolver, message: "Press gh or gd. Press q or ctrl+c to quit."}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

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

func (m model) View() tea.View {
	return tea.NewView(fmt.Sprintf("%s\n\nBindings:\n  gh  go home\n  gd  go dashboard\n", m.message))
}

func main() {
	m, err := newModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build model: %v\n", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run program: %v\n", err)
		os.Exit(1)
	}
}
