package straw

import (
	"strconv"
	"testing"
)

const (
	benchmarkTargetAction benchmarkAction = -1
	benchmarkPrefixAction benchmarkAction = -2
	benchmarkLongAction   benchmarkAction = -3
)

type benchmarkAction int

var benchmarkBindingCounts = []int{10, 100, 1_000, 10_000}

var benchmarkAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&()*+,-./:;<=>?@[]^_{|}")

var benchmarkResultSink Result[benchmarkAction]
var benchmarkResolverSink *Resolver[benchmarkAction]
var benchmarkErrorSink error
var benchmarkTimeoutSink Timeout[benchmarkAction]

// benchmarkKeyPress builds a root text key.
func benchmarkKeyPress(text string) Key {
	return Text(text)
}

// benchmarkCodePress builds a root special key.
func benchmarkCodePress(code rune) Key {
	return Code(code)
}

// benchmarkCountName returns a stable sub-benchmark name for a binding count.
func benchmarkCountName(count int) string {
	return strconv.Itoa(count)
}

// benchmarkGeneratedSequence returns a deterministic three-key filler sequence.
func benchmarkGeneratedSequence(index int) Seq {
	base := len(benchmarkAlphabet)
	first := benchmarkAlphabet[index%base]
	second := benchmarkAlphabet[(index/base)%base]
	third := benchmarkAlphabet[(index/(base*base))%base]
	return Sequence(Text(string(first)), Text(string(second)), Text(string(third)))
}

// benchmarkGeneratedBindings creates deterministic filler bindings.
func benchmarkGeneratedBindings(count int, startAction int) []Binding[benchmarkAction] {
	bindings := make([]Binding[benchmarkAction], count)
	candidate := 0
	for i := range count {
		sequence := benchmarkGeneratedSequence(candidate)
		for seqHasPrefix(sequence, TextSequence("g")) {
			candidate++
			sequence = benchmarkGeneratedSequence(candidate)
		}
		bindings[i] = Bind(benchmarkAction(startAction+i), sequence)
		candidate++
	}
	return bindings
}

// benchmarkExactSingleKeyBindings creates bindings with one single-key target.
func benchmarkExactSingleKeyBindings(count int) []Binding[benchmarkAction] {
	bindings := []Binding[benchmarkAction]{
		Bind(benchmarkTargetAction, Sequence(Text("~"))),
	}
	bindings = append(bindings, benchmarkGeneratedBindings(count-1, 0)...)
	return bindings
}

// benchmarkMultiKeyBindings creates bindings with a known multi-key target.
func benchmarkMultiKeyBindings(count int) []Binding[benchmarkAction] {
	bindings := []Binding[benchmarkAction]{
		Bind(benchmarkTargetAction, TextSequence("gh")),
	}
	bindings = append(bindings, benchmarkGeneratedBindings(count-1, 0)...)
	return bindings
}

// benchmarkPendingPrefixBindings creates bindings where g is only a prefix.
func benchmarkPendingPrefixBindings(count int) []Binding[benchmarkAction] {
	bindings := []Binding[benchmarkAction]{
		Bind(benchmarkLongAction, TextSequence("gh")),
	}
	bindings = append(bindings, benchmarkGeneratedBindings(count-1, 0)...)
	return bindings
}

// benchmarkAmbiguousPrefixBindings creates bindings where g is both match and prefix.
func benchmarkAmbiguousPrefixBindings(count int) []Binding[benchmarkAction] {
	bindings := []Binding[benchmarkAction]{
		Bind(benchmarkPrefixAction, TextSequence("g")),
		Bind(benchmarkLongAction, TextSequence("gh")),
	}
	bindings = append(bindings, benchmarkGeneratedBindings(count-2, 0)...)
	return bindings
}

