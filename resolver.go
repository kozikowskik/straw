package straw

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

var nextResolverID uint64

type resolverTimeoutMsg struct {
	resolverID uint64
	generation uint64
}

// Resolver tracks key sequence state for application-owned actions.
type Resolver[A comparable] struct {
	options         resolverOptions
	index           *bindingIndex[A]
	id              uint64
	generation      uint64
	pendingSeq      Seq
	pendingMatch    Binding[A]
	hasPendingMatch bool
}

// New validates resolver options and builds a resolver.
func New[A comparable](bindings []Binding[A], opts ...Option) (*Resolver[A], error) {
	options, err := buildResolverOptions(opts)
	if err != nil {
		return nil, err
	}

	index, err := newBindingIndex(bindings)
	if err != nil {
		return nil, err
	}

	nextResolverID++

	return &Resolver[A]{options: options, index: index, id: nextResolverID}, nil
}

// Update accepts Bubble Tea messages and returns the resolver result plus command.
func (r *Resolver[A]) Update(msg tea.Msg) (Result[A], tea.Cmd) {
	if timeout, ok := msg.(resolverTimeoutMsg); ok {
		return r.handleTimeout(timeout)
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return idleResult[A](), nil
	}

	key := keyPressMsgToKey(keyMsg)
	if len(r.pendingSeq) > 0 && seqContains(r.options.cancelKeys, key) {
		canceled := cloneSeq(r.pendingSeq)
		r.clearPending()
		return canceledResult[A](key, canceled), nil
	}

	attempted := append(cloneSeq(r.pendingSeq), key)
	status := r.index.lookup(attempted)

	if status.hasMatch && !status.hasPrefix {
		r.clearPending()
		return matchedResult(status.binding, key), nil
	}

	if status.hasPrefix {
		r.pendingSeq = attempted
		r.hasPendingMatch = status.hasMatch
		if status.hasMatch {
			r.pendingMatch = status.binding
		}
		r.generation++
		return pendingResult[A](key, attempted), r.timeoutCommand()
	}

	passThrough := len(r.pendingSeq) == 0 || r.options.failedPendingPassThrough
	r.clearPending()
	return unmatchedResult[A](key, attempted, passThrough), nil
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() {
	r.clearPending()
}

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool {
	return len(r.pendingSeq) > 0
}

// handleTimeout resolves or ignores timeout messages for the current pending sequence.
func (r *Resolver[A]) handleTimeout(timeout resolverTimeoutMsg) (Result[A], tea.Cmd) {
	if !r.acceptsTimeout(timeout) {
		return idleResult[A](), nil
	}
	if r.hasPendingMatch {
		return r.resolvePendingMatch()
	}
	return r.cancelPendingTimeout()
}

// acceptsTimeout reports whether a timeout message belongs to the active pending generation.
func (r *Resolver[A]) acceptsTimeout(timeout resolverTimeoutMsg) bool {
	return timeout.resolverID == r.id && timeout.generation == r.generation && len(r.pendingSeq) > 0
}

// resolvePendingMatch accepts the exact binding that was waiting for possible continuation.
func (r *Resolver[A]) resolvePendingMatch() (Result[A], tea.Cmd) {
	binding := r.pendingMatch
	r.clearPending()
	return matchedResult(binding, Key{}), nil
}

// cancelPendingTimeout cancels a pending prefix that did not have an exact match to resolve.
func (r *Resolver[A]) cancelPendingTimeout() (Result[A], tea.Cmd) {
	canceled := cloneSeq(r.pendingSeq)
	r.clearPending()
	return canceledResult[A](Key{}, canceled), nil
}

// clearPending removes pending sequence state and invalidates outstanding timeout messages.
func (r *Resolver[A]) clearPending() {
	r.pendingSeq = nil
	r.pendingMatch = Binding[A]{}
	r.hasPendingMatch = false
	r.generation++
}

// timeoutCommand returns a command that resolves the current pending generation after the timeout.
func (r *Resolver[A]) timeoutCommand() tea.Cmd {
	resolverID := r.id
	generation := r.generation
	duration := r.options.timeout
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return resolverTimeoutMsg{resolverID: resolverID, generation: generation}
	})
}

// keyPressMsgToKey converts Bubble Tea key press data into the resolver's key model.
func keyPressMsgToKey(msg tea.KeyPressMsg) Key {
	key := msg.Key()
	if key.Mod != 0 {
		return Modified(key.Code, key.Mod)
	}
	if key.Text != "" {
		return Text(key.Text)
	}
	return Code(key.Code)
}
