package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	straw "github.com/kozikowskik/straw/bubbletea/v2"
)

type action int

const (
	goHome action = iota + 1
	goHelp
)

type model struct {
	resolver *straw.Resolver[action]
	message  string
}

func newModel() (model, error) {
	resolver, err := straw.New(
		[]straw.Binding[action]{
			straw.Bind(goHome, straw.TextSequence("g"), straw.Description("go home after timeout")),
			straw.Bind(goHelp, straw.TextSequence("gh"), straw.Description("go help immediately")),
		},
		straw.WithTimeout(750*time.Millisecond),
		straw.WithCancelKeys(straw.Code(straw.KeyEsc), straw.Modified('c', straw.ModCtrl)),
	)
	if err != nil {
		return model{}, err
	}

	return model{resolver: resolver, message: "Press g, then wait for timeout, or press gh. Esc and ctrl+c cancel pending input."}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.resolver.Update(msg)

	switch {
	case result.Match(goHome):
		m.message = "timeout resolved: go home"
		return m, cmd
	case result.Match(goHelp):
		m.message = "matched: go help"
		return m, cmd
	case result.IsPending():
		m.message = "pending: press h before timeout, or esc to cancel"
	case result.IsCanceled():
		m.message = "pending sequence canceled"
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
	return tea.NewView(fmt.Sprintf("%s\n\nBindings:\n  g   go home after 750ms timeout\n  gh  go help immediately\n\nCancel pending input:\n  esc\n  ctrl+c\n", m.message))
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
