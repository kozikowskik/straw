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

func TestNewAcceptsSpaceCodeBinding(t *testing.T) {
	resolver, err := New([]Binding[testAction]{
		Bind(testGoHome, Sequence(Code(tea.KeySpace))),
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if resolver == nil {
		t.Fatal("New() resolver = nil, want resolver")
	}
}

func TestNewRejectsInvalidBindings(t *testing.T) {
	tests := []struct {
		name     string
		bindings []Binding[testAction]
		wantErr  error
	}{
		{
			name:     "empty sequence",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence())},
			wantErr:  ErrInvalidBinding,
		},
		{
			name:     "empty text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "multi-rune text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("gg")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "whitespace text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text(" ")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unicode whitespace text key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Text("\u00a0")))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "printable code key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Code('g')))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unicode printable code key",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Code('é')))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "unknown key kind",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Key{}))},
			wantErr:  ErrInvalidKey,
		},
		{
			name:     "modified key without modifier",
			bindings: []Binding[testAction]{Bind(testGoHome, Sequence(Modified('c', 0)))},
			wantErr:  ErrInvalidKey,
		},
		{
			name: "duplicate sequence",
			bindings: []Binding[testAction]{
				Bind(testGoHome, TextSequence("gh")),
				Bind(testCopyLine, TextSequence("gh")),
			},
			wantErr: ErrDuplicateSequence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.bindings)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCollectsMultipleValidationErrors(t *testing.T) {
	_, err := New([]Binding[testAction]{
		Bind(testGoHome, Sequence()),
		Bind(testCopyLine, Sequence(Text(""))),
		Bind(testGoHome, TextSequence("gh")),
		Bind(testCopyLine, TextSequence("gh")),
	})
	if !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("New() error = %v, want ErrInvalidBinding", err)
	}
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("New() error = %v, want ErrInvalidKey", err)
	}
	if !errors.Is(err, ErrDuplicateSequence) {
		t.Fatalf("New() error = %v, want ErrDuplicateSequence", err)
	}
}
