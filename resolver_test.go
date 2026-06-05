package straw

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestNewAcceptsEmptyBindings(t *testing.T) {
	tests := []struct {
		name     string
		bindings []Binding[testAction]
	}{
		{name: "nil", bindings: nil},
		{name: "empty slice", bindings: []Binding[testAction]{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := New(tt.bindings)
			if err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
			if resolver == nil {
				t.Fatal("New() resolver = nil, want resolver")
			}
			if resolver.Pending() {
				t.Fatal("Pending() = true, want false")
			}
		})
	}
}

func TestNewRejectsInvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "zero", timeout: 0},
		{name: "negative", timeout: -time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New[testAction](nil, WithTimeout(tt.timeout))
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
			}
		})
	}
}

func TestNewAcceptsResolverOptions(t *testing.T) {
	resolver, err := New[testAction](nil,
		WithTimeout(250*time.Millisecond),
		WithCancelKeys(Code(tea.KeyEsc), Modified('c', tea.ModCtrl)),
		WithFailedPendingPassThrough(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if resolver == nil {
		t.Fatal("New() resolver = nil, want resolver")
	}
}
