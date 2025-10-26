# Internal Package Contracts

**Note**: This is a CLI tool with no external REST/GraphQL API. This directory documents internal Go package interfaces for development and testing purposes.

## Package Architecture

```
cmd/gcode-optimizer (main package)
         │
         ├─► internal/parser     (GCode parsing, modal state)
         ├─► internal/optimizer   (Move classification, filtering)
         ├─► internal/writer      (GCode output generation)
         └─► internal/progress    (Progress reporting, statistics)
```

---

## internal/parser

### Purpose
Parse Snapmaker Luban GCode files, extract header metadata, track modal state throughout file processing.

### Public Interface

```go
package parser

import (
    "io"
    "github.com/256dpi/gcode"
)

// HeaderMetadata contains parsed Snapmaker Luban header information
type HeaderMetadata struct {
    FileType        string
    ToolHead        string
    Machine         string
    TotalLines      int64
    EstimatedTimeSec float64
    IsRotate        bool

    MaxX, MinX float64
    MaxY, MinY float64
    MaxZ, MinZ float64
    MaxB, MinB float64

    WorkSpeed int
    JogSpeed  int
}

// ModalState tracks current position and parameters during file processing
type ModalState struct {
    X float64
    Y float64
    Z float64
    B float64
    F float64
}

// Parser handles GCode file parsing and modal state management
type Parser struct {
    file     *gcode.File
    header   HeaderMetadata
    state    ModalState
    warnings []string
}

// NewParser creates a parser from an io.Reader
func NewParser(r io.Reader) (*Parser, error)

// Header returns the parsed header metadata
func (p *Parser) Header() HeaderMetadata

// File returns the underlying gcode.File
func (p *Parser) File() *gcode.File

// State returns the current modal state
func (p *Parser) State() ModalState

// UpdateState updates modal state from a GCode line
func (p *Parser) UpdateState(line gcode.Line)

// ResetState reinitializes modal state from header metadata
func (p *Parser) ResetState()

// ScanMinZ finds the minimum Z value in all G1 commands
func (p *Parser) ScanMinZ() (float64, error)

// Warnings returns any parsing warnings (e.g., missing header fields)
func (p *Parser) Warnings() []string
```

### Contract Tests

```go
func TestNewParser(t *testing.T) {
    // Valid Snapmaker Luban file
    // Missing header fields (should warn but proceed)
    // Malformed file (should error)
}

func TestHeaderParsing(t *testing.T) {
    // Parse all header fields correctly
    // Handle missing optional fields
    // Detect 4-axis via is_rotate flag
}

func TestModalStateTracking(t *testing.T) {
    // Update X, Y, Z coordinates
    // Preserve modal values when not specified
    // Initialize Z from max_z header value
}

func TestScanMinZ(t *testing.T) {
    // Find deepest Z value in G1 commands
    // Handle files with no G1 commands
    // Handle files with only G0 (rapid moves)
}
```

---

## internal/optimizer

### Purpose
Classify moves, apply optimization strategies, calculate intersection points for move splitting.

### Public Interface

```go
package optimizer

import "github.com/256dpi/gcode"

// MoveClassification categorizes a G1 move relative to depth threshold
type MoveClassification int

const (
    Shallow MoveClassification = iota
    Deep
    CrossingEnter
    CrossingLeave
    NonCutting
)

// OptimizationStrategy defines how crossing moves are handled
type OptimizationStrategy int

const (
    Conservative OptimizationStrategy = iota  // Preserve crossing moves
    Aggressive                                // Split crossing moves
)

// IntersectionPoint represents where a move crosses the threshold
type IntersectionPoint struct {
    X, Y, Z float64
    T       float64  // Parametric parameter
}

// Optimizer applies depth-based filtering to GCode moves
type Optimizer struct {
    threshold float64
    strategy  OptimizationStrategy
}

// NewOptimizer creates an optimizer with specified threshold and strategy
func NewOptimizer(minZ, allowance float64, strategy OptimizationStrategy) *Optimizer

// Threshold returns the calculated depth threshold
func (o *Optimizer) Threshold() float64

// ClassifyMove categorizes a G1 move
func (o *Optimizer) ClassifyMove(startZ, endZ float64) MoveClassification

// CalculateIntersection finds where a move crosses the threshold
func (o *Optimizer) CalculateIntersection(startX, startY, startZ, endX, endY, endZ float64) (IntersectionPoint, error)

// ShouldPreserve determines if a line should be included in output
func (o *Optimizer) ShouldPreserve(classification MoveClassification) bool

// SplitMove generates two GCode lines from a crossing move
// Returns (moveToIntersection, moveFromIntersection)
func (o *Optimizer) SplitMove(line gcode.Line, intersection IntersectionPoint, classification MoveClassification) (gcode.Line, gcode.Line, error)
```

