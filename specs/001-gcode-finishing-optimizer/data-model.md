# Data Model: GCode Finishing Pass Optimizer

**Phase**: 1 (Design & Contracts)
**Date**: 2025-10-26
**Purpose**: Define core data structures, relationships, and state transitions

## Core Entities

### 1. ModalState

**Description**: Tracks current position and parameter values throughout GCode file processing. In GCode, coordinates and parameters not explicitly specified in a command persist from previous commands (modal programming).

**Fields**:
```go
type ModalState struct {
    X float64  // Current X position (mm)
    Y float64  // Current Y position (mm)
    Z float64  // Current Z position/depth (mm, negative = below surface)
    B float64  // Current B rotation (degrees, 4-axis only)
    F float64  // Current feed rate (mm/min)
}
```

**Initialization Rules**:
- **Z**: Initialize from header `max_z` value if present, otherwise `0.0`
- **X, Y, B**: Initialize to `0.0`
- **F**: Initialize to `0.0` (will be set by first move command with feed rate)

**Update Behavior**:
- Only fields present in current GCode command update their values
- Absent fields retain previous values (modal behavior)
- Updates occur before move classification/processing

**Validation**:
- No explicit validation (GCode files may have any coordinate values)
- Division-by-zero checks occur during move splitting calculations

---

### 2. MoveClassification

**Description**: Classification of a G1 cutting move relative to the depth threshold, determines optimization action.

**Enum**:
```go
type MoveClassification int

const (
    Shallow MoveClassification = iota  // Both start and end above threshold → Remove
    Deep                                // Both start and end below/at threshold → Preserve
    CrossingEnter                       // Starts above, ends below/at threshold → Split (keep deep portion)
    CrossingLeave                       // Starts below/at, ends above threshold → Split (keep deep portion)
    NonCutting                          // Not a G1 command → Preserve as-is
)
```

**Classification Logic**:
```go
func ClassifyMove(startZ, endZ, threshold float64) MoveClassification {
    startDeep := startZ <= threshold
    endDeep := endZ <= threshold

    if !startDeep && !endDeep {
        return Shallow  // Both points above threshold
    }
    if startDeep && endDeep {
        return Deep  // Both points at/below threshold
    }
    if !startDeep && endDeep {
        return CrossingEnter  // Entering deep zone
    }
    return CrossingLeave  // Leaving deep zone
}
```

**State Transitions**:
- No state machine; classification is stateless per-move decision
- Classification drives optimization action (remove, preserve, split)

---

### 3. OptimizationStrategy

**Description**: Approach for handling moves that cross the depth threshold.

**Enum**:
```go
type OptimizationStrategy int

const (
    Conservative OptimizationStrategy = iota  // Preserve entire crossing moves
    Aggressive                                // Split crossing moves at threshold
)
```

**Strategy Behaviors**:

| Classification | Conservative Action | Aggressive Action |
|----------------|---------------------|-------------------|
| Shallow | Remove | Remove |
| Deep | Preserve | Preserve |
| CrossingEnter | **Preserve entire move** | **Split at threshold, keep deep portion** |
| CrossingLeave | **Preserve entire move** | **Split at threshold, keep deep portion** |
| NonCutting | Preserve | Preserve |

**Selection**:
- Controlled by `--strategy` CLI flag
- Default: `Aggressive`
- Validated at startup against allowed values

---

### 4. IntersectionPoint

**Description**: Calculated intersection of a move with the depth threshold plane.

**Structure**:
```go
type IntersectionPoint struct {
    X float64  // X coordinate at intersection
    Y float64  // Y coordinate at intersection
    Z float64  // Z coordinate (equals threshold exactly)
    T float64  // Parametric parameter (0 < t < 1)
}
```

**Calculation**: See spec.md FR-013 for authoritative parametric linear interpolation formula. Implementation follows the exact algorithm specified in requirements:

