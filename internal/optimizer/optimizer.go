// Package optimizer provides core optimization logic for GCode finishing pass files.
// It classifies moves, applies optimization strategies, and calculates intersection points
// for move splitting.
package optimizer

// MoveClassification categorizes a move (G0 or G1) relative to the depth threshold.
// This classification determines what optimization action to take.
type MoveClassification int

const (
	// Shallow: Both start and end points above threshold → Remove move
	Shallow MoveClassification = iota
	// Deep: Both start and end points at/below threshold → Preserve move
	Deep
	// CrossingEnter: Starts above threshold, ends below/at threshold → Split or preserve based on strategy
	CrossingEnter
	// CrossingLeave: Starts below/at threshold, ends above threshold → Split or preserve based on strategy
	CrossingLeave
	// NonCutting: Not a G1 command → Always preserve as-is
	NonCutting
)

// String returns the string representation of the MoveClassification.
func (mc MoveClassification) String() string {
	switch mc {
	case Shallow:
		return "Shallow"
	case Deep:
		return "Deep"
	case CrossingEnter:
		return "CrossingEnter"
	case CrossingLeave:
		return "CrossingLeave"
	case NonCutting:
		return "NonCutting"
	default:
		return "Unknown"
	}
}

// OptimizationStrategy defines how moves that cross the depth threshold are handled.
type OptimizationStrategy int

const (
	// Conservative: Preserve entire crossing moves (safer, less optimization)
	Conservative OptimizationStrategy = iota
	// Aggressive: Split crossing moves at threshold intersection point (maximum time savings)
	Aggressive
)

// String returns the string representation of the OptimizationStrategy.
func (os OptimizationStrategy) String() string {
	switch os {
	case Conservative:
		return "conservative"
	case Aggressive:
		return "aggressive"
	default:
		return "unknown"
	}
}

// IntersectionPoint represents where a move crosses the depth threshold plane.
// Calculated using parametric linear interpolation.
type IntersectionPoint struct {
	X float64 // X coordinate at intersection
	Y float64 // Y coordinate at intersection
	Z float64 // Z coordinate (equals threshold exactly)
	T float64 // Parametric parameter (0 < t < 1)
}

// Optimizer applies depth-based filtering to GCode moves.
type Optimizer struct {
	threshold float64
	strategy  OptimizationStrategy
}