### Contract Tests

```go
func TestClassifyMove(t *testing.T) {
    // Table-driven tests for all classification types
    // Boundary conditions (exactly at threshold)
    // Edge cases (start == end)
}

func TestCalculateIntersection(t *testing.T) {
    // Parametric interpolation accuracy
    // Edge cases (horizontal move, vertical move)
    // Division by zero handling
    // Out-of-range t values
}

func TestOptimizationStrategies(t *testing.T) {
    // Conservative: preserve crossing moves
    // Aggressive: split crossing moves
    // Consistent behavior for Shallow/Deep
}

func TestSplitMove(t *testing.T) {
    // Correct G1 commands generated
    // Feed rate preserved
    // Coordinate precision (3-4 decimals)
    // Both enter and leave directions
}
```

---

## internal/writer

### Purpose
Write optimized GCode to output file, preserve header and formatting.

### Public Interface

```go
package writer

import (
    "io"
    "github.com/256dpi/gcode"
)

// Writer handles GCode output with format preservation
type Writer struct {
    w io.Writer
}

// NewWriter creates a writer for the given io.Writer
func NewWriter(w io.Writer) *Writer

// WriteHeader writes header lines verbatim
func (w *Writer) WriteHeader(header []string) error

// WriteLine writes a single GCode line
func (w *Writer) WriteLine(line gcode.Line) error

// WriteFile writes an entire gcode.File
func (w *Writer) WriteFile(file *gcode.File) error
```

### Contract Tests

```go
func TestWriteHeader(t *testing.T) {
    // Preserve header exactly as-is
    // Handle multiple header lines
    // Handle empty header
}

func TestWriteLine(t *testing.T) {
    // Format G1 commands correctly
    // Preserve comments
    // Handle lines with multiple codes
}

func TestFormatPreservation(t *testing.T) {
    // Round-trip: parse → write → parse should be identical
    // Coordinate precision maintained
    // Comment formatting preserved
}
```

---

## internal/progress

### Purpose
Report progress during processing, calculate statistics, display results.

### Public Interface

```go
package progress

import "time"

// Reporter tracks and displays optimization progress
type Reporter struct {
    totalLines     int64
    processedLines int64
    startTime      time.Time
    lastUpdate     time.Time
}

// NewReporter creates a progress reporter
func NewReporter(totalLines int64) *Reporter

// Update updates progress and displays if criteria met (2s OR 10k lines)
func (r *Reporter) Update(linesProcessed int64)

// Finish displays final progress
func (r *Reporter) Finish()

// OptimizationResult contains statistics from optimization
type OptimizationResult struct {
    TotalInputLines      int64
    InputFileSizeBytes   int64
    LinesProcessed       int64
    LinesRemoved         int64
    LinesPreserved       int64
    LinesSplit           int64
    MinZ                 float64
    Threshold            float64
    TotalOutputLines     int64
    OutputFileSizeBytes  int64
    ReductionPercent     float64
    EstimatedTimeSavingsSec float64
    ProcessingDurationSec   float64
    LinesPerSecond          float64
}

// ResultFormatter formats and displays OptimizationResult
type ResultFormatter struct{}

// Format returns formatted result string
func (rf *ResultFormatter) Format(result OptimizationResult) string

// Display prints formatted result to stdout
func (rf *ResultFormatter) Display(result OptimizationResult)
```

### Contract Tests

```go
func TestProgressReporter(t *testing.T) {
    // Update frequency (2s OR 10k lines)
    // ETA calculation accuracy
    // Handle unknown total lines (streaming mode)
}

func TestResultFormatting(t *testing.T) {
    // All metrics displayed
    // Percentage calculations correct
    // Time formatting readable (e.g., "142.5 minutes")
}

func TestStatisticsCalculation(t *testing.T) {
    // Reduction percentage
    // Lines per second throughput
    // Time savings estimation
}
```

