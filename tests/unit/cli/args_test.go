package cli_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
)

func TestValidateArgs(t *testing.T) {
	// Create temp directory for tests
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.cnc")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		args    *cli.Args
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid args with existing input file",
			args: &cli.Args{
				InputFile:  existingFile,
				OutputFile: filepath.Join(tmpDir, "output.cnc"),
				Allowance:  1.0,
				Strategy:   "safe",
			},
			wantErr: false,
		},
		{
			name: "Input file does not exist",
			args: &cli.Args{
				InputFile:  filepath.Join(tmpDir, "nonexistent.cnc"),
				OutputFile: filepath.Join(tmpDir, "output.cnc"),
				Allowance:  1.0,
				Strategy:   "safe",
			},
			wantErr: true,
			errMsg:  "input file does not exist",
		},
		{
			name: "Output directory does not exist",
			args: &cli.Args{
				InputFile:  existingFile,
				OutputFile: "/nonexistent/directory/output.cnc",
				Allowance:  1.0,
				Strategy:   "safe",
			},
			wantErr: true,
			errMsg:  "output directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.ValidateArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message should contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestShouldShowHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"--help flag", []string{"--help"}, true},
		{"-h flag", []string{"-h"}, true},
		{"--help with other args", []string{"--help", "foo", "bar"}, true},
		{"--help in middle", []string{"foo", "--help", "bar"}, true},
		{"No help flag", []string{"foo", "bar"}, false},
		{"Empty args", []string{}, false},
		{"Similar but not help", []string{"--helper"}, false},
		{"Just -h alone", []string{"-h"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.ShouldShowHelp(tt.args)
			if got != tt.want {
				t.Errorf("ShouldShowHelp(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestShouldShowVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"--version flag", []string{"--version"}, true},
		{"-v flag", []string{"-v"}, true},
		{"--version with other args", []string{"--version", "foo", "bar"}, true},
		{"--version in middle", []string{"foo", "--version", "bar"}, true},
		{"No version flag", []string{"foo", "bar"}, false},
		{"Empty args", []string{}, false},
		{"Similar but not version", []string{"--verbose"}, false},
		{"Just -v alone", []string{"-v"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.ShouldShowVersion(tt.args)
			if got != tt.want {
				t.Errorf("ShouldShowVersion(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestGetHelpText(t *testing.T) {
	help := cli.GetHelpText()

	// Test structure
	if help == "" {
		t.Fatal("GetHelpText() returned empty string")
	}

	// Test required sections (matching contract)
	requiredStrings := []string{
		"GCode Finishing Pass Optimizer",
		"Usage:",
		"snapmaker-cnc-finisher",
		"<input-file>",
		"<allowance>",
		"<output-file>",
		"Positional Arguments:",
		"Optional Flags:",
		"--force",
		"--strategy",
		"--help",
		"--version",
		"Examples:",
		"github.com/chrisns/snapmaker-cnc-finisher",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(help, required) {
			t.Errorf("Help text missing required string: %q", required)
		}
	}

	// Test that all strategy options are documented
	strategies := []string{"safe", "all-axes", "split", "aggressive"}
	for _, strategy := range strategies {
		if !strings.Contains(help, strategy) {
			t.Errorf("Help text missing strategy: %q", strategy)
		}
	}
}

func TestGetVersionText(t *testing.T) {
	version := cli.GetVersionText()

	// Test structure
	if version == "" {
		t.Fatal("GetVersionText() returned empty string")
	}

	// Test required elements
	requiredStrings := []string{
		"snapmaker-cnc-finisher",
		"version",
		"Built with Go",
		"Platform:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(version, required) {
			t.Errorf("Version text missing required string: %q", required)
		}
	}

	// Verify runtime info is present
	if !strings.Contains(version, runtime.Version()) {
		t.Error("Version text should contain Go runtime version")
	}

	if !strings.Contains(version, runtime.GOOS) {
		t.Error("Version text should contain OS name")
	}

	if !strings.Contains(version, runtime.GOARCH) {
		t.Error("Version text should contain architecture")
	}
}
