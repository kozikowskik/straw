package straw

import tea "charm.land/bubbletea/v2"

// Resolver tracks key sequence state for application-owned actions.
type Resolver[A comparable] struct {
	options    resolverOptions
	index      *bindingIndex[A]
	pendingSeq Seq
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

	return &Resolver[A]{options: options, index: index}, nil
}

// Update accepts Bubble Tea messages and returns the resolver result plus command.
func (r *Resolver[A]) Update(msg tea.Msg) (Result[A], tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return idleResult[A](), nil
	}

	key := keyPressMsgToKey(keyMsg)
	attempted := append(cloneSeq(r.pendingSeq), key)
	status := r.index.lookup(attempted)

	if status.hasMatch && !status.hasPrefix {
		r.pendingSeq = nil
		return matchedResult(status.binding, key), nil
	}

	if status.hasPrefix {
		r.pendingSeq = attempted
		return pendingResult[A](key, attempted), r.timeoutCommand()
	}

	r.pendingSeq = nil
	return idleResult[A](), nil
}

// Reset clears any pending sequence state.
func (r *Resolver[A]) Reset() {
	r.pendingSeq = nil
}

// Pending reports whether the resolver is waiting for more keys.
func (r *Resolver[A]) Pending() bool {
	return len(r.pendingSeq) > 0
}

// timeoutCommand returns the placeholder command used until real timeout messages are implemented.
func (r *Resolver[A]) timeoutCommand() tea.Cmd {
	return func() tea.Msg { return nil }
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
