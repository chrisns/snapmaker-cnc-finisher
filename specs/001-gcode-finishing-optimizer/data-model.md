# Data Model: GCode Finishing Pass Optimizer

**Feature**: 001-gcode-finishing-optimizer
**Date**: 2025-10-26
**Phase**: Phase 1 - Design

This document defines the core entities, their attributes, relationships, validation rules, and state transitions for the GCode Finishing Pass Optimizer.

---

## Entity Diagram

```
┌─────────────────────┐
│   GCodeFile         │
├─────────────────────┤
│ - Path              │
│ - Header            │
│ - Lines []Line      │
│ - Metadata          │
└──────┬──────────────┘
       │
       │ contains *
       ▼
┌─────────────────────┐         ┌──────────────────┐
│   Line              │────────>│  Command         │
├─────────────────────┤ has *   ├──────────────────┤
│ - Number            │         │ - Type           │
│ - Raw               │         │ - Letter         │
│ - Commands []Cmd    │         │ - Value          │
│ - Comment           │         │ - Comment        │
│ - IsFiltered        │         └──────────────────┘
└─────────────────────┘
       │
       │ updates
       ▼
┌─────────────────────┐
│   Statistics        │
├─────────────────────┤
│ - TotalLines        │
│ - ProcessedLines    │
│ - RemovedLines      │
│ - BytesIn           │
│ - BytesOut          │
│ - TimeSaved         │
└─────────────────────┘
```

---

## Entity 1: GCodeFile

Represents a GCode CNC file (input or output).

### Attributes

| Name | Type | Description | Constraints | Default |
|------|------|-------------|-------------|---------|
| `Path` | `string` | Absolute file path | Non-empty, valid path | - |
| `Header` | `FileHeader` | Parsed header metadata | - | - |
| `Metadata` | `FileMetadata` | File statistics | - | Auto-calculated |

### Relationships

- **Contains** 0..* `Line` entities (streaming, not loaded into memory)

### Validation Rules

- `Path` MUST exist (for input files) or be writable (for output files)
- `Path` MUST have `.cnc`, `.gcode`, or `.nc` extension (warning if not)
- File size MUST be < 10GB (practical limit per SC-006)

### State Transitions

```
[Unopened] --Open()--> [Reading] --Parse()--> [Parsed] --Close()--> [Closed]
                             |
                             +--Error()--> [Error]
```

---

## Entity 2: FileHeader

Snapmaker Luban GCode header metadata.

### Attributes

| Name | Type | Description | Constraints | Default |
|------|------|-------------|-------------|---------|
| `MinZ` | `*float64` | Minimum Z-axis value (mm) | Nullable | nil |
| `MaxZ` | `*float64` | Maximum Z-axis value (mm) | Nullable | nil |
| `AxisConfig` | `AxisConfig` | 3-axis or 4-axis CNC | Enum: `Axis3`, `Axis4` | `Axis3` |
| `ZReference` | `ZReferenceMethod` | How Z=0 was determined | Enum (see below) | `ZRefMaterialSurface` |
| `Comments` | `[]string` | Header comment lines | - | Empty slice |

### Enumerations

**AxisConfig**:
- `Axis3` - X, Y, Z axes
- `Axis4` - X, Y, Z, B axes (rotary)

**ZReferenceMethod** (per FR-004):
- `ZRefHeaderMetadata` - Detected from header `min_z`/`max_z` fields
- `ZRefMachineOrigin` - Fallback to machine work origin interpretation
- `ZRefMaterialSurface` - Final fallback (Z=0 = top of material)

### Validation Rules

- If `MinZ` and `MaxZ` both present, `MinZ` MUST be ≤ `MaxZ`
- `AxisConfig` MUST be detected from presence of "B" commands in file

### Parsing Logic

```
1. Scan first 50 lines for comments starting with ";"
2. Extract ";MIN_Z:<value>" and ";MAX_Z:<value>" if present
3. Detect axis config from first occurrence of "B<value>" command
4. Set ZReference based on detection success:
   - If MinZ/MaxZ found → ZRefHeaderMetadata
   - Else if work origin heuristics apply → ZRefMachineOrigin
   - Else → ZRefMaterialSurface (default)
5. Log ZReference method to console (per FR-004)
```

