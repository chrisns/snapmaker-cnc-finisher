package cli

import (
	"fmt"
	"strings"
	"time"
)

// FormatNumber adds thousands separators (12450 -> "12,450")
func FormatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Use strings.Builder for efficient string construction
	str := fmt.Sprintf("%d", n)
	length := len(str)

	var result strings.Builder
	result.Grow(length + length/3) // Pre-allocate space for digits + commas

	for i, digit := range str {
		result.WriteRune(digit)
		remaining := length - i - 1
		if remaining > 0 && remaining%3 == 0 {
			result.WriteRune(',')
		}
	}

	return result.String()
}

// FormatBytes formats byte counts with thousands separators
func FormatBytes(n int64) string {
	return FormatNumber(int(n))
}

// FormatDuration formats duration in human-readable form (3.2s, 1m 15s, etc.)
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	minutes := seconds / 60
	secs := seconds % 60

	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}

	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// FormatThroughput formats lines per second with appropriate precision
func FormatThroughput(linesPerSecond float64) string {
	if linesPerSecond < 1000 {
		return fmt.Sprintf("%.0f", linesPerSecond)
	}
	// For high throughput, use thousands separator
	return FormatNumber(int(linesPerSecond))
}