// benchmarkLongSharedPrefixBindings creates bindings that share a long common prefix.
func benchmarkLongSharedPrefixBindings(count int) []Binding[benchmarkAction] {
	prefix := TextSequence("abcdefghijkl")
	base := len(benchmarkAlphabet)
	bindings := make([]Binding[benchmarkAction], 0, count)
	for i := range count {
		sequence := appendKey(prefix, Text(string(benchmarkAlphabet[i%base])))
		sequence = appendKey(sequence, Text(string(benchmarkAlphabet[(i/base)%base])))
		bindings = append(bindings, Bind(benchmarkAction(i), sequence))
	}
	return bindings
}

// benchmarkNewResolver builds a resolver and fails the benchmark on setup errors.
func benchmarkNewResolver(b *testing.B, bindings []Binding[benchmarkAction], opts ...Option) *Resolver[benchmarkAction] {
	b.Helper()
	resolver, err := New(bindings, opts...)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	return resolver
}

// benchmarkPreparePendingPrefix seeds pending state without per-iteration setup allocation.
func benchmarkPreparePendingPrefix(resolver *Resolver[benchmarkAction], sequence Seq) Timeout[benchmarkAction] {
	resolver.pendingSeq = sequence
	resolver.pendingMatch = Binding[benchmarkAction]{}
	resolver.hasPendingMatch = false
	resolver.generation++
	return Timeout[benchmarkAction]{resolverID: resolver.id, generation: resolver.generation}
}

// benchmarkPreparePendingMatch seeds an ambiguous pending match without invoking Update.
func benchmarkPreparePendingMatch(resolver *Resolver[benchmarkAction], sequence Seq, binding Binding[benchmarkAction]) Timeout[benchmarkAction] {
	resolver.pendingSeq = sequence
	resolver.pendingMatch = binding
	resolver.hasPendingMatch = true
	resolver.generation++
	return Timeout[benchmarkAction]{resolverID: resolver.id, generation: resolver.generation}
}

// BenchmarkNew measures resolver validation, cloning, duplicate detection, and index construction.
func BenchmarkNew(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkGeneratedBindings(count, 0)
		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				resolver, err := New(bindings)
				benchmarkResolverSink = resolver
				benchmarkErrorSink = err
				if err != nil {
					b.Fatalf("New() error = %v", err)
				}
			}
		})
	}
}

// BenchmarkUpdateExactSingleKey measures a direct single-key binding match.
func BenchmarkUpdateExactSingleKey(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkExactSingleKeyBindings(count)
		resolver := benchmarkNewResolver(b, bindings)
		msg := benchmarkKeyPress("~")

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				result, timeout = resolver.UpdateKey(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.Match(benchmarkTargetAction) || resolver.Pending() || timeout.Scheduled() {
				b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), !timeout.Scheduled())
			}
		})
	}
}

// BenchmarkUpdateExactMultiKeyFinal measures seeded pending state plus the final exact key.
func BenchmarkUpdateExactMultiKeyFinal(b *testing.B) {
	counts := []int{10, 100, 1_000}
	for _, count := range counts {
		bindings := benchmarkMultiKeyBindings(count)
		pendingSeq := TextSequence("g")
		finalMsg := benchmarkKeyPress("h")
		b.Run(benchmarkCountName(count), func(b *testing.B) {
			resolver := benchmarkNewResolver(b, bindings)
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				benchmarkPreparePendingPrefix(resolver, pendingSeq)
				result, timeout = resolver.UpdateKey(finalMsg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.Match(benchmarkTargetAction) || resolver.Pending() || timeout.Scheduled() {
				b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), !timeout.Scheduled())
			}
		})
	}
}

// BenchmarkUpdateUnmatchedIdle measures an idle key that matches no binding or prefix.
func BenchmarkUpdateUnmatchedIdle(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkGeneratedBindings(count, 0)
		resolver := benchmarkNewResolver(b, bindings)
		msg := benchmarkCodePress(KeyEsc)

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				result, timeout = resolver.UpdateKey(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.IsUnmatched() || resolver.Pending() || !result.PassThrough() {
				b.Fatalf("Update() unmatched/pending/passThrough = %v/%v/%v, want true/false/true", result.IsUnmatched(), resolver.Pending(), result.PassThrough())
			}
		})
	}
}