---

## Entity 3: Line

Represents a single line from the GCode file.

### Attributes

| Name | Type | Description | Constraints | Default |
|------|------|-------------|-------------|---------|
| `Number` | `int` | Line number (1-indexed) | > 0 | - |
| `Raw` | `string` | Original line text | - | - |
| `Commands` | `[]Command` | Parsed GCode commands | 0..* | Empty |
| `Comment` | `string` | Inline comment (if any) | - | "" |
| `IsFiltered` | `bool` | Removed by optimizer | - | false |

### Relationships

- **Has** 0..* `Command` entities (parsed from `Raw`)

### Validation Rules

- `Raw` length MUST be < 2048 characters (GCode convention)
- `Number` MUST increment sequentially (if validated)

### State Transitions

```
[Unparsed] --Parse()--> [Parsed] --Filter()--> [Filtered/Retained]
                             |
                             +--Error()--> [ParseError]
```

### Lifecycle

Lines are **not** stored in memory. They are:
1. Read via `bufio.Scanner`
2. Parsed into `Command` list
3. Evaluated for filtering
4. Written to output (if retained)
5. Discarded

Only `Statistics` are retained in memory.

---

## Entity 4: Command

Represents a single GCode command (e.g., "G1", "X10.5", "F1500").

### Attributes

| Name | Type | Description | Constraints | Default |
|------|------|-------------|-------------|---------|
| `Letter` | `string` | Command letter (G, M, X, Y, Z, B, F, etc.) | Single uppercase letter | - |
| `Value` | `float64` | Numeric parameter | - | 0.0 |
| `Comment` | `string` | Inline comment (empty if none) | - | "" |

### Validation Rules

- `Letter` MUST be single uppercase letter (A-Z)
- Common letters: G, M (machine codes), X, Y, Z, B (axes), F (feed rate), S (spindle speed)

### Command Types

Derived from `Letter` + `Value` combination:

| Type | Letter | Example | Description |
|------|--------|---------|-------------|
| **Rapid Move** | G | `G0` | Non-cutting rapid positioning |
| **Linear Move** | G | `G1` | Cutting move with feed rate |
| **Machine Code** | M | `M3`, `M5` | Spindle on/off, coolant, etc. |
| **Coordinate** | X/Y/Z/B | `X10.5` | Axis position |
| **Feed Rate** | F | `F1500` | Cutting speed (mm/min) |
| **Comment** | ; | `;comment` | Not a command (handled separately) |

### Filtering Logic

A `Line` is **removed** if:
1. It contains a `G1` command (cutting move), AND
2. Its `Z` coordinate value is **shallower** than `allowance` threshold, AND
3. Multi-axis strategy rules apply (see Strategy enum below)

A `Line` is **preserved** if:
- It contains `G0` (rapid move) - per edge case
- It contains `M` codes (machine commands) - per FR-006
- It is a comment line - per FR-006
- Its `Z` depth exceeds `allowance` (requires material removal)

---

## Entity 5: Statistics

Tracks optimization results.

### Attributes

| Name | Type | Description | Constraints | Default |
|------|------|-------------|-------------|---------|
| `TotalLines` | `int` | Lines in input file | ≥ 0 | 0 |
| `ProcessedLines` | `int` | Lines processed so far | ≤ TotalLines | 0 |
| `RemovedLines` | `int` | Lines filtered out | ≤ TotalLines | 0 |
| `RetainedLines` | `int` | Lines written to output | = TotalLines - RemovedLines | 0 |
| `BytesIn` | `int64` | Input file size (bytes) | ≥ 0 | 0 |
| `BytesOut` | `int64` | Output file size (bytes) | ≥ 0 | 0 |
| `EstimatedTimeSaved` | `time.Duration` | Calculated time savings | ≥ 0 | 0 |
| `StartTime` | `time.Time` | Processing start | - | - |
| `EndTime` | `time.Time` | Processing end | ≥ StartTime | - |

### Derived Fields

