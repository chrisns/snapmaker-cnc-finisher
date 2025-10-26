# Research: GCode Finishing Pass Optimizer

**Phase**: 0 (Outline & Research)
**Date**: 2025-10-26
**Purpose**: Resolve technical unknowns and establish best practices for implementation

## Research Questions

### Q1: GCode Parsing Library Selection

**Decision**: Use `github.com/256dpi/gcode` v0.3.0

**Rationale**:
- **Official Go Package**: Well-documented at pkg.go.dev with clear API
- **Pure Go**: No CGO dependencies, supports static binary distribution (Constitution Principle II)
- **Proven Functionality**: Provides exactly what we need:
  - `ParseFile(io.Reader)` for file parsing
  - `WriteFile(io.Writer, *File)` for output
  - Structured types: `File`, `Line`, `GCode` with `Letter`, `Value`, `Comment` fields
  - Modal programming support (coordinates persist across lines)
- **MIT Licensed**: Compatible with open-source project
- **Mature**: Published May 2021, stable API

**Alternatives Considered**:
- **Custom parser**: Rejected due to complexity and time investment; GCode syntax is well-defined but has edge cases
- **github.com/mauroalderete/gcode-core**: More recent but less documented, unclear API stability

**Implementation Notes**:
- Library's `GCode` struct uses `float64` for `Value` - perfect for coordinate calculations
- `ParseLine(string)` available for incremental processing if memory constraints require streaming
- No built-in filtering - we implement optimization logic, library handles I/O

**Sources**:
- https://pkg.go.dev/github.com/256dpi/gcode
- https://github.com/256dpi/gcode

---

### Q2: Modal State Management Pattern

**Decision**: Use struct-based state machine with explicit initialization from header metadata

**Rationale**:
- **GCode Specification**: Coordinates and parameters not specified in a command persist from previous commands (modal programming)
- **Initialization Strategy**: Read header metadata (max_z, min_z) during parse, initialize modal state before processing commands
- **Type Safety**: Go struct with `float64` fields for X, Y, Z, B coordinates and F (feed rate) parameter

**Pattern**:
```go
type ModalState struct {
    X float64  // Current X position
    Y float64  // Current Y position
    Z float64  // Current Z position (initialized from max_z header or 0)
    B float64  // Current B rotation (4-axis support)
    F float64  // Current feed rate
}

// Update from GCode line
func (m *ModalState) Update(line gcode.Line) {
    for _, code := range line.Codes {
        switch code.Letter {
        case "X": m.X = code.Value
        case "Y": m.Y = code.Value
        case "Z": m.Z = code.Value
        case "B": m.B = code.Value
        case "F": m.F = code.Value
        }
    }
}
```

**Alternatives Considered**:
- **Map-based state**: Rejected due to lack of type safety and additional lookup overhead
- **Require explicit coordinates**: Rejected because real GCode files use modal programming heavily (see freya.cnc analysis)

**Sources**:
- Spec clarification Q&A #2 (modal state tracking)
- Spec clarification Q&A #5 (initial state from header)
- Analysis of freya.cnc showing G1 commands without explicit Z coordinates

---

### Q3: Parametric Linear Interpolation for Move Splitting

**Decision**: Standard parametric line equation with explicit threshold intersection calculation

**Rationale**:
- **Mathematical Foundation**: Parametric form: `P(t) = P_start + t(P_end - P_start)` where `0 ≤ t ≤ 1`
- **Threshold Intersection**: Solve for `t` where `Z(t) = threshold`:
  ```
  threshold = Z_start + t(Z_end - Z_start)
  t = (threshold - Z_start) / (Z_end - Z_start)
  ```
- **Intersection Point**: `(X₀, Y₀, Z₀)` where:
  - `X₀ = X_start + t(X_end - X_start)`
  - `Y₀ = Y_start + t(Y_end - Y_start)`
  - `Z₀ = threshold` (exact, no floating-point drift)

**Edge Cases**:
- **Division by zero**: If `Z_end == Z_start`, move doesn't cross threshold vertically (check endpoints only)
- **Out of range t**: If `t < 0` or `t > 1`, move doesn't intersect within segment (shouldn't occur with proper threshold classification)
- **Precision**: Maintain 3-4 decimal places per spec (use `fmt.Sprintf("%.4f", value)`)

**Alternatives Considered**:
- **Approximate splitting**: Rejected due to potential toolpath errors
- **Conservative whole-move preservation**: Implemented as separate strategy option

