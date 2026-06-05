package straw

import tea "charm.land/bubbletea/v2"

// Resolver tracks key sequence state for application-owned actions.
type Resolver[A comparable] struct {
	options resolverOptions
}

// New validates resolver options and builds a resolver.
func New[A comparable](bindings []Binding[A], opts ...Option) (*Resolver[A], error) {
	options, err := buildResolverOptions(opts)
	if err != nil {
		return nil, err
	}

	return &Resolver[A]{options: options}, nil
}

// Update accepts Bubble Tea messages and returns the resolver result plus command.
func (r *Resolver[A]) Update(msg tea.Msg) (Result[A], tea.Cmd) {
	return idleResult[A](), nil
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() {}

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool {
	return false
}
