package straw

import "testing"

var allocationResultSink Result[testAction]
var allocationResolverSink *Resolver[testAction]
var allocationErrorSink error
var allocationTimeoutSink Timeout[testAction]
var allocationNextChoicesSink []NextChoice[testAction]

func TestUpdateKeyAvoidsDuplicateSequenceAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("allocation counts are not stable under the race detector")
	}

	pendingG := TextSequence("g")
	tests := []struct {
		name      string
		bindings  []Binding[testAction]
		prepare   func(*Resolver[testAction])
		key       Key
		wantState State
		maxAllocs float64
	}{
		{
			name:      "exact single key",
			bindings:  []Binding[testAction]{Bind(testGoHome, TextSequence("g"))},
			key:       Text("g"),
			wantState: Matched,
			maxAllocs: 0,
		},
		{
			name:      "idle unmatched key",
			bindings:  []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
			key:       Text("x"),
			wantState: Unmatched,
			maxAllocs: 0,
		},
		{
			name:      "pending prefix",
			bindings:  []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
			key:       Text("g"),
			wantState: Pending,
			maxAllocs: 1,
		},
		{
			name:     "pending final match",
			bindings: []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
			prepare: func(resolver *Resolver[testAction]) {
				resolver.pendingSeq = pendingG
			},
			key:       Text("h"),
			wantState: Matched,
			maxAllocs: 0,
		},
		{
			name:     "failed pending key",
			bindings: []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
			prepare: func(resolver *Resolver[testAction]) {
				resolver.pendingSeq = pendingG
			},
			key:       Text("x"),
			wantState: Unmatched,
			maxAllocs: 0,
		},
		{
			name:     "cancel pending key",
			bindings: []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))},
			prepare: func(resolver *Resolver[testAction]) {
				resolver.pendingSeq = pendingG
			},
			key:       Code(KeyEsc),
			wantState: Canceled,
			maxAllocs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New(tt.bindings)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			allocs := testing.AllocsPerRun(100, func() {
				resolver.Reset()
				if tt.prepare != nil {
					tt.prepare(resolver)
				}
				result, timeout := resolver.UpdateKey(tt.key)
				allocationResultSink = result
				allocationTimeoutSink = timeout
				if result.State() != tt.wantState {
					t.Fatalf("UpdateKey() state = %v, want %v", result.State(), tt.wantState)
				}
			})

			if allocs > tt.maxAllocs {
				t.Fatalf("UpdateKey() allocations = %.0f, want <= %.0f", allocs, tt.maxAllocs)
			}
		})
	}
}

func TestNewAvoidsDuplicateBindingSequenceCopies(t *testing.T) {
	if raceEnabled {
		t.Skip("allocation counts are not stable under the race detector")
	}

	bindings := []Binding[testAction]{Bind(testGoHome, TextSequence("gh"))}

	allocs := testing.AllocsPerRun(100, func() {
		resolver, err := New(bindings)
		allocationResolverSink = resolver
		allocationErrorSink = err
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
	})

	if allocs > 12 {
		t.Fatalf("New() allocations = %.0f, want <= 12", allocs)
	}
}

func TestNextChoicesAllocationProfile(t *testing.T) {
	if raceEnabled {
		t.Skip("allocation counts are not stable under the race detector")
	}

	bindings := []Binding[testAction]{
		Bind(testGoHome, TextSequence("gh")),
		Bind(testCopyLine, TextSequence("gd")),
		Bind(testDeleteLine, TextSequence("yy")),
	}
	resolver, err := New(bindings)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	allocs := testing.AllocsPerRun(100, func() {
		choices := resolver.NextChoices()
		allocationNextChoicesSink = choices
		if len(choices) != 2 {
			t.Fatalf("NextChoices() length = %d, want 2", len(choices))
		}
	})

	if allocs > 5 {
		t.Fatalf("NextChoices() allocations = %.0f, want <= 5", allocs)
	}
}
