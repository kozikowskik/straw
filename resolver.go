package straw

import (
	"sync/atomic"
	"time"
)

var nextResolverID uint64

// Timeout is an opaque token returned by UpdateKey when an adapter should schedule a timeout.
type Timeout[A comparable] struct {
	resolverID uint64
	generation uint64
	duration   time.Duration
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

	id := atomic.AddUint64(&nextResolverID, 1)

	return &Resolver[A]{options: options, index: index, id: id}, nil
}

// UpdateKey accepts one version-neutral key press and returns a resolver result plus optional timeout token.
func (r *Resolver[A]) UpdateKey(key Key) (Result[A], Timeout[A]) {
	if len(r.pendingSeq) > 0 && seqContains(r.options.cancelKeys, key) {
		result := canceledResult[A](key, r.pendingSeq)
		r.clearPending()
		return result, Timeout[A]{}
	}

	status := r.index.lookupWithKey(r.pendingSeq, key)

	if status.hasMatch && !status.hasPrefix {
		r.clearPending()
		return matchedResult(status.binding, key), Timeout[A]{}
	}

	if status.hasPrefix {
		attempted := appendKey(r.pendingSeq, key)
		r.pendingSeq = attempted
		r.hasPendingMatch = status.hasMatch
		if status.hasMatch {
			r.pendingMatch = status.binding
		}
		r.generation++
		return pendingResult[A](key, attempted), r.timeoutToken()
	}

	passThrough := len(r.pendingSeq) == 0 || r.options.failedPendingPassThrough
	if len(r.pendingSeq) == 0 {
		r.clearPending()
		return unmatchedKeyResult[A](key, passThrough), Timeout[A]{}
	}
	result := unmatchedTailResult[A](key, r.pendingSeq, passThrough)
	r.clearPending()
	return result, Timeout[A]{}
}

// UpdateTimeout resolves or ignores timeout tokens for the current pending sequence.
func (r *Resolver[A]) UpdateTimeout(timeout Timeout[A]) Result[A] {
	if !r.acceptsTimeout(timeout) {
		return idleResult[A]()
	}
	if r.hasPendingMatch {
		return r.resolvePendingMatch()
	}
	return r.cancelPendingTimeout()
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() {
	r.clearPending()
}

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool {
	return len(r.pendingSeq) > 0
}

// PendingSequence returns a safe copy of the active pending sequence.
func (r *Resolver[A]) PendingSequence() Seq {
	return cloneSeq(r.pendingSeq)
}

// Scheduled reports whether this timeout token should be scheduled by an adapter.
func (t Timeout[A]) Scheduled() bool {
	return t.resolverID != 0
}

// Duration returns the timeout duration configured on the resolver that emitted this token.
func (t Timeout[A]) Duration() time.Duration {
	return t.duration
}

// acceptsTimeout reports whether a timeout token belongs to the active pending generation.
func (r *Resolver[A]) acceptsTimeout(timeout Timeout[A]) bool {
	return timeout.resolverID == r.id && timeout.generation == r.generation && len(r.pendingSeq) > 0
}

// resolvePendingMatch accepts the exact binding that was waiting for possible continuation.
func (r *Resolver[A]) resolvePendingMatch() Result[A] {
	binding := r.pendingMatch
	r.clearPending()
	return matchedResult(binding, Key{})
}

// cancelPendingTimeout cancels a pending prefix that did not have an exact match to resolve.
func (r *Resolver[A]) cancelPendingTimeout() Result[A] {
	result := canceledResult[A](Key{}, r.pendingSeq)
	r.clearPending()
	return result
}

// clearPending removes pending sequence state and invalidates outstanding timeout messages.
func (r *Resolver[A]) clearPending() {
	r.pendingSeq = nil
	r.pendingMatch = Binding[A]{}
	r.hasPendingMatch = false
	r.generation++
}

// timeoutToken returns an opaque token for the current pending generation.
func (r *Resolver[A]) timeoutToken() Timeout[A] {
	return Timeout[A]{resolverID: r.id, generation: r.generation, duration: r.options.timeout}
}
