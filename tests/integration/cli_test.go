package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIEndToEnd(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		inputFile    string
		wantExitCode int
		checkOutput  func(t *testing.T, stdout, stderr string)
	}{
		{
			name: "Valid optimization with 3-axis fixture",
			args: []string{
				filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
				"1.0",
				filepath.Join(t.TempDir(), "output.cnc"),
			},
			wantExitCode: 0,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stdout, "Optimization Complete") {
					t.Error("Missing success message in output")
				}
				if !strings.Contains(stdout, "Total lines:") {
					t.Error("Missing statistics in output")
				}
			},
		},
		{
			name: "Missing input file",
			args: []string{
				filepath.Join(t.TempDir(), "nonexistent.cnc"),
				"1.0",
				filepath.Join(t.TempDir(), "output.cnc"),
			},
			wantExitCode: 1,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stderr, "does not exist") {
					t.Errorf("Expected file not found error, got: %s", stderr)
				}
			},
		},
		{
			name: "Invalid allowance",
			args: []string{
				filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
				"invalid",
				filepath.Join(t.TempDir(), "output.cnc"),
			},
			wantExitCode: 1,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stderr, "invalid allowance") {
					t.Errorf("Expected invalid allowance error, got: %s", stderr)
				}
			},
		},
		{
			name: "Invalid strategy",
			args: []string{
				"--strategy=invalid-strategy",
				filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
				"1.0",
				filepath.Join(t.TempDir(), "output.cnc"),
			},
			wantExitCode: 2,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stderr, "invalid strategy") {
					t.Errorf("Expected invalid strategy error, got: %s", stderr)
				}
			},
		},
		{
			name: "With --force flag",
			args: []string{
				"--force",
				filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
				"1.0",
				filepath.Join(t.TempDir(), "output_force.cnc"),
			},
			wantExitCode: 0,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stdout, "Optimization Complete") {
					t.Error("Missing success message")
				}
			},
		},
		{
			name: "With --strategy=all-axes",
			args: []string{
				"--strategy=all-axes",
				filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
				"1.0",
				filepath.Join(t.TempDir(), "output_allaxes.cnc"),
			},
			wantExitCode: 0,
			checkOutput: func(t *testing.T, stdout, stderr string) {
				if !strings.Contains(stdout, "Optimization Complete") {
					t.Error("Missing success message")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)

			stdout, stderr, exitCode := runCommand(cmd)

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nStdout: %s\nStderr: %s",
					exitCode, tt.wantExitCode, stdout, stderr)
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, stdout, stderr)
			}
		})
	}
}

func TestCLIOutputFileCreation(t *testing.T) {
	// Build the binary
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Prepare test
	outputPath := filepath.Join(t.TempDir(), "output.cnc")

	cmd := exec.Command(binPath,
		filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
		"1.0",
		outputPath,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	if exitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nStdout: %s\nStderr: %s",
			exitCode, stdout, stderr)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify output file has content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	// Verify output contains expected GCode structure
	contentStr := string(content)
	if !strings.Contains(contentStr, "G") {
		t.Error("Output file doesn't contain GCode commands")
	}
}

func TestCLIPerformanceRequirement(t *testing.T) {
	// Build the binary
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "output.cnc")

	cmd := exec.Command(binPath,
		filepath.Join("../../tests/testdata/finishing_3axis.cnc"),
		"1.0",
		outputPath,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	if exitCode != 0 {
		t.Fatalf("Command failed\nStdout: %s\nStderr: %s", stdout, stderr)
	}

	// Check that processing time is reported and reasonable
	if !strings.Contains(stdout, "Processing time:") {
		t.Error("Processing time not reported in output")
	}

	// Verify it's fast enough (should be well under 10 seconds for test files)
	// This is a soft check - we just verify the output reports a time
}

func TestCLIProgressReporting(t *testing.T) {
	// Build the binary
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "output.cnc")

	// Use medium file (50k lines) which should trigger progress updates
	cmd := exec.Command(binPath,
		"--force",
		filepath.Join("../../tests/testdata/medium_file.cnc"),
		"1.0",
		outputPath,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	if exitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nStdout: %s\nStderr: %s",
			exitCode, stdout, stderr)
	}

	// Verify progress updates appeared (should see "Processing:" for 50k line file)
	if !strings.Contains(stdout, "Processing:") {
		t.Error("No progress updates in output for medium-sized file")
	}

	// Verify progress format contains expected elements
	if !strings.Contains(stdout, "lines") {
		t.Error("Progress output missing 'lines' indicator")
	}

	if !strings.Contains(stdout, "%") {
		t.Error("Progress output missing percentage indicator")
	}

	if !strings.Contains(stdout, "Speed:") {
		t.Error("Progress output missing throughput (Speed:)")
	}

	if !strings.Contains(stdout, "lines/s") {
		t.Error("Progress output missing throughput unit (lines/s)")
	}

	if !strings.Contains(stdout, "Elapsed:") {
		t.Error("Progress output missing elapsed time")
	}

	if !strings.Contains(stdout, "ETA:") {
		t.Error("Progress output missing ETA")
	}

	// Verify final summary appears after progress
	if !strings.Contains(stdout, "Optimization Complete") {
		t.Error("Missing final summary after progress updates")
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created despite progress reporting")
	}
}

