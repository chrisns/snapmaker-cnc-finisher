package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
)

// generateLargeGCode creates a realistic GCode file with specified number of lines
// for testing progress reporting and large file handling
func main() {
	lines := flag.Int("lines", 10000000, "Number of lines to generate")
	output := flag.String("output", "tests/testdata/large_file.cnc", "Output file path")
	flag.Parse()

	file, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Track actual line count
	lineCount := 0

	// Write Snapmaker Luban header
	header := []string{
		";Header Start",
		";header_type: 3dp",
		";tool_head: Snapmaker 2.0 CNC 1.5W",
		";machine: Snapmaker 2.0 A350",
		";work_speed: 1000",
		";min_x: -10.0",
		";max_x: 100.0",
		";min_y: -10.0",
		";max_y: 100.0",
		";min_z: -5.0",
		";max_z: 0.0",
		";min_b: 0.0",
		";max_b: 0.0",
		";estimated_time: 7200",
		";Header End",
		"",
		"M3 S12000 ;Start spindle",
		"G0 Z5.0 ;Raise tool",
		"",
	}

	for _, line := range header {
		fmt.Fprintln(writer, line)
		lineCount++
	}

	// Generate realistic toolpath lines
	// Mix of shallow cuts (< 1.0mm) and deep cuts (> 1.0mm)
	// Patterns: linear moves, circular arcs, rapid positioning

	x, y, z := 0.0, 0.0, 0.0
	feedRate := 1500.0

	// Calculate target number of move lines
	footer := []string{
		"",
		"G0 Z10.0 ;Raise tool",
		"M5 ;Stop spindle",
		"G0 X0 Y0 ;Return to origin",
		"M2 ;End program",
	}
	targetMoveLines := *lines - len(header) - len(footer)

	for i := 0; i < targetMoveLines; i++ {
		// Every 100th line is a comment
		if i%100 == 0 {
			fmt.Fprintf(writer, "; Layer %d - progress checkpoint\n", i/100)
			lineCount++
			continue
		}

		// Every 50th line is a rapid move (always preserved)
		if i%50 == 0 {
			z += 5.0
			fmt.Fprintf(writer, "G0 Z%.3f\n", math.Min(z, 5.0))
			z = math.Min(z, 5.0)
			lineCount++
			continue
		}

		// Alternate between shallow and deep cuts
		// 60% shallow (should be filtered), 40% deep (should be kept)
		if i%5 < 3 {
			// Shallow cut (Z between -0.9mm and 0mm)
			z = -0.9 + (float64(i%10) * 0.1)
		} else {
			// Deep cut (Z between -1.5mm and -5.0mm)
			z = -1.5 - (float64(i%10) * 0.3)
		}

		// Circular motion pattern
		angle := float64(i) * 0.01
		x = 50.0 + 40.0*math.Cos(angle)
		y = 50.0 + 40.0*math.Sin(angle)

		// Vary feed rate occasionally
		if i%1000 == 0 {
			feedRate = 1000.0 + float64(i%5)*200.0
			fmt.Fprintf(writer, "G1 X%.3f Y%.3f Z%.3f F%.1f\n", x, y, z, feedRate)
		} else {
			fmt.Fprintf(writer, "G1 X%.3f Y%.3f Z%.3f\n", x, y, z)
		}
		lineCount++
	}

	// Footer
	for _, line := range footer {
		fmt.Fprintln(writer, line)
		lineCount++
	}

	fmt.Printf("Generated %d lines in %s\n", lineCount, *output)
}
