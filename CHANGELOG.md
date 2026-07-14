# Changelog

All notable changes to `straw` will be documented in this file.

## v0.2.0 - Unreleased

- Added `Resolver.PendingSequence()` for reading the active pending key sequence safely.
- Added `Resolver.NextChoices()` and `NextChoice[A]` for rendering immediate which-key-style choices from resolver state.
- Added `Key.Label()` for stable display labels across text, special, and modified keys.
- Forwarded resolver query APIs through the Bubble Tea v1 and v2 adapter resolvers.
- Documented resolver query APIs and added tests and benchmarks for the new behavior.

## v0.1.2 - 2026-06-17

- Cleaned up README and user-facing docs.
- Removed maintainer-focused benchmark instructions from the README.
- Added guidance for nested Bubble Tea models and per-screen resolvers.
- Reworded current limitations so general docs do not need release-version updates.

## v0.1.0 - 2026-06-16

- Initial public release of the version-neutral resolver core.
- Bubble Tea v2 adapter at `github.com/kozikowskik/straw/bubbletea/v2`.
- Bubble Tea v1 adapter at `github.com/kozikowskik/straw/bubbletea/v1`.
- Vim-like pending sequence timeout, cancel, reset, and stale timeout behavior.
- Typed action bindings, structured key helpers, pass-through handling, examples, docs, tests, and benchmark smoke coverage.
