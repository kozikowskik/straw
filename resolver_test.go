package straw

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestNewAcceptsEmptyBindings verifies empty binding lists create an inert resolver.
func TestNewAcceptsEmptyBindings(t *testing.T) {
	tests := []struct {
		name     string
		bindings []Binding[testAction]
	}{
		{name: "nil", bindings: nil},
		{name: "empty slice", bindings: []Binding[testAction]{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New(tt.bindings)
			if err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
			if resolver == nil {
				t.Fatal("New() resolver = nil, want resolver")
			}
			if resolver.Pending() {
				t.Fatal("Pending() = true, want false")
			}
		})
	}
}

// TestNewRejectsInvalidTimeout verifies invalid timeout values are rejected with ErrInvalidOption.
func TestNewRejectsInvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "zero", timeout: 0},
		{name: "negative", timeout: -time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New[testAction](nil, WithTimeout(tt.timeout))
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
			}
		})
	}
}

// TestNewAcceptsResolverOptions verifies supported options can be applied during construction.
func TestNewAcceptsResolverOptions(t *testing.T) {
	resolver, err := New[testAction](nil,
		WithTimeout(250*time.Millisecond),
		WithCancelKeys(Code(tea.KeyEsc), Modified('c', tea.ModCtrl)),
		WithFailedPendingPassThrough(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if resolver == nil {
		t.Fatal("New() resolver = nil, want resolver")
	}
}

// TestNewAcceptsSpaceCodeBinding verifies space is represented with Code rather than Text.
func TestNewAcceptsSpaceCodeBinding(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, Sequence(Code(tea.KeySpace))),
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if resolver == nil {
		t.Fatal("New() resolver = nil, want resolver")
	}
}

// TestNewRejectsInvalidBindings verifies binding validation rejects invalid configuration.
func TestNewRejectsInvalidBindings(t *testing.T) {
	tests := []struct {
		name     string
		bindings []Binding[testAction]
		wantErr  error
	}{
		{
			name:     "empty sequence",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence())},
			wantErr:  ErrInvalidBinding,
		},
		{
			name:     "empty text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "multi-rune text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("gg")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "whitespace text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text(" ")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unicode whitespace text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("\u00a0")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "printable code key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Code('g')))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unicode printable code key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Code('é')))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unknown key kind",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Key{}))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "modified key without modifier",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Modified('c', 0)))},
			wantErr:  ErrInvalidKey,
		},
		{
			name: "duplicate sequence",
			bindings: []Binding[testAction]{
				Bind(testGoHome, TextSequence("gh")),
				Bind(testCopyLine, TextSequence("gh")),
			},
			wantErr: ErrDuplicateSequence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.bindings)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewCollectsMultipleValidationErrors verifies joined validation errors remain inspectable.
func TestNewCollectsMultipleValidationErrors(t *testing.T) {
	_, err := New([]Binding[testAction]{
		Bind(testGoHome, Sequence()),
		Bind(testCopyLine, Sequence(Text(""))),
		Bind(testGoHome, TextSequence("gh")),
		Bind(testCopyLine, TextSequence("gh")),
	})
	if !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("New() error = %v, want ErrInvalidBinding", err)
	}
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("New() error = %v, want ErrInvalidKey", err)
	}
	if !errors.Is(err, ErrDuplicateSequence) {
		t.Fatalf("New() error = %v, want ErrDuplicateSequence", err)
	}
}

// TestResolverMatchesSimpleSequence verifies a multi-key sequence moves from pending to matched.
func TestResolverMatchesSimpleSequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(keyPress("g"))
	if cmd == nil {
		t.Fatal("first Update() cmd = nil, want timeout command")
	}
	if !result.IsPending() || !resolver.Pending() {
		t.Fatalf("after g: result pending = %v, resolver pending = %v", result.IsPending(), resolver.Pending())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))

	result, cmd = resolver.Update(keyPress("h"))
	if cmd != nil {
		t.Fatal("second Update() cmd is not nil, want nil")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("after h: Match = %v, Pending = %v", result.Match(testGoHome), resolver.Pending())
	}
	if result.Key() != Text("h") {
		t.Fatalf("after h: Key() = %#v, want text h", result.Key())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("gh"))
}

// TestResolverIgnoresNonKeyMessages verifies unrelated Bubble Tea messages do not affect state.
func TestResolverIgnoresNonKeyMessages(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(struct{}{})
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsIdle() || resolver.Pending() {
		t.Fatalf("Update() idle = %v, resolver pending = %v, want idle/false", result.IsIdle(), resolver.Pending())
	}
}

