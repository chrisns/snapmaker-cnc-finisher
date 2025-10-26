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
