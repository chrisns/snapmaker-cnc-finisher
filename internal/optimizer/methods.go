package optimizer

import (
	"fmt"
	"math"

	"github.com/256dpi/gcode"
)

// NewOptimizer creates an optimizer with specified threshold and strategy.
// threshold = minZ + allowance
func NewOptimizer(minZ, allowance float64, strategy OptimizationStrategy) *Optimizer {
	return &Optimizer{
		threshold: minZ + allowance,
		strategy:  strategy,
	}
}

// Threshold returns the calculated depth threshold.
func (o *Optimizer) Threshold() float64 {
	return o.threshold
}

// Strategy returns the optimization strategy.
func (o *Optimizer) Strategy() OptimizationStrategy {
	return o.strategy
}

// ClassifyMove categorizes a move (G0 or G1) based on start/end Z relative to threshold.
func (o *Optimizer) ClassifyMove(startZ, endZ float64) MoveClassification {
	startDeep := startZ <= o.threshold
	endDeep := endZ <= o.threshold

	if !startDeep && !endDeep {
		return Shallow // Both points above threshold
	}
	if startDeep && endDeep {
		return Deep // Both points at/below threshold
	}
	if !startDeep && endDeep {
		return CrossingEnter // Entering deep zone
	}
	return CrossingLeave // Leaving deep zone
}

// CalculateIntersection finds where a move crosses the threshold using parametric interpolation.
// Returns error if move doesn't cross threshold or if division by zero.
func (o *Optimizer) CalculateIntersection(startX, startY, startZ, endX, endY, endZ float64) (IntersectionPoint, error) {
	deltaZ := endZ - startZ

	// Check for horizontal move (no Z change)
	if math.Abs(deltaZ) < 1e-9 {
		return IntersectionPoint{}, fmt.Errorf("move does not cross threshold vertically (deltaZ ≈ 0)")
	}

	// Calculate parametric parameter t
	t := (o.threshold - startZ) / deltaZ

	// Validate t is in range (0, 1) - move must cross threshold within segment
	if t <= 0 || t >= 1 {
		return IntersectionPoint{}, fmt.Errorf("intersection parameter t=%v out of range (0,1)", t)
	}

	// Calculate intersection point using parametric interpolation
	point := IntersectionPoint{
		X: startX + t*(endX-startX),
		Y: startY + t*(endY-startY),
		Z: o.threshold, // Exact threshold value
		T: t,
	}

	return point, nil
}

// ShouldPreserve determines if a line should be included in output based on classification and strategy.
// Returns true if the line should be preserved as-is (no splitting).
func (o *Optimizer) ShouldPreserve(classification MoveClassification) bool {
	switch classification {
	case Shallow:
		return false // Always remove shallow moves
	case Deep:
		return true // Always preserve deep moves
	case NonCutting:
		return true // Always preserve non-cutting commands
	case CrossingEnter, CrossingLeave:
		// Conservative: preserve entire crossing move
		// Aggressive: will split (return false to indicate splitting needed)
		return o.strategy == Conservative
	default:
		return true // Unknown classification, preserve to be safe
	}
}

// SplitMove generates new GCode lines for a crossing move (G0 or G1).
// For G1 moves: preserves feed rate in split segments.
// For G0 moves: no feed rate (rapid positioning).
// For CrossingEnter: returns (moveToIntersection, moveFromIntersection)
// For CrossingLeave: returns (moveToIntersection, empty)
func (o *Optimizer) SplitMove(line gcode.Line, intersection IntersectionPoint, classification MoveClassification, startX, startY, startZ float64) (gcode.Line, gcode.Line, error) {
	// Extract move type (G0 or G1) and feed rate from original line
	feedRate := 0.0
	gValue := -1.0

	for _, code := range line.Codes {
		if code.Letter == "G" && (code.Value == 0 || code.Value == 1) {
			gValue = code.Value
		}
		if code.Letter == "F" {
			feedRate = code.Value
		}
	}

	if gValue != 0 && gValue != 1 {
		return gcode.Line{}, gcode.Line{}, fmt.Errorf("line is not a G0 or G1 move")
	}

	// Get end coordinates from original line
	endX, endY, endZ := startX, startY, startZ
	for _, code := range line.Codes {
		switch code.Letter {
		case "X":
			endX = code.Value
		case "Y":
			endY = code.Value
		case "Z":
			endZ = code.Value
		}
	}

	var line1, line2 gcode.Line

	if classification == CrossingEnter {
		// Start above, end below: preserve intersection → end
		line1 = gcode.Line{
			Codes: []gcode.GCode{
				{Letter: "G", Value: gValue}, // Use detected G value (0 or 1)
				{Letter: "X", Value: math.Round(intersection.X*10000) / 10000}, // 4 decimal places
				{Letter: "Y", Value: math.Round(intersection.Y*10000) / 10000},
				{Letter: "Z", Value: math.Round(intersection.Z*10000) / 10000},
			},
		}
		// Only add feed rate for G1 moves (G0 rapid moves don't use feed rates)
		if gValue == 1 && feedRate > 0 {
			line1.Codes = append(line1.Codes, gcode.GCode{Letter: "F", Value: feedRate})
		}

		line2 = gcode.Line{
			Codes: []gcode.GCode{
				{Letter: "G", Value: gValue},
				{Letter: "X", Value: endX},
				{Letter: "Y", Value: endY},
				{Letter: "Z", Value: endZ},
			},
		}
		if gValue == 1 && feedRate > 0 {
			line2.Codes = append(line2.Codes, gcode.GCode{Letter: "F", Value: feedRate})
		}
	} else if classification == CrossingLeave {
		// Start below, end above: preserve start → intersection
		line1 = gcode.Line{
			Codes: []gcode.GCode{
				{Letter: "G", Value: gValue},
				{Letter: "X", Value: math.Round(intersection.X*10000) / 10000},
				{Letter: "Y", Value: math.Round(intersection.Y*10000) / 10000},
				{Letter: "Z", Value: math.Round(intersection.Z*10000) / 10000},
			},
		}
		if gValue == 1 && feedRate > 0 {
			line1.Codes = append(line1.Codes, gcode.GCode{Letter: "F", Value: feedRate})
		}
		// line2 remains empty (discard shallow portion)
	}

	return line1, line2, nil
}
