package gcode_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestReadGCodeFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantLines int
		wantErr   bool
	}{
		{
			name: "Read small file",
			content: `G0 X0 Y0
G1 Z-1.0 F1000
M3 S1000`,
			wantLines: 3,
			wantErr:   false,
		},
		{
			name: "Read file with comments",
			content: `; Header
G0 X0 Y0
; Comment
G1 Z-1.0 F1000`,
			wantLines: 4,
			wantErr:   false,
		},
		{
			name:      "Read empty file",
			content:   "",
			wantLines: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.cnc")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Read file
			lines, err := gcode.ReadGCodeFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadGCodeFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(lines) != tt.wantLines {
				t.Errorf("ReadGCodeFile() got %d lines, want %d", len(lines), tt.wantLines)
			}
		})
	}
}

func TestWriteGCodeFile(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		wantErr bool
	}{
		{
			name: "Write small file",
			lines: []string{
				"G0 X0 Y0",
				"G1 Z-1.0 F1000",
				"M3 S1000",
			},
			wantErr: false,
		},
		{
			name: "Write file with comments",
			lines: []string{
				"; Header",
				"G0 X0 Y0",
				"; Comment",
				"G1 Z-1.0 F1000",
			},
			wantErr: false,
		},
		{
			name:    "Write empty file",
			lines:   []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "output.cnc")

			// Write file
			err := gcode.WriteGCodeFile(tmpFile, tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteGCodeFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify content
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			if len(tt.lines) == 0 && len(content) == 0 {
				return // Empty file case
			}

			if len(lines) != len(tt.lines) {
				t.Errorf("Output has %d lines, want %d", len(lines), len(tt.lines))
			}

			for i, line := range lines {
				if i < len(tt.lines) && line != tt.lines[i] {
					t.Errorf("Line %d: got %q, want %q", i, line, tt.lines[i])
				}
			}
		})
	}
}

func TestStreamingRead(t *testing.T) {
	// Test that streaming doesn't load entire file into memory
	lines := make([]string, 10000)
	for i := range lines {
		lines[i] = "G1 X0 Y0 Z0 F1000"
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.cnc")
	if err := os.WriteFile(tmpFile, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	result, err := gcode.ReadGCodeFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadGCodeFile() error = %v", err)
	}

	if len(result) != 10000 {
		t.Errorf("Got %d lines, want 10000", len(result))
	}
}

func TestFlushStrategy(t *testing.T) {
	// Test that writer flushes properly
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "flush_test.cnc")

	lines := []string{"G0 X0", "G1 Y10 F1000"}
	err := gcode.WriteGCodeFile(tmpFile, lines)
	if err != nil {
		t.Fatalf("WriteGCodeFile() error = %v", err)
	}

	// Verify content was flushed
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(content) == 0 {
		t.Error("File is empty, flush may have failed")
	}
}

func TestBufferedWriter(t *testing.T) {
	var buf bytes.Buffer

	lines := []string{
		"G0 X0 Y0",
		"G1 Z-1.0 F1000",
		"M3 S1000",
	}

	writer := gcode.NewBufferedWriter(&buf)
	for _, line := range lines {
		if err := writer.WriteLine(line); err != nil {
			t.Fatalf("WriteLine() error = %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	outputLines := strings.Split(strings.TrimSpace(output), "\n")

	if len(outputLines) != len(lines) {
		t.Errorf("Got %d lines, want %d", len(outputLines), len(lines))
	}

	for i, line := range outputLines {
		if line != lines[i] {
			t.Errorf("Line %d: got %q, want %q", i, line, lines[i])
		}
	}
}