1. Calculate intersection parameter: `t = (threshold - Z_start) / (Z_end - Z_start)`
2. Calculate intersection point: `X₀ = X_start + t(X_end - X_start)`, `Y₀ = Y_start + t(Y_end - Y_start)`, `Z₀ = threshold`
3. Validate `0 < t < 1` (indicates move crosses threshold within segment)
4. Handle edge case: `deltaZ ≈ 0` (horizontal move, does not cross threshold vertically)

**Precision**:
- Coordinates formatted to 3-4 decimal places per spec FR-013
- Use `fmt.Sprintf("%.4f", value)` for output

**Edge Cases**:
- `deltaZ ≈ 0`: Move doesn't cross threshold vertically (classify as Shallow or Deep based on Z value)
- `t < 0` or `t > 1`: Shouldn't occur if classification is correct; return error for debugging

---

### 5. OptimizationResult

**Description**: Statistics and metrics from optimization process.

**Structure**:
```go
type OptimizationResult struct {
    // Input metrics
    TotalInputLines   int64
    InputFileSizeBytes int64

    // Processing metrics
    LinesProcessed    int64
    LinesRemoved      int64
    LinesPreserved    int64
    LinesSplit        int64  // Number of moves that were split (aggressive mode)

    // Depth analysis
    MinZ              float64  // Minimum Z value found in file
    Threshold         float64  // Calculated threshold (min_z + allowance)

    // Output metrics
    TotalOutputLines   int64
    OutputFileSizeBytes int64
    ReductionPercent   float64  // (LinesRemoved / TotalInputLines) * 100

    // Time savings estimate
    EstimatedTimeSavingsSec float64  // Sum of (distance / feed_rate) for removed moves

    // Performance
    ProcessingDurationSec float64
    LinesPerSecond       float64
}
```

**Calculation Rules**:
- **ReductionPercent**: `(LinesRemoved / TotalInputLines) × 100`
- **EstimatedTimeSavingsSec**: For each removed G1 move, calculate:
  ```
  distance = sqrt((Δx)² + (Δy)² + (Δz)²)
  time_sec = (distance / feed_rate_mm_per_min) * 60
  total = sum(time_sec for all removed moves)
  ```
- **LinesPerSecond**: `TotalInputLines / ProcessingDurationSec`

**Display Format** (Console Output):
```
Optimization Complete
━━━━━━━━━━━━━━━━━━━━
Depth Analysis:
  Min Z: -12.037mm
  Threshold: -11.037mm (1.0mm allowance)

Processing Summary:
  Total lines: 8,233,531
  Lines removed: 3,127,845 (38.0%)
  Lines preserved: 5,105,686
  Moves split: 1,234 (aggressive strategy)

Output:
  File size: 412.3 MB → 255.1 MB (38.1% reduction)
  Estimated time savings: 142.5 minutes

Performance:
  Processing time: 6.3 seconds
  Throughput: 1,306,908 lines/sec
```

---

### 6. HeaderMetadata

**Description**: Parsed Snapmaker Luban header information.

**Structure**:
```go
type HeaderMetadata struct {
    FileType        string   // e.g., "cnc"
    ToolHead        string   // e.g., "standardCNCToolheadForSM2"
    Machine         string   // e.g., "Snapmaker 2.0 A350"
    TotalLines      int64    // file_total_lines
    EstimatedTimeSec float64 // estimated_time(s)
    IsRotate        bool     // is_rotate (4-axis detection)

    // Bounding box
    MaxX, MinX float64
    MaxY, MinY float64
    MaxZ, MinZ float64
    MaxB, MinB float64  // Rotation axis (if is_rotate)

    // Other
    WorkSpeed int  // work_speed(mm/minute)
    JogSpeed  int  // jog_speed(mm/minute)
}
```

**Parsing Rules**:
- Header lines start with `;` (comment character in GCode)
- Format: `;key: value` or `;key(unit): value`
- Examples from freya.cnc:
  ```
  ;min_z(mm): -12.037
  ;max_z(mm): 97.5
  ;is_rotate: true
  ;file_total_lines: 8233531
  ```

