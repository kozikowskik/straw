package straw

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
)

const defaultTimeout = time.Second

type resolverOptions struct {
	timeout                  time.Duration
	cancelKeys               Seq
	failedPendingPassThrough bool
}

// Option configures resolver behavior.
type Option func(*resolverOptions)

// WithTimeout configures how long an ambiguous sequence can stay pending.
func WithTimeout(duration time.Duration) Option {
	return func(opts *resolverOptions) {
		opts.timeout = duration
	}
}

// WithCancelKeys configures keys that cancel a pending sequence.
func WithCancelKeys(keys ...Key) Option {
	return func(opts *resolverOptions) {
		opts.cancelKeys = cloneSeq(keys)
	}
}

// WithFailedPendingPassThrough configures whether failed pending keys pass through.
func WithFailedPendingPassThrough(enabled bool) Option {
	return func(opts *resolverOptions) {
		opts.failedPendingPassThrough = enabled
	}
}

func defaultResolverOptions() resolverOptions {
	return resolverOptions{
		timeout:    defaultTimeout,
		cancelKeys: Sequence(Code(tea.KeyEsc)),
	}
}

func buildResolverOptions(options []Option) (resolverOptions, error) {
	resolved := defaultResolverOptions()
	for _, option := range options {
		option(&resolved)
	}
	if resolved.timeout <= 0 {
		return resolved, fmt.Errorf("%w: timeout must be greater than zero", ErrInvalidOption)
	}
	return resolved, nil
}
