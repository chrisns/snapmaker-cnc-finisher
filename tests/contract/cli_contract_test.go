package contract_test

import (
	"strings"
	"testing"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
)

func TestCLIArgumentsParsing(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(*testing.T, *cli.Args)
	}{
		{
			name:    "Valid arguments: input, allowance, output",
			args:    []string{"input.cnc", "1.0", "output.cnc"},
			wantErr: false,
			check: func(t *testing.T, args *cli.Args) {
				if args.InputFile != "input.cnc" {
					t.Errorf("InputFile = %q, want %q", args.InputFile, "input.cnc")
				}
				if args.Allowance != 1.0 {
					t.Errorf("Allowance = %v, want %v", args.Allowance, 1.0)
				}
				if args.OutputFile != "output.cnc" {
					t.Errorf("OutputFile = %q, want %q", args.OutputFile, "output.cnc")
				}
				if args.Strategy != "safe" {
					t.Errorf("Strategy = %q, want %q (default)", args.Strategy, "safe")
				}
				if args.Force {
					t.Error("Force should be false by default")
				}
			},
		},
		{
			name:    "Too few arguments",
			args:    []string{"input.cnc", "1.0"},
			wantErr: true,
		},
		{
			name:    "No arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Invalid allowance (non-numeric)",
			args:    []string{"input.cnc", "invalid", "output.cnc"},
			wantErr: true,
		},
		{
			name:    "Negative allowance",
			args:    []string{"input.cnc", "-1.0", "output.cnc"},
			wantErr: true,
		},
		{
			name:    "With --force flag",
			args:    []string{"--force", "input.cnc", "1.0", "output.cnc"},
			wantErr: false,
			check: func(t *testing.T, args *cli.Args) {
				if !args.Force {
					t.Error("Force should be true when --force flag provided")
				}
			},
		},
		{
			name:    "With --strategy flag (safe)",
			args:    []string{"--strategy=safe", "input.cnc", "1.0", "output.cnc"},
			wantErr: false,
			check: func(t *testing.T, args *cli.Args) {
				if args.Strategy != "safe" {
					t.Errorf("Strategy = %q, want %q", args.Strategy, "safe")
				}
			},
		},
		{
			name:    "With --strategy flag (all-axes)",
			args:    []string{"--strategy=all-axes", "input.cnc", "1.0", "output.cnc"},
			wantErr: false,
			check: func(t *testing.T, args *cli.Args) {
				if args.Strategy != "all-axes" {
					t.Errorf("Strategy = %q, want %q", args.Strategy, "all-axes")
				}
			},
		},
		{
			name:    "With both flags",
			args:    []string{"--force", "--strategy=aggressive", "input.cnc", "2.5", "output.cnc"},
			wantErr: false,
			check: func(t *testing.T, args *cli.Args) {
				if !args.Force {
					t.Error("Force should be true")
				}
				if args.Strategy != "aggressive" {
					t.Errorf("Strategy = %q, want %q", args.Strategy, "aggressive")
				}
				if args.Allowance != 2.5 {
					t.Errorf("Allowance = %v, want %v", args.Allowance, 2.5)
				}
			},
		},
		{
			name:    "Invalid strategy value",
			args:    []string{"--strategy=invalid", "input.cnc", "1.0", "output.cnc"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := cli.ParseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, args)
			}
		})
	}
}

func TestCLIHelpFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantHelp bool
	}{
		{
			name:     "--help flag",
			args:     []string{"--help"},
			wantHelp: true,
		},
		{
			name:     "-h flag",
			args:     []string{"-h"},
			wantHelp: true,
		},
		{
			name:     "--help with other args",
			args:     []string{"--help", "input.cnc", "1.0", "output.cnc"},
			wantHelp: true,
		},
		{
			name:     "No help flag",
			args:     []string{"input.cnc", "1.0", "output.cnc"},
			wantHelp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := cli.ShouldShowHelp(tt.args)
			if help != tt.wantHelp {
				t.Errorf("ShouldShowHelp() = %v, want %v", help, tt.wantHelp)
			}
		})
	}
}

func TestCLIVersionFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantVersion bool
	}{
		{
			name:        "--version flag",
			args:        []string{"--version"},
			wantVersion: true,
		},
		{
			name:        "-v flag",
			args:        []string{"-v"},
			wantVersion: true,
		},
		{
			name:        "--version with other args",
			args:        []string{"--version", "input.cnc", "1.0", "output.cnc"},
			wantVersion: true,
		},
		{
			name:        "No version flag",
			args:        []string{"input.cnc", "1.0", "output.cnc"},
			wantVersion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := cli.ShouldShowVersion(tt.args)
			if version != tt.wantVersion {
				t.Errorf("ShouldShowVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestCLIHelpMessage(t *testing.T) {
	help := cli.GetHelpText()

	// Verify help message contains required sections from CLI contract
	requiredStrings := []string{
		"Usage:",
		"snapmaker-cnc-finisher",
		"<input-file>",
		"<allowance>",
		"<output-file>",
		"--force",
		"--strategy",
		"--help",
		"--version",
		"Examples:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(help, required) {
			t.Errorf("Help text missing required string: %q", required)
		}
	}
}

func TestCLIVersionMessage(t *testing.T) {
	version := cli.GetVersionText()

	// Verify version message contains required information
	requiredStrings := []string{
		"snapmaker-cnc-finisher",
		"version",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(version, required) {
			t.Errorf("Version text missing required string: %q", required)
		}
	}
}
