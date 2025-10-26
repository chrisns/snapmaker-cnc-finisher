# Feature Specification: GCode Finishing Pass Optimizer

**Feature Branch**: `001-gcode-finishing-optimizer`
**Created**: 2025-10-26
**Status**: Draft
**Input**: User description: "I want a simple go application that will read the gcode cnc files provided that were produced by snapmaker luban. The command should take the arguments of the finishing.cnc file. The first argument is the file to load, the second is the allowance, assume that there has already been a rough cut that has reduced the material to this apart from the last in this case 1mm, so we can reduce any cuts from the gcode that do not address this since we'll just be cutting into air for most the time. The third argument is the output file. It should then create a new gcode file that is a optimised version of the finishing.cnc that skips cutting anything that the roughcut would have already done."

## Clarifications

### Session 2025-10-26

- Q: Output file conflict resolution - What should happen when the output file already exists? → A: Prompt user for confirmation (interactive mode), but support optional --force flag to overwrite without prompting (for automation/scripting)
- Q: How to determine which cuts are needed for finishing? → A: Find the minimum Z value (deepest cut) in the file, then calculate threshold = min_z + allowance. Only preserve cuts at or below this threshold. This ensures we only keep the final material layer that needs finishing, not the upper layers already removed by the rough cut.
- Q: Estimated time savings calculation method - How should time savings be calculated? → A: Calculate initial estimate by summing machining time of removed G1 moves (distance ÷ feed rate for each move). Tool should support optional user feedback on actual job completion times to refine the estimation algorithm over time.
- Q: GCode format validation strictness - How strict should Snapmaker Luban format validation be? → A: Validate Snapmaker header presence, issue console warning if missing or malformed, but proceed with processing if the file is still parseable as GCode.
- Q: How should moves that cross the depth threshold be handled? → A: Split the move at the threshold intersection point using parametric linear interpolation. Preserve only the portion of the move that goes into the finishing zone (Z ≤ threshold). Maintain the original feed rate without adjustment, as feed rate is modal in GCode and applies continuously along the toolpath.
- Q: Move splitting coordinate calculation edge case - When splitting a G1 move at the threshold intersection point (FR-013), if the starting Z coordinate is not explicitly stated in the GCode line (due to modal programming where Z remains from a previous command), how should the tool determine the start position for interpolation calculations? → A: Track modal state of all coordinates (X, Y, Z, B) throughout file processing and use last known values for any coordinates not explicitly specified in the current command.
- Q: Strategy flag definition - The scope section mentions an optional "--strategy" flag, but the specification doesn't define what optimization strategies should be supported. What strategies should the tool implement? → A: Support two strategies: "conservative" (only remove moves entirely above threshold) and "aggressive" (remove moves + split threshold-crossing moves as currently specified). Aggressive is the default.
- Q: Invalid strategy value handling - What should happen when a user provides an invalid value for the --strategy flag (e.g., --strategy=medium or --strategy=xyz)? → A: Reject with error message listing valid options: "Invalid strategy 'xyz'. Valid options are: conservative, aggressive"
- Q: Initial modal state for coordinates - When the tool starts processing, before encountering the first explicit coordinate values, what should be the initial values for modal state tracking (X, Y, Z, B)? → A: Initialize from header metadata: Z from max_z (or 0 if max_z not in header), X/Y/B default to 0. This ensures correct move calculations even when early G1 commands use modal programming.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic GCode Optimization (Priority: P1)

A CNC operator has completed a rough cut with a 1.0mm allowance and needs to run a finishing pass. The finishing pass GCode file contains the complete toolpath including depths already handled by the rough cut. They need to remove redundant air-cutting operations to save machining time.

**Why this priority**: This is the core value proposition - reducing machining time by eliminating redundant operations. Without this, the tool has no purpose.

**Independent Test**: Can be fully tested by providing a finishing GCode file and allowance value, then verifying the output file has fewer operations and runs faster while maintaining the same final surface quality.

**Acceptance Scenarios**:

1. **Given** a valid finishing GCode file (produced by Snapmaker Luban) and an allowance value of 1.0mm, **When** the operator runs the optimization command, **Then** an optimized GCode file is created with reduced line count
2. **Given** a finishing pass with operations at depths shallower than the rough cut allowance, **When** the optimization runs, **Then** those shallow operations are removed from the output
3. **Given** a 3-axis or 4-axis CNC GCode file, **When** the optimization runs, **Then** the tool correctly identifies the axis configuration and processes depth (Z-axis) commands appropriately
4. **Given** a GCode file with moves that cross the depth threshold, **When** the optimization runs, **Then** the tool splits those moves at the threshold intersection point and preserves only the deep portion that requires finishing

---

### User Story 2 - Progress Monitoring (Priority: P2)

During optimization of large GCode files, the operator wants to see real-time progress updates to understand how much time reduction they're achieving and ensure the tool is working correctly.

**Why this priority**: Provides transparency and confidence in the optimization process, especially important for large files that may take time to process.

**Independent Test**: Can be tested by running the tool and observing console output shows progress updates with line counts and file size metrics throughout processing.

**Acceptance Scenarios**:

1. **Given** a GCode file being optimized, **When** the tool processes the file, **Then** progress updates display on console showing number of lines processed
2. **Given** the optimization completes, **When** the final summary is shown, **Then** it includes total lines removed and final file size reduction percentage
3. **Given** a very large GCode file, **When** processing occurs, **Then** progress updates appear regularly (not just at start and end)

---

### User Story 3 - Error Handling (Priority: P3)

When provided with invalid inputs (non-existent files, invalid allowance values, or unsupported GCode formats), the operator needs clear error messages explaining what went wrong and how to fix it.

**Why this priority**: Essential for usability but doesn't affect core functionality. Prevents frustration and improves user experience when mistakes happen.

**Independent Test**: Can be tested by providing various invalid inputs and verifying appropriate error messages are displayed.

**Acceptance Scenarios**:

1. **Given** a non-existent input file path, **When** the command runs, **Then** a clear error message indicates the file was not found
2. **Given** an invalid allowance value (negative or non-numeric), **When** the command runs, **Then** an error message explains valid allowance format
3. **Given** an invalid --strategy flag value (e.g., --strategy=medium), **When** the command runs, **Then** an error message lists the valid strategy options (conservative, aggressive)
4. **Given** a GCode file with missing or malformed Snapmaker Luban header, **When** optimization attempts, **Then** a console warning is displayed but processing continues if the file is parseable as GCode
5. **Given** a completely unparseable file (not GCode format), **When** optimization attempts, **Then** an error indicates the file cannot be parsed

---

### Edge Cases

