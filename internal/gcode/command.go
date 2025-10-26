package gcode

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/256dpi/gcode"
)

// Command represents a parsed GCode command
type Command struct {
	Letter  string             // Command letter (G, M, etc.)
	Value   int                // Command number (e.g., 1 for G1)
	Params  map[string]float64 // Parameters (X, Y, Z, F, S, etc.)
	Comment string             // Comment text if line is a comment
}

// ParseCommand parses a GCode line into a Command struct
func ParseCommand(input string) (Command, error) {
	input = strings.TrimSpace(input)

	// Empty line
	if input == "" {
		return Command{}, nil
	}

	// Comment line
	if strings.HasPrefix(input, ";") {
		return Command{Comment: input}, nil
	}

	// Parse using gcode library
	parsed, err := gcode.ParseLine(input)
	if err != nil {
		return Command{}, fmt.Errorf("failed to parse line: %w", err)
	}

	// Extract commands (G, M, etc.)
	cmd := Command{
		Params: make(map[string]float64),
	}

	for _, code := range parsed.Codes {
		letter := code.Letter

		// Handle command codes (G, M)
		if letter == "G" || letter == "M" || letter == "T" {
			cmd.Letter = letter
			cmd.Value = int(code.Value)
		} else {
			// Handle parameters (X, Y, Z, F, S, etc.)
			cmd.Params[letter] = code.Value
		}
	}

	// Handle comment
	if parsed.Comment != "" {
		cmd.Comment = parsed.Comment
	}

	return cmd, nil
}

// IsRapidMove returns true if this is a G0 rapid move command
func (c Command) IsRapidMove() bool {
	return c.Letter == "G" && c.Value == 0
}

// IsCuttingMove returns true if this is a G1 cutting move command with feed rate
func (c Command) IsCuttingMove() bool {
	if c.Letter != "G" || c.Value != 1 {
		return false
	}
	_, hasFeedRate := c.Params["F"]
	return hasFeedRate || len(c.Params) > 0 // G1 with any params is a cutting move
}

// IsMachineCode returns true if this is an M-code
func (c Command) IsMachineCode() bool {
	return c.Letter == "M"
}

// IsComment returns true if this is a comment line
func (c Command) IsComment() bool {
	return c.Comment != "" && c.Letter == ""
}

// HasParam returns true if the command has the specified parameter
func (c Command) HasParam(param string) bool {
	_, ok := c.Params[param]
	return ok
}

// GetParam returns the value of the specified parameter or 0 if not present
func (c Command) GetParam(param string) float64 {
	return c.Params[param]
}

// String returns a string representation of the command
func (c Command) String() string {
	if c.IsComment() {
		return c.Comment
	}

	if c.Letter == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(c.Letter)
	sb.WriteString(strconv.Itoa(c.Value))

	// Add parameters in consistent order
	paramOrder := []string{"X", "Y", "Z", "B", "A", "F", "S", "I", "J", "K", "P", "Q", "R"}
	for _, param := range paramOrder {
		if val, ok := c.Params[param]; ok {
			sb.WriteString(" ")
			sb.WriteString(param)
			sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
		}
	}

	// Add any remaining parameters not in the standard order
	for param, val := range c.Params {
		found := false
		for _, p := range paramOrder {
			if p == param {
				found = true
				break
			}
		}
		if !found {
			sb.WriteString(" ")
			sb.WriteString(param)
			sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
		}
	}

	if c.Comment != "" {
		sb.WriteString(" ")
		sb.WriteString(c.Comment)
	}

	return sb.String()
}
