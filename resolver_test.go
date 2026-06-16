package straw

import (
	"errors"
	"sync"
	"testing"
	"time"
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

// TestDefaultTimeoutIsHalfSecond verifies the omitted timeout favors responsive ambiguity resolution.
func TestDefaultTimeoutIsHalfSecond(t *testing.T) {
	options := defaultResolverOptions()

	if options.timeout != 500*time.Millisecond {
		t.Fatalf("default timeout = %v, want 500ms", options.timeout)
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

// TestNewRejectsNilOption verifies nil resolver options return a validation error instead of panicking.
func TestNewRejectsNilOption(t *testing.T) {
	_, err := New[testAction](nil, nil)
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

// TestNewCanRunConcurrently verifies resolver IDs can be assigned safely from concurrent constructors.
func TestNewCanRunConcurrently(t *testing.T) {
	const workers = 32

	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("New() error = %v, want nil", err)
		}
	}
}

// TestNewRejectsInvalidCancelKeys verifies cancel-key options use the same key contract as bindings.
func TestNewRejectsInvalidCancelKeys(t *testing.T) {
	tests := []struct {
		name string
		key  Key
	}{
		{name: "empty text key", key: Text("")},
		{name: "multi-rune text key", key: Text("gg")},
		{name: "printable code key", key: Code('g')},
		{name: "zero code key", key: Code(0)},
		{name: "modified key without modifier", key: Modified('c', 0)},
		{name: "unknown key kind", key: Key{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New[testAction](nil, WithCancelKeys(tt.key))
			if !errors.Is(err, ErrInvalidKey) {
				t.Fatalf("New() error = %v, want ErrInvalidKey", err)
			}
		})
	}
}

// TestNewAcceptsResolverOptions verifies supported options can be applied during construction.
func TestNewAcceptsResolverOptions(t *testing.T) {
	resolver, err := New[testAction](nil,
		WithTimeout(250*time.Millisecond),
		WithCancelKeys(Code(KeyEsc), Modified('c', ModCtrl)),
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
		Bind(testGoHome, Sequence(Code(KeySpace))),
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
			name:     "zero code key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Code(0)))},
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

	result, timeout := resolver.UpdateKey(keyPress("g"))
	if !timeout.Scheduled() {
		t.Fatal("first UpdateKey() timeout was not scheduled")
	}
	if !result.IsPending() || !resolver.Pending() {
		t.Fatalf("after g: result pending = %v, resolver pending = %v", result.IsPending(), resolver.Pending())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))

	result, timeout = resolver.UpdateKey(keyPress("h"))
	if timeout.Scheduled() {
		t.Fatal("second UpdateKey() timeout was scheduled")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("after h: Match = %v, Pending = %v", result.Match(testGoHome), resolver.Pending())
	}
	if result.Key() != Text("h") {
		t.Fatalf("after h: Key() = %#v, want text h", result.Key())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("gh"))
}

// TestResolverRejectsForgedTimeout verifies arbitrary timeout tokens do not affect state.
func TestResolverRejectsForgedTimeout(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := resolver.UpdateTimeout(Timeout[testAction]{})
	if !result.IsIdle() || resolver.Pending() {
		t.Fatalf("UpdateTimeout() idle = %v, resolver pending = %v, want idle/false", result.IsIdle(), resolver.Pending())
	}
}

// TestResolverMatchesSingleKeySequence verifies exact one-key bindings match immediately.
func TestResolverMatchesSingleKeySequence(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, timeout := resolver.UpdateKey(keyPress("g"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
	}
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("Match = %v, Pending = %v, want true/false", result.Match(testGoHome), resolver.Pending())
	}
}

// TestResolverMatchesSpecialAndModifiedKeys verifies non-text root keys match.
func TestResolverMatchesSpecialAndModifiedKeys(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding[testAction]
		key     Key
	}{
		{
			name:    "special key",
			binding: Bind(testGoHome, Sequence(Code(KeyEsc))),
			key:     Code(KeyEsc),
		},
		{
			name:    "modified key",
			binding: Bind(testGoHome, Sequence(Modified('c', ModCtrl))),
			key:     Modified('c', ModCtrl),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New([]Binding[testAction]{tt.binding})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, timeout := resolver.UpdateKey(tt.key)
			if timeout.Scheduled() {
				t.Fatal("UpdateKey() timeout was scheduled")
			}
			if !result.Match(testGoHome) {
				t.Fatal("Match(testGoHome) = false, want true")
			}
		})
	}
}

