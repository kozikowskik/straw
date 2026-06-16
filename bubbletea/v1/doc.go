// Package v1 adapts straw's version-neutral resolver to Bubble Tea v1 messages and commands.
//
// KeyMsg values with one rune are converted to text keys. Ctrl KeyType constants become ctrl-modified keys, and Alt key messages become alt-modified keys. Pasted and multi-rune key messages are ignored so pasted text does not accidentally trigger shortcuts.
package v1
