# Changelog

All notable changes to `straw` will be documented in this file.

## v0.2.0 - Unreleased

### Added

- Added `Resolver.PendingSequence()` for reading the active pending key sequence safely.
- Added `Resolver.NextChoices()` and `NextChoice[A]` for rendering immediate which-key-style choices from resolver state.
- Added `Resolver.PendingStatus()` and `PendingStatus[A]` for grouped pending-sequence inspection, including the exact binding waiting behind an ambiguous prefix.
- Added `Key.Label()` for stable display labels across text, special, and modified keys.
- Forwarded resolver query APIs through the Bubble Tea v1 and v2 adapter resolvers.

### Changed

- Raised the minimum Go version from 1.25.0 to 1.26.0.
- Updated `charm.land/bubbletea/v2` from v2.0.7 to v2.1.1 and updated indirect dependencies.

### Documentation

- Documented pending-sequence inspection and added a runnable `Resolver.PendingStatus()` example.
- Added guidance for nested Bubble Tea models and per-screen resolvers.
- Simplified the README and clarified current limitations. Benchmark instructions are available in [CONTRIBUTING.md](CONTRIBUTING.md#benchmarks).

## v0.1.1 - 2026-06-16

- Clarified the distinction between ordered key sequences and modified keys, and documented that simultaneous non-modifier chords are not supported.

## v0.1.0 - 2026-06-16

- Initial public release of the version-neutral resolver core.
- Bubble Tea v2 adapter at `github.com/kozikowskik/straw/bubbletea/v2`.
- Bubble Tea v1 adapter at `github.com/kozikowskik/straw/bubbletea/v1`.
- Vim-like pending sequence timeout, cancel, reset, and stale timeout behavior.
- Typed action bindings, structured key helpers, pass-through handling, examples, docs, tests, and benchmark smoke coverage.
