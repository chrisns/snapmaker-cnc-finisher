package cli

import (
	"fmt"
	"os"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

// InvalidStrategyError represents an invalid filtering strategy error
type InvalidStrategyError struct {
	Strategy string
}

func (e *InvalidStrategyError) Error() string {
	return fmt.Sprintf("invalid strategy: %s", e.Strategy)
}

// PrintWarning prints a warning message to stderr
// Format: "WARNING: <message>"
func PrintWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", message)
}

// PrintSummary prints optimization statistics to stdout
func PrintSummary(stats *optimizer.Statistics) {
	fmt.Println("\n=== Optimization Complete ===")
	fmt.Println()

	// Line statistics (with thousands separators for readability)
	fmt.Printf("Total lines:     %s\n", FormatNumber(stats.TotalLines))
	fmt.Printf("Removed lines:   %s\n", FormatNumber(stats.RemovedLines))
	fmt.Printf("Kept lines:      %s\n", FormatNumber(stats.KeptLines()))
	fmt.Printf("Line reduction:  %.1f%%\n", stats.LineReductionPercent())
	fmt.Println()

	// File size statistics (with thousands separators)
	fmt.Printf("Input size:      %s bytes\n", FormatBytes(stats.BytesIn))
	fmt.Printf("Output size:     %s bytes\n", FormatBytes(stats.BytesOut))
	fmt.Printf("Size reduction:  %.1f%%\n", stats.FileSizeReductionPercent())
	fmt.Println()

	// Time statistics
	fmt.Printf("Estimated time saved:  %s\n", FormatDuration(stats.EstimatedTimeSaved))
	fmt.Printf("Processing time:       %s\n", FormatDuration(stats.ProcessingTime))
	fmt.Println()
}

// PrintError prints an error message to stderr and returns the appropriate exit code
// Exit codes:
//
//	0 - No error (nil error)
//	1 - General error (file I/O, parsing, etc.)
//	2 - Invalid arguments or strategy
func PrintError(err error) int {
	if err == nil {
		return 0
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)

	// Determine exit code based on error type
	switch err.(type) {
	case *InvalidStrategyError:
		return 2
	default:
		return 1
	}
}
