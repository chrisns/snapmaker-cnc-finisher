package cli_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

func TestPrintSummary(t *testing.T) {
	tests := []struct {
		name       string
		stats      *optimizer.Statistics
		wantOutput []string // Strings that should appear in output
	}{
		{
			name: "Typical optimization results",
			stats: &optimizer.Statistics{
				TotalLines:         1000,
				RemovedLines:       250,
				BytesIn:            50000,
				BytesOut:           37500,
				EstimatedTimeSaved: 5 * time.Minute,
				ProcessingTime:     100 * time.Millisecond,
			},
			wantOutput: []string{
				"1,000",  // Total lines (formatted with comma)
				"250",    // Removed lines
				"750",    // Kept lines
				"25.0%",  // Line reduction
				"50,000", // Input size (formatted with comma)
				"37,500", // Output size (formatted with comma)
				"5m 0s",  // Time saved (formatted with space)
				"0.1s",   // Processing time (formatted as seconds)
			},
		},
		{
			name: "No optimization (all lines kept)",
			stats: &optimizer.Statistics{
				TotalLines:         500,
				RemovedLines:       0,
				BytesIn:            25000,
				BytesOut:           25000,
				EstimatedTimeSaved: 0,
				ProcessingTime:     50 * time.Millisecond,
			},
			wantOutput: []string{
				"500",  // Total lines
				"0",    // Removed lines
				"500",  // Kept lines
				"0.0%", // No reduction
			},
		},
		{
			name: "High optimization (75% removed)",
			stats: &optimizer.Statistics{
				TotalLines:         4000,
				RemovedLines:       3000,
				BytesIn:            200000,
				BytesOut:           50000,
				EstimatedTimeSaved: 15 * time.Minute,
				ProcessingTime:     250 * time.Millisecond,
			},
			wantOutput: []string{
				"4,000",   // Total lines (formatted with comma)
				"3,000",   // Removed lines (formatted with comma)
				"1,000",   // Kept lines (formatted with comma)
				"75.0%",   // Line reduction
				"200,000", // Input size (formatted with comma)
				"50,000",  // Output size (formatted with comma)
				"15m 0s",  // Time saved (formatted with space)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cli.PrintSummary(tt.stats)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check all expected strings are present
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("PrintSummary() output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantExitCode int
		wantOutput   string
	}{
		{
			name:         "Generic error",
			err:          os.ErrNotExist,
			wantExitCode: 1,
			wantOutput:   "file does not exist",
		},
		{
			name:         "Custom error message",
			err:          &cli.InvalidStrategyError{Strategy: "invalid-strategy"},
			wantExitCode: 2,
			wantOutput:   "invalid-strategy",
		},
		{
			name:         "File operation error",
			err:          os.ErrPermission,
			wantExitCode: 1,
			wantOutput:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			exitCode := cli.PrintError(tt.err)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check exit code
			if exitCode != tt.wantExitCode {
				t.Errorf("PrintError() exit code = %d, want %d", exitCode, tt.wantExitCode)
			}

			// Check error message is present
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("PrintError() output missing %q\nGot:\n%s", tt.wantOutput, output)
			}
		})
	}
}

func TestPrintErrorNilError(t *testing.T) {
	// Edge case: nil error should not panic
	exitCode := cli.PrintError(nil)
	if exitCode != 0 {
		t.Errorf("PrintError(nil) exit code = %d, want 0", exitCode)
	}
}

func TestPrintWarning(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		args       []interface{}
		wantOutput string
	}{
		{
			name:       "Simple warning message",
			format:     "This is a test warning",
			args:       nil,
			wantOutput: "WARNING: This is a test warning",
		},
		{
			name:       "Warning with formatting",
			format:     "Skipping line %d: %s",
			args:       []interface{}{42, "malformed"},
			wantOutput: "WARNING: Skipping line 42: malformed",
		},
		{
			name:       "Warning with multiple parameters",
			format:     "Using default feed rate %v mm/min",
			args:       []interface{}{1000.0},
			wantOutput: "WARNING: Using default feed rate 1000 mm/min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			cli.PrintWarning(tt.format, tt.args...)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Check warning message is present
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("PrintWarning() output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}