**Usage**:
- **Initial Modal State**: Use `max_z` for Z initialization
- **4-Axis Detection**: Use `is_rotate` to determine if B axis tracking needed
- **Progress ETA**: Use `file_total_lines` for accurate progress percentage
- **Validation Warning**: Check for `tool_head` containing "CNC" to confirm file type

**Validation** (Spec FR-002):
- If header missing or malformed: Issue console warning, proceed if file is parseable
- Missing `max_z`: Default to `0.0` for Z initialization
- Missing `file_total_lines`: Progress display without ETA

---

## Relationships

### Entity Relationship Diagram

```
┌─────────────────┐
│ HeaderMetadata  │────┐
└─────────────────┘    │
                       │ Initializes
                       ▼
                 ┌─────────────┐
                 │ ModalState  │
                 └─────────────┘
                       │
                       │ Tracks position for
                       ▼
          ┌──────────────────────────┐
          │ Move (G1 command)        │
          └──────────────────────────┘
                       │
                       │ Classified by
                       ▼
            ┌─────────────────────────┐
            │ MoveClassification      │
            └─────────────────────────┘
                       │
            ┌──────────┴───────────┐
            │                      │
            ▼                      ▼
   ┌────────────────┐     ┌────────────────────┐
   │ Shallow/Deep   │     │ CrossingEnter/Leave│
   │ (Simple action)│     └────────────────────┘
   └────────────────┘              │
                                   │ If Aggressive
                                   ▼
                         ┌─────────────────────┐
                         │ IntersectionPoint   │
                         └─────────────────────┘
                                   │
                                   │ Generates
                                   ▼
                         ┌─────────────────────┐
                         │ Split Move Commands │
                         └─────────────────────┘
```

---

## State Transitions

### File Processing State Machine

```
┌─────────┐
│  START  │
└────┬────┘
     │
     ▼
┌─────────────────────┐
│ Parse Header        │
│ Extract metadata    │
└────┬────────────────┘
     │
     ▼
┌─────────────────────┐
│ Initialize          │
│ ModalState from     │
│ header (max_z → Z)  │
└────┬────────────────┘
     │
     ▼
┌─────────────────────┐
│ First Pass:         │
│ Scan for min_z      │──────┐ Calculate threshold
└────┬────────────────┘      │ (min_z + allowance)
     │                        │
     │◄───────────────────────┘
     ▼
┌─────────────────────┐
│ Second Pass:        │
│ Process each line   │──────┐
└────┬────────────────┘      │
     │                        │
     │  For each line:        │
     │  ┌──────────────────┐  │
     │  │ Update ModalState│  │
     │  └──────┬───────────┘  │
     │         │              │
     │         ▼              │
     │  ┌──────────────────┐  │
     │  │ Classify Move    │  │
     │  └──────┬───────────┘  │
     │         │              │
     │         ▼              │
     │  ┌──────────────────┐  │
     │  │ Apply Strategy   │  │
     │  │ (Remove/Preserve/│  │
     │  │  Split)          │  │
     │  └──────┬───────────┘  │
     │         │              │
     │         ▼              │
     │  ┌──────────────────┐  │
     │  │ Write to Output  │  │
     │  └──────────────────┘  │
     │                        │
     │◄───────────────────────┘
     │
     ▼
┌─────────────────────┐
│ Calculate Stats     │
│ Display Results     │
└────┬────────────────┘
     │
     ▼
┌─────────┐
│   END   │
└─────────┘
```

### Move Processing Decision Tree