| Name | Type | Formula |
|------|------|---------|
| `PercentageRemoved` | `float64` | `(RemovedLines / TotalLines) * 100` |
| `PercentageReduction` | `float64` | `((BytesIn - BytesOut) / BytesIn) * 100` |
| `ProcessingDuration` | `time.Duration` | `EndTime - StartTime` |

### Validation Rules

- `RemovedLines + RetainedLines` MUST equal `TotalLines`
- `BytesOut` SHOULD be ≤ `BytesIn` (optimizer never increases size)
- `EstimatedTimeSaved` calculation per FR-010: sum of `distance ÷ feedRate` for each removed `G1` move

### Time Savings Calculation

For each **removed** `G1` line:
```
distance = sqrt((ΔX)² + (ΔY)² + (ΔZ)² + (ΔB)²)  # Euclidean distance
feedRate = F value (mm/min) from command or previous line
timeSaved = distance / feedRate  # in minutes
```

Accumulate across all removed lines.

**Edge Cases**:
- If `F` not specified on line, use last known feed rate from previous lines
- If no feed rate ever specified (malformed file), assume 1000 mm/min (log warning)

---

## Entity 6: FilterStrategy

Enum defining how multi-axis moves are handled when only Z exceeds allowance.

### Enum Values

| Value | Description | Behavior |
|-------|-------------|----------|
| `StrategySafe` | **Default** - Conservative | Preserve entire move if Z-axis component exceeds threshold |
| `StrategyAllAxes` | All-axes check | Preserve only if **all** axes indicate finishing work (beyond allowance) |
| `StrategySplit` | Decomposition | Attempt to split multi-axis move into single-axis commands |
| `StrategyAggressive` | Maximum removal | Remove entire move if Z is shallow, even if X/Y/B exceed threshold |

### Mapping to CLI Flag

Command-line `--strategy` flag:
- `"safe"` → `StrategySafe` (default)
- `"all-axes"` → `StrategyAllAxes`
- `"split"` → `StrategySplit`
- `"aggressive"` → `StrategyAggressive`

### Validation Rules

- Invalid strategy string MUST return error: "Invalid strategy '<value>'. Must be one of: safe, all-axes, split, aggressive"

---

## Entity 7: ProgressUpdate

Real-time processing status (per SC-005).

### Attributes

| Name | Type | Description |
|------|------|-------------|
| `LinesProcessed` | `int` | Current line count |
| `PercentComplete` | `float64` | `(LinesProcessed / TotalLines) * 100` |
| `LinesRemoved` | `int` | Running total of filtered lines |
| `ElapsedTime` | `time.Duration` | Time since start |
| `EstimatedTimeRemaining` | `time.Duration` | Projected time to completion |

### Update Frequency

Per SC-005: "Every 10,000 lines processed OR every 2 seconds, whichever is more frequent"

```go
if lineCount % 10000 == 0 || time.Since(lastUpdate) >= 2*time.Second {
    reportProgress(...)
}
```

---

## Relationships Summary

```
GCodeFile (1) --contains--> (*) Line
Line (1) --has--> (*) Command
Line (*) --updates--> (1) Statistics
FilterStrategy (1) --applies to--> (*) Line
ProgressUpdate (*) --aggregates--> (1) Statistics
```

---

## Data Flow

```
1. [Input] GCodeFile opened → FileHeader parsed
2. [Streaming] For each Line:
   a. Parse into Commands
   b. Extract Z coordinate (if G1 command)
   c. Compare Z vs. Allowance + apply FilterStrategy
   d. If retained: write to Output
   e. If filtered: update Statistics.RemovedLines, accumulate TimeSaved
   f. Update Statistics.ProcessedLines
   g. Emit ProgressUpdate (if frequency threshold met)
3. [Finalize] Close Output, calculate final Statistics, display summary
```

---

## Persistence

**File System Only**:
- Input GCode file: read-only
- Output GCode file: write-only
- No database, no configuration files

**In-Memory State**:
- `Statistics` - single instance
- `FileHeader` - single instance
- Current `Line` and `Commands` - transient (garbage collected after processing)

---

## Validation Summary

