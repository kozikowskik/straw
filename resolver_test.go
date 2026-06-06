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

// keyPress builds a printable Bubble Tea key message for resolver tests.
func keyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}
