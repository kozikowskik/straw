package v1

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kozikowskik/straw"
)

type Key = straw.Key
type Seq = straw.Seq
type Binding[A comparable] = straw.Binding[A]
type BindingOption = straw.BindingOption
type Option = straw.Option
type Result[A comparable] = straw.Result[A]
type NextChoice[A comparable] = straw.NextChoice[A]
type State = straw.State
type Mod = straw.Mod

const (
	Idle      = straw.Idle
	Pending   = straw.Pending
	Matched   = straw.Matched
	Unmatched = straw.Unmatched
	Canceled  = straw.Canceled

	ModAlt   = straw.ModAlt
	ModCtrl  = straw.ModCtrl
	ModShift = straw.ModShift
	ModMeta  = straw.ModMeta
	ModHyper = straw.ModHyper
	ModSuper = straw.ModSuper

	KeyBackspace = straw.KeyBackspace
	KeyTab       = straw.KeyTab
	KeyEnter     = straw.KeyEnter
	KeyEsc       = straw.KeyEsc
	KeySpace     = straw.KeySpace
	KeyUp        = straw.KeyUp
	KeyDown      = straw.KeyDown
	KeyRight     = straw.KeyRight
	KeyLeft      = straw.KeyLeft
	KeyHome      = straw.KeyHome
	KeyEnd       = straw.KeyEnd
	KeyPgUp      = straw.KeyPgUp
	KeyPgDown    = straw.KeyPgDown
	KeyDelete    = straw.KeyDelete
	KeyInsert    = straw.KeyInsert
	KeyF1        = straw.KeyF1
	KeyF2        = straw.KeyF2
	KeyF3        = straw.KeyF3
	KeyF4        = straw.KeyF4
	KeyF5        = straw.KeyF5
	KeyF6        = straw.KeyF6
	KeyF7        = straw.KeyF7
	KeyF8        = straw.KeyF8
	KeyF9        = straw.KeyF9
	KeyF10       = straw.KeyF10
	KeyF11       = straw.KeyF11
	KeyF12       = straw.KeyF12
)

func Text(value string) Key                     { return straw.Text(value) }
func TextSequence(value string) Seq             { return straw.TextSequence(value) }
func Code(code rune) Key                        { return straw.Code(code) }
func Modified(code rune, mod Mod) Key           { return straw.Modified(code, mod) }
func Sequence(keys ...Key) Seq                  { return straw.Sequence(keys...) }
func Description(text string) BindingOption     { return straw.Description(text) }
func WithTimeout(duration time.Duration) Option { return straw.WithTimeout(duration) }
func Bind[A comparable](action A, sequence Seq, opts ...BindingOption) Binding[A] {
	return straw.Bind(action, sequence, opts...)
}
func WithCancelKeys(keys ...Key) Option { return straw.WithCancelKeys(keys...) }
func WithFailedPendingPassThrough(enabled bool) Option {
	return straw.WithFailedPendingPassThrough(enabled)
}
func ShouldPassThrough[A comparable](result Result[A]) bool { return straw.ShouldPassThrough(result) }

// Resolver adapts the version-neutral straw resolver to Bubble Tea v1 messages and commands.
type Resolver[A comparable] struct {
	core *straw.Resolver[A]
}

// New validates resolver options and builds a Bubble Tea v1 resolver adapter.
func New[A comparable](bindings []Binding[A], opts ...Option) (*Resolver[A], error) {
	core, err := straw.New(bindings, opts...)
	if err != nil {
		return nil, err
	}
	return &Resolver[A]{core: core}, nil
}