// TestModifierMatchingIsExact verifies extra modifiers do not match a narrower binding.
func TestModifierMatchingIsExact(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, Sequence(Modified('c', ModCtrl)))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, _ := resolver.UpdateKey(Modified('c', ModCtrl|ModAlt))
	if result.Match(testGoHome) {
		t.Fatal("ctrl+alt+c matched ctrl+c binding, want no match")
	}
	if !result.IsUnmatched() || !result.PassThrough() {
		t.Fatalf("result unmatched/pass-through = %v/%v, want true/true", result.IsUnmatched(), result.PassThrough())
	}
}

// keyPress builds a root text key for resolver tests.
func keyPress(text string) Key {
	return Text(text)
}

// TestDirectUnmatchedKeyPassesThrough verifies idle unmatched keys can fall back to host handling.
func TestDirectUnmatchedKeyPassesThrough(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, TextSequence("gh"))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, timeout := resolver.UpdateKey(keyPress("j"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(keyPress("x"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(keyPress("x"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(Code(KeyEsc))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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
		Bind(testCopyLine, Sequence(Text("g"), Code(KeyEsc))),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(Code(KeyEsc))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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
		WithCancelKeys(Modified('c', ModCtrl)),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(Modified('c', ModCtrl))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
	}
	if !result.IsCanceled() || resolver.Pending() {
		t.Fatalf("result canceled = %v, resolver pending = %v", result.IsCanceled(), resolver.Pending())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestCancelKeyDoesNotCancelWhileIdle verifies cancel keys can still be normal bindings when idle.
func TestCancelKeyDoesNotCancelWhileIdle(t *testing.T) {
	resolver, err := New([]Binding[testAction]{Bind(testGoHome, Sequence(Code(KeyEsc)))})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, _ := resolver.UpdateKey(Code(KeyEsc))
	if !result.Match(testGoHome) {
		t.Fatalf("Match(testGoHome) = false, want true")
	}
}

// TestCancelKeysCanBeDisabled verifies callers can bind cancel-key input when cancellation is off.
func TestCancelKeysCanBeDisabled(t *testing.T) {
	resolver, err := New(
		[]Binding[testAction]{Bind(testGoHome, Sequence(Text("g"), Code(KeyEsc)))},
		WithCancelKeys(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.UpdateKey(keyPress("g"))
	result, _ := resolver.UpdateKey(Code(KeyEsc))
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

	result, timeout := resolver.UpdateKey(keyPress("g"))
	if !timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was not scheduled")
	}
	if !result.IsPending() || result.Match(testGoHome) {
		t.Fatalf("after g: pending = %v, match short = %v, want pending true and match false", result.IsPending(), result.Match(testGoHome))
	}

	result, timeout = resolver.UpdateKey(keyPress("h"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
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

	resolver.UpdateKey(keyPress("g"))
	result := resolver.UpdateTimeout(Timeout[testAction]{resolverID: resolver.id, generation: resolver.generation})
	if !result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("timeout match short = %v, pending = %v, want true/false", result.Match(testGoHome), resolver.Pending())
	}
}

// TestReturnedTimeoutTokenResolvesPendingMatch verifies timeout tokens can resolve pending matches.
func TestReturnedTimeoutTokenResolvesPendingMatch(t *testing.T) {
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

	_, timeout := resolver.UpdateKey(keyPress("g"))
	if !timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was not scheduled")
	}
	if timeout.Duration() != time.Nanosecond {
		t.Fatalf("timeout duration = %v, want 1ns", timeout.Duration())
	}

	result := resolver.UpdateTimeout(timeout)
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

	resolver.UpdateKey(keyPress("g"))
	result := resolver.UpdateTimeout(Timeout[testAction]{resolverID: resolver.id, generation: resolver.generation})
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

	resolver.UpdateKey(keyPress("g"))
	stale := Timeout[testAction]{resolverID: resolver.id, generation: resolver.generation - 1}
	result := resolver.UpdateTimeout(stale)
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

	first.UpdateKey(keyPress("g"))
	second.UpdateKey(keyPress("g"))
	foreign := Timeout[testAction]{resolverID: first.id, generation: second.generation}
	result := second.UpdateTimeout(foreign)
	if !result.IsIdle() || !second.Pending() {
		t.Fatalf("foreign timeout idle = %v, pending = %v, want true/true", result.IsIdle(), second.Pending())
	}
}

// TestResetClearsPendingAndInvalidatesTimeout verifies reset discards pending state and old timers.
func TestResetClearsPendingAndInvalidatesTimeout(t *testing.T) {
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

	_, timeout := resolver.UpdateKey(keyPress("g"))
	if !timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was not scheduled")
	}

	resolver.Reset()
	if resolver.Pending() {
		t.Fatal("Pending() = true after Reset(), want false")
	}

	result := resolver.UpdateTimeout(timeout)
	if !result.IsIdle() || result.Match(testGoHome) {
		t.Fatalf("reset timeout idle/match = %v/%v, want true/false", result.IsIdle(), result.Match(testGoHome))
	}
}

// TestNewCopiesBindingsForResolverUse verifies caller mutations after New do not affect matching.
func TestNewCopiesBindingsForResolverUse(t *testing.T) {
	binding := Bind(testGoHome, TextSequence("gh"))
	bindings := []Binding[testAction]{binding}
	resolver, err := New(bindings)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	bindings[0] = Bind(testCopyLine, TextSequence("yy"))
	binding.sequence[0] = Text("x")

	resolver.UpdateKey(keyPress("g"))
	result, timeout := resolver.UpdateKey(keyPress("h"))
	if timeout.Scheduled() {
		t.Fatal("UpdateKey() timeout was scheduled")
	}
	if !result.Match(testGoHome) || result.Match(testCopyLine) {
		t.Fatalf("match go-home/copy-line = %v/%v, want true/false", result.Match(testGoHome), result.Match(testCopyLine))
	}
}

// TestOverlappingPrefixChainPreservesDeepContinuation verifies three-level overlaps keep waiting.
func TestOverlappingPrefixChainPreservesDeepContinuation(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("g")),
		Bind(testCopyLine, TextSequence("gh")),
		Bind(testDeleteLine, TextSequence("gha")),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result, timeout := resolver.UpdateKey(keyPress("g"))
	if !timeout.Scheduled() || !result.IsPending() || result.Match(testGoHome) {
		t.Fatalf("after g: timeout scheduled/pending/match short = %v/%v/%v, want true/true/false", timeout.Scheduled(), result.IsPending(), result.Match(testGoHome))
	}

	result, timeout = resolver.UpdateKey(keyPress("h"))
	if !timeout.Scheduled() || !result.IsPending() || result.Match(testCopyLine) {
		t.Fatalf("after gh: timeout scheduled/pending/match middle = %v/%v/%v, want true/true/false", timeout.Scheduled(), result.IsPending(), result.Match(testCopyLine))
	}

	result, timeout = resolver.UpdateKey(keyPress("a"))
	if timeout.Scheduled() {
		t.Fatal("after gha timeout was scheduled")
	}
	if !result.Match(testDeleteLine) || resolver.Pending() {
		t.Fatalf("after gha match deep/pending = %v/%v, want true/false", result.Match(testDeleteLine), resolver.Pending())
	}
}

// TestOverlappingPrefixChainTimeoutResolvesNearestPendingMatch verifies timeout uses the current exact match.
func TestOverlappingPrefixChainTimeoutResolvesNearestPendingMatch(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, TextSequence("g")),
		Bind(testCopyLine, TextSequence("gh")),
		Bind(testDeleteLine, TextSequence("gha")),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resolver.UpdateKey(keyPress("g"))
	resolver.UpdateKey(keyPress("h"))
	result := resolver.UpdateTimeout(Timeout[testAction]{resolverID: resolver.id, generation: resolver.generation})
	if !result.Match(testCopyLine) || result.Match(testGoHome) || resolver.Pending() {
		t.Fatalf("timeout match middle/short/pending = %v/%v/%v, want true/false/false", result.Match(testCopyLine), result.Match(testGoHome), resolver.Pending())
	}
}
