package straw

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

type testAction int

const (
	testGoHome testAction = iota + 1
	testCopyLine
	testDeleteLine
)

// TestBindStoresActionSequenceAndDescription verifies Bind preserves required binding data.
func TestBindStoresActionSequenceAndDescription(t *testing.T) {
	sequence := TextSequence("gh")
	binding := Bind(testGoHome, sequence, Description("go home"))

	if binding.Action() != testGoHome {
		t.Fatalf("Action() = %v, want %v", binding.Action(), testGoHome)
	}

	assertSeqEqual(t, binding.Sequence(), sequence)

	if binding.Description() != "go home" {
		t.Fatalf("Description() = %q, want %q", binding.Description(), "go home")
	}
}

// TestBindAllowsNoDescription verifies description metadata is optional.
func TestBindAllowsNoDescription(t *testing.T) {
	binding := Bind(testCopyLine, TextSequence("yy"))

	if binding.Action() != testCopyLine {
		t.Fatalf("Action() = %v, want %v", binding.Action(), testCopyLine)
	}

	if binding.Description() != "" {
		t.Fatalf("Description() = %q, want empty string", binding.Description())
	}
}

// TestBindingSequenceReturnsCopy verifies callers cannot mutate binding-owned sequences.
func TestBindingSequenceReturnsCopy(t *testing.T) {
	original := TextSequence("gh")
	binding := Bind(testGoHome, original)

	original[0] = Text("x")
	assertSeqEqual(t, binding.Sequence(), TextSequence("gh"))

	returned := binding.Sequence()
	returned[0] = Text("x")
	assertSeqEqual(t, binding.Sequence(), TextSequence("gh"))
}

// TestTextSequenceSplitsByRune verifies Unicode text is split into single-rune keys.
func TestTextSequenceSplitsByRune(t *testing.T) {
	assertSeqEqual(t, TextSequence("gé"), Sequence(Text("g"), Text("é")))
}

// TestSequenceReturnsCopy verifies Sequence does not retain caller-owned slices.
func TestSequenceReturnsCopy(t *testing.T) {
	keys := []Key{Text("g"), Text("h")}
	sequence := Sequence(keys...)

	keys[0] = Text("x")
	assertSeqEqual(t, sequence, TextSequence("gh"))
}

// TestSequenceSupportsExplicitKeyConstructors verifies explicit Text keys match text sequences.
func TestSequenceSupportsExplicitKeyConstructors(t *testing.T) {
	sequence := Sequence(Text("g"), Text("h"))

	assertSeqEqual(t, sequence, TextSequence("gh"))
}

// TestCodeAndModifiedKeysCanBeStoredInSequence verifies non-text key builders compose in sequences.
func TestCodeAndModifiedKeysCanBeStoredInSequence(t *testing.T) {
	sequence := Sequence(Code(tea.KeyEsc), Modified('c', tea.ModCtrl))

	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}

	if sequence[0] != Code(tea.KeyEsc) {
		t.Fatalf("sequence[0] = %#v, want esc code key", sequence[0])
	}

	if sequence[1] != Modified('c', tea.ModCtrl) {
		t.Fatalf("sequence[1] = %#v, want ctrl+c modified key", sequence[1])
	}
}

// assertSeqEqual compares sequences while producing focused test failures.
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