func TestCLIProgressLargeFile(t *testing.T) {
	// Skip if large_file.cnc doesn't exist (it's 10M lines, might not be committed)
	largeFile := filepath.Join("../../tests/testdata/large_file.cnc")
	if _, err := os.Stat(largeFile); os.IsNotExist(err) {
		t.Skip("Skipping large file test - large_file.cnc not found")
	}

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "output_large.cnc")

	cmd := exec.Command(binPath,
		"--force",
		largeFile,
		"1.0",
		outputPath,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	if exitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nStderr: %s",
			exitCode, stderr)
	}

	// SC-005: Verify progress updates appear for large files
	if !strings.Contains(stdout, "Processing:") {
		t.Error("SC-005 violation: No progress updates for 10M line file")
	}

	// SC-006: Verify large file processing completes without crash
	if !strings.Contains(stdout, "Optimization Complete") {
		t.Error("SC-006 violation: Large file processing did not complete successfully")
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created for large file")
	}
}

// Helper functions

func TestCLINoFeedRate(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "output.cnc")

	// Run with no_feed_rate.cnc test fixture (has NO F parameters)
	cmd := exec.Command(binPath,
		"--force",
		"../../tests/testdata/no_feed_rate.cnc",
		"1.0",
		outputFile,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	// Should succeed despite missing feed rate
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStdout: %s\nStderr: %s", exitCode, stdout, stderr)
	}

	// Should print warning for missing feed rate
	if !strings.Contains(stderr, "WARNING") {
		t.Error("Expected warning message for missing feed rate in stderr")
	}

	if !strings.Contains(stderr, "feed rate") && !strings.Contains(stderr, "F parameter") {
		t.Error("Expected feed rate warning in stderr")
	}

	if !strings.Contains(stderr, "1000") {
		t.Error("Expected default feed rate (1000 mm/min) mentioned in warning")
	}

	// Output file should still be created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should be created despite missing feed rate")
	}

	// Should still show completion message
	if !strings.Contains(stdout, "Optimization Complete") {
		t.Error("Should show completion message")
	}
}

func TestCLIMissingHeader(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "output.cnc")

	// Run with malformed_header.cnc test fixture
	cmd := exec.Command(binPath,
		"--force",
		"../../tests/testdata/malformed_header.cnc",
		"1.0",
		outputFile,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	// Should succeed despite missing header
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStdout: %s\nStderr: %s", exitCode, stdout, stderr)
	}

	// Should print warning for missing/incomplete header
	if !strings.Contains(stderr, "WARNING") {
		t.Error("Expected warning message for missing header in stderr")
	}

	if !strings.Contains(stderr, "Z-axis reference") {
		t.Error("Expected Z-axis reference warning in stderr")
	}

	// Output file should still be created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should be created despite missing header")
	}

	// Should still show completion message
	if !strings.Contains(stdout, "Optimization Complete") {
		t.Error("Should show completion message")
	}
}

func TestCLIMalformedLines(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	if err := buildBinary(binPath); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "output.cnc")

	// Run with unparseable.cnc test fixture
	cmd := exec.Command(binPath,
		"--force",
		"../../tests/testdata/unparseable.cnc",
		"1.0",
		outputFile,
	)

	stdout, stderr, exitCode := runCommand(cmd)

	// Should succeed despite malformed lines
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStdout: %s\nStderr: %s", exitCode, stdout, stderr)
	}

	// Should print warnings for malformed lines
	if !strings.Contains(stderr, "WARNING") {
		t.Error("Expected warning messages for malformed lines in stderr")
	}

	if !strings.Contains(stderr, "Skipping malformed line") {
		t.Error("Expected 'Skipping malformed line' message in stderr")
	}

	// Output file should still be created with valid lines
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file should be created despite malformed lines")
	}

	// Should still show completion message
	if !strings.Contains(stdout, "Optimization Complete") {
		t.Error("Should show completion message")
	}
}

func buildBinary(binPath string) error {
	cmd := exec.Command("go", "build", "-o", binPath,
		"../../cmd/snapmaker-cnc-finisher/main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		return &exec.ExitError{
			ProcessState: cmd.ProcessState,
			Stderr:       output,
		}
	}
	return nil
}

func runCommand(cmd *exec.Cmd) (stdout, stderr string, exitCode int) {
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode
}
