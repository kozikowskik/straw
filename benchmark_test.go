package straw

import (
	"strconv"
	"testing"

	tea "charm.land/bubbletea/v2"
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
var benchmarkCmdSink tea.Cmd

// benchmarkKeyPress builds a printable Bubble Tea key press message.
func benchmarkKeyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

// benchmarkCodePress builds a special-key Bubble Tea key press message.
func benchmarkCodePress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
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
func benchmarkPreparePendingPrefix(resolver *Resolver[benchmarkAction], sequence Seq) resolverTimeoutMsg {
	resolver.pendingSeq = sequence
	resolver.pendingMatch = Binding[benchmarkAction]{}
	resolver.hasPendingMatch = false
	resolver.generation++
	return resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation}
}

// benchmarkPreparePendingMatch seeds an ambiguous pending match without invoking Update.
func benchmarkPreparePendingMatch(resolver *Resolver[benchmarkAction], sequence Seq, binding Binding[benchmarkAction]) resolverTimeoutMsg {
	resolver.pendingSeq = sequence
	resolver.pendingMatch = binding
	resolver.hasPendingMatch = true
	resolver.generation++
	return resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation}
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
			var cmd tea.Cmd
			for range b.N {
				result, cmd = resolver.Update(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
			if !result.Match(benchmarkTargetAction) || resolver.Pending() || cmd != nil {
				b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), cmd == nil)
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
			var cmd tea.Cmd
			for range b.N {
				benchmarkPreparePendingPrefix(resolver, pendingSeq)
				result, cmd = resolver.Update(finalMsg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
			if !result.Match(benchmarkTargetAction) || resolver.Pending() || cmd != nil {
				b.Fatalf("Update() match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkTargetAction), resolver.Pending(), cmd == nil)
			}
		})
	}
}

// BenchmarkUpdateUnmatchedIdle measures an idle key that matches no binding or prefix.
func BenchmarkUpdateUnmatchedIdle(b *testing.B) {
	for _, count := range benchmarkBindingCounts {
		bindings := benchmarkGeneratedBindings(count, 0)
		resolver := benchmarkNewResolver(b, bindings)
		msg := benchmarkCodePress(tea.KeyEsc)

		b.Run(benchmarkCountName(count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var cmd tea.Cmd
			for range b.N {
				result, cmd = resolver.Update(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
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
		finalMsg := benchmarkCodePress(tea.KeyTab)
		b.Run(benchmarkCountName(count), func(b *testing.B) {
			resolver := benchmarkNewResolver(b, bindings)
			b.ReportAllocs()
			b.ResetTimer()
			var result Result[benchmarkAction]
			var cmd tea.Cmd
			for range b.N {
				benchmarkPreparePendingPrefix(resolver, pendingSeq)
				result, cmd = resolver.Update(finalMsg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
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
			var cmd tea.Cmd
			for range b.N {
				resolver.Reset()
				result, cmd = resolver.Update(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
			if !result.IsPending() || !resolver.Pending() || cmd == nil {
				b.Fatalf("Update() pending/resolverPending/cmdNil = %v/%v/%v, want true/true/false", result.IsPending(), resolver.Pending(), cmd == nil)
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
			var cmd tea.Cmd
			for range b.N {
				resolver.Reset()
				result, cmd = resolver.Update(msg)
			}
			b.StopTimer()
			benchmarkResultSink = result
			benchmarkCmdSink = cmd
			if !result.IsPending() || !resolver.Pending() || cmd == nil {
				b.Fatalf("Update() pending/resolverPending/cmdNil = %v/%v/%v, want true/true/false", result.IsPending(), resolver.Pending(), cmd == nil)
			}
		})
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
	var cmd tea.Cmd
	for range b.N {
		timeout := benchmarkPreparePendingMatch(resolver, pendingSeq, pendingMatch)
		result, cmd = resolver.Update(timeout)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.Match(benchmarkPrefixAction) || resolver.Pending() || cmd != nil {
		b.Fatalf("timeout match/pending/cmdNil = %v/%v/%v, want true/false/true", result.Match(benchmarkPrefixAction), resolver.Pending(), cmd == nil)
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
	var cmd tea.Cmd
	for range b.N {
		timeout := benchmarkPreparePendingPrefix(resolver, pendingSeq)
		result, cmd = resolver.Update(timeout)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.IsCanceled() || resolver.Pending() || cmd != nil {
		b.Fatalf("timeout canceled/pending/cmdNil = %v/%v/%v, want true/false/true", result.IsCanceled(), resolver.Pending(), cmd == nil)
	}
}

// BenchmarkTimeoutIgnoreStale measures stale timeout handling as idle/no-op behavior.
func BenchmarkTimeoutIgnoreStale(b *testing.B) {
	bindings := benchmarkPendingPrefixBindings(100)
	resolver := benchmarkNewResolver(b, bindings)
	_, pendingCmd := resolver.Update(benchmarkKeyPress("g"))
	benchmarkCmdSink = pendingCmd
	staleTimeout := resolverTimeoutMsg{resolverID: resolver.id, generation: resolver.generation - 1}

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[benchmarkAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(staleTimeout)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.IsIdle() || !resolver.Pending() || cmd != nil {
		b.Fatalf("stale timeout idle/pending/cmdNil = %v/%v/%v, want true/true/true", result.IsIdle(), resolver.Pending(), cmd == nil)
	}
}
