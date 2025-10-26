package optimizer

import (
	"math"
	"time"
)

// Statistics tracks optimization metrics
type Statistics struct {
	TotalLines          int           // Total lines processed
	RemovedLines        int           // Lines filtered out
	BytesIn             int64         // Input file size in bytes
	BytesOut            int64         // Output file size in bytes
	EstimatedTimeSaved  time.Duration // Estimated machining time saved
	ProcessingTime      time.Duration // Time spent processing
}

// NewStatistics creates a new Statistics instance
func NewStatistics() *Statistics {
	return &Statistics{}
}

// KeptLines returns the number of lines retained after filtering
func (s *Statistics) KeptLines() int {
	return s.TotalLines - s.RemovedLines
}

// LineReductionPercent returns the percentage of lines removed
func (s *Statistics) LineReductionPercent() float64 {
	if s.TotalLines == 0 {
		return 0.0
	}
	return (float64(s.RemovedLines) / float64(s.TotalLines)) * 100.0
}

// FileSizeReductionPercent returns the percentage of file size reduced
func (s *Statistics) FileSizeReductionPercent() float64 {
	if s.BytesIn == 0 {
		return 0.0
	}
	return (float64(s.BytesIn-s.BytesOut) / float64(s.BytesIn)) * 100.0
}

// CalculateTimeSaved calculates the time saved for a move between two 3D points
// given the feed rate in mm/min. Returns the time as a Duration.
//
// Uses Euclidean distance formula: sqrt((x2-x1)^2 + (y2-y1)^2 + (z2-z1)^2)
// Time = Distance / FeedRate (converted to seconds)
//
// If feedRate is 0 or negative, uses default of 1000 mm/min
func CalculateTimeSaved(x1, y1, z1, x2, y2, z2, feedRate float64) time.Duration {
	// Calculate Euclidean distance
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// Use default feed rate if invalid
	if feedRate <= 0 {
		feedRate = 1000.0 // Default: 1000 mm/min
	}

	// Calculate time: distance (mm) / feed rate (mm/min) = time (min)
	// Convert to seconds
	timeMinutes := distance / feedRate
	timeSeconds := timeMinutes * 60.0

	return time.Duration(timeSeconds * float64(time.Second))
}
