package straw

import "strings"

type keyKind int

const (
	keyKindText keyKind = iota + 1
	keyKindCode
	keyKindModified
)

// Mod describes keyboard modifiers supported by straw's version-neutral key model.
type Mod uint

const (
	ModAlt Mod = 1 << iota
	ModCtrl
	ModShift
	ModMeta
	ModHyper
	ModSuper
)

const (
	KeyBackspace = '\b'
	KeyTab       = '\t'
	KeyEnter     = '\r'
	KeyEsc       = '\x1b'
	KeySpace     = ' '
	KeyUp        = 0xF700 + iota
	KeyDown
	KeyRight
	KeyLeft
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

var codeLabels = map[rune]string{
	KeyBackspace: "backspace",
	KeyTab:       "tab",
	KeyEnter:     "enter",
	KeyEsc:       "esc",
	KeySpace:     "space",
	KeyUp:        "up",
	KeyDown:      "down",
	KeyRight:     "right",
	KeyLeft:      "left",
	KeyHome:      "home",
	KeyEnd:       "end",
	KeyPgUp:      "pgup",
	KeyPgDown:    "pgdown",
	KeyDelete:    "delete",
	KeyInsert:    "insert",
	KeyF1:        "f1",
	KeyF2:        "f2",
	KeyF3:        "f3",
	KeyF4:        "f4",
	KeyF5:        "f5",
	KeyF6:        "f6",
	KeyF7:        "f7",
	KeyF8:        "f8",
	KeyF9:        "f9",
	KeyF10:       "f10",
	KeyF11:       "f11",
	KeyF12:       "f12",
}

// Key describes one version-neutral key press in a binding sequence.
type Key struct {
	kind keyKind
	text string
	code rune
	mod  Mod
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

// Code builds a special key.
func Code(code rune) Key {
	return Key{kind: keyKindCode, code: code}
}

// Modified builds a key with modifiers such as ctrl or alt.
func Modified(code rune, mod Mod) Key {
	return Key{kind: keyKindModified, code: code, mod: mod}
}

// Label returns stable display text for this key.
func (k Key) Label() string {
	switch k.kind {
	case keyKindText:
		return k.text
	case keyKindCode:
		return codeLabel(k.code)
	case keyKindModified:
		parts := modifierLabels(k.mod)
		base := codeLabel(k.code)
		if base == "" && k.code != 0 {
			base = string(k.code)
		}
		if base != "" {
			parts = append(parts, base)
		}
		return strings.Join(parts, "+")
	default:
		return ""
	}
}

func codeLabel(code rune) string {
	return codeLabels[code]
}

func modifierLabels(mod Mod) []string {
	labels := make([]string, 0, 6)
	if mod&ModCtrl != 0 {
		labels = append(labels, "ctrl")
	}
	if mod&ModAlt != 0 {
		labels = append(labels, "alt")
	}
	if mod&ModShift != 0 {
		labels = append(labels, "shift")
	}
	if mod&ModMeta != 0 {
		labels = append(labels, "meta")
	}
	if mod&ModHyper != 0 {
		labels = append(labels, "hyper")
	}
	if mod&ModSuper != 0 {
		labels = append(labels, "super")
	}
	return labels
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
		if opt == nil {
			continue
		}
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

// appendKey returns an owned copy of sequence with key appended.
func appendKey(sequence Seq, key Key) Seq {
	appended := make(Seq, len(sequence)+1)
	copy(appended, sequence)
	appended[len(sequence)] = key
	return appended
}
