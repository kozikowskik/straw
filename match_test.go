package straw

import "testing"

func TestBindingIndexLookupDoesNotAllocatePerBinding(t *testing.T) {
	bindings := benchmarkGeneratedBindings(100, 0)
	index, err := newBindingIndex(bindings)
	if err != nil {
		t.Fatalf("newBindingIndex() error = %v", err)
	}

	sequence := Sequence(Code(KeyEsc))
	allocs := testing.AllocsPerRun(100, func() {
		status := index.lookup(sequence)
		if status.hasMatch || status.hasPrefix {
			t.Fatalf("lookup() = match/prefix %v/%v, want false/false", status.hasMatch, status.hasPrefix)
		}
	})

	if allocs != 0 {
		t.Fatalf("lookup() allocations = %v, want 0", allocs)
	}
}
