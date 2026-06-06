package straw

import tea "charm.land/bubbletea/v2"

type keyKind int

const (
	keyKindText keyKind = iota + 1
	keyKindCode
	keyKindModified
)

// Key describes one Bubble Tea key press in a binding sequence.
type Key struct {
	kind keyKind
	text string
	code rune
	mod  tea.KeyMod
}

// Seq is an ordered key sequence, such as g then h.
type Seq []Key

// Text builds a regular text key.
func Text(value string) Key {
	return Key{kind: keyKindText, text: value}
}

// TextSequence expands a string into one text key per rune.
func TextSequence(value string) Seq {
	sequence := make(Seq, 0, len(value))
	for _, r := range value {
		sequence = append(sequence, Text(string(r)))
	}
	return sequence
}

// Code builds a Bubble Tea special key.
func Code(code rune) Key {
	return Key{kind: keyKindCode, code: code}
}

// Modified builds a key with Bubble Tea modifiers such as ctrl or alt.
func Modified(code rune, mod tea.KeyMod) Key {
	return Key{kind: keyKindModified, code: code, mod: mod}
}

// Sequence builds a sequence from explicit keys.
func Sequence(keys ...Key) Seq {
	return cloneSeq(keys)
}

// Binding connects an application-owned action to a key sequence.
type Binding[A comparable] struct {
	action      A
	sequence    Seq
	description string
}

type bindingOptions struct {
	description string
}

// BindingOption configures optional binding metadata.
type BindingOption func(*bindingOptions)

// Description adds human-readable metadata to a binding.
func Description(text string) BindingOption {
	return func(opts *bindingOptions) {
		opts.description = text
	}
}

// Bind builds a binding without validating it.
func Bind[A comparable](action A, sequence Seq, opts ...BindingOption) Binding[A] {
	options := bindingOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return Binding[A]{
		action:      action,
		sequence:    cloneSeq(sequence),
		description: options.description,
	}
}

// Action returns the application-owned action attached to the binding.
func (b Binding[A]) Action() A {
	return b.action
}

// Sequence returns a copy of the binding's key sequence.
func (b Binding[A]) Sequence() Seq {
	return cloneSeq(b.sequence)
}

// Description returns optional human-readable binding metadata.
func (b Binding[A]) Description() string {
	return b.description
}

// cloneSeq returns an independent copy of a key sequence.
func cloneSeq(sequence []Key) Seq {
	cloned := make(Seq, len(sequence))
	copy(cloned, sequence)
	return cloned
}