**Sources**:
- Spec FR-013 (parametric linear interpolation requirements)
- Standard computational geometry references

---

### Q4: CLI Argument Parsing Best Practices (Go 1.21)

**Decision**: Use stdlib `flag` package with positional arguments via `flag.Args()`

**Rationale**:
- **Constitution Compliance**: Minimal dependencies, prefer stdlib (Principle V)
- **Sufficient Functionality**: Supports named flags (`--force`, `--strategy=aggressive`) and captures remaining arguments
- **Go 1.21 Compatibility**: Stable API, no changes needed for future Go versions
- **POSIX-style**: Flags before positional arguments: `./tool --force --strategy=conservative input.cnc 1.0 output.cnc`

**Pattern**:
```go
var (
    force = flag.Bool("force", false, "Overwrite output without confirmation")
    strategy = flag.String("strategy", "aggressive", "Optimization strategy (conservative|aggressive)")
)

func main() {
    flag.Parse()

    args := flag.Args()
    if len(args) != 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s [flags] <input.cnc> <allowance> <output.cnc>\n", os.Args[0])
        flag.PrintDefaults()
        os.Exit(1)
    }

    inputPath := args[0]
    allowance := parseFloat(args[1])
    outputPath := args[2]

    // Validate strategy
    if *strategy != "conservative" && *strategy != "aggressive" {
        fmt.Fprintf(os.Stderr, "Invalid strategy '%s'. Valid options are: conservative, aggressive\n", *strategy)
        os.Exit(1)
    }
}
```

**Alternatives Considered**:
- **Cobra/Viper**: Rejected due to heavy dependencies contradicting constitution
- **spf13/pflag**: Rejected as stdlib `flag` is sufficient for our use case
- **Positional-first parsing**: Rejected because Go's `flag` package expects flags before positional args (standard POSIX pattern)

**Sources**:
- https://pkg.go.dev/flag
- https://gobyexample.com/command-line-flags
- Go 1.21 CLI best practices articles (2025 search results)

---

### Q5: Table-Driven Testing Strategy (Go 1.21/1.22+)

**Decision**: Use subtests with table-driven pattern, leverage Go 1.22 loop variable scoping

**Rationale**:
- **Go Community Standard**: Table-driven tests are idiomatic Go (see official Go Wiki)
- **Test Coverage**: Easily add edge cases, regression tests, and boundary conditions
- **Parallel Execution**: Subtests can run in parallel with `t.Parallel()`
- **Go 1.22 Update**: Loop variable scoping fix eliminates need for `tc := tc` pattern

**Pattern**:
```go
func TestMoveFiltering(t *testing.T) {
    tests := []struct {
        name      string
        startZ    float64
        endZ      float64
        threshold float64
        want      FilterAction
    }{
        {"shallow move above threshold", -5.0, -4.0, -9.0, Remove},
        {"deep move below threshold", -11.0, -10.0, -9.0, Preserve},
        {"crossing move entering deep", -8.0, -10.0, -9.0, Split},
        {"crossing move leaving deep", -10.0, -8.0, -9.0, Split},
        {"move exactly at threshold start", -9.0, -10.0, -9.0, Preserve},
        {"move exactly at threshold end", -10.0, -9.0, -9.0, Preserve},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Go 1.22+: no need for tt := tt
            got := ClassifyMove(tt.startZ, tt.endZ, tt.threshold)
            if got != tt.want {
                t.Errorf("ClassifyMove(%v, %v, %v) = %v, want %v",
                    tt.startZ, tt.endZ, tt.threshold, got, tt.want)
            }
        })
    }
}
```

**Best Practices Applied**:
- **Descriptive names**: Test case `name` field for clear failure messages
- **Independent tests**: Each case runs in isolation
- **Edge case coverage**: Include boundary conditions, threshold equality, crossing directions
- **Subtests since Go 1.7**: Better granularity and reporting with `go test -v`

**Alternatives Considered**:
- **Separate test functions**: Rejected due to code duplication
- **Map-based tables**: Rejected because iteration order is undefined (slice order is stable)

**Sources**:
- https://go.dev/wiki/TableDrivenTests
- https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
- Go 1.22 release notes (loop variable scoping change)

---

### Q6: Progress Reporting and ETA Calculation

**Decision**: Use elapsed time ratio with periodic console updates

