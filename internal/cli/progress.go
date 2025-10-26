package cli

import (
	"fmt"
	"io"
	"time"
)

// ProgressTracker tracks processing progress and calculates statistics
type ProgressTracker struct {
	totalLines   int
	currentLine  int
	linesRemoved int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalLines int) *ProgressTracker {
	if totalLines <= 0 {
		totalLines = 1 // Minimum to avoid division by zero
	}
	return &ProgressTracker{
		totalLines: totalLines,
	}
}

// TotalLines returns the estimated total line count
func (p *ProgressTracker) TotalLines() int {
	return p.totalLines
}

// UpdateTotalEstimate allows adjusting the estimate during processing
func (p *ProgressTracker) UpdateTotalEstimate(newTotal int) {
	if newTotal > p.totalLines {
		p.totalLines = newTotal
	}
}

// Update updates the current progress
func (p *ProgressTracker) Update(currentLine, linesRemoved int) {
	p.currentLine = currentLine
	p.linesRemoved = linesRemoved
}

// ShouldUpdate determines if progress should be displayed
// Per SC-005: Every 10,000 lines processed OR every 2 seconds
func (p *ProgressTracker) ShouldUpdate(lastUpdateLine int, timeSinceLastUpdate time.Duration) bool {
	linesSinceUpdate := p.currentLine - lastUpdateLine
	return linesSinceUpdate >= 10000 || timeSinceLastUpdate >= 2*time.Second
}

// PercentComplete calculates completion percentage
func (p *ProgressTracker) PercentComplete() float64 {
	if p.totalLines == 0 {
		return 0.0
	}
	return (float64(p.currentLine) / float64(p.totalLines)) * 100.0
}

// EstimatedTimeRemaining calculates projected time to completion
func (p *ProgressTracker) EstimatedTimeRemaining(elapsed time.Duration) time.Duration {
	if p.currentLine == 0 || p.currentLine >= p.totalLines {
		return 0 // Clamp to 0 when at or past end
	}

	linesPerSecond := float64(p.currentLine) / elapsed.Seconds()
	remainingLines := p.totalLines - p.currentLine

	if linesPerSecond == 0 || remainingLines <= 0 {
		return 0
	}

	remainingSeconds := float64(remainingLines) / linesPerSecond
	return time.Duration(remainingSeconds * float64(time.Second))
}

// Display outputs progress to the given writer with single-line overwrite (\\r)
// Format: "Processing: 45,230 / 100,000 lines (45.2%) | Removed: 12,450 | Speed: 14,096 lines/s | Elapsed: 3.2s | ETA: 3.8s"
func (p *ProgressTracker) Display(w io.Writer, elapsed time.Duration) {
	eta := p.EstimatedTimeRemaining(elapsed)
	percent := p.PercentComplete()

	// Calculate throughput (lines per second)
	throughput := 0.0
	if elapsed.Seconds() > 0 {
		throughput = float64(p.currentLine) / elapsed.Seconds()
	}

	fmt.Fprintf(w, "\rProcessing: %s / %s lines (%.1f%%) | Removed: %s | Speed: %s lines/s | Elapsed: %s | ETA: %s",
		FormatNumber(p.currentLine),
		FormatNumber(p.totalLines),
		percent,
		FormatNumber(p.linesRemoved),
		FormatThroughput(throughput),
		FormatDuration(elapsed),
		FormatDuration(eta),
	)
}
