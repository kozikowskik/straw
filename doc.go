// Package straw provides a version-neutral key sequence resolver for terminal applications.
//
// Bubble Tea emits one key message at a time. Straw turns those key presses into higher-level resolver results: pending prefixes, matched bindings, unmatched input, canceled pending input, and idle updates.
//
// Applications define their own comparable action type, bind actions to key sequences with Bind, then pass keys to Resolver.UpdateKey. The resolver reports a Result so the application can decide which action to run and whether unmatched keys should continue to normal host key handling.
//
// Use package github.com/kozikowskik/straw/bubbletea/v2 for Bubble Tea v2 applications and github.com/kozikowskik/straw/bubbletea/v1 for Bubble Tea v1 applications. Those adapter packages translate Bubble Tea key messages to straw keys and convert pending sequence timeouts into Bubble Tea commands.
package straw