// BenchmarkUpdateUnmatchedPending measures seeded pending state plus a failed key.
func BenchmarkUpdateUnmatchedPending(b *testing.B) {
	counts := []int{10, 100, 1_000}
	for _, count := range counts {
		bindings := benchmarkPendingPrefixBindings(count)
		pendingSeq := TextSequence("g")
		finalMsg := benchmarkCodePress(KeyTab)
		b.Run(benchmarkCountName(count), func(b *testing.B) {
			resolver := benchmarkNewResolver(b, bindings)
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				benchmarkPreparePendingPrefix(resolver, pendingSeq)
				result, timeout = resolver.UpdateKey(finalMsg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.IsUnmatched() || resolver.Pending() || result.PassThrough() {
				b.Fatalf("Update() unmatched/pending/passThrough = %v/%v/%v, want true/false/false", result.IsUnmatched(), resolver.Pending(), result.PassThrough())
			}
		})
	}
}

// BenchmarkUpdatePendingPrefix measures reset state plus a prefix key that starts pending.
func BenchmarkUpdatePendingPrefix(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkPendingPrefixBindings(count)
		resolver := benchmarkNewResolver(b, bindings)
		msg := benchmarkKeyPress("g")

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				resolver.Reset()
				result, timeout = resolver.UpdateKey(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.IsPending() || !resolver.Pending() || !timeout.Scheduled() {
				b.Fatalf("Update() pending/resolverPending/cmdNil = %v/%v/%v, want true/true/false", result.IsPending(), resolver.Pending(), !timeout.Scheduled())
			}
		})
	}
}

// BenchmarkUpdateAmbiguousPrefix measures reset state plus an ambiguous prefix key.
func BenchmarkUpdateAmbiguousPrefix(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkAmbiguousPrefixBindings(count)
		resolver := benchmarkNewResolver(b, bindings)
		msg := benchmarkKeyPress("g")

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				resolver.Reset()
				result, timeout = resolver.UpdateKey(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.IsPending() || !resolver.Pending() || !timeout.Scheduled() {
				b.Fatalf("Update() pending/resolverPending/cmdNil = %v/%v/%v, want true/true/false", result.IsPending(), resolver.Pending(), !timeout.Scheduled())
			}
		})
	}
}

// BenchmarkUpdateLongSharedPrefix measures lookup when many bindings share the same long prefix.
func BenchmarkUpdateLongSharedPrefix(b *testing.B) {
	counts := []int{10, 100, 1_000}
	for _, count := range counts {
		bindings := benchmarkLongSharedPrefixBindings(count)
		resolver := benchmarkNewResolver(b, bindings)
		pendingSeq := TextSequence("abcdefghijkl")
		msg := benchmarkKeyPress(string(benchmarkAlphabet[0]))

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var timeout Timeout[benchmarkAction]
			for range b.N {
				benchmarkPreparePendingPrefix(resolver, pendingSeq)
				result, timeout = resolver.UpdateKey(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkTimeoutSink = timeout
			if !result.IsPending() || !resolver.Pending() || !timeout.Scheduled() {
				b.Fatalf("Update() pending/resolverPending/cmd = %v/%v/%v, want true/true/true", result.IsPending(), resolver.Pending(), timeout.Scheduled())
			}
		})
	}
}

