package straw

import (
	"errors"
	"fmt"
	"time"
)

const defaultTimeout = 500 * time.Millisecond

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

// defaultResolverOptions builds the baseline configuration used when callers omit options.
func defaultResolverOptions() resolverOptions {
	return resolverOptions{
		timeout:    defaultTimeout,
		cancelKeys: Sequence(Code(KeyEsc)),
	}
}

// buildResolverOptions applies caller options and validates the resolved configuration.
func buildResolverOptions(options []Option) (resolverOptions, error) {
	resolved := defaultResolverOptions()
	var errs []error
	for index, option := range options {
		if option == nil {
			errs = append(errs, fmt.Errorf("%w: option %d is nil", ErrInvalidOption, index))
			continue
		}
		option(&resolved)
	}

	if resolved.timeout <= 0 {
		errs = append(errs, fmt.Errorf("%w: timeout must be greater than zero", ErrInvalidOption))
	}
	for index, key := range resolved.cancelKeys {
		if err := validateKey(key); err != nil {
			errs = append(errs, fmt.Errorf("cancel key %d: %w", index, err))
		}
	}
	if len(errs) > 0 {
		return resolved, errors.Join(errs...)
	}
	return resolved, nil
}
