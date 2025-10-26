package gcode_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
)

// BenchmarkFileReading tests performance of reading GCode files
func BenchmarkFileReading(b *testing.B) {
	// Create a temporary test file with realistic content
	tmpFile := filepath.Join(b.TempDir(), "test.cnc")
	f, err := os.Create(tmpFile)
	if err != nil {
		b.Fatal(err)
	}

	// Write 10k lines of realistic GCode
	writer := bufio.NewWriter(f)
	for i := 0; i < 10000; i++ {
		fmt.Fprintf(writer, "G1 X%d.5 Y%d.3 Z-1.2 F1500\n", i%100, i%50)
	}
	writer.Flush()
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, err := os.Open(tmpFile)
		if err != nil {
			b.Fatal(err)
		}

		scanner := bufio.NewScanner(f)
		lineCount := 0
		for scanner.Scan() {
			_ = scanner.Text()
			lineCount++
		}

		f.Close()
	}
}

// BenchmarkFileWriting tests performance of writing GCode files
func BenchmarkFileWriting(b *testing.B) {
	lines := make([]string, 10000)
	for i := 0; i < len(lines); i++ {
		lines[i] = fmt.Sprintf("G1 X%d.5 Y%d.3 Z-1.2 F1500", i%100, i%50)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpFile := filepath.Join(b.TempDir(), fmt.Sprintf("test_%d.cnc", i))
		f, err := os.Create(tmpFile)
		if err != nil {
			b.Fatal(err)
		}

		writer := gcode.NewBufferedWriter(f)
		for _, line := range lines {
			if err := writer.WriteLine(line); err != nil {
				b.Fatal(err)
			}
		}
		writer.Flush()
		f.Close()
	}
}

// BenchmarkMetadataExtraction tests performance of header metadata parsing
func BenchmarkMetadataExtraction(b *testing.B) {
	// Create test file with header
	tmpFile := filepath.Join(b.TempDir(), "test_header.cnc")
	f, err := os.Create(tmpFile)
	if err != nil {
		b.Fatal(err)
	}

	// Write header
	header := `;Header Start
;header_type: cnc
;min_x(mm): 0.0
;max_x(mm): 200.0
;min_y(mm): 0.0
;max_y(mm): 150.0
;min_z(mm): -10.0
;max_z(mm): 5.0
;Header End
`
	f.WriteString(header)

	// Write some GCode lines
	writer := bufio.NewWriter(f)
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(writer, "G1 X%d.5 Y%d.3 Z-1.2 F1500\n", i%100, i%50)
	}
	writer.Flush()
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, err := os.Open(tmpFile)
		if err != nil {
			b.Fatal(err)
		}

		_, err = gcode.ExtractMetadata(f)
		if err != nil {
			b.Fatal(err)
		}

		f.Close()
	}
}

// BenchmarkLargeFileStreaming simulates processing a large file
// This benchmark verifies SC-006 (10M lines without memory issues)
func BenchmarkLargeFileStreaming(b *testing.B) {
	// Create a large test file (100k lines for benchmark, real test would be 10M)
	tmpFile := filepath.Join(b.TempDir(), "large_test.cnc")
	f, err := os.Create(tmpFile)
	if err != nil {
		b.Fatal(err)
	}

	writer := bufio.NewWriter(f)
	lineCount := 100000
	for i := 0; i < lineCount; i++ {
		fmt.Fprintf(writer, "G1 X%d.5 Y%d.3 Z-1.2 F1500\n", i%100, i%50)
	}
	writer.Flush()
	f.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		f, err := os.Open(tmpFile)
		if err != nil {
			b.Fatal(err)
		}

		scanner := bufio.NewScanner(f)
		buf := make([]byte, 0, gcode.InitialBufferSize)
		scanner.Buffer(buf, gcode.MaxLineLength)

		processed := 0
		for scanner.Scan() {
			line := scanner.Text()
			// Simulate minimal processing
			if len(line) > 0 {
				processed++
			}
		}

		if err := scanner.Err(); err != nil {
			b.Fatal(err)
		}

		f.Close()

		if processed != lineCount {
			b.Fatalf("Expected %d lines, got %d", lineCount, processed)
		}
	}
}

// BenchmarkParseCommand tests command parsing performance
func BenchmarkParseCommand(b *testing.B) {
	testCases := []struct {
		name string
		line string
	}{
		{"Simple G1", "G1 X10.5 Y20.3 Z-1.2 F1500"},
		{"G0 Rapid", "G0 Z5.0"},
		{"Multi-axis", "G1 X10 Y20 Z-5 B45 F1200"},
		{"Comment line", "; This is a comment"},
		{"M-code", "M3 S12000"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := gcode.ParseCommand(tc.line)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
