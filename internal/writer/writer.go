// Package writer handles GCode output writing with format preservation.
package writer

import (
	"fmt"
	"io"

	"github.com/256dpi/gcode"
)

// Writer handles GCode output with format preservation.
type Writer struct {
	w io.Writer
}

// NewWriter creates a writer for the given io.Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteLine writes a single GCode line.
func (wr *Writer) WriteLine(line gcode.Line) error {
	_, err := fmt.Fprintf(wr.w, "%s\n", line.String())
	return err
}

// WriteFile writes an entire gcode.File.
func (wr *Writer) WriteFile(file *gcode.File) error {
	return gcode.WriteFile(wr.w, file)
}