| Entity | Required Fields | Constraints | Error Handling |
|--------|-----------------|-------------|----------------|
| GCodeFile | Path | File exists (input), writable (output) | Return error with clear message |
| FileHeader | AxisConfig | Valid enum | Default to Axis3, log warning |
| Line | Number, Raw | Number > 0, Raw < 2048 chars | Skip malformed lines, log warning |
| Command | Letter, Value | Letter in A-Z, Value is float | Parse error → skip line, log warning |
| Statistics | All fields | Arithmetic consistency | Internal validation, panic if violated |
| FilterStrategy | - | Valid enum value | Return error if invalid CLI flag |

---

## Example Instances

### Example 1: Simple G1 Command Line

**Raw Line**: `G1 X10.5 Y20.3 Z-0.5 F1500`

**Parsed Structure**:
```go
Line{
    Number: 42,
    Raw: "G1 X10.5 Y20.3 Z-0.5 F1500",
    Commands: []Command{
        {Letter: "G", Value: 1.0, Comment: ""},
        {Letter: "X", Value: 10.5, Comment: ""},
        {Letter: "Y", Value: 20.3, Comment: ""},
        {Letter: "Z", Value: -0.5, Comment: ""},
        {Letter: "F", Value: 1500.0, Comment: ""},
    },
    Comment: "",
    IsFiltered: false,  // Depends on allowance threshold
}
```

**Filtering Decision** (assuming allowance = 1.0mm):
- Z = -0.5mm (absolute value 0.5mm)
- 0.5mm < 1.0mm → **Line is FILTERED** (shallow cut already handled by rough pass)

### Example 2: Comment Line

**Raw Line**: `; Set spindle speed to 12000 RPM`

**Parsed Structure**:
```go
Line{
    Number: 10,
    Raw: "; Set spindle speed to 12000 RPM",
    Commands: []Command{},
    Comment: " Set spindle speed to 12000 RPM",
    IsFiltered: false,  // Always retained per FR-006
}
```

### Example 3: Rapid Move

**Raw Line**: `G0 Z5.0`

**Parsed Structure**:
```go
Line{
    Number: 15,
    Raw: "G0 Z5.0",
    Commands: []Command{
        {Letter: "G", Value: 0.0, Comment: ""},
        {Letter: "Z", Value: 5.0, Comment: ""},
    },
    Comment: "",
    IsFiltered: false,  // G0 always retained (edge case rule)
}
```

---

## State Machine: Line Processing

```
          ┌─────────┐
          │  Start  │
          └────┬────┘
               │
               ▼
      ┌────────────────┐
      │  Read Raw Line │
      └────┬───────────┘
           │
           ▼
    ┌──────────────┐
    │  Parse Line  │
    └──┬───────────┘
       │
       ├─ Comment? ────> [Retain & Write]
       │
       ├─ G0 Rapid? ──> [Retain & Write]
       │
       ├─ M-Code? ────> [Retain & Write]
       │
       ├─ G1 + Z? ────> ┌─────────────────┐
       │                 │ Compare Z vs.   │
       │                 │ Allowance +     │
       │                 │ Apply Strategy  │
       │                 └──┬──────────────┘
       │                    │
       │                    ├─ Z shallow? ──> [Filter & Skip]
       │                    │
       │                    └─ Z deep? ────> [Retain & Write]
       │
       └─ Other ──────────> [Retain & Write]
```

---

## Test Data Requirements

Per Principle III (TDD), create test fixtures in `tests/testdata/`:

1. **finishing_3axis.cnc** - Standard 3-axis file with mixed depths
2. **finishing_4axis.cnc** - 4-axis file with B-axis rotary commands
3. **malformed_header.cnc** - Missing Snapmaker header (triggers warning)
4. **large_file.cnc** - 1M+ lines to test streaming and progress updates
5. **all_shallow.cnc** - All cuts < allowance (tests maximum filtering)
6. **all_deep.cnc** - All cuts > allowance (tests minimal filtering)
7. **no_feed_rate.cnc** - Missing F values (tests fallback logic)

---

**Generated By**: Claude Code (Sonnet 4.5)
**References**: [spec.md](./spec.md), [research.md](./research.md)
