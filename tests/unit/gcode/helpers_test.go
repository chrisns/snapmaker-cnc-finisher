package gcode_test

import (
	"os"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

// commandsEqual compares two Command structs for equality
// Used by multiple test files for assertion
func commandsEqual(a, b gcode.Command) bool {
	if a.Letter != b.Letter || a.Value != b.Value || a.Comment != b.Comment {
		return false
	}
	if len(a.Params) != len(b.Params) {
		return false
	}
	for k, v := range a.Params {
		if b.Params[k] != v {
			return false
		}
	}
	return true
}

// failingWriter always returns errors - useful for testing error paths
type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, os.ErrClosed
}

// errorReader always returns errors - useful for testing read error paths
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, os.ErrClosed
}
