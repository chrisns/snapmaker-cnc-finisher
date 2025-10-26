package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Version information (set during build with -ldflags)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Args contains parsed command-line arguments
type Args struct {
	InputFile  string
	OutputFile string
	Allowance  float64
	Strategy   string
	Force      bool
}

// ParseArgs parses command-line arguments
// Expected format: [--force] [--strategy=STRATEGY] <input> <allowance> <output>
func ParseArgs(args []string) (*Args, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments provided")
	}

	// Create flag set
	fs := flag.NewFlagSet("snapmaker-cnc-finisher", flag.ContinueOnError)

	result := &Args{
		Strategy: "safe", // Default strategy
	}

	// Define flags
	fs.BoolVar(&result.Force, "force", false, "Overwrite output file without prompting")
	fs.StringVar(&result.Strategy, "strategy", "safe", "Filtering strategy (safe, all-axes, split, aggressive)")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Get positional arguments (after flags)
	positional := fs.Args()

	// Validate we have exactly 3 positional arguments
	if len(positional) != 3 {
		return nil, fmt.Errorf("expected 3 arguments (input, allowance, output), got %d", len(positional))
	}

	result.InputFile = positional[0]
	result.OutputFile = positional[2]

	// Parse allowance
	allowance, err := strconv.ParseFloat(positional[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid allowance value %q: must be a number", positional[1])
	}

	// Validate allowance is non-negative
	if allowance < 0 {
		return nil, fmt.Errorf("allowance must be non-negative, got %v", allowance)
	}

	result.Allowance = allowance

	// Validate strategy (will be validated again when parsed to enum, but check early)
	validStrategies := map[string]bool{
		"safe":       true,
		"all-axes":   true,
		"split":      true,
		"aggressive": true,
	}

	if !validStrategies[result.Strategy] {
		return nil, &InvalidStrategyError{Strategy: result.Strategy}
	}

	return result, nil
}

// ValidateArgs validates that the parsed arguments are valid
// Checks that input file exists and output directory is writable
func ValidateArgs(args *Args) error {
	// Check input file exists
	if _, err := os.Stat(args.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", args.InputFile)
	} else if err != nil {
		return fmt.Errorf("failed to check input file: %w", err)
	}

	// Check output directory exists
	outputDir := filepath.Dir(args.OutputFile)
	if outputDir == "." || outputDir == "" {
		outputDir = "."
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", outputDir)
	} else if err != nil {
		return fmt.Errorf("failed to check output directory: %w", err)
	}

	return nil
}

// ShouldShowHelp checks if --help or -h flag is present
func ShouldShowHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

// ShouldShowVersion checks if --version or -v flag is present
func ShouldShowVersion(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-v" {
			return true
		}
	}
	return false
}

// GetHelpText returns the help message text
func GetHelpText() string {
	var sb strings.Builder

	sb.WriteString("GCode Finishing Pass Optimizer\n\n")
	sb.WriteString("Usage: snapmaker-cnc-finisher <input-file> <allowance> <output-file> [FLAGS]\n\n")

	sb.WriteString("Positional Arguments:\n")
	sb.WriteString("  input-file     Path to input GCode file (finishing pass from Snapmaker Luban)\n")
	sb.WriteString("  allowance      Remaining material depth in mm after rough cut (e.g., 1.0)\n")
	sb.WriteString("  output-file    Path for optimized output GCode file\n\n")

	sb.WriteString("Optional Flags:\n")
	sb.WriteString("  --force, -f              Overwrite output file without confirmation\n")
	sb.WriteString("  --strategy=<value>, -s   Multi-axis move handling strategy (default: safe)\n")
	sb.WriteString("                           Allowed values: safe, all-axes, split, aggressive\n")
	sb.WriteString("  --help, -h               Display this help message\n")
	sb.WriteString("  --version, -v            Display version information\n\n")

	sb.WriteString("Examples:\n")
	sb.WriteString("  snapmaker-cnc-finisher finishing.cnc 1.0 output.cnc\n")
	sb.WriteString("  snapmaker-cnc-finisher finishing.cnc 1.5 output.cnc --force\n")
	sb.WriteString("  snapmaker-cnc-finisher finishing.cnc 0.5 output.cnc --strategy=aggressive\n\n")

	sb.WriteString("For more information, visit: https://github.com/chrisns/snapmaker-cnc-finisher\n")

	return sb.String()
}

// GetVersionText returns the version information text
func GetVersionText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("snapmaker-cnc-finisher version %s\n", Version))
	sb.WriteString(fmt.Sprintf("Built with Go %s\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH))

	if GitCommit != "unknown" {
		sb.WriteString(fmt.Sprintf("Git commit: %s\n", GitCommit))
	}

	if BuildDate != "unknown" {
		sb.WriteString(fmt.Sprintf("Build date: %s\n", BuildDate))
	}

	return sb.String()
}
