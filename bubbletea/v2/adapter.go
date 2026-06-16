package v2

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/kozikowskik/straw"
)

type Key = straw.Key
type Seq = straw.Seq
type Binding[A comparable] = straw.Binding[A]
type BindingOption = straw.BindingOption
type Option = straw.Option
type Result[A comparable] = straw.Result[A]
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

func Text(value string) Key           { return straw.Text(value) }
func TextSequence(value string) Seq   { return straw.TextSequence(value) }
func Code(code rune) Key              { return straw.Code(code) }
func Modified(code rune, mod Mod) Key { return straw.Modified(code, mod) }
func Sequence(keys ...Key) Seq        { return straw.Sequence(keys...) }
func Bind[A comparable](action A, sequence Seq, opts ...BindingOption) Binding[A] {
	return straw.Bind(action, sequence, opts...)
}
func Description(text string) BindingOption     { return straw.Description(text) }
func WithTimeout(duration time.Duration) Option { return straw.WithTimeout(duration) }
func WithCancelKeys(keys ...Key) Option         { return straw.WithCancelKeys(keys...) }
func WithFailedPendingPassThrough(enabled bool) Option {
	return straw.WithFailedPendingPassThrough(enabled)
}
func ShouldPassThrough[A comparable](result Result[A]) bool { return straw.ShouldPassThrough(result) }

// Resolver adapts the version-neutral straw resolver to Bubble Tea v2 messages and commands.
type Resolver[A comparable] struct {
	core *straw.Resolver[A]
}

// New validates resolver options and builds a Bubble Tea v2 resolver adapter.
func New[A comparable](bindings []Binding[A], opts ...Option) (*Resolver[A], error) {
	core, err := straw.New(bindings, opts...)
	if err != nil {
		return nil, err
	}
	return &Resolver[A]{core: core}, nil
}

// Update accepts Bubble Tea v2 messages and returns a straw result plus v2 command.
func (r *Resolver[A]) Update(msg tea.Msg) (Result[A], tea.Cmd) {
	if timeout, ok := msg.(timeoutMsg[A]); ok {
		return r.core.UpdateTimeout(timeout.token), nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return zeroResult[A](), nil
	}

	result, timeout := r.core.UpdateKey(keyPressMsgToKey(keyMsg))
	return result, r.timeoutCommand(timeout)
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() { r.core.Reset() }

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool { return r.core.Pending() }

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

func keyPressMsgToKey(msg tea.KeyPressMsg) Key {
	key := msg.Key()
	if key.Text != "" {
		return Text(key.Text)
	}
	if hasModifiers(key.Mod) {
		return Modified(codeFromV2(key.Code), modFromV2(key.Mod))
	}
	return Code(codeFromV2(key.Code))
}

func hasModifiers(mod tea.KeyMod) bool {
	return mod != 0
}

func modFromV2(mod tea.KeyMod) Mod {
	var converted Mod
	if mod&tea.ModAlt != 0 {
		converted |= ModAlt
	}
	if mod&tea.ModCtrl != 0 {
		converted |= ModCtrl
	}
	if mod&tea.ModShift != 0 {
		converted |= ModShift
	}
	if mod&tea.ModMeta != 0 {
		converted |= ModMeta
	}
	if mod&tea.ModHyper != 0 {
		converted |= ModHyper
	}
	if mod&tea.ModSuper != 0 {
		converted |= ModSuper
	}
	return converted
}

func codeFromV2(code rune) rune {
	switch code {
	case tea.KeyEsc:
		return KeyEsc
	case tea.KeySpace:
		return KeySpace
	case tea.KeyTab:
		return KeyTab
	case tea.KeyEnter:
		return KeyEnter
	case tea.KeyUp:
		return KeyUp
	case tea.KeyDown:
		return KeyDown
	case tea.KeyRight:
		return KeyRight
	case tea.KeyLeft:
		return KeyLeft
	case tea.KeyHome:
		return KeyHome
	case tea.KeyEnd:
		return KeyEnd
	case tea.KeyPgUp:
		return KeyPgUp
	case tea.KeyPgDown:
		return KeyPgDown
	case tea.KeyDelete:
		return KeyDelete
	case tea.KeyInsert:
		return KeyInsert
	case tea.KeyF1:
		return KeyF1
	case tea.KeyF2:
		return KeyF2
	case tea.KeyF3:
		return KeyF3
	case tea.KeyF4:
		return KeyF4
	case tea.KeyF5:
		return KeyF5
	case tea.KeyF6:
		return KeyF6
	case tea.KeyF7:
		return KeyF7
	case tea.KeyF8:
		return KeyF8
	case tea.KeyF9:
		return KeyF9
	case tea.KeyF10:
		return KeyF10
	case tea.KeyF11:
		return KeyF11
	case tea.KeyF12:
		return KeyF12
	default:
		return code
	}
}

func zeroResult[A comparable]() Result[A] {
	var result Result[A]
	return result
}
