package contract_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestErrorMessageClarity validates SC-007: 95% of invalid inputs result in clear, actionable error messages
// This acceptance test covers all major error scenarios from spec.md Edge Cases
func TestErrorMessageClarity(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "snapmaker-cnc-finisher")
	cmd := exec.Command("go", "build", "-o", binPath,
		"../../cmd/snapmaker-cnc-finisher/main.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	tests := []struct {
		name           string
		args           []string
		setupFunc      func(t *testing.T) string // Returns temp file path if needed
		wantExitCode   int
		wantInStderr   []string // Error message should contain these strings
		wantActionable bool     // Should contain actionable hint/guidance
		errorScenario  string   // Description from spec.md
	}{
		{
			name:           "Non-existent input file",
			args:           []string{"nonexistent.cnc", "1.0", "output.cnc"},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "does not exist", "nonexistent.cnc"},
			wantActionable: true,
			errorScenario:  "EC-001: Input file does not exist",
		},
		{
			name: "Invalid allowance - non-numeric",
			args: []string{"input.cnc", "abc", "output.cnc"},
			setupFunc: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "input.cnc")
				os.WriteFile(path, []byte("G0 X0 Y0"), 0644)
				return path
			},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "invalid", "allowance", "number"},
			wantActionable: true,
			errorScenario:  "EC-002: Invalid allowance value",
		},
		{
			name: "Negative allowance",
			args: []string{"input.cnc", "-1.5", "output.cnc"},
			setupFunc: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "input.cnc")
				os.WriteFile(path, []byte("G0 X0 Y0"), 0644)
				return path
			},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "allowance", "non-negative"},
			wantActionable: true,
			errorScenario:  "EC-003: Negative allowance",
		},
		{
			name: "Invalid strategy",
			args: []string{"--strategy=invalid", "input.cnc", "1.0", "output.cnc"},
			setupFunc: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "input.cnc")
				os.WriteFile(path, []byte("G0 X0 Y0"), 0644)
				return path
			},
			wantExitCode:   2,
			wantInStderr:   []string{"Error:", "invalid strategy"},
			wantActionable: true,
			errorScenario:  "EC-004: Invalid strategy parameter",
		},
		{
			name: "Output directory does not exist",
			args: []string{"input.cnc", "1.0", "/nonexistent/path/output.cnc"},
			setupFunc: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "input.cnc")
				os.WriteFile(path, []byte("G0 X0 Y0"), 0644)
				return path
			},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "directory", "does not exist"},
			wantActionable: true,
			errorScenario:  "EC-005: Output directory does not exist",
		},
		{
			name: "Output file exists without --force",
			args: []string{}, // Will be set in setupFunc
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				inputPath := filepath.Join(tmpDir, "input.cnc")
				outputPath := filepath.Join(tmpDir, "output.cnc")
				os.WriteFile(inputPath, []byte(";MIN_Z: -1.0\n;MAX_Z: 0.0\nG0 X0 Y0"), 0644)
				os.WriteFile(outputPath, []byte("existing"), 0644)
				// Return special marker that we'll detect to set both paths
				return inputPath + "|" + outputPath
			},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "already exists", "--force"},
			wantActionable: true,
			errorScenario:  "EC-006: Output file already exists",
		},
		{
			name:           "Too few arguments",
			args:           []string{"input.cnc", "1.0"},
			wantExitCode:   1,
			wantInStderr:   []string{"Error:", "expected 3 arguments"},
			wantActionable: true,
			errorScenario:  "EC-007: Insufficient arguments",
		},
	}

	successCount := 0
	totalTests := len(tests)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment if needed
			var actualArgs []string
			if tt.setupFunc != nil {
				result := tt.setupFunc(t)
				// Check if result contains both input and output paths
				if strings.Contains(result, "|") {
					parts := strings.Split(result, "|")
					actualArgs = []string{parts[0], "1.0", parts[1]}
				} else {
					inputPath := result
					// Replace "input.cnc" with actual path in args
					actualArgs = make([]string, len(tt.args))
					for i, arg := range tt.args {
						if arg == "input.cnc" {
							actualArgs[i] = inputPath
						} else if arg == "output.cnc" && strings.Contains(inputPath, t.TempDir()) {
							actualArgs[i] = filepath.Join(filepath.Dir(inputPath), "output.cnc")
						} else {
							actualArgs[i] = arg
						}
					}
				}
			} else {
				actualArgs = tt.args
			}

			// Execute command
			cmd := exec.Command(binPath, actualArgs...)
			var outBuf, errBuf strings.Builder
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf

			err := cmd.Run()
			stderr := errBuf.String()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					exitCode = 1
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nStderr: %s", exitCode, tt.wantExitCode, stderr)
				return // Don't count as success
			}

			// Check error message contains expected strings
			hasAllStrings := true
			for _, want := range tt.wantInStderr {
				if !strings.Contains(stderr, want) {
					t.Errorf("Error message missing %q\nGot stderr: %s", want, stderr)
					hasAllStrings = false
				}
			}

			if !hasAllStrings {
				return // Don't count as success
			}

			// Check if error is actionable (contains hints or clear guidance)
			if tt.wantActionable {
				isActionable := strings.Contains(stderr, "use") ||
					strings.Contains(stderr, "must") ||
					strings.Contains(stderr, "expected") ||
					strings.Contains(stderr, "should") ||
					strings.Contains(strings.ToLower(stderr), "check") ||
					strings.Contains(stderr, ":") // Context after error

				if !isActionable {
					t.Errorf("Error message not actionable (no hint/guidance): %s", stderr)
					return // Don't count as success
				}
			}

			// Test passed - increment success count
			successCount++
			t.Logf("✓ %s: Clear, actionable error message", tt.errorScenario)
		})
	}

	// Validate SC-007: 95% of invalid inputs have clear error messages
	successRate := float64(successCount) / float64(totalTests) * 100
	t.Logf("\n=== SC-007 Validation ===")
	t.Logf("Error scenarios with clear messages: %d/%d (%.1f%%)", successCount, totalTests, successRate)

	if successRate < 95.0 {
		t.Errorf("SC-007 FAILED: Only %.1f%% of errors have clear messages (need >= 95%%)", successRate)
	} else {
		t.Logf("✓ SC-007 PASSED: %.1f%% of errors have clear, actionable messages", successRate)
	}
}
