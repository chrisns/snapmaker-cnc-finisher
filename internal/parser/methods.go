package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/256dpi/gcode"
)

// NewParser creates a parser from an io.Reader.
// It parses the GCode file, extracts header metadata, and initializes modal state.
func NewParser(r io.Reader) (*Parser, error) {
	file, err := gcode.ParseFile(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GCode file: %w", err)
	}

	p := &Parser{
		file:     file,
		header:   HeaderMetadata{},
		state:    ModalState{},
		warnings: []string{},
	}

	// Parse header metadata from comment lines
	p.parseHeader()

	// Initialize modal state from header
	p.ResetState()

	return p, nil
}

// Header returns the parsed header metadata.
func (p *Parser) Header() HeaderMetadata {
	return p.header
}

// File returns the underlying gcode.File.
func (p *Parser) File() *gcode.File {
	return p.file
}

// State returns the current modal state.
func (p *Parser) State() ModalState {
	return p.state
}

// Warnings returns any parsing warnings (e.g., missing header fields).
func (p *Parser) Warnings() []string {
	return p.warnings
}

// parseHeader extracts HeaderMetadata from comment lines in the GCode file.
// Snapmaker Luban format: ;key: value or ;key(unit): value
func (p *Parser) parseHeader() {
	for _, line := range p.file.Lines {
		if line.Comment == "" {
			continue
		}

		// Parse header line format: ;key: value or ;key(unit): value
		comment := strings.TrimSpace(line.Comment)

		// Split on first colon
		parts := strings.SplitN(comment, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove unit suffix from key (e.g., "max_z(mm)" -> "max_z")
		if idx := strings.Index(key, "("); idx != -1 {
			key = key[:idx]
		}

		// Parse based on key
		switch key {
		case "file_type":
			p.header.FileType = value
		case "tool_head":
			p.header.ToolHead = value
		case "machine":
			p.header.Machine = value
		case "file_total_lines":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				p.header.TotalLines = v
			}
		case "estimated_time":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.EstimatedTimeSec = v
			}
		case "is_rotate":
			p.header.IsRotate = (value == "true")
		case "max_x":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MaxX = v
			}
		case "min_x":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MinX = v
			}
		case "max_y":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MaxY = v
			}
		case "min_y":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MinY = v
			}
		case "max_z":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MaxZ = v
			}
		case "min_z":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MinZ = v
			}
		case "max_b":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MaxB = v
			}
		case "min_b":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				p.header.MinB = v
			}
		case "work_speed":
			if v, err := strconv.Atoi(value); err == nil {
				p.header.WorkSpeed = v
			}
		case "jog_speed":
			if v, err := strconv.Atoi(value); err == nil {
				p.header.JogSpeed = v
			}
		}
	}

	// Validate header (check for CNC tool head)
	if p.header.ToolHead != "" && !strings.Contains(strings.ToLower(p.header.ToolHead), "cnc") {
		p.warnings = append(p.warnings, fmt.Sprintf("Warning: Tool head '%s' may not be a CNC tool head", p.header.ToolHead))
	}
	if p.header.ToolHead == "" {
		p.warnings = append(p.warnings, "Warning: Missing tool_head in header")
	}
}

// ResetState reinitializes modal state from header metadata.
// Z is initialized from max_z (or 0 if not present), X/Y/B/F default to 0.
func (p *Parser) ResetState() {
	p.state = ModalState{
		X: 0.0,
		Y: 0.0,
		Z: p.header.MaxZ, // Initialize Z from header max_z
		B: 0.0,
		F: 0.0,
	}
}

// UpdateState updates modal state from a GCode line.
// Only coordinates/parameters present in the line update their values.
func (p *Parser) UpdateState(line gcode.Line) {
	for _, code := range line.Codes {
		switch code.Letter {
		case "X":
			p.state.X = code.Value
		case "Y":
			p.state.Y = code.Value
		case "Z":
			p.state.Z = code.Value
		case "B":
			p.state.B = code.Value
		case "F":
			p.state.F = code.Value
		}
	}
}

// ScanMinZ finds the minimum Z value in all G1 commands in the file.
// Returns error if no G1 commands with Z coordinates are found.
func (p *Parser) ScanMinZ() (float64, error) {
	minZ := 0.0
	found := false

	// Reset state for scanning
	p.ResetState()

	for _, line := range p.file.Lines {
		// Update modal state first
		p.UpdateState(line)

		// Check if this is a G0 or G1 command
		isMove := false
		for _, code := range line.Codes {
			if code.Letter == "G" && (code.Value == 0 || code.Value == 1) {
				isMove = true
				break
			}
		}

		if !isMove {
			continue
		}

		// Use current modal state Z value
		if !found {
			minZ = p.state.Z
			found = true
		} else if p.state.Z < minZ {
			minZ = p.state.Z
		}
	}

	if !found {
		return 0.0, fmt.Errorf("no G0/G1 moves found in file")
	}

	// Reset state after scanning
	p.ResetState()

	return minZ, nil
}
