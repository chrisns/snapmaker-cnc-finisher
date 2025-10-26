package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/chrisns/snapmaker-cnc-finisher/internal/optimizer"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/parser"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/progress"
	"github.com/chrisns/snapmaker-cnc-finisher/internal/writer"
)

var (
	force    = flag.Bool("force", false, "Overwrite output file without confirmation")
	strategy = flag.String("strategy", "aggressive", "Optimization strategy (conservative|aggressive)")
	version  = flag.Bool("version", false, "Show version information")
	help     = flag.Bool("help", false, "Show help message")
)

const versionString = "1.0.0"

func main() {
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("gcode-optimizer version %s\n", versionString)
		os.Exit(0)
	}

	// Handle help flag
	if *help {
		printHelp()
		os.Exit(0)
	}

	// Get positional arguments
	args := flag.Args()
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Error: Expected 3 arguments, got %d\n\n", len(args))
		printUsage()
		os.Exit(1)
	}

	inputPath := args[0]
	allowanceStr := args[1]
	outputPath := args[2]

	// Validate allowance
	allowance, err := strconv.ParseFloat(allowanceStr, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Allowance must be a non-negative number, got: %s\n", allowanceStr)
		os.Exit(1)
	}
	if allowance < 0 {
		fmt.Fprintf(os.Stderr, "Error: Allowance must be a non-negative number, got: %v\n", allowance)
		os.Exit(1)
	}

	// Validate strategy
	var optStrategy optimizer.OptimizationStrategy
	switch *strategy {
	case "conservative":
		optStrategy = optimizer.Conservative
	case "aggressive":
		optStrategy = optimizer.Aggressive
	default:
		fmt.Fprintf(os.Stderr, "Invalid strategy '%s'. Valid options are: conservative, aggressive\n", *strategy)
		os.Exit(1)
	}

	// Check input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file not found: %s\n", inputPath)
		os.Exit(1)
	}

	// Check if output file exists (unless --force)
	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("Output file exists: %s\nOverwrite? (y/n): ", outputPath)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled.")
				os.Exit(0)
			}
		}
	}

	// Run optimization
	if err := optimizeGCodeFile(inputPath, outputPath, allowance, optStrategy); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func optimizeGCodeFile(inputPath, outputPath string, allowance float64, strategy optimizer.OptimizationStrategy) error {
	// Get input file size
	inputFileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}
	inputFileSize := inputFileInfo.Size()

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Parse GCode file
	fmt.Println("Parsing GCode file...")
	p, err := parser.NewParser(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse GCode file: %w", err)
	}

	// Display warnings
	for _, warning := range p.Warnings() {
		fmt.Fprintf(os.Stderr, "%s\n", warning)
	}

	// Scan for minimum Z value
	fmt.Println("Analyzing depth...")
	minZ, err := p.ScanMinZ()
	if err != nil {
		return fmt.Errorf("failed to find minimum Z: %w", err)
	}

	// Calculate threshold
	threshold := minZ + allowance

	// Display depth analysis
	fmt.Printf("\nDepth Analysis:\n")
	fmt.Printf("  Min Z: %.3fmm\n", minZ)
	fmt.Printf("  Threshold: %.3fmm (%.1fmm allowance)\n\n", threshold, allowance)

	// Create optimizer
	opt := optimizer.NewOptimizer(minZ, allowance, strategy)

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create writer
	wr := writer.NewWriter(outputFile)

	// Create progress reporter
	totalLines := p.Header().TotalLines
	if totalLines == 0 {
		totalLines = int64(len(p.File().Lines))
	}
	reporter := progress.NewReporter(totalLines)

	// Process file
	fmt.Println("Optimizing...")
	startTime := time.Now()
	linesProcessed := int64(0)
	linesRemoved := int64(0)
	linesPreserved := int64(0)
	linesSplit := int64(0)

	// Reset modal state for processing
	p.ResetState()

	for _, line := range p.File().Lines {
		linesProcessed++
		reporter.Update(linesProcessed)

		// Track start position before updating state
		startZ := p.State().Z

		// Update modal state
		p.UpdateState(line)

		// Check if this is a G1 cutting move
		isG1 := false
		for _, code := range line.Codes {
			if code.Letter == "G" && code.Value == 1 {
				isG1 = true
				break
			}
		}

		if !isG1 {
			// Not a G1 move - preserve as-is
			if err := wr.WriteLine(line); err != nil {
				return fmt.Errorf("failed to write line: %w", err)
			}
			linesPreserved++
			continue
		}

		// Get end Z from current state (after update)
		endZ := p.State().Z

		// Classify the move
		classification := opt.ClassifyMove(startZ, endZ)

		// Check if we should preserve this move
		if opt.ShouldPreserve(classification) {
			if err := wr.WriteLine(line); err != nil {
				return fmt.Errorf("failed to write line: %w", err)
			}
			linesPreserved++
		} else {
			linesRemoved++
		}
	}

	reporter.Finish()
	processingDuration := time.Since(startTime)

	// Get output file size
	outputFileInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}
	outputFileSize := outputFileInfo.Size()

	// Calculate statistics
	result := progress.OptimizationResult{
		TotalInputLines:       linesProcessed,
		InputFileSizeBytes:    inputFileSize,
		LinesProcessed:        linesProcessed,
		LinesRemoved:          linesRemoved,
		LinesPreserved:        linesPreserved,
		LinesSplit:            linesSplit,
		MinZ:                  minZ,
		Threshold:             threshold,
		TotalOutputLines:      linesPreserved + linesSplit,
		OutputFileSizeBytes:   outputFileSize,
		ReductionPercent:      float64(linesRemoved) / float64(linesProcessed) * 100,
		ProcessingDurationSec: processingDuration.Seconds(),
		LinesPerSecond:        float64(linesProcessed) / processingDuration.Seconds(),
	}

	// Display formatted results
	formatter := progress.ResultFormatter{}
	formatter.Display(result)

	fmt.Printf("\n✓ Optimization complete! Output written to: %s\n", outputPath)

	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: gcode-optimizer [flags] <input.cnc> <allowance> <output.cnc>\n\n")
	fmt.Fprintf(os.Stderr, "Arguments:\n")
	fmt.Fprintf(os.Stderr, "  input.cnc    Path to finishing pass GCode file\n")
	fmt.Fprintf(os.Stderr, "  allowance    Material thickness left after rough cut (mm)\n")
	fmt.Fprintf(os.Stderr, "  output.cnc   Path for optimized output file\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func printHelp() {
	fmt.Println("GCode Finishing Pass Optimizer")
	fmt.Println("==============================")
	fmt.Println()
	printUsage()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gcode-optimizer finishing.cnc 1.0 finishing-opt.cnc")
	fmt.Println("  gcode-optimizer --strategy=conservative finishing.cnc 1.0 output.cnc")
	fmt.Println("  gcode-optimizer --force finishing.cnc 1.0 output.cnc")
	fmt.Println()
	fmt.Println("For more information: https://github.com/chrisns/snapmaker-cnc-finisher")
}
