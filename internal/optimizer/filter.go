package optimizer

import "github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"

// ShouldFilterByDepth determines if a cutting move at given Z depth should be filtered
// based on the allowance threshold. Returns true if the move is shallow enough to remove.
//
// Z-axis convention: Positive Z increases upward from reference point
// A move is considered "shallow" if Z > (reference - allowance)
func ShouldFilterByDepth(z float64, allowance float64, meta *gcode.Metadata) bool {
	return meta.IsShallowDepth(z, allowance)
}

// ShouldAlwaysPreserve returns true if the command should always be kept
// regardless of depth filtering. This includes:
// - G0 rapid moves (positioning, not cutting)
// - M-codes (machine control: spindle, coolant, etc.)
// - Comments (documentation)
// - Empty commands
func ShouldAlwaysPreserve(cmd gcode.Command) bool {
	// Preserve rapid moves (G0)
	if cmd.IsRapidMove() {
		return true
	}

	// Preserve machine codes (M-codes)
	if cmd.IsMachineCode() {
		return true
	}

	// Preserve comments
	if cmd.IsComment() {
		return true
	}

	// Preserve empty commands
	if cmd.Letter == "" && cmd.Comment == "" {
		return true
	}

	return false
}

// ShouldFilterMove determines if a move should be filtered based on depth and strategy
// This is the main filtering decision function that combines depth checking,
// command type checking, and multi-axis strategy application
func ShouldFilterMove(cmd gcode.Command, allowance float64, meta *gcode.Metadata, strategy FilterStrategy) bool {
	// Never filter commands that should always be preserved
	if ShouldAlwaysPreserve(cmd) {
		return false
	}

	// Only filter cutting moves (G1)
	if !cmd.IsCuttingMove() {
		return false
	}

	// Get Z parameter
	z, hasZ := cmd.Params["Z"]
	if !hasZ {
		return false // No Z movement, don't filter
	}

	// Check if depth is shallow
	isShallow := ShouldFilterByDepth(z, allowance, meta)
	if !isShallow {
		return false // Deep cut, always keep
	}

	// Apply strategy for multi-axis moves
	return applyStrategy(cmd, strategy)
}

// applyStrategy applies the filtering strategy to determine if a shallow move should be filtered
func applyStrategy(cmd gcode.Command, strategy FilterStrategy) bool {
	// Check if this is a multi-axis move (has X, Y, A, or B in addition to Z)
	isMultiAxis := false
	for param := range cmd.Params {
		if param == "X" || param == "Y" || param == "A" || param == "B" {
			isMultiAxis = true
			break
		}
	}

	switch strategy {
	case StrategySafe:
		// Only filter if it's a Z-only move (not multi-axis)
		return !isMultiAxis

	case StrategyAllAxes:
		// Filter based on Z depth regardless of other axes
		return true

	case StrategySplit:
		// For now, treat like Safe (split strategy requires move decomposition - advanced)
		return !isMultiAxis

	case StrategyAggressive:
		// Filter all shallow moves, even multi-axis
		return true

	default:
		// Unknown strategy, use safe default
		return !isMultiAxis
	}
}