**Rationale**:
- **User Requirement**: Spec FR-009 requires progress updates showing lines processed and estimated completion
- **ETA Formula**: `ETA = (elapsed_time / lines_processed) × (total_lines - lines_processed)`
- **Update Frequency**: Spec SC-005 requires updates every 10,000 lines OR every 2 seconds (whichever is more frequent)
- **Streaming Consideration**: If total line count unknown, display lines processed without ETA

**Pattern**:
```go
type ProgressReporter struct {
    totalLines    int64
    processedLines int64
    startTime     time.Time
    lastUpdate    time.Time
}

func (p *ProgressReporter) Update(linesProcessed int64) {
    p.processedLines = linesProcessed
    now := time.Now()

    // Update every 2 seconds or 10k lines
    if now.Sub(p.lastUpdate) < 2*time.Second && linesProcessed%10000 != 0 {
        return
    }

    p.lastUpdate = now
    elapsed := now.Sub(p.startTime)

    if p.totalLines > 0 {
        remaining := p.totalLines - p.processedLines
        eta := time.Duration(float64(elapsed) / float64(p.processedLines) * float64(remaining))
        fmt.Printf("\rProcessed: %d/%d lines (%.1f%%) - ETA: %s",
            p.processedLines, p.totalLines,
            100.0*float64(p.processedLines)/float64(p.totalLines),
            eta.Round(time.Second))
    } else {
        fmt.Printf("\rProcessed: %d lines", p.processedLines)
    }
}
```

**Alternatives Considered**:
- **Third-party progress bars**: Rejected to minimize dependencies
- **Fixed update interval only**: Rejected because spec requires dual criteria (time OR line count)

**Sources**:
- Spec FR-009, SC-005 (progress requirements)
- Go stdlib `time` package patterns

---

### Q7: File I/O and Format Preservation

**Decision**: Use `gcode` library's WriteFile with original header/comment preservation

**Rationale**:
- **Format Compatibility**: Must preserve Snapmaker Luban header and metadata (spec FR-007)
- **Library Support**: `gcode.WriteFile(writer, file)` maintains line structure and comments
- **Streaming Not Required**: Memory-conscious but not mandatory streaming; Go can handle 10M line files in-memory with ~1-2GB RAM
- **Header Preservation**: Parse header during initial scan, include verbatim in output

**Implementation Notes**:
- Read entire file into `*gcode.File` structure
- Preserve header lines (starting with `;`) exactly as-is
- Filter/modify only G1 cutting moves based on strategy
- Preserve all G0 (rapid), M-codes, comments, configuration commands
- Write output with same formatting

**Memory Estimate**:
- 10M lines × ~50 bytes/line average = ~500MB raw text
- Parsed structure overhead: ~200MB
- Total: ~700MB peak memory for largest expected files (well within modern system limits)

**Alternatives Considered**:
- **Streaming line-by-line**: Rejected as premature optimization; adds complexity without proven need
- **Custom writer**: Rejected because `gcode.WriteFile` handles formatting correctly

**Sources**:
- `gcode` library documentation
- Spec FR-007 (format preservation), SC-006 (10M line handling)

---

## Summary of Technical Decisions

| Area | Decision | Key Rationale |
|------|----------|---------------|
| **Parsing** | `github.com/256dpi/gcode` v0.3.0 | Pure Go, well-documented, stable API |
| **Modal State** | Struct-based state machine | Type-safe, initialized from header metadata |
| **Move Splitting** | Parametric linear interpolation | Mathematically accurate, preserves toolpath integrity |
| **CLI Parsing** | Stdlib `flag` package | Minimal dependencies, sufficient functionality |
| **Testing** | Table-driven subtests (Go 1.22+) | Idiomatic, comprehensive coverage, parallel-safe |
| **Progress** | Time-ratio ETA with dual update criteria | Meets spec requirements (2s OR 10k lines) |
| **File I/O** | In-memory processing with `gcode.WriteFile` | Format preservation, acceptable memory footprint |

## Next Steps (Phase 1)

1. **Data Model Design** → `data-model.md`
   - Define `ModalState`, `MoveClassification`, `OptimizationStrategy` types
   - Specify `OptimizationResult` statistics structure

2. **API Contracts** → `contracts/`
   - Not applicable (CLI tool, no external API)
   - Document internal package interfaces instead

3. **Quickstart Guide** → `quickstart.md`
   - Installation instructions (download binary, go install)
   - Basic usage examples
   - Common workflows (rough + finish optimization)
