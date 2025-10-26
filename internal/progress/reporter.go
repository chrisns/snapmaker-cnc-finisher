package progress

import (
	"fmt"
	"time"
)

// ProgressReporter tracks and displays optimization progress.
type ProgressReporter struct {
	totalLines     int64
	processedLines int64
	startTime      time.Time
	lastUpdate     time.Time
}

// NewReporter creates a progress reporter.
// If totalLines is 0 or unknown, progress displays without ETA.
func NewReporter(totalLines int64) *ProgressReporter {
	now := time.Now()
	return &ProgressReporter{
		totalLines:     totalLines,
		processedLines: 0,
		startTime:      now,
		lastUpdate:     now,
	}
}

// Update updates progress and displays if criteria met (2s OR 10k lines).
// Criteria: Update every 2 seconds OR every 10,000 lines, whichever is more frequent.
func (r *ProgressReporter) Update(linesProcessed int64) {
	r.processedLines = linesProcessed
	now := time.Now()

	// Check if we should display update
	timeSinceLastUpdate := now.Sub(r.lastUpdate)
	linesSinceLastUpdate := linesProcessed % 10000

	// Update if: 2 seconds elapsed OR reached 10k line boundary
	shouldUpdate := timeSinceLastUpdate >= 2*time.Second || linesSinceLastUpdate == 0

	if !shouldUpdate {
		return
	}

	r.lastUpdate = now
	elapsed := now.Sub(r.startTime)

	// Display progress based on whether total is known
	if r.totalLines > 0 {
		// Calculate percentage and ETA
		percent := float64(r.processedLines) / float64(r.totalLines) * 100

		// ETA calculation: (elapsed / processed) × (total - processed)
		if r.processedLines > 0 {
			remaining := r.totalLines - r.processedLines
			eta := time.Duration(float64(elapsed) / float64(r.processedLines) * float64(remaining))

			fmt.Printf("\rProcessed: %d/%d lines (%.1f%%) - ETA: %s    ",
				r.processedLines, r.totalLines, percent, eta.Round(time.Second))
		} else {
			fmt.Printf("\rProcessed: %d/%d lines (%.1f%%)    ",
				r.processedLines, r.totalLines, percent)
		}
	} else {
		// No total known - just show lines processed
		fmt.Printf("\rProcessed: %d lines (%.1fs elapsed)    ",
			r.processedLines, elapsed.Seconds())
	}
}

// Finish displays final progress line with newline.
func (r *ProgressReporter) Finish() {
	if r.processedLines > 0 {
		elapsed := time.Since(r.startTime)
		fmt.Printf("\rProcessed: %d lines in %.1fs\n", r.processedLines, elapsed.Seconds())
	}
}
