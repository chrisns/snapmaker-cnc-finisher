package integration

import (
	"testing"
)

// T021: Integration test for end-to-end optimization (aggressive strategy)
func TestEndToEndOptimizationAggressive(t *testing.T) {
	t.Skip("CLI not yet implemented - TDD: test written first")
	// Test will:
	// 1. Run optimizer on freya-subset.cnc with 1.0mm allowance, aggressive strategy
	// 2. Verify output file has fewer lines than input
	// 3. Verify moves are correctly filtered/split
	// 4. Verify output is valid GCode
}

// T022: Integration test for end-to-end optimization (conservative strategy)
func TestEndToEndOptimizationConservative(t *testing.T) {
	t.Skip("CLI not yet implemented - TDD: test written first")
	// Test will:
	// 1. Run optimizer on freya-subset.cnc with 1.0mm allowance, conservative strategy
	// 2. Verify crossing moves are preserved entirely
	// 3. Verify output file size is larger than aggressive but smaller than input
}

// T023: Integration test for 3-axis vs 4-axis detection
func TestAxisDetection(t *testing.T) {
	t.Skip("CLI not yet implemented - TDD: test written first")
	// Test will:
	// 1. Test file with is_rotate: true - verify B axis tracking
	// 2. Test file with is_rotate: false - verify no B axis errors
}
