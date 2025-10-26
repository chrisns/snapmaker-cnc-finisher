// Package progress provides progress reporting and statistics tracking
// for GCode optimization operations.
package progress

// OptimizationResult contains statistics and metrics from the optimization process.
type OptimizationResult struct {
	// Input metrics
	TotalInputLines    int64
	InputFileSizeBytes int64

	// Processing metrics
	LinesProcessed int64
	LinesRemoved   int64
	LinesPreserved int64
	LinesSplit     int64 // Number of moves that were split (aggressive mode)

	// Depth analysis
	MinZ      float64 // Minimum Z value found in file
	Threshold float64 // Calculated threshold (min_z + allowance)

	// Output metrics
	TotalOutputLines    int64
	OutputFileSizeBytes int64
	ReductionPercent    float64 // (LinesRemoved / TotalInputLines) * 100

	// Time savings estimate
	EstimatedTimeSavingsSec float64 // Sum of (distance / feed_rate) for removed moves

	// Performance
	ProcessingDurationSec float64
	LinesPerSecond        float64
}
