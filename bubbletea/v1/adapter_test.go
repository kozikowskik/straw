package v1

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kozikowskik/straw"
)

type testAction int

const (
	testGoHome testAction = iota + 1
	testCopyLine
)

func TestUpdateMatchesTextSequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if !result.IsPending() || cmd == nil {
		t.Fatalf("after g pending/cmd = %v/%v, want true/non-nil", result.IsPending(), cmd != nil)
	}

	result, cmd = resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if !result.Match(testGoHome) || cmd != nil {
		t.Fatalf("after h match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
	}
}

func TestUpdateMatchesSpecialCtrlAndAltKeys(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyMsg
	}{
		{
			name:    "special key",
			binding: Bind(testGoHome, Sequence(Code(straw.KeyEsc))),
			message: tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name:    "ctrl key",
			binding: Bind(testGoHome, Sequence(Modified('c', straw.ModCtrl))),
			message: tea.KeyMsg{Type: tea.KeyCtrlC},
		},
		{
			name:    "alt text key",
			binding: Bind(testGoHome, Sequence(Modified('g', straw.ModAlt))),
			message: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}, Alt: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New([]Binding[testAction]{tt.binding})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, cmd := resolver.Update(tt.message)
			if !result.Match(testGoHome) || cmd != nil {
				t.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
			}
		})
	}
}

func TestUpdateMatchesNormalTabAndEnterAsSpecialKeys(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyMsg
	}{
		{
			name:    "tab",
			binding: Bind(testGoHome, Sequence(Code(straw.KeyTab))),
			message: tea.KeyMsg{Type: tea.KeyTab},
		},
		{
			name:    "enter",
			binding: Bind(testGoHome, Sequence(Code(straw.KeyEnter))),
			message: tea.KeyMsg{Type: tea.KeyEnter},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New([]Binding[testAction]{tt.binding})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, cmd := resolver.Update(tt.message)
			if !result.Match(testGoHome) || cmd != nil {
				t.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
			}
		})
	}
}

func TestUpdateV1ModifierBehavior(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyMsg
	}{
		{
			name:    "uppercase text is text, not shift modifier",
			binding: Bind(testGoHome, Sequence(Text("G"))),
			message: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}},
		},
		{
			name:    "alt enter is modified special key",
			binding: Bind(testGoHome, Sequence(Modified(straw.KeyEnter, straw.ModAlt))),
			message: tea.KeyMsg{Type: tea.KeyEnter, Alt: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New([]Binding[testAction]{tt.binding})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, cmd := resolver.Update(tt.message)
			if !result.Match(testGoHome) || cmd != nil {
				t.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
			}
		})
	}
}

func TestPasteAndMultiRuneMessagesAreIgnored(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, msg := range []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}, Paste: true},
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g', 'h'}},
	} {
		result, cmd := resolver.Update(msg)
		if !result.IsIdle() || cmd != nil || resolver.Pending() {
			t.Fatalf("Update(%#v) idle/cmdNil/pending = %v/%v/%v, want true/true/false", msg, result.IsIdle(), cmd == nil, resolver.Pending())
		}
	}
}

func TestTimeoutCommandResolvesPendingMatch(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{
			Bind(testGoHome, TextSequence("g")),
			Bind(testCopyLine, TextSequence("gh")),
		},
		WithTimeout(time.Nanosecond),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cmd == nil {
		t.Fatal("Update() cmd = nil, want timeout command")
	}

	result, next := resolver.Update(cmd())
	if !result.Match(testGoHome) || next != nil || resolver.Pending() {
		t.Fatalf("timeout match/nextNil/pending = %v/%v/%v, want true/true/false", result.Match(testGoHome), next == nil, resolver.Pending())
	}
}

func TestDefaultCancelKeyCancelsPendingSequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	result, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !result.IsCanceled() || cmd != nil || resolver.Pending() {
		t.Fatalf("esc canceled/cmdNil/pending = %v/%v/%v, want true/true/false", result.IsCanceled(), cmd == nil, resolver.Pending())
	}
}

func TestConfiguredCancelKeyCancelsPendingSequence(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
		WithCancelKeys(Modified('c', straw.ModCtrl)),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	result, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !result.IsCanceled() || cmd != nil || resolver.Pending() {
		t.Fatalf("ctrl+c canceled/cmdNil/pending = %v/%v/%v, want true/true/false", result.IsCanceled(), cmd == nil, resolver.Pending())
	}
}

func TestResetClearsPendingState(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	resolver.Reset()
	if resolver.Pending() {
		t.Fatal("Pending() = true after Reset(), want false")
	}

	result, cmd := resolver.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if !result.IsUnmatched() || !result.PassThrough() || cmd != nil {
		t.Fatalf("after reset h unmatched/passThrough/cmdNil = %v/%v/%v, want true/true/true", result.IsUnmatched(), result.PassThrough(), cmd == nil)
	}
}
