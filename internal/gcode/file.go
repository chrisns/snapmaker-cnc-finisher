package gcode

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// ReadGCodeFile reads a GCode file and returns all lines
// Uses bufio.Scanner for memory-efficient streaming
func ReadGCodeFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines (default is 64KB)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line length

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return lines, nil
}

// WriteGCodeFile writes lines to a GCode file with buffering
// Flush strategy: flush every 1000 lines OR on completion
func WriteGCodeFile(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush() // Ensure flush on error via defer

	for i, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line %d: %w", i, err)
		}

		// Flush every 1000 lines for memory efficiency
		if (i+1)%1000 == 0 {
			if err := writer.Flush(); err != nil {
				return fmt.Errorf("failed to flush at line %d: %w", i, err)
			}
		}
	}

	return nil
}

// BufferedWriter wraps a buffered writer for incremental GCode writing
type BufferedWriter struct {
	writer    *bufio.Writer
	lineCount int
}

// NewBufferedWriter creates a new buffered writer for GCode files
func NewBufferedWriter(w io.Writer) *BufferedWriter {
	return &BufferedWriter{
		writer: bufio.NewWriter(w),
	}
}

// WriteLine writes a single line to the buffer
func (bw *BufferedWriter) WriteLine(line string) error {
	if _, err := bw.writer.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write line: %w", err)
	}

	bw.lineCount++

	// Auto-flush every 1000 lines
	if bw.lineCount%1000 == 0 {
		if err := bw.writer.Flush(); err != nil {
			return fmt.Errorf("failed to auto-flush: %w", err)
		}
	}

	return nil
}

// Flush ensures all buffered data is written
func (bw *BufferedWriter) Flush() error {
	if err := bw.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}
	return nil
}

// LineCount returns the number of lines written
func (bw *BufferedWriter) LineCount() int {
	return bw.lineCount
}
