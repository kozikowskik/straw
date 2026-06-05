package straw

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

type bindingIndex[A comparable] struct {
	bindings []Binding[A]
}

// newBindingIndex validates bindings and stores safe copies for resolver use.
func newBindingIndex[A comparable](bindings []Binding[A]) (*bindingIndex[A], error) {
	var errs []error
	seen := map[string]int{}
	cloned := make([]Binding[A], len(bindings))

	for bindingIndex, binding := range bindings {
		sequence := binding.Sequence()
		if len(sequence) == 0 {
			errs = append(errs, fmt.Errorf("%w: binding %d has empty sequence", ErrInvalidBinding, bindingIndex))
		}
		for keyIndex, key := range sequence {
			if err := validateKey(key); err != nil {
				errs = append(errs, fmt.Errorf("binding %d sequence %d: %w", bindingIndex, keyIndex, err))
			}
		}

		fingerprint := seqFingerprint(sequence)
		if first, ok := seen[fingerprint]; ok {
			errs = append(errs, fmt.Errorf("%w: binding %d duplicates binding %d: %s", ErrDuplicateSequence, bindingIndex, first, fingerprint))
		} else {
			seen[fingerprint] = bindingIndex
		}

		cloned[bindingIndex] = Bind(binding.Action(), sequence, Description(binding.Description()))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &bindingIndex[A]{bindings: cloned}, nil
}

// validateKey enforces the public key-builder contract at resolver construction time.
func validateKey(key Key) error {
	switch key.kind {
	case keyKindText:
		if key.text == "" {
			return fmt.Errorf("%w: Text must not be empty", ErrInvalidKey)
		}
		if utf8.RuneCountInString(key.text) != 1 {
			return fmt.Errorf("%w: Text(%q) must contain exactly one rune", ErrInvalidKey, key.text)
		}
		r, _ := utf8.DecodeRuneInString(key.text)
		if unicode.IsSpace(r) {
			return fmt.Errorf("%w: whitespace text keys are invalid", ErrInvalidKey)
		}
	case keyKindCode:
		if isPrintableRegularKeyCode(key.code) {
			return fmt.Errorf("%w: Code(%q) is printable; use Text", ErrInvalidKey, key.code)
		}
	case keyKindModified:
		if key.mod == 0 {
			return fmt.Errorf("%w: Modified key must include at least one modifier", ErrInvalidKey)
		}
	default:
		return fmt.Errorf("%w: unknown key kind", ErrInvalidKey)
	}
	return nil
}

// isPrintableRegularKeyCode catches text-like keys that should use Text instead of Code.
func isPrintableRegularKeyCode(code rune) bool {
	return code != tea.KeySpace && unicode.IsPrint(code)
}

// seqFingerprint builds a stable identity for detecting duplicate binding sequences.
func seqFingerprint(sequence Seq) string {
	var builder strings.Builder
	for _, key := range sequence {
		builder.WriteString(keyFingerprint(key))
		builder.WriteByte(' ')
	}
	return builder.String()
}

// keyFingerprint encodes one key into the sequence fingerprint format.
func keyFingerprint(key Key) string {
	switch key.kind {
	case keyKindText:
		return fmt.Sprintf("text:%s", key.text)
	case keyKindCode:
		return fmt.Sprintf("code:%d", key.code)
	case keyKindModified:
		return fmt.Sprintf("modified:%d:%d", key.code, key.mod)
	default:
		return "unknown"
	}
}
