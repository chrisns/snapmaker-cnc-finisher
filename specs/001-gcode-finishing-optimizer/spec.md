# Feature Specification: GCode Finishing Pass Optimizer

**Feature Branch**: `001-gcode-finishing-optimizer`
**Created**: 2025-10-26
**Status**: Draft
**Input**: User description: "I want a simple go application that will read the gcode cnc files provided that were produced by snapmaker luban. The command should take the arguments of the finishing.cnc file. The first argument is the file to load, the second is the allowance, assume that there has already been a rough cut that has reduced the material to this apart from the last in this case 1mm, so we can reduce any cuts from the gcode that do not address this since we'll just be cutting into air for most the time. The third argument is the output file. It should then create a new gcode file that is a optimised version of the finishing.cnc that skips cutting anything that the roughcut would have already done."

## Clarifications

### Session 2025-10-26

- Q: Output file conflict resolution - What should happen when the output file already exists? → A: Prompt user for confirmation (interactive mode), but support optional --force flag to overwrite without prompting (for automation/scripting)
- Q: Z-axis reference point for depth calculations - What is Z=0 reference point? → A: Auto-detect from GCode header metadata (min_z/max_z fields); fallback to machine work origin if header incomplete; final fallback to material surface convention (Z=0 = top surface). Console alerts must show which method was used.
- Q: Estimated time savings calculation method - How should time savings be calculated? → A: Calculate initial estimate by summing machining time of removed G1 moves (distance ÷ feed rate for each move). Tool should support optional user feedback on actual job completion times to refine the estimation algorithm over time.
- Q: GCode format validation strictness - How strict should Snapmaker Luban format validation be? → A: Validate Snapmaker header presence, issue console warning if missing or malformed, but proceed with processing if the file is still parseable as GCode.
- Q: Multi-axis move preservation logic - How should multi-axis moves be handled when only Z exceeds allowance? → A: Default to preserving entire move if Z-axis component exceeds threshold (safe). Provide command-line options for alternative strategies: preserve only if all axes indicate finishing work, split moves into single-axis commands, or remove entire move if Z is shallow.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic GCode Optimization (Priority: P1)

A CNC operator has completed a rough cut with a 1.0mm allowance and needs to run a finishing pass. The finishing pass GCode file contains the complete toolpath including depths already handled by the rough cut. They need to remove redundant air-cutting operations to save machining time.

**Why this priority**: This is the core value proposition - reducing machining time by eliminating redundant operations. Without this, the tool has no purpose.

**Independent Test**: Can be fully tested by providing a finishing GCode file and allowance value, then verifying the output file has fewer operations and runs faster while maintaining the same final surface quality.

**Acceptance Scenarios**:

1. **Given** a valid finishing GCode file (produced by Snapmaker Luban) and an allowance value of 1.0mm, **When** the operator runs the optimization command, **Then** an optimized GCode file is created with reduced line count
2. **Given** a finishing pass with operations at depths shallower than the rough cut allowance, **When** the optimization runs, **Then** those shallow operations are removed from the output
3. **Given** a 3-axis or 4-axis CNC GCode file, **When** the optimization runs, **Then** the tool correctly identifies the axis configuration and processes depth (Z-axis) commands appropriately
4. **Given** a 4-axis GCode file with multi-axis moves, **When** the operator specifies different --strategy options, **Then** the tool applies the corresponding multi-axis move preservation logic ('safe', 'all-axes', 'split', or 'aggressive')

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
3. **Given** a GCode file with missing or malformed Snapmaker Luban header, **When** optimization attempts, **Then** a console warning is displayed but processing continues if the file is parseable as GCode
4. **Given** a completely unparseable file (not GCode format), **When** optimization attempts, **Then** an error indicates the file cannot be parsed

---

### Edge Cases