---

## Integration Contract

### File Processing Pipeline

```go
// High-level workflow (implemented in cmd/gcode-optimizer/main.go)
func OptimizeGCodeFile(inputPath, outputPath string, allowance float64, strategy string, force bool) error {
    // 1. Parse input file
    parser := parser.NewParser(inputFile)

    // 2. Scan for min_z
    minZ := parser.ScanMinZ()

    // 3. Create optimizer
    optimizer := optimizer.NewOptimizer(minZ, allowance, strategy)

    // 4. Initialize progress reporter
    reporter := progress.NewReporter(parser.Header().TotalLines)

    // 5. Process each line
    for _, line := range parser.File().Lines {
        parser.UpdateState(line)
        classification := optimizer.ClassifyMove(...)

        if optimizer.ShouldPreserve(classification) {
            writer.WriteLine(line)
        } else if classification == CrossingEnter || classification == CrossingLeave {
            if strategy == Aggressive {
                line1, line2 := optimizer.SplitMove(...)
                writer.WriteLine(line1)
                writer.WriteLine(line2)
            } else {
                writer.WriteLine(line)
            }
        }

        reporter.Update(lineNumber)
    }

    // 6. Display results
    formatter := progress.ResultFormatter{}
    formatter.Display(result)

    return nil
}
```

### End-to-End Contract Test

```go
func TestOptimizationPipeline(t *testing.T) {
    tests := []struct {
        name       string
        inputFile  string  // Path to test fixture
        allowance  float64
        strategy   string
        wantLines  int64   // Expected output line count
        wantReduction float64  // Expected reduction %
    }{
        {
            name:       "freya.cnc subset with 1mm allowance",
            inputFile:  "fixtures/freya-subset.cnc",
            allowance:  1.0,
            strategy:   "aggressive",
            wantLines:  5000,  // Example
            wantReduction: 38.0,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Run full optimization
            // Verify output line count
            // Verify reduction percentage
            // Verify output file is valid GCode
        })
    }
}
```

---

## Error Handling Contract

All packages follow consistent error handling:

1. **Return errors, don't panic** (unless truly unrecoverable)
2. **Wrap errors with context**: Use `fmt.Errorf("context: %w", err)`
3. **User-facing errors**: Clear, actionable messages
4. **Internal errors**: Include debugging info (line numbers, values)

### Example Error Messages

```go
// User-facing (from main)
fmt.Fprintf(os.Stderr, "Error: Input file not found: %s\n", inputPath)
fmt.Fprintf(os.Stderr, "Error: Invalid allowance value: %s\n", allowanceArg)

// Internal (from packages)
return fmt.Errorf("failed to parse header field %q: %w", field, err)
return fmt.Errorf("intersection calculation failed (t=%f out of range): %w", t, err)
```

---

## Testing Strategy

### Unit Tests (`tests/unit/`)
- **parser_test.go**: Header parsing, modal state tracking
- **optimizer_test.go**: Move classification, threshold calculations
- **move_test.go**: Intersection calculations, split accuracy
- **writer_test.go**: GCode formatting, round-trip preservation
- **progress_test.go**: ETA calculations, statistics

### Integration Tests (`tests/integration/`)
- **cli_test.go**: End-to-end CLI workflows
  - Valid inputs → successful optimization
  - Invalid inputs → appropriate error messages
  - File overwrite confirmation prompts
  - Strategy flag behavior
- **fixtures/**: Sample GCode files
  - `freya-subset.cnc`: Subset of freya.cnc for fast tests
  - `simple-3axis.cnc`: Basic 3-axis test case
  - `4axis-rotary.cnc`: 4-axis test with B rotation
  - `threshold-crossing.cnc`: Moves that cross threshold

### Contract Tests
- Verify package interfaces match this specification
- Ensure consistent error handling
- Validate integration points between packages

---

## Version Compatibility

- **Go Version**: 1.21+
- **github.com/256dpi/gcode**: v0.3.0
- **Breaking Changes**: None planned for v1.x
- **Deprecations**: TBD based on usage feedback