- What happens when the allowance value is 0? (Should process file normally but may not remove many operations)
- What happens when the finishing pass depth is entirely within the allowance threshold? (Should skip most/all cutting operations)
- How does the tool handle non-cutting GCode commands (M-codes, comments, header information)? (Should preserve them in output)
- What happens with rapid moves (G0) vs. cutting moves (G1) at shallow depths? (Should preserve rapid moves even if shallow, only remove cutting moves)
- How are moves handled when they cross the depth threshold? (Depends on strategy: aggressive splits moves using parametric linear interpolation - see FR-013; conservative preserves entire crossing moves - see FR-005)
- What happens if output file already exists? (Should prompt for confirmation; --force flag allows overwrite without prompting for automation)
- What happens with extremely large files (millions of lines)? (Should process efficiently with memory-conscious in-memory processing using efficient data structures)
- What happens when --strategy flag is not specified? (Defaults to aggressive strategy with move splitting enabled)
- What happens with invalid --strategy flag values? (Tool rejects with error listing valid options: conservative, aggressive)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tool MUST accept exactly three command-line arguments: input file path, allowance value (numeric), and output file path, with optional flags: --force (bypass overwrite confirmation) and --strategy=<conservative|aggressive> (optimization strategy, default: aggressive)
- **FR-002**: Tool MUST read and parse GCode files produced by Snapmaker Luban with support for both 3-axis and 4-axis configurations, validating Snapmaker header presence and issuing console warning if missing/malformed while proceeding if file is still parseable
- **FR-003**: Tool MUST identify the axis configuration (3-axis vs 4-axis) from the GCode file header or command structure
- **FR-004**: Tool MUST scan all G1 commands in the file to determine the minimum Z value (deepest cut). Tool MUST calculate the depth threshold as: **threshold = min_z + allowance**. Example: If the finishing pass cuts from Z=0 down to Z=-10mm with a 1.0mm allowance, then min_z=-10mm and threshold=-9.0mm. Only cutting moves with Z ≤ -9.0mm are preserved (these represent the final 1mm of material that needs finishing). Moves with Z > -9.0mm are shallow (already removed by rough cut) and should be removed or split (see FR-013). Z-axis convention: Negative Z values represent depth below material surface; more negative = deeper cuts. Tool MUST display console message showing detected min_z and calculated threshold.
- **FR-005**: Tool MUST handle cutting moves (G1 commands) based on their relationship to the depth threshold calculated in FR-004, with behavior determined by the selected strategy (see FR-001):
  - **Conservative strategy** (--strategy=conservative):
    - **Both start and end points shallow** (both Z > threshold): Remove entire move
    - **Both start and end points deep** (both Z ≤ threshold): Preserve entire move
    - **Move crosses threshold**: Preserve entire move (safer approach, includes some redundant cutting but eliminates risk of split-move errors)
  - **Aggressive strategy** (--strategy=aggressive, default):
    - **Both start and end points shallow** (both Z > threshold): Remove entire move
    - **Both start and end points deep** (both Z ≤ threshold): Preserve entire move
    - **Move crosses threshold**: Split move at threshold intersection point using parametric linear interpolation (see FR-013)
  - **Feed rate preservation**: Maintain original feed rate (F parameter) without adjustment when splitting moves (feed rate is modal in GCode and applies continuously)
- **FR-006**: Tool MUST preserve all non-cutting commands including rapid moves (G0), machine codes (M-codes), header comments, and configuration commands
- **FR-007**: Tool MUST preserve the original GCode file structure including header information and metadata
- **FR-008**: Tool MUST write the optimized GCode to the specified output file path, prompting for confirmation if the file exists (see FR-001 for --force flag behavior)
- **FR-009**: Tool MUST display progress updates to console during processing including lines processed and estimated completion. Progress ETA calculated using: (elapsed_time / lines_processed) × (total_lines - lines_processed). If total line count is unavailable in header metadata, display lines processed without ETA.
- **FR-010**: Tool MUST report final statistics including total lines removed, percentage reduction, file size before/after, and estimated time savings (calculated by summing machining time of removed G1 moves using distance ÷ feed rate formula)
- **FR-011**: Tool MUST validate all inputs before processing and provide clear error messages for invalid inputs. Validation includes: file path existence, allowance value (must be numeric and non-negative), and --strategy flag value (must be "conservative" or "aggressive" if specified). Invalid --strategy values MUST produce error message format: "Invalid strategy '<value>'. Valid options are: conservative, aggressive"
- **FR-012**: Tool MUST handle file I/O errors gracefully with appropriate error messages
- **FR-013**: For G1 moves that cross the depth threshold when using aggressive strategy, tool MUST split them using parametric linear interpolation:
  1. **Modal state initialization**: Before processing commands, initialize modal state from header metadata: Z from max_z value in header (or 0 if max_z not present), X/Y/B default to 0. Feed rate (F) initializes to 0 (will be set by first move command).
  2. **Modal state tracking**: Tool MUST maintain current state of all coordinates (X, Y, Z, B) and parameters (F) throughout file processing. When a coordinate is not explicitly specified in a command, use the last known value from modal state.
  3. **Calculate intersection parameter**: `t = (threshold - Z_start) / (Z_end - Z_start)` where 0 < t < 1 indicates the move crosses threshold
  4. **Calculate intersection point**: `X₀ = X_start + t(X_end - X_start)`, `Y₀ = Y_start + t(Y_end - Y_start)`, `Z₀ = threshold`
  5. **Entering deep zone** (start shallow Z > threshold, end deep Z ≤ threshold): Output move from intersection point to end point only, discarding shallow portion. Example: `G1 X7.5 Y15 Z-9.0 F1000` then `G1 X10 Y20 Z-10`
  6. **Leaving deep zone** (start deep Z ≤ threshold, end shallow Z > threshold): Output move from start point to intersection point only, discarding shallow portion
  7. **Coordinate precision**: Maintain 3-4 decimal places for intersection point coordinates
  8. **Feed rate handling**: Preserve original feed rate without modification (GCode feed rate is modal and continues across split moves)

