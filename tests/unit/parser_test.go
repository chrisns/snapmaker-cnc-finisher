package unit

import (
	"strings"
	"testing"

	_ "github.com/chrisns/snapmaker-cnc-finisher/internal/parser"
)

// T013: Unit test for header parsing - verify HeaderMetadata extraction from Snapmaker Luban header
func TestHeaderParsing(t *testing.T) {
	t.Skip("Parser not yet implemented - TDD: test written first")
	_ = strings.NewReader
}

// T014: Unit test for modal state initialization - verify Z initialized from max_z, X/Y/B default to 0
func TestModalStateInitialization(t *testing.T) {
	t.Skip("Parser not yet implemented - TDD: test written first")
}

// T015: Unit test for modal state updates - verify coordinates update only when specified, others persist
func TestModalStateUpdates(t *testing.T) {
	t.Skip("Parser not yet implemented - TDD: test written first")
	// Test will verify that when a G1 command specifies only X, the Y and Z persist from previous state
}

// T016: Unit test for ScanMinZ - verify deepest Z value found in G1 commands
func TestScanMinZ(t *testing.T) {
	t.Skip("Parser not yet implemented - TDD: test written first")
}