- What happens when the allowance value is 0? (Should process file normally but may not remove many operations)
- What happens when the finishing pass depth is entirely within the allowance threshold? (Should skip most/all cutting operations)
- How does the tool handle non-cutting GCode commands (M-codes, comments, header information)? (Should preserve them in output)
- What happens with rapid moves (G0) vs. cutting moves (G1) at shallow depths? (Should preserve rapid moves even if shallow, only remove cutting moves)
- How are multi-axis moves handled when only Z is beyond allowance? (Configurable via --strategy flag: 'safe' [default] preserves entire move if Z exceeds threshold; 'all-axes' requires all axes indicate finishing; 'split' attempts single-axis decomposition; 'aggressive' removes if Z is shallow)
- What happens if output file already exists? (Should prompt for confirmation; --force flag allows overwrite without prompting for automation)
- What happens with extremely large files (millions of lines)? (Should process efficiently with memory-conscious streaming)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tool MUST accept exactly three command-line arguments: input file path, allowance value (numeric), and output file path, with optional flags: --force (bypass overwrite confirmation), --strategy (set multi-axis move handling: 'safe' [default], 'all-axes', 'split', 'aggressive'; case-insensitive; invalid values exit with code 1 and display error message)
- **FR-002**: Tool MUST read and parse GCode files produced by Snapmaker Luban with support for both 3-axis and 4-axis configurations, validating Snapmaker header presence and issuing console warning if missing/malformed while proceeding if file is still parseable
- **FR-003**: Tool MUST identify the axis configuration (3-axis vs 4-axis) from the GCode file header or command structure
- **FR-004**: Tool MUST determine Z-axis reference point by: (1) auto-detecting from GCode header metadata (min_z/max_z), (2) falling back to machine work origin interpretation if metadata incomplete, (3) final fallback to material surface convention (Z=0 = top surface). Tool MUST display console alert indicating which method was used. Tool MUST then analyze Z-axis depth commands (G1 Z values) and compare them against the specified allowance threshold using this determined reference point. Z-axis convention: Positive Z increases upward from reference point. "Shallow depth" means the Z-value is greater than (reference_point - allowance). Example: If reference=0 and allowance=1.0mm, only cutting moves with Z ≤ -1.0mm are preserved (moves at Z > -1.0mm are considered shallow and removed).
- **FR-005**: Tool MUST remove cutting moves (G1 commands with feed rates) that occur at depths shallower than the allowance threshold, with strategy-based handling for multi-axis moves: 'safe' (default) preserves entire move if Z exceeds threshold; 'all-axes' preserves only if all axes indicate finishing work; 'split' attempts to split into single-axis commands; 'aggressive' removes entire move if Z is shallow
- **FR-006**: Tool MUST preserve all non-cutting commands including rapid moves (G0), machine codes (M-codes), header comments, and configuration commands
- **FR-007**: Tool MUST preserve the original GCode file structure including header information and metadata
- **FR-008**: Tool MUST write the optimized GCode to the specified output file path, prompting for confirmation if the file exists (see FR-001 for --force flag behavior)
- **FR-009**: Tool MUST display progress updates to console during processing including lines processed and estimated completion. Progress ETA calculated using: (elapsed_time / lines_processed) × (total_lines - lines_processed). If total line count is unknown (streaming mode), display lines processed without ETA.
- **FR-010**: Tool MUST report final statistics including total lines removed, percentage reduction, file size before/after, and estimated time savings (calculated by summing machining time of removed G1 moves using distance ÷ feed rate formula)
- **FR-011**: Tool MUST validate all inputs before processing and provide clear error messages for invalid inputs
- **FR-012**: Tool MUST handle file I/O errors gracefully with appropriate error messages

### Key Entities

- **GCode File**: Input/output file containing CNC machine instructions with commands, coordinates, and metadata
- **GCode Command**: Individual instruction line with command type (G0, G1, M3, etc.), coordinates (X, Y, Z, B), and parameters
- **Cutting Move**: GCode command that performs material removal (typically G1 commands with feed rates and Z-axis depth changes)
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
