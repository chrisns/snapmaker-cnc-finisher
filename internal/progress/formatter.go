package progress

import (
	"fmt"
	"strings"
)

// ResultFormatter formats and displays OptimizationResult.
type ResultFormatter struct{}

// Format returns formatted result string with nice formatting.
func (rf *ResultFormatter) Format(result OptimizationResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("\nOptimization Complete\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	// Depth Analysis
	sb.WriteString("Depth Analysis:\n")
	sb.WriteString(fmt.Sprintf("  Min Z: %.3fmm\n", result.MinZ))
	sb.WriteString(fmt.Sprintf("  Threshold: %.3fmm (%.1fmm allowance)\n\n",
		result.Threshold, result.Threshold-result.MinZ))

	// Processing Summary
	sb.WriteString("Processing Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total lines: %s\n", formatNumber(result.TotalInputLines)))
	sb.WriteString(fmt.Sprintf("  Lines removed: %s (%.1f%%)\n",
		formatNumber(result.LinesRemoved), result.ReductionPercent))
	sb.WriteString(fmt.Sprintf("  Lines preserved: %s\n", formatNumber(result.LinesPreserved)))
	if result.LinesSplit > 0 {
		sb.WriteString(fmt.Sprintf("  Moves split: %s (aggressive strategy)\n", formatNumber(result.LinesSplit)))
	}
	sb.WriteString("\n")

	// Output
	sb.WriteString("Output:\n")
	sb.WriteString(fmt.Sprintf("  File size: %s → %s (%.1f%% reduction)\n",
		formatBytes(result.InputFileSizeBytes),
		formatBytes(result.OutputFileSizeBytes),
		(1.0-float64(result.OutputFileSizeBytes)/float64(result.InputFileSizeBytes))*100))

	if result.EstimatedTimeSavingsSec > 0 {
		sb.WriteString(fmt.Sprintf("  Estimated time savings: %.1f minutes\n\n",
			result.EstimatedTimeSavingsSec/60))
	} else {
		sb.WriteString("\n")
	}

	// Performance
	sb.WriteString("Performance:\n")
	sb.WriteString(fmt.Sprintf("  Processing time: %.1f seconds\n", result.ProcessingDurationSec))
	sb.WriteString(fmt.Sprintf("  Throughput: %s lines/sec\n",
		formatNumber(int64(result.LinesPerSecond))))

	return sb.String()
}

// Display prints formatted result to stdout.
func (rf *ResultFormatter) Display(result OptimizationResult) {
	fmt.Print(rf.Format(result))
}

// formatNumber adds thousand separators to numbers.
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

// formatBytes formats byte sizes in human-readable form.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