// Update accepts Bubble Tea v1 messages and returns a straw result plus v1 command.
func (r *Resolver[A]) Update(msg tea.Msg) (Result[A], tea.Cmd) {
	if timeout, ok := msg.(timeoutMsg[A]); ok {
		return r.core.UpdateTimeout(timeout.token), nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return zeroResult[A](), nil
	}

	key, ok := keyMsgToKey(keyMsg)
	if !ok {
		return zeroResult[A](), nil
	}

	result, timeout := r.core.UpdateKey(key)
	return result, r.timeoutCommand(timeout)
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() { r.core.Reset() }

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool { return r.core.Pending() }

// PendingSequence returns a safe copy of the active pending sequence.
func (r *Resolver[A]) PendingSequence() Seq { return r.core.PendingSequence() }

// NextChoices returns the immediate keys available from the current pending sequence.
func (r *Resolver[A]) NextChoices() []NextChoice[A] { return r.core.NextChoices() }

type timeoutMsg[A comparable] struct {
	token straw.Timeout[A]
}

func (r *Resolver[A]) timeoutCommand(timeout straw.Timeout[A]) tea.Cmd {
	if !timeout.Scheduled() {
		return nil
	}
	return tea.Tick(timeout.Duration(), func(time.Time) tea.Msg {
		return timeoutMsg[A]{token: timeout}
	})
}

func keyMsgToKey(msg tea.KeyMsg) (Key, bool) {
	if msg.Paste {
		return Key{}, false
	}
	if msg.Alt {
		if len(msg.Runes) == 1 {
			return Modified(msg.Runes[0], ModAlt), true
		}
		if code, ok := codeFromV1(msg.Type); ok {
			return Modified(code, ModAlt), true
		}
	}
	if msg.Type == tea.KeyRunes {
		if len(msg.Runes) != 1 {
			return Key{}, false
		}
		return Text(string(msg.Runes[0])), true
	}
	if code, ok := codeFromV1(msg.Type); ok {
		return Code(code), true
	}
	if code, ok := ctrlFromV1(msg.Type); ok {
		return Modified(code, ModCtrl), true
	}
	return Key{}, false
}

func ctrlFromV1(keyType tea.KeyType) (rune, bool) {
	switch keyType {
	case tea.KeyCtrlA:
		return 'a', true
	case tea.KeyCtrlB:
		return 'b', true
	case tea.KeyCtrlC:
		return 'c', true
	case tea.KeyCtrlD:
		return 'd', true
	case tea.KeyCtrlE:
		return 'e', true
	case tea.KeyCtrlF:
		return 'f', true
	case tea.KeyCtrlG:
		return 'g', true
	case tea.KeyCtrlH:
		return 'h', true
	case tea.KeyCtrlI:
		return 'i', true
	case tea.KeyCtrlJ:
		return 'j', true
	case tea.KeyCtrlK:
		return 'k', true
	case tea.KeyCtrlL:
		return 'l', true
	case tea.KeyCtrlM:
		return 'm', true
	case tea.KeyCtrlN:
		return 'n', true
	case tea.KeyCtrlO:
		return 'o', true
	case tea.KeyCtrlP:
		return 'p', true
	case tea.KeyCtrlQ:
		return 'q', true
	case tea.KeyCtrlR:
		return 'r', true
	case tea.KeyCtrlS:
		return 's', true
	case tea.KeyCtrlT:
		return 't', true
	case tea.KeyCtrlU:
		return 'u', true
	case tea.KeyCtrlV:
		return 'v', true
	case tea.KeyCtrlW:
		return 'w', true
	case tea.KeyCtrlX:
		return 'x', true
	case tea.KeyCtrlY:
		return 'y', true
	case tea.KeyCtrlZ:
		return 'z', true
	default:
		return 0, false
	}
}

func codeFromV1(keyType tea.KeyType) (rune, bool) {
	switch keyType {
	case tea.KeyBackspace:
		return KeyBackspace, true
	case tea.KeyTab:
		return KeyTab, true
	case tea.KeyEnter:
		return KeyEnter, true
	case tea.KeyEsc:
		return KeyEsc, true
	case tea.KeySpace:
		return KeySpace, true
	case tea.KeyUp:
		return KeyUp, true
	case tea.KeyDown:
		return KeyDown, true
	case tea.KeyRight:
		return KeyRight, true
	case tea.KeyLeft:
		return KeyLeft, true
	case tea.KeyHome:
		return KeyHome, true
	case tea.KeyEnd:
		return KeyEnd, true
	case tea.KeyPgUp:
		return KeyPgUp, true
	case tea.KeyPgDown:
		return KeyPgDown, true
	case tea.KeyDelete:
		return KeyDelete, true
	case tea.KeyInsert:
		return KeyInsert, true
	case tea.KeyF1:
		return KeyF1, true
	case tea.KeyF2:
		return KeyF2, true
	case tea.KeyF3:
		return KeyF3, true
	case tea.KeyF4:
		return KeyF4, true
	case tea.KeyF5:
		return KeyF5, true
	case tea.KeyF6:
		return KeyF6, true
	case tea.KeyF7:
		return KeyF7, true
	case tea.KeyF8:
		return KeyF8, true
	case tea.KeyF9:
		return KeyF9, true
	case tea.KeyF10:
		return KeyF10, true
	case tea.KeyF11:
		return KeyF11, true
	case tea.KeyF12:
		return KeyF12, true
	default:
		return 0, false
	}
}

func zeroResult[A comparable]() Result[A] {
	var result Result[A]
	return result
}
