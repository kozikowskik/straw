package straw

import "errors"

// ErrInvalidBinding identifies binding configuration that cannot be used.
var ErrInvalidBinding = errors.New("invalid binding")

// ErrInvalidKey identifies a key value that cannot be matched safely.
var ErrInvalidKey = errors.New("invalid key")

// ErrDuplicateSequence identifies two bindings with the same key sequence.
var ErrDuplicateSequence = errors.New("duplicate sequence")

// ErrInvalidOption identifies resolver options with invalid values.
var ErrInvalidOption = errors.New("invalid option")
