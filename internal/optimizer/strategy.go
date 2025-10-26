package optimizer

import (
	"fmt"
	"strings"
)

// FilterStrategy defines how multi-axis moves should be handled during optimization
type FilterStrategy int

const (
	// StrategySafe: Only filter if Z is shallow AND all other axes stationary
	// Most conservative - preserves complex toolpaths
	StrategySafe FilterStrategy = iota

	// StrategyAllAxes: Filter based on Z depth regardless of other axes
	// Moderate - removes shallow cuts even with XY/B motion
	StrategyAllAxes

	// StrategySplit: Split multi-axis moves, filter Z-only portions
	// Advanced - requires move decomposition
	StrategySplit

	// StrategyAggressive: Filter all shallow moves, even multi-axis
	// Maximum reduction - may affect surface quality on complex parts
	StrategyAggressive
)

// String returns the string representation of the strategy
func (s FilterStrategy) String() string {
	switch s {
	case StrategySafe:
		return "safe"
	case StrategyAllAxes:
		return "all-axes"
	case StrategySplit:
		return "split"
	case StrategyAggressive:
		return "aggressive"
	default:
		return "unknown"
	}
}

// ParseFilterStrategy parses a string into a FilterStrategy
// Returns error if the string doesn't match any valid strategy
func ParseFilterStrategy(s string) (FilterStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "safe":
		return StrategySafe, nil
	case "all-axes":
		return StrategyAllAxes, nil
	case "split":
		return StrategySplit, nil
	case "aggressive":
		return StrategyAggressive, nil
	default:
		return StrategySafe, fmt.Errorf("invalid filter strategy: %q (valid: safe, all-axes, split, aggressive)", s)
	}
}
