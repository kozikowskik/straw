package v2

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

var benchmarkResultSink Result[testAction]
var benchmarkCmdSink tea.Cmd

func benchmarkNewResolver(b *testing.B, bindings []Binding[testAction], opts ...Option) *Resolver[testAction] {
	b.Helper()
	resolver, err := New(bindings, opts...)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	return resolver
}

func BenchmarkUpdateTextKey(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	msg := tea.KeyPressMsg(tea.Key{Text: "g", Code: 'g'})

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[testAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.Match(testGoHome) || cmd != nil {
		b.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
	}
}

func BenchmarkUpdateSpecialKey(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[testAction]{Bind(testGoHome, Sequence(Code(KeyEsc)))})
	msg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[testAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.Match(testGoHome) || cmd != nil {
		b.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
	}
}

func BenchmarkUpdateModifiedKey(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[testAction]{Bind(testGoHome, Sequence(Modified('c', ModCtrl)))})
	msg := tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl})

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[testAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.Match(testGoHome) || cmd != nil {
		b.Fatalf("Update() match/cmdNil = %v/%v, want true/true", result.Match(testGoHome), cmd == nil)
	}
}

func BenchmarkUpdateNonKeyMessage(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	msg := struct{}{}

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[testAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.IsIdle() || cmd != nil {
		b.Fatalf("Update() idle/cmdNil = %v/%v, want true/true", result.IsIdle(), cmd == nil)
	}
}

func BenchmarkUpdateTimeoutMessage(b *testing.B) {
	resolver := benchmarkNewResolver(b, []Binding[testAction]{Bind(testGoHome, TextSequence("g"))})
	msg := timeoutMsg[testAction]{}

	b.ReportAllocs()
	b.ResetTimer()
	var result Result[testAction]
	var cmd tea.Cmd
	for range b.N {
		result, cmd = resolver.Update(msg)
	}
	b.StopTimer()
	benchmarkResultSink = result
	benchmarkCmdSink = cmd
	if !result.IsIdle() || cmd != nil {
		b.Fatalf("Update() idle/cmdNil = %v/%v, want true/true", result.IsIdle(), cmd == nil)
	}
}