// TestResolverMatchesSingleKeySequence verifies exact one-key bindings match immediately.
func TestResolverMatchesSingleKeySequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(keyPress("g"))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("Match = %v, Pending = %v, want true/false", result.Match(testGoHome), resolver.Pending())
	}
}

// TestResolverMatchesSpecialAndModifiedKeys verifies key conversion for non-text key presses.
func TestResolverMatchesSpecialAndModifiedKeys(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		message tea.KeyPressMsg
	}{
		{
			name:    "special key",
			binding: Bind(testGoHome, Sequence(Code(tea.KeyEsc))),
			message: tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}),
		},
		{
			name:    "modified key",
			binding: Bind(testGoHome, Sequence(Modified('c', tea.ModCtrl))),
			message: tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New([]Binding[testAction]{tt.binding})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, cmd := resolver.Update(tt.message)
			if cmd != nil {
				t.Fatal("Update() cmd is not nil, want nil")
			}
			if !result.Match(testGoHome) {
				t.Fatal("Match(testGoHome) = false, want true")
			}
		})
	}
}

// TestModifierMatchingIsExact verifies extra modifiers do not match a narrower binding.
func TestModifierMatchingIsExact(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, Sequence(Modified('c', tea.ModCtrl)))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl | tea.ModAlt}))
	if result.Match(testGoHome) {
		t.Fatal("ctrl+alt+c matched ctrl+c binding, want no match")
	}
	if !result.IsUnmatched() || !result.PassThrough() {
		t.Fatalf("result unmatched/pass-through = %v/%v, want true/true", result.IsUnmatched(), result.PassThrough())
	}
}

// keyPress builds a printable Bubble Tea key message for resolver tests.
func keyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

// TestDirectUnmatchedKeyPassesThrough verifies idle unmatched keys can fall back to host handling.
func TestDirectUnmatchedKeyPassesThrough(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(keyPress("j"))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsUnmatched() || !result.PassThrough() {
		t.Fatalf("result unmatched/pass-through = %v/%v, want true/true", result.IsUnmatched(), result.PassThrough())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("j"))
}

// TestFailedPendingSequenceIsConsumedByDefault verifies failed chords do not pass through by default.
func TestFailedPendingSequenceIsConsumedByDefault(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(keyPress("x"))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsUnmatched() || result.PassThrough() {
		t.Fatalf("result unmatched/pass-through = %v/%v, want true/false", result.IsUnmatched(), result.PassThrough())
	}
	if resolver.Pending() {
		t.Fatal("Pending() = true, want false")
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("gx"))
}

