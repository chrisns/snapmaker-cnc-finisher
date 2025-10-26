package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/cli"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/gcode"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
)

func main() {
	exitCode := run(os.Args[1:])
	os.Exit(exitCode)
}

func run(args []string) int {
	// Check for help flag first
	if cli.ShouldShowHelp(args) {
		fmt.Print(cli.GetHelpText())
		return 0
	}

	// Check for version flag
	if cli.ShouldShowVersion(args) {
		fmt.Print(cli.GetVersionText())
		return 0
	}

	// Parse command-line arguments
	parsedArgs, err := cli.ParseArgs(args)
	if err != nil {
		return cli.PrintError(err)
	}

	// Validate arguments
	if err := cli.ValidateArgs(parsedArgs); err != nil {
		return cli.PrintError(err)
	}

	// Parse filter strategy
	strategy, err := optimizer.ParseFilterStrategy(parsedArgs.Strategy)
	if err != nil {
		return cli.PrintError(&cli.InvalidStrategyError{Strategy: parsedArgs.Strategy})
	}

	// Check if output file exists and --force not provided
	if !parsedArgs.Force {
		if _, err := os.Stat(parsedArgs.OutputFile); err == nil {
			return cli.PrintError(fmt.Errorf("output file already exists: %s (use --force to overwrite)", parsedArgs.OutputFile))
		}
	}

	// Start timing
	startTime := time.Now()

	// Read input file and extract metadata
	inputFile, err := os.Open(parsedArgs.InputFile)
	if err != nil {
		return cli.PrintError(fmt.Errorf("failed to open input file: %w", err))
	}
	defer inputFile.Close()

	metadata, err := gcode.ExtractMetadata(inputFile)
	if err != nil {
		return cli.PrintError(fmt.Errorf("failed to extract metadata: %w", err))
	}

	// Print warning if header is missing or incomplete
	if metadata.ZReference != gcode.ZRefMetadata {
		cli.PrintWarning("%s", metadata.ZRefMessage)
	}

	// Reopen input file for reading (metadata extraction consumed it)
	inputFile.Close()
	inputFile, err = os.Open(parsedArgs.InputFile)
	if err != nil {
		return cli.PrintError(fmt.Errorf("failed to reopen input file: %w", err))
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(parsedArgs.OutputFile)
	if err != nil {
		return cli.PrintError(fmt.Errorf("failed to create output file: %w", err))
	}
	defer outputFile.Close()

	// Initialize statistics
	stats := optimizer.NewStatistics()

	// Get input file size
	if fileInfo, err := inputFile.Stat(); err == nil {
		stats.BytesIn = fileInfo.Size()
	}

	// Create buffered writer
	writer := gcode.NewBufferedWriter(outputFile)

	// Process file line by line using scanner
	scanner := bufio.NewScanner(inputFile)
	buf := make([]byte, 0, gcode.InitialBufferSize)
	scanner.Buffer(buf, gcode.MaxLineLength)

	// Initialize progress tracking with better estimation
	var progressTracker *cli.ProgressTracker
	var lastProgressUpdate int
	var lastProgressTime time.Time
	var progressDisplayed bool

	var lastX, lastY, lastZ float64
	var currentFeedRate float64
	var feedRateWarned bool // Track if we've warned about using default feed rate

	for scanner.Scan() {
		line := scanner.Text()
		stats.TotalLines++

		// Initialize progress tracker with improved heuristic
		if progressTracker == nil && stats.TotalLines == 1 {
			// Better estimate: GCode files average 30-50 bytes per line
			// Use 35 as middle ground
			estimatedLines := int(stats.BytesIn / 35)
			if estimatedLines < 1000 {
				estimatedLines = 1000 // Minimum reasonable estimate
			}
			progressTracker = cli.NewProgressTracker(estimatedLines)
			lastProgressTime = startTime
		}

		// Update progress display if needed (every 10k lines or 2 seconds)
		if progressTracker != nil {
			progressTracker.Update(stats.TotalLines, stats.RemovedLines)
			timeSinceLastUpdate := time.Since(lastProgressTime)

			// Dynamically adjust estimate if we exceed it
			if stats.TotalLines > progressTracker.TotalLines() {
				progressTracker.UpdateTotalEstimate(stats.TotalLines + int(stats.TotalLines/10))
			}

			if progressTracker.ShouldUpdate(lastProgressUpdate, timeSinceLastUpdate) {
				elapsed := time.Since(startTime)
				progressTracker.Display(os.Stdout, elapsed)
				progressDisplayed = true
				lastProgressUpdate = stats.TotalLines
				lastProgressTime = time.Now()
			}
		}

		// Parse command
		cmd, err := gcode.ParseCommand(line)
		if err != nil {
			// Skip malformed lines with warning but continue processing
			cli.PrintWarning("Skipping malformed line %d: %v", stats.TotalLines, err)
			continue
		}

		// Check if this is a cutting move and we need feed rate
		if cmd.Letter == "G" && (cmd.Value == 1 || cmd.Value == 2 || cmd.Value == 3) {
			// This is a cutting move (G1, G2, G3)
			// Warn if no feed rate has been specified yet
			if currentFeedRate == 0 && !feedRateWarned {
				cli.PrintWarning("No feed rate (F parameter) found in GCode file, using default %v mm/min for time calculations", optimizer.DefaultFeedRate)
				feedRateWarned = true
			}
		}

		// Decide whether to filter this line
		shouldFilter := optimizer.ShouldFilterMove(cmd, parsedArgs.Allowance, metadata, strategy)

		if shouldFilter {
			stats.RemovedLines++

			// Calculate time saved for this move (BEFORE updating position)
			// Get the new position from the command
			newX := lastX // Default to current position if parameter not present
			newY := lastY
			newZ := lastZ

			if x, hasX := cmd.Params["X"]; hasX {
				newX = x
			}
			if y, hasY := cmd.Params["Y"]; hasY {
				newY = y
			}
			if z, hasZ := cmd.Params["Z"]; hasZ {
				newZ = z
			}

			// Calculate time savings from last position to new position
			timeSaved := optimizer.CalculateTimeSaved(
				lastX, lastY, lastZ,
				newX, newY, newZ,
				currentFeedRate,
			)
			stats.EstimatedTimeSaved += timeSaved
		} else {
			// Keep this line - update position state and write to output
			if x, hasX := cmd.Params["X"]; hasX {
				lastX = x
			}
			if y, hasY := cmd.Params["Y"]; hasY {
				lastY = y
			}
			if z, hasZ := cmd.Params["Z"]; hasZ {
				lastZ = z
			}
			if f, hasF := cmd.Params["F"]; hasF {
				currentFeedRate = f
			}

			if err := writer.WriteLine(line); err != nil {
				return cli.PrintError(fmt.Errorf("failed to write output: %w", err))
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return cli.PrintError(fmt.Errorf("error reading input file: %w", err))
	}

	// Flush remaining lines
	if err := writer.Flush(); err != nil {
		return cli.PrintError(fmt.Errorf("failed to flush output: %w", err))
	}

	// Get output file size
	if fileInfo, err := outputFile.Stat(); err == nil {
		stats.BytesOut = fileInfo.Size()
	}

	// Record processing time
	stats.ProcessingTime = time.Since(startTime)

	// Clear progress line (only if it was actually displayed)
	if progressDisplayed {
		// Clear with enough spaces to cover typical progress line (~120 chars)
		fmt.Print("\r" + string(make([]byte, 120)) + "\r")
	}

	// Print summary
	cli.PrintSummary(stats)

	return 0
}
