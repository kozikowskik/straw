package straw

// State identifies the outcome of a resolver update.
type State int

const (
	// Idle means the resolver did not produce key handling behavior.
	Idle State = iota
	// Pending means the resolver is waiting for more keys in a sequence.
	Pending
	// Matched means the resolver matched a binding.
	Matched
	// Unmatched means the resolver did not match a binding.
	Unmatched
	// Canceled means a pending sequence was canceled.
	Canceled
)

// Result describes the outcome of a resolver update.
type Result[A comparable] struct {
	state        State
	binding      Binding[A]
	hasBinding   bool
	key          Key
	sequence     Seq
	sequenceKey  bool
	sequenceTail bool
	passThrough  bool
}

// idleResult builds the default no-op result.
func idleResult[A comparable]() Result[A] {
	return Result[A]{state: Idle}
}

// pendingResult records a partial sequence that may match after more keys.
func pendingResult[A comparable](key Key, sequence Seq) Result[A] {
	return Result[A]{state: Pending, key: key, sequence: sequence}
}

// matchedResult records the binding matched by the latest key.
func matchedResult[A comparable](binding Binding[A], key Key) Result[A] {
	return Result[A]{state: Matched, binding: binding, hasBinding: true, key: key, sequence: binding.sequence}
}

// unmatchedResult records a failed sequence and whether the latest key should pass through.
func unmatchedResult[A comparable](key Key, sequence Seq, passThrough bool) Result[A] {
	return Result[A]{state: Unmatched, key: key, sequence: sequence, passThrough: passThrough}
}

// unmatchedTailResult records a failed pending sequence without eagerly appending the latest key.
func unmatchedTailResult[A comparable](key Key, sequence Seq, passThrough bool) Result[A] {
	return Result[A]{state: Unmatched, key: key, sequence: sequence, sequenceTail: true, passThrough: passThrough}
}

// unmatchedKeyResult records an idle unmatched key without allocating a one-key sequence.
func unmatchedKeyResult[A comparable](key Key, passThrough bool) Result[A] {
	return Result[A]{state: Unmatched, key: key, sequenceKey: true, passThrough: passThrough}
}

// canceledResult records a pending sequence canceled by the latest key.
func canceledResult[A comparable](key Key, sequence Seq) Result[A] {
	return Result[A]{state: Canceled, key: key, sequence: sequence}
}

// Match reports whether the result matched the given action.
func (r Result[A]) Match(action A) bool {
	return r.hasBinding && r.binding.Action() == action
}

// Binding returns the matched binding, if any.
func (r Result[A]) Binding() (Binding[A], bool) {
	return r.binding, r.hasBinding
}

// State returns the result state.
func (r Result[A]) State() State {
	return r.state
}

// IsIdle reports whether the result state is Idle.
func (r Result[A]) IsIdle() bool {
	return r.state == Idle
}

// IsPending reports whether the result state is Pending.
func (r Result[A]) IsPending() bool {
	return r.state == Pending
}

// IsMatched reports whether the result state is Matched.
func (r Result[A]) IsMatched() bool {
	return r.state == Matched
}

// IsUnmatched reports whether the result state is Unmatched.
func (r Result[A]) IsUnmatched() bool {
	return r.state == Unmatched
}

// IsCanceled reports whether the result state is Canceled.
func (r Result[A]) IsCanceled() bool {
	return r.state == Canceled
}

// PassThrough reports whether the latest key should be handled by the host application.
func (r Result[A]) PassThrough() bool {
	return r.passThrough
}

// ShouldPassThrough reports whether host key handling should run for a result.
func ShouldPassThrough[A comparable](result Result[A]) bool {
	return result.IsUnmatched() && result.PassThrough()
}

// Key returns the latest key that contributed to this result.
func (r Result[A]) Key() Key {
	return r.key
}

// Sequence returns a copy of the key sequence associated with this result.
func (r Result[A]) Sequence() Seq {
	if r.sequenceKey {
		return Sequence(r.key)
	}
	if r.sequenceTail {
		return appendKey(r.sequence, r.key)
	}
	return cloneSeq(r.sequence)
}
