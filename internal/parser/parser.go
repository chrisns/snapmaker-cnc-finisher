// Package parser provides GCode file parsing, header metadata extraction,
// and modal state tracking for CNC GCode processing.
package parser

import (
	"github.com/256dpi/gcode"
)

// HeaderMetadata contains parsed Snapmaker Luban header information.
// The header provides critical metadata for optimization including
// bounding box dimensions, total line count, and 4-axis detection.
type HeaderMetadata struct {
	FileType         string  // e.g., "cnc"
	ToolHead         string  // e.g., "standardCNCToolheadForSM2"
	Machine          string  // e.g., "Snapmaker 2.0 A350"
	TotalLines       int64   // file_total_lines
	EstimatedTimeSec float64 // estimated_time(s)
	IsRotate         bool    // is_rotate (4-axis detection)

	// Bounding box
	MaxX, MinX float64
	MaxY, MinY float64
	MaxZ, MinZ float64
	MaxB, MinB float64 // Rotation axis (if is_rotate)

	// Other
	WorkSpeed int // work_speed(mm/minute)
	JogSpeed  int // jog_speed(mm/minute)
}

// ModalState tracks current position and parameters throughout GCode file processing.
// In GCode modal programming, coordinates and parameters not explicitly specified
// in a command persist from previous commands.
//
// Initialization Rules:
//   - Z: Initialize from header max_z value if present, otherwise 0.0
//   - X, Y, B: Initialize to 0.0
//   - F: Initialize to 0.0 (will be set by first move command with feed rate)
//
// Update Behavior:
//   - Only fields present in current GCode command update their values
//   - Absent fields retain previous values (modal behavior)
//   - Updates occur before move classification/processing
type ModalState struct {
	X float64 // Current X position (mm)
	Y float64 // Current Y position (mm)
	Z float64 // Current Z position/depth (mm, negative = below surface)
	B float64 // Current B rotation (degrees, 4-axis only)
	F float64 // Current feed rate (mm/min)
}

// Parser handles GCode file parsing and modal state management.
type Parser struct {
	file     *gcode.File
	header   HeaderMetadata
	state    ModalState
	warnings []string
}

// GetLine returns a specific line from the file by index.
func (p *Parser) GetLine(index int) gcode.Line {
	if index >= 0 && index < len(p.file.Lines) {
		return p.file.Lines[index]
	}
	return gcode.Line{}
}
