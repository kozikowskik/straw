package straw

import "testing"

// TestIdleResultContract verifies idle results expose no match or pass-through behavior.
func TestIdleResultContract(t *testing.T) {
	result := idleResult[testAction]()

	if result.State() != Idle || !result.IsIdle() {
		t.Fatalf("idle result state = %v, IsIdle = %v", result.State(), result.IsIdle())
	}
	if result.Match(testGoHome) {
		t.Fatal("Match() = true, want false")
	}
	if _, ok := result.Binding(); ok {
		t.Fatal("Binding() ok = true, want false")
	}
	if result.PassThrough() {
		t.Fatal("PassThrough() = true, want false")
	}
	assertSeqEqual(t, result.Sequence(), nil)
}

// TestPendingResultContract verifies pending results expose partial sequence state.
func TestPendingResultContract(t *testing.T) {
	result := pendingResult[testAction](Text("g"), TextSequence("g"))

	if !result.IsPending() || result.State() != Pending {
		t.Fatalf("pending state = %v, IsPending = %v", result.State(), result.IsPending())
	}
	if result.Key() != Text("g") {
		t.Fatalf("Key() = %#v, want text g", result.Key())
	}
	if result.Match(testGoHome) {
		t.Fatal("Match() = true, want false")
	}
	if _, ok := result.Binding(); ok {
		t.Fatal("Binding() ok = true, want false")
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestMatchedResultContract verifies matched results expose the matching binding and key.
func TestMatchedResultContract(t *testing.T) {
	binding := Bind(testGoHome, TextSequence("gh"), Description("go home"))
	result := matchedResult(binding, Text("h"))

	if !result.IsMatched() || result.State() != Matched {
		t.Fatalf("matched state = %v, IsMatched = %v", result.State(), result.IsMatched())
	}
	if !result.Match(testGoHome) {
		t.Fatal("Match(testGoHome) = false, want true")
	}
	if result.Match(testCopyLine) {
		t.Fatal("Match(testCopyLine) = true, want false")
	}
	gotBinding, ok := result.Binding()
	if !ok {
		t.Fatal("Binding() ok = false, want true")
	}
	if gotBinding.Description() != "go home" {
		t.Fatalf("Binding().Description() = %q, want go home", gotBinding.Description())
	}
	if result.Key() != Text("h") {
		t.Fatalf("Key() = %#v, want text h", result.Key())
	}
	assertSeqEqual(t, result.Sequence(), TextSequence("gh"))
}

// TestUnmatchedAndCanceledResultContracts verifies failed and canceled result state contracts.
func TestUnmatchedAndCanceledResultContracts(t *testing.T) {
	unmatched := unmatchedResult[testAction](Text("x"), TextSequence("gx"), true)
	if !unmatched.IsUnmatched() || !unmatched.PassThrough() {
		t.Fatalf("unmatched state/pass-through = %v/%v, want unmatched/true", unmatched.State(), unmatched.PassThrough())
	}
	assertSeqEqual(t, unmatched.Sequence(), TextSequence("gx"))

	canceled := canceledResult[testAction](Code(KeyEsc), TextSequence("g"))
	if !canceled.IsCanceled() || canceled.PassThrough() {
		t.Fatalf("canceled state/pass-through = %v/%v, want canceled/false", canceled.State(), canceled.PassThrough())
	}
	assertSeqEqual(t, canceled.Sequence(), TextSequence("g"))
}

// TestResultSequenceReturnsCopy verifies callers cannot mutate result-owned sequences.
func TestResultSequenceReturnsCopy(t *testing.T) {
	result := pendingResult[testAction](Text("g"), TextSequence("g"))

	sequence := result.Sequence()
	sequence[0] = Text("x")

	assertSeqEqual(t, result.Sequence(), TextSequence("g"))
}

// TestShouldPassThroughRequiresUnmatchedPassThrough verifies host handling only applies to unmatched pass-through results.
func TestShouldPassThroughRequiresUnmatchedPassThrough(t *testing.T) {
	tests := []struct {
		name   string
		result Result[testAction]
		want   bool
	}{
		{name: "idle", result: idleResult[testAction](), want: false},
		{name: "pending", result: pendingResult[testAction](Text("g"), TextSequence("g")), want: false},
		{name: "matched", result: matchedResult(Bind(testGoHome, TextSequence("gh")), Text("h")), want: false},
		{name: "unmatched without pass-through", result: unmatchedResult[testAction](Text("x"), TextSequence("gx"), false), want: false},
		{name: "unmatched with pass-through", result: unmatchedResult[testAction](Text("x"), TextSequence("x"), true), want: true},
		{name: "canceled", result: canceledResult[testAction](Code(KeyEsc), TextSequence("g")), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldPassThrough(tt.result); got != tt.want {
				t.Fatalf("ShouldPassThrough() = %v, want %v", got, tt.want)
			}
		})
	}
}
