// Package v2 adapts straw's version-neutral resolver to Bubble Tea v2 messages and commands.
//
// KeyPressMsg values are converted into straw keys. KeyReleaseMsg and unrelated messages are ignored.
// ShiftedCode, BaseCode, IsRepeat, and extended key fields are not part of straw's current matching model.
package v2
