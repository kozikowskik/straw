package straw

// State identifies the outcome of a resolver update.
type State int

const (
	// Idle means the resolver did not produce key handling behavior.
	Idle State = iota
)

// Result describes the outcome of a resolver update.
type Result[A comparable] struct {
	state State
}

// idleResult builds the default no-op result used before matching behavior exists.
func idleResult[A comparable]() Result[A] {
	return Result[A]{state: Idle}
}