// BenchmarkUpdateLongSequenceFinal measures final-key lookup for a long sequence.
func BenchmarkUpdateLongSequenceFinal(b *testing.B) {
	sequence := TextSequence("abcdefghijklmnop")
	pendingSeq := sequence[:len(sequence)-1]
	finalMsg := sequence[len(sequence)-1]
	resolver := benchmarkNewResolver(b, []Binding[benchmarkAction]{Bind(benchmarkTargetAction, sequence)})

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var timeout Timeout[benchmarkAction]
	for range b.N {
		benchmarkPreparePendingPrefix(resolver, pendingSeq)
		result, timeout = resolver.UpdateKey(finalMsg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkTimeoutSink = timeout
	if !result.Match(benchmarkTargetAction) || resolver.Pending() || timeout.Scheduled() {
		b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), !timeout.Scheduled())
	}
}

// BenchmarkUpdateUnicodeText measures matching a Unicode text key.
func BenchmarkUpdateUnicodeText(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[benchmarkAction]{Bind(benchmarkTargetAction, TextSequence("λ"))})
	msg := benchmarkKeyPress("λ")

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var timeout Timeout[benchmarkAction]
	for range b.N {
		result, timeout = resolver.UpdateKey(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkTimeoutSink = timeout
	if !result.Match(benchmarkTargetAction) || resolver.Pending() || timeout.Scheduled() {
		b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), !timeout.Scheduled())
	}
}

// BenchmarkTimeoutResolvePendingMatch measures seeded pending match state plus timeout handling.
func BenchmarkTimeoutResolvePendingMatch(b *testing.B) {
	bindings := benchmarkAmbiguousPrefixBindings(100)
	resolver := benchmarkNewResolver(b, bindings)
	pendingSeq := TextSequence("g")
	pendingMatch := bindings[0]
	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var timeout Timeout[benchmarkAction]
	for range b.N {
		timeout = benchmarkPreparePendingMatch(resolver, pendingSeq, pendingMatch)
		result = resolver.UpdateTimeout(timeout)
		timeout = Timeout[benchmarkAction]{}
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkTimeoutSink = timeout
	if !result.Match(benchmarkPrefixAction) || resolver.Pending() || timeout.Scheduled() {
		b.Fatalf("timeout match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkPrefixAction), resolver.Pending(), !timeout.Scheduled())
	}
}

// BenchmarkTimeoutCancelPendingPrefix measures seeded pending prefix state plus timeout handling.
func BenchmarkTimeoutCancelPendingPrefix(b *testing.B) {
	bindings := benchmarkPendingPrefixBindings(100)
	resolver := benchmarkNewResolver(b, bindings)
	pendingSeq := TextSequence("g")
	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var timeout Timeout[benchmarkAction]
	for range b.N {
		timeout = benchmarkPreparePendingPrefix(resolver, pendingSeq)
		result = resolver.UpdateTimeout(timeout)
		timeout = Timeout[benchmarkAction]{}
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkTimeoutSink = timeout
	if !result.IsCanceled() || resolver.Pending() || timeout.Scheduled() {
		b.Fatalf("timeout canceled/pending/cmdNil = %v/%v/%v, want true/false/true", result.IsCanceled(), resolver.Pending(), !timeout.Scheduled())
	}
}

// BenchmarkTimeoutIgnoreStale measures stale timeout handling as idle/no-op behavior.
func BenchmarkTimeoutIgnoreStale(b *testing.B) {
	bindings := benchmarkPendingPrefixBindings(100)
	resolver := benchmarkNewResolver(b, bindings)
	_, pendingTimeout := resolver.UpdateKey(benchmarkKeyPress("g"))
	benchmarkTimeoutSink = pendingTimeout
	staleTimeout := Timeout[benchmarkAction]{resolverID: resolver.id, generation: resolver.generation - 1}

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var timeout Timeout[benchmarkAction]
	for range b.N {
		result = resolver.UpdateTimeout(staleTimeout)
		timeout = Timeout[benchmarkAction]{}
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkTimeoutSink = timeout
	if !result.IsIdle() || !resolver.Pending() || timeout.Scheduled() {
		b.Fatalf("stale timeout idle/pending/cmdNil = %v/%v/%v, want true/true/true", result.IsIdle(), resolver.Pending(), !timeout.Scheduled())
	}
}