// TestFailedPendingSequenceCanPassThrough verifies failed chords can be delegated to host handling.
func TestFailedPendingSequenceCanPassThrough(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
		WithFailedPendingPassThrough(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(keyPress("x"))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsUnmatched() || !result.PassThrough() {
		t.Fatalf("result unmatched/pass-through = %v/%v, want true/true", result.IsUnmatched(), result.PassThrough())
	}
	if resolver.Pending() {
		t.Fatal("Pending() = true, want false")
	}
}

// TestDefaultCancelKeyCancelsPendingSequence verifies escape cancels an in-progress sequence.
func TestDefaultCancelKeyCancelsPendingSequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsCanceled() || resolver.Pending() {
		t.Fatalf("result canceled = %v, resolver pending = %v", result.IsCanceled(), resolver.Pending())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestDefaultCancelKeyTakesPrecedenceWhilePending verifies escape cancels instead of matching a longer binding.
func TestDefaultCancelKeyTakesPrecedenceWhilePending(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("gh")),
		Bind(testCopyLine, Sequence(Text("g"), Code(tea.KeyEsc))),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsCanceled() || result.Match(testCopyLine) {
		t.Fatalf("result canceled/match copy-line = %v/%v, want true/false", result.IsCanceled(), result.Match(testCopyLine))
	}
	if resolver.Pending() {
		t.Fatal("Pending() = true, want false")
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestConfiguredCancelKeyCancelsPendingSequence verifies custom cancel keys replace the default.
func TestConfiguredCancelKeyCancelsPendingSequence(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
		WithCancelKeys(Modified('c', tea.ModCtrl)),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.IsCanceled() || resolver.Pending() {
		t.Fatalf("result canceled = %v, resolver pending = %v", result.IsCanceled(), resolver.Pending())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestCancelKeyDoesNotCancelWhileIdle verifies cancel keys can still be normal bindings when idle.
func TestCancelKeyDoesNotCancelWhileIdle(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, Sequence(Code(tea.KeyEsc)))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if !result.Match(testGoHome) {
		t.Fatalf("Match(testGoHome) = false, want true")
	}
}

// TestCancelKeysCanBeDisabled verifies callers can bind cancel-key input when cancellation is off.
func TestCancelKeysCanBeDisabled(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{Bind(testGoHome, Sequence(Text("g"), Code(tea.KeyEsc)))},
		WithCancelKeys(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, _ := resolver.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if !result.Match(testGoHome) {
		t.Fatalf("Match(testGoHome) = false, want true")
	}
}

// TestAmbiguousMatchWaitsForContinuation verifies exact-prefix ambiguity waits for another key.
func TestAmbiguousMatchWaitsForContinuation(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("g")),
		Bind(testCopyLine, TextSequence("gh")),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, cmd := resolver.Update(keyPress("g"))
	if cmd == nil {
		t.Fatal("Update() cmd = nil, want timeout command")
	}
	if !result.IsPending() || result.Match(testGoHome) {
		t.Fatalf("after g: pending = %v, match short = %v, want pending true and match false", result.IsPending(), result.Match(testGoHome))
	}

	result, cmd = resolver.Update(keyPress("h"))
	if cmd != nil {
		t.Fatal("Update() cmd is not nil, want nil")
	}
	if !result.Match(testCopyLine) {
		t.Fatalf("Match(testCopyLine) = false, want true")
	}
}

// TestAmbiguousMatchResolvesShortBindingOnTimeout verifies timeout accepts a pending exact match.
func TestAmbiguousMatchResolvesShortBindingOnTimeout(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("g")),
		Bind(testCopyLine, TextSequence("gh")),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, cmd := resolver.Update(resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation})
	if cmd != nil {
		t.Fatal("timeout Update() cmd is not nil, want nil")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("timeout match short = %v, pending = %v, want true/false", result.Match(testGoHome), resolver.Pending())
	}
}

// TestReturnedTimeoutCommandResolvesPendingMatch verifies timeout commands emit usable messages.
func TestReturnedTimeoutCommandResolvesPendingMatch(t *testing.T) {
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

	_, cmd := resolver.Update(keyPress("g"))
	if cmd == nil {
		t.Fatal("Update() cmd = nil, want timeout command")
	}

	result, next := resolver.Update(cmd())
	if next != nil {
		t.Fatal("timeout Update() cmd is not nil, want nil")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("timeout command match short = %v, pending = %v, want true/false", result.Match(testGoHome), resolver.Pending())
	}
}

// TestPurePrefixCancelsOnTimeout verifies timeout cancels a prefix with no pending exact match.
func TestPurePrefixCancelsOnTimeout(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	result, _ := resolver.Update(resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation})
	if !result.IsCanceled() || resolver.Pending() {
		t.Fatalf("timeout canceled = %v, pending = %v, want true/false", result.IsCanceled(), resolver.Pending())
	}
}

// TestStaleTimeoutMessagesAreIgnored verifies old timeout commands cannot affect newer pending state.
func TestStaleTimeoutMessagesAreIgnored(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.Update(keyPress("g"))
	stale := resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation - 1}
	result, cmd := resolver.Update(stale)
	if cmd != nil {
		t.Fatal("stale timeout cmd is not nil, want nil")
	}
	if !result.IsIdle() || !resolver.Pending() {
		t.Fatalf("stale timeout idle = %v, pending = %v, want true/true", result.IsIdle(), resolver.Pending())
	}
}

// TestOtherResolverTimeoutMessagesAreIgnored verifies timeout identity is resolver-specific.
func TestOtherResolverTimeoutMessagesAreIgnored(t *testing.T) {
	first, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("first New() error = %v", err)
	}
	second, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("second New() error = %v", err)
	}

	first.Update(keyPress("g"))
	second.Update(keyPress("g"))
	foreign := resolverTimeoutMsg{resolverID: first.id, generation: second.generation}
	result, cmd := second.Update(foreign)
	if cmd != nil {
		t.Fatal("foreign timeout cmd is not nil, want nil")
	}
	if !result.IsIdle() || !second.Pending() {
		t.Fatalf("foreign timeout idle = %v, pending = %v, want true/true", result.IsIdle(), second.Pending())
	}
}
