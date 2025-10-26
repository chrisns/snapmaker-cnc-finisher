package gcode_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

func TestReadGCodeFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "File does not exist",
			setup: func() string {
				return "/nonexistent/path/to/file.cnc"
			},
			wantErr: true,
		},
		{
			name: "Directory instead of file",
			setup: func() string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			_, err := gcode.ReadGCodeFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadGCodeFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteGCodeFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		lines   []string
		wantErr bool
	}{
		{
			name:    "Invalid path - directory does not exist",
			path:    "/nonexistent/directory/output.cnc",
			lines:   []string{"G0 X0 Y0"},
			wantErr: true,
		},
		{
			name: "Path is a directory",
			path: func() string {
				tmpDir := t.TempDir()
				return tmpDir
			}(),
			lines:   []string{"G0 X0 Y0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gcode.WriteGCodeFile(tt.path, tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteGCodeFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBufferedWriterErrors(t *testing.T) {
	t.Run("Flush to failing writer", func(t *testing.T) {
		// Create a writer that always fails
		failWriter := &failingWriter{}
		writer := gcode.NewBufferedWriter(failWriter)

		// Write a line (this will succeed because it's buffered)
		writer.WriteLine("G0 X0 Y0")

		// Flush should fail
		err := writer.Flush()
		if err == nil {
			t.Error("Expected error flushing to failing writer, got nil")
		}
	})

	t.Run("Auto-flush to failing writer after 1000 lines", func(t *testing.T) {
		// Create a writer that always fails
		failWriter := &failingWriter{}
		writer := gcode.NewBufferedWriter(failWriter)

		// Write 1000 lines to trigger auto-flush
		var err error
		for i := 0; i < 1000; i++ {
			err = writer.WriteLine("G0 X0 Y0")
			if err != nil {
				break
			}
		}

		// Should get error from auto-flush
		if err == nil {
			t.Error("Expected error from auto-flush to failing writer, got nil")
		}
	})
}

func TestParseCommandErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Invalid GCode syntax - garbage input",
			input:   "INVALID GCODE !@#$%",
			wantErr: true,
		},
		{
			name:    "Malformed command letter without value",
			input:   "G",
			wantErr: true,
		},
		{
			name:    "Invalid parameter value",
			input:   "G1 XABC",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gcode.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractMetadataErrors(t *testing.T) {
	t.Run("Reader that returns error", func(t *testing.T) {
		// Create a reader that returns an error
		errorReader := &errorReader{}
		_, err := gcode.ExtractMetadata(errorReader)
		if err == nil {
			t.Error("Expected error from failing reader, got nil")
		}
	})

	t.Run("Handles malformed MIN_Z value gracefully", func(t *testing.T) {
		input := `;MIN_Z: not_a_number
G0 X0 Y0`
		r := strings.NewReader(input)
		meta, err := gcode.ExtractMetadata(r)

		// Should not error, but should fall back to surface convention
		if err != nil {
			t.Errorf("ExtractMetadata() unexpected error = %v", err)
		}

		// Should fall back to ZRefSurface since MIN_Z couldn't be parsed
		if meta.ZReference != gcode.ZRefSurface {
			t.Errorf("Expected ZRefSurface fallback, got %v", meta.ZReference)
		}
	})

	t.Run("Handles malformed MAX_Z value gracefully", func(t *testing.T) {
		input := `;MAX_Z: invalid
G0 X0 Y0`
		r := strings.NewReader(input)
		meta, err := gcode.ExtractMetadata(r)

		// Should not error, but should fall back to surface convention
		if err != nil {
			t.Errorf("ExtractMetadata() unexpected error = %v", err)
		}

		// Should fall back to ZRefSurface since MAX_Z couldn't be parsed
		if meta.ZReference != gcode.ZRefSurface {
			t.Errorf("Expected ZRefSurface fallback, got %v", meta.ZReference)
		}
	})
}

func TestFileIOSuccessCases(t *testing.T) {
	t.Run("Read and write round-trip preserves content", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.cnc")
		outputFile := filepath.Join(tmpDir, "output.cnc")

		originalLines := []string{
			"; Header",
			"G0 X0 Y0",
			"G1 Z-1.0 F1000",
			"M3 S1000",
			"; Footer",
		}

		// Write original file
		err := gcode.WriteGCodeFile(inputFile, originalLines)
		if err != nil {
			t.Fatalf("Failed to write input file: %v", err)
		}

		// Read it back
		lines, err := gcode.ReadGCodeFile(inputFile)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		// Write to output file
		err = gcode.WriteGCodeFile(outputFile, lines)
		if err != nil {
			t.Fatalf("Failed to write output file: %v", err)
		}

		// Read output file
		finalLines, err := gcode.ReadGCodeFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		// Verify content matches
		if len(finalLines) != len(originalLines) {
			t.Errorf("Line count mismatch: got %d, want %d", len(finalLines), len(originalLines))
		}

		for i, line := range finalLines {
			if i < len(originalLines) && line != originalLines[i] {
				t.Errorf("Line %d mismatch: got %q, want %q", i, line, originalLines[i])
			}
		}
	})

	t.Run("BufferedWriter handles large number of lines", func(t *testing.T) {
		var buf bytes.Buffer
		writer := gcode.NewBufferedWriter(&buf)

		// Write 2000 lines to trigger multiple auto-flushes
		lineCount := 2000
		for i := 0; i < lineCount; i++ {
			if err := writer.WriteLine("G1 X0 Y0 Z0 F1000"); err != nil {
				t.Fatalf("WriteLine() error at line %d: %v", i, err)
			}
		}

		if err := writer.Flush(); err != nil {
			t.Fatalf("Flush() error: %v", err)
		}

		if writer.LineCount() != lineCount {
			t.Errorf("LineCount() = %d, want %d", writer.LineCount(), lineCount)
		}

		// Verify all lines were written
		output := buf.String()
		outputLines := strings.Split(strings.TrimSpace(output), "\n")
		if len(outputLines) != lineCount {
			t.Errorf("Output has %d lines, want %d", len(outputLines), lineCount)
		}
	})
}

func TestParseCommandComplexCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, gcode.Command)
	}{
		{
			name:    "Empty line returns empty command",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, cmd gcode.Command) {
				if cmd.Letter != "" || cmd.Value != 0 || len(cmd.Params) != 0 {
					t.Errorf("Expected empty command, got %+v", cmd)
				}
			},
		},
		{
			name:    "Whitespace-only line returns empty command",
			input:   "   \t  ",
			wantErr: false,
			check: func(t *testing.T, cmd gcode.Command) {
				if cmd.Letter != "" || cmd.Value != 0 || len(cmd.Params) != 0 {
					t.Errorf("Expected empty command, got %+v", cmd)
				}
			},
		},
		{
			name:    "Comment with leading/trailing whitespace",
			input:   "  ; This is a comment  ",
			wantErr: false,
			check: func(t *testing.T, cmd gcode.Command) {
				if !cmd.IsComment() {
					t.Error("Expected comment command")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := gcode.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}