```
                      ┌──────────────┐
                      │ GCode Line   │
                      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │ Is G1 move?  │
                      └──┬───────┬───┘
                 No      │       │      Yes
            ┌────────────┘       └────────────┐
            │                                  │
            ▼                                  ▼
    ┌──────────────┐                  ┌──────────────────┐
    │ Preserve as-is│                  │ Update ModalState│
    └───────────────┘                  └────────┬─────────┘
                                                │
                                                ▼
                                       ┌─────────────────┐
                                       │ Classify Move   │
                                       └────┬────────────┘
                          ┌──────────┬──────┴───────┬────────────┐
                          │          │              │            │
                          ▼          ▼              ▼            ▼
                   ┌──────────┐ ┌────────┐  ┌─────────────┐ ┌─────────────┐
                   │ Shallow  │ │  Deep  │  │CrossingEnter│ │CrossingLeave│
                   └────┬─────┘ └───┬────┘  └──────┬──────┘ └──────┬──────┘
                        │           │              │               │
                        ▼           ▼              │               │
                   ┌────────┐  ┌─────────┐         │               │
                   │ Remove │  │Preserve │         │               │
                   └────────┘  └─────────┘         │               │
                                                   │               │
                                    ┌──────────────┴───────────────┘
                                    │
                                    ▼
                             ┌──────────────┐
                             │ Check Strategy│
                             └──┬────────┬───┘
                       Conservative│    │Aggressive
                                   │    │
                        ┌──────────┘    └──────────┐
                        │                          │
                        ▼                          ▼
                  ┌──────────┐            ┌──────────────────┐
                  │ Preserve │            │ Calculate        │
                  │ Entire   │            │ IntersectionPoint│
                  │ Move     │            └────────┬─────────┘
                  └──────────┘                     │
                                                   ▼
                                          ┌────────────────┐
                                          │ Split Move:    │
                                          │ - G1 to inter. │
                                          │ - G1 from inter│
                                          └────────────────┘
```

---

## Validation Rules

### Input Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| Input File Path | Must exist and be readable | `Error: Input file not found: <path>` |
| Allowance | Must be numeric and ≥ 0 | `Error: Allowance must be a non-negative number, got: <value>` |
| Output File Path | Parent directory must exist | `Error: Output directory does not exist: <parent>` |
| Output File (existing) | Prompt for confirmation unless `--force` | `Output file exists. Overwrite? (y/n)` |
| Strategy Flag | Must be "conservative" or "aggressive" | `Invalid strategy '<value>'. Valid options are: conservative, aggressive` |

### Processing Validation

| Check | Condition | Action |
|-------|-----------|--------|
| Header Validation | Missing or malformed Snapmaker header | Console warning, proceed if parseable |
| No G1 Moves | File contains no G1 commands with Z coordinates | Warning: "No cutting moves found, output will be identical to input" |
| Threshold Outside Range | `threshold > max_z` (allowance too large) | Warning: "Threshold above maximum Z, most moves will be removed" |
| Division by Zero | `end.Z == start.Z` during split | Classify as Shallow or Deep based on Z value, don't split |
| Out-of-Range t | `t ≤ 0` or `t ≥ 1` during intersection | Error (indicates classification bug), preserve move as fallback |

---

## Data Flow

### End-to-End Processing

```
Input File (freya.cnc)
         │
         ▼
   ┌─────────────┐
   │ Parse Header│
   └──────┬──────┘
          │ HeaderMetadata
          ▼
   ┌──────────────────┐
   │ Initialize State │
   └──────┬───────────┘
          │ ModalState (Z from max_z)
          ▼
   ┌──────────────────┐
   │ First Pass:      │
   │ Find min_z       │
   └──────┬───────────┘
          │ min_z = -12.037
          │ allowance = 1.0
          │ threshold = -11.037
          ▼
   ┌──────────────────┐
   │ Second Pass:     │
   │ Process Lines    │
   └──────┬───────────┘
          │
          │ For each line:
          │   - Update ModalState
          │   - Classify Move
          │   - Apply Strategy
          │   - Accumulate Stats
          ▼
   ┌──────────────────┐
   │ Write Output     │
   └──────┬───────────┘
          │ Optimized GCode
          ▼
   ┌──────────────────┐
   │ Display Results  │
   └──────────────────┘
          │ OptimizationResult
          ▼
    Output File (freya-opt.cnc)
```

---

## Next Steps (Contracts & Quickstart)

1. **Contracts** → `contracts/`
   - This is a CLI tool with no external API
   - Internal package interfaces documented in code via Go doc comments
   - No OpenAPI/GraphQL schema needed

2. **Quickstart** → `quickstart.md`
   - Installation (download binary, `go install`)
   - Basic usage examples
   - Common workflows
