package v2

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
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

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	if !result.IsPending() || cmd == nil {
		t.Fatalf("after g pending/cmd = %v/%v, want true/non-nil", result.IsPending(), cmd != nil)
	}

	result, cmd = resolver.Update(tea.KeyPressMsg(tea.Key{Text: "h", Code: 'h'}))
	if !result.Match(testGoHome) || cmd != nil {
		t.Fatalf("after h match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
	}
}

func TestResolverQueryAPIsForwardToCore(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("gh"), Description("go home")),
		Bind(testCopyLine, TextSequence("gd"), Description("go dashboard")),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	assertSeqEqual(t, resolver.PendingSequence(), TextSequence("g"))

	choices := resolver.NextChoices()
	if len(choices) != 2 {
		t.Fatalf("NextChoices() length = %d, want 2", len(choices))
	}
	if choices[0].Key != Text("h") || !choices[0].HasBinding || choices[0].Binding.Description() != "go home" {
		t.Fatalf("first choice = %#v, want h binding go home", choices[0])
	}
	if choices[1].Key != Text("d") || !choices[1].HasBinding || choices[1].Binding.Description() != "go dashboard" {
		t.Fatalf("second choice = %#v, want d binding go dashboard", choices[1])
	}
}

func assertSeqEqual(t *testing.T, got Seq, want Seq) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("sequence length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestUpdateMatchesSpecialAndModifiedKeys(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyPressMsg
	}{
		{
			name:    "special key",
			binding: Bind(testGoHome, Sequence(Code(straw.KeyEsc))),
			message: tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}),
		},
		{
			name:    "modified key",
			binding: Bind(testGoHome, Sequence(Modified('c', straw.ModCtrl))),
			message: tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}),
		},
		{
			name:    "combined modifiers",
			binding: Bind(testGoHome, Sequence(Modified('c', straw.ModCtrl|straw.ModAlt))),
			message: tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl | tea.ModAlt}),
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

func TestUpdateV2ModifierBehavior(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyPressMsg
	}{
		{
			name:    "uppercase text without modifier is text",
			binding: Bind(testGoHome, Sequence(Text("G"))),
			message: tea.KeyPressMsg(tea.Key{Text: "G", Code: 'G'}),
		},
		{
			name:    "shifted printable text with explicit shift is text",
			binding: Bind(testGoHome, Sequence(Text("A"))),
			message: tea.KeyPressMsg(tea.Key{Text: "A", Code: 'a', Mod: tea.ModShift}),
		},
		{
			name:    "shift tab is modified special key",
			binding: Bind(testGoHome, Sequence(Modified(straw.KeyTab, straw.ModShift))),
			message: tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}),
		},
		{
			name:    "alt enter is modified special key",
			binding: Bind(testGoHome, Sequence(Modified(straw.KeyEnter, straw.ModAlt))),
			message: tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModAlt}),
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

func TestHasModifiersReportsWhetherAnyModifierBitIsSet(t *testing.T) {
	if hasModifiers(0) {
		t.Fatal("hasModifiers(0) = true, want false")
	}
	if !hasModifiers(tea.ModShift) {
		t.Fatal("hasModifiers(tea.ModShift) = false, want true")
	}
	if !hasModifiers(tea.ModCtrl | tea.ModAlt) {
		t.Fatal("hasModifiers(tea.ModCtrl | tea.ModAlt) = false, want true")
	}
}

func TestUpdateIgnoresNonKeyAndKeyReleaseMessages(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))

	for _, msg := range []tea.Msg{struct{}{}, tea.KeyReleaseMsg(tea.Key{Text: "h", Code: 'h'})} {
		result, cmd := resolver.Update(msg)
		if !result.IsIdle() || cmd != nil || !resolver.Pending() {
			t.Fatalf("Update(%T) idle/cmdNil/pending = %v/%v/%v, want true/true/true", msg, result.IsIdle(), cmd == nil, resolver.Pending())
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

	_, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	if cmd == nil {
		t.Fatal("Update() cmd = nil, want timeout command")
	}

	result, next := resolver.Update(cmd())
	if !result.Match(testGoHome) || next != nil || resolver.Pending() {
		t.Fatalf("timeout match/nextNil/pending = %v/%v/%v, want true/true/false", result.Match(testGoHome), next == nil, resolver.Pending())
	}
}

func TestStaleAndForeignTimeoutMessagesAreIgnored(t *testing.T) {
	first, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("first New() error = %v", err)
	}
	second, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("second New() error = %v", err)
	}

	_, firstCmd := first.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	second.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))

	result, cmd := second.Update(firstCmd())
	if !result.IsIdle() || cmd != nil || !second.Pending() {
		t.Fatalf("foreign timeout idle/cmdNil/pending = %v/%v/%v, want true/true/true", result.IsIdle(), cmd == nil, second.Pending())
	}
}

func TestDefaultCancelKeyCancelsPendingSequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
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

	_, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if !result.IsCanceled() || cmd != nil || resolver.Pending() {
		t.Fatalf("ctrl+c canceled/cmdNil/pending = %v/%v/%v, want true/true/false", result.IsCanceled(), cmd == nil, resolver.Pending())
	}
}

func TestResetClearsPendingState(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'}))
	if cmd == nil || !resolver.Pending() {
		t.Fatalf("after g cmd/pending = %v/%v, want non-nil/true", cmd != nil, resolver.Pending())
	}

	resolver.Reset()
	if resolver.Pending() {
		t.Fatal("Pending() = true after Reset(), want false")
	}

	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Text: "h", Code: 'h'}))
	if !result.IsUnmatched() || !result.PassThrough() || cmd != nil {
		t.Fatalf("after reset h unmatched/passThrough/cmdNil = %v/%v/%v, want true/true/true", result.IsUnmatched(), result.PassThrough(), cmd == nil)
	}
}