### Key Entities

- **GCode File**: Input/output file containing CNC machine instructions with commands, coordinates, and metadata
- **GCode Command**: Individual instruction line with command type (G0, G1, M3, etc.), coordinates (X, Y, Z, B), and parameters
- **Cutting Move**: GCode command that performs material removal (typically G1 commands with feed rates and Z-axis depth changes)
- **Modal State**: Current position and parameter values maintained throughout file processing. In GCode, coordinates and parameters not explicitly specified in a command persist from previous commands. Tool tracks X, Y, Z, B coordinates and F (feed rate) parameter to enable correct move splitting calculations.
- **Optimization Strategy**: Approach for handling moves that cross the depth threshold. Conservative strategy preserves entire crossing moves (safer, less optimization). Aggressive strategy splits crossing moves at threshold intersection point (maximum time savings, requires precise interpolation).
- **Allowance Threshold**: Numeric value representing remaining material depth after rough cut (e.g., 1.0mm)
- **Optimization Statistics**: Data structure tracking lines removed, file size changes, estimated time savings (calculated from removed G1 move distances and feed rates), and processing metrics

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can process a finishing GCode file and receive an optimized output file in under 10 seconds for files up to 100,000 lines
- **SC-002**: Optimized files reduce total machining time by at least 20% compared to original finishing pass for typical rough+finish workflows (measured as reduction in wall-clock time for CNC to execute the GCode from start to completion)
- **SC-003**: Optimized GCode produces identical final surface quality as original finishing pass when verified by test cuts
- **SC-004**: Tool correctly identifies and processes both 3-axis and 4-axis CNC GCode files with 100% accuracy
- **SC-005**: Console progress updates appear at minimum every 10,000 lines processed or every 2 seconds, whichever is more frequent
- **SC-006**: Tool handles GCode files up to 10 million lines without running out of memory or crashing
- **SC-007**: 95% of invalid inputs result in clear, actionable error messages rather than crashes or unclear failures
- **SC-008**: File size reduction averages 15-40% for typical finishing passes (typical = finishing passes with 0.5-2.0mm allowances on flat or contoured surfaces; outliers outside 15-40% range should be investigated for toolpath anomalies)

## Assumptions

- Rough cut has been completed with consistent allowance across the entire part
- Finishing GCode files follow Snapmaker Luban output format conventions
- Z-axis represents depth in all GCode files (standard for CNC machining); reference point is auto-detected from header or uses fallback conventions
- Allowance is provided in millimeters matching the GCode unit system
- Users have basic command-line proficiency to run the tool
- Output file path is writable and has sufficient disk space
- GCode files are text-based and human-readable (not binary)
- Feed rates and spindle speeds in original GCode are appropriate and should be preserved
- The rough cut left uniform allowance (not variable depth)

## Dependencies

- File system access for reading input and writing output files
- Console/terminal for displaying progress and results
- GCode parsing capability (library or custom implementation to be determined)

## Scope Boundaries

### In Scope
- Reading Snapmaker Luban GCode files
- Parsing 3-axis and 4-axis GCode commands
- Identifying and removing redundant cutting moves based on depth
- Preserving file structure and non-cutting commands
- Progress reporting and statistics
- Time savings estimation based on removed G1 move calculations (distance ÷ feed rate)
- Command-line interface with three arguments (plus optional --force and --strategy flags)
- Error handling for common failure modes

### Out of Scope
- GCode simulation or visualization
- Support for non-Snapmaker GCode formats (other CAM software)
- Rough cut GCode analysis (only finishing file is processed)
- GCode validation for machine safety or correctness
- Toolpath optimization beyond depth-based filtering (e.g., move ordering, velocity optimization)
- Graphical user interface
- Batch processing multiple files
- Configuration files or saved preferences
- Integration with CAM software or machine controllers
