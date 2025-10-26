package gcode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ZReference indicates how the Z-axis reference was determined
type ZReference int

const (
	ZRefMetadata      ZReference = iota // From GCode header metadata (min_z/max_z)
	ZRefMachineOrigin                   // Fallback to machine work origin
	ZRefSurface                         // Fallback to material surface convention (Z=0 = top)

	// HeaderScanLines is the maximum number of lines to scan for metadata
	HeaderScanLines = 50
)

// Metadata contains extracted header information from a GCode file
type Metadata struct {
	MinZ        float64    // Minimum Z value from header
	MaxZ        float64    // Maximum Z value from header
	Is4Axis     bool       // True if B-axis commands detected
	ZReference  ZReference // How Z-axis reference was determined
	ZRefMessage string     // Human-readable message about Z reference method
}

// ExtractMetadata scans the first 50 lines of a GCode file to extract header metadata
func ExtractMetadata(r io.Reader) (*Metadata, error) {
	scanner := bufio.NewScanner(r)
	meta := &Metadata{}

	hasMinZ := false
	hasMaxZ := false
	lineCount := 0

	// Scan first N lines for header metadata
	for scanner.Scan() && lineCount < HeaderScanLines {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Check for MIN_Z in header
		if strings.HasPrefix(line, ";MIN_Z:") || strings.Contains(line, "MIN_Z:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val := strings.TrimSpace(parts[len(parts)-1])
				if minZ, err := strconv.ParseFloat(val, 64); err == nil {
					meta.MinZ = minZ
					hasMinZ = true
				}
			}
		}

		// Check for MAX_Z in header
		if strings.HasPrefix(line, ";MAX_Z:") || strings.Contains(line, "MAX_Z:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val := strings.TrimSpace(parts[len(parts)-1])
				if maxZ, err := strconv.ParseFloat(val, 64); err == nil {
					meta.MaxZ = maxZ
					hasMaxZ = true
				}
			}
		}

		// Check for B-axis commands (4-axis indicator)
		if strings.Contains(line, "B") && !strings.HasPrefix(line, ";") {
			// Parse to check if it's actually a B parameter
			if cmd, err := ParseCommand(line); err == nil {
				if _, hasB := cmd.Params["B"]; hasB {
					meta.Is4Axis = true
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	// Determine Z reference method based on metadata availability
	// Check if we have complete metadata (both min and max)
	hasCompleteMetadata := hasMinZ && hasMaxZ

	if hasCompleteMetadata {
		meta.ZReference = ZRefMetadata
		meta.ZRefMessage = "Using Z-axis reference from GCode header metadata"
	} else if hasMinZ || hasMaxZ {
		// Incomplete metadata - fall back to machine origin
		meta.ZReference = ZRefMachineOrigin
		meta.ZRefMessage = "Z-axis reference: falling back to machine work origin (metadata incomplete)"
	} else {
		// No metadata at all - use surface convention
		meta.ZReference = ZRefSurface
		meta.ZRefMessage = "Z-axis reference: using material surface convention (Z=0 = top surface)"
	}

	return meta, nil
}

// GetZReference returns the Z-axis reference point value
func (m *Metadata) GetZReference() float64 {
	switch m.ZReference {
	case ZRefMetadata:
		// Use the material surface (MaxZ) as reference
		return m.MaxZ
	case ZRefMachineOrigin:
		// Use machine origin (0)
		return 0.0
	case ZRefSurface:
		// Use surface convention (0)
		return 0.0
	default:
		return 0.0
	}
}

// IsShallowDepth determines if a Z value is shallower than the allowance threshold
// Z-axis convention: Positive Z increases upward from reference point
// Shallow means Z > (reference - allowance)
func (m *Metadata) IsShallowDepth(z float64, allowance float64) bool {
	reference := m.GetZReference()
	threshold := reference - allowance
	return z > threshold
}
