# CLI Interface Contract

**Feature**: 001-gcode-finishing-optimizer
**Date**: 2025-10-26
**Version**: 1.0.0

This document defines the command-line interface contract for the GCode Finishing Pass Optimizer, including arguments, flags, exit codes, and output formats.

---

## Command Signature

```
snapmaker-cnc-finisher <input-file> <allowance> <output-file> [FLAGS]
```

---

## Positional Arguments

### Argument 1: `input-file`

- **Type**: File path (string)
- **Required**: Yes
- **Description**: Path to the input GCode file (finishing pass produced by Snapmaker Luban)
- **Constraints**:
  - File MUST exist
  - File MUST be readable
  - File SHOULD have extension `.cnc`, `.gcode`, or `.nc` (warning if not)
- **Example**: `finishing_pass.cnc`

### Argument 2: `allowance`

- **Type**: Numeric (float64)
- **Required**: Yes
- **Description**: Remaining material depth in millimeters after rough cut (e.g., 1.0mm)
- **Constraints**:
  - MUST be numeric (integer or decimal)
  - MUST be ≥ 0.0
  - Unit: millimeters (assumed to match GCode unit system)
- **Example**: `1.0`, `0.5`, `2.0`

### Argument 3: `output-file`

- **Type**: File path (string)
- **Required**: Yes
- **Description**: Path for the optimized GCode output file
- **Constraints**:
  - Parent directory MUST exist
  - Path MUST be writable
  - If file exists, user will be prompted for confirmation (unless `--force` flag provided)
- **Example**: `finishing_pass_optimized.cnc`

---

## Optional Flags

### `--force` / `-f`

- **Type**: Boolean flag
- **Default**: `false`
- **Description**: Overwrite output file without confirmation prompt (for automation/scripting)
- **Usage**: `--force` or `-f`

### `--strategy` / `-s`

- **Type**: String enum
- **Default**: `"safe"`
- **Allowed Values**: `safe`, `all-axes`, `split`, `aggressive`
- **Description**: Multi-axis move preservation strategy when only Z-axis exceeds allowance threshold
- **Behavior**:
  - `safe`: Preserve entire move if Z-axis component exceeds threshold (default, conservative)
  - `all-axes`: Preserve only if all axes indicate finishing work
  - `split`: Attempt to split multi-axis move into single-axis commands
  - `aggressive`: Remove entire move if Z is shallow, even if other axes exceed threshold
- **Usage**: `--strategy=safe` or `-s aggressive`

### `--help` / `-h`

- **Type**: Boolean flag
- **Description**: Display help message and exit
- **Precedence**: Overrides all other arguments; tool exits after displaying help

### `--version` / `-v`

- **Type**: Boolean flag
- **Description**: Display version information and exit
- **Precedence**: Overrides all other arguments except `--help`; tool exits after displaying version

---

## Usage Examples

### Example 1: Basic Usage

```bash
snapmaker-cnc-finisher finishing.cnc 1.0 output.cnc
```

**Effect**: Processes `finishing.cnc` with 1.0mm allowance, writes optimized output to `output.cnc`, prompts if output file exists.

### Example 2: Force Overwrite

```bash
snapmaker-cnc-finisher finishing.cnc 1.5 output.cnc --force
```

**Effect**: Overwrites `output.cnc` without prompting.

### Example 3: Custom Strategy

```bash
snapmaker-cnc-finisher finishing.cnc 0.5 output.cnc --strategy=aggressive
```

**Effect**: Uses aggressive strategy to maximize line removal.

### Example 4: Combined Flags

```bash
snapmaker-cnc-finisher finishing.cnc 2.0 output.cnc -f -s all-axes
```

**Effect**: Force overwrite + all-axes strategy.

### Example 5: Help Display

```bash
snapmaker-cnc-finisher --help
```

**Output**:
```
GCode Finishing Pass Optimizer v1.0.0

Usage: snapmaker-cnc-finisher <input-file> <allowance> <output-file> [FLAGS]

Positional Arguments:
  input-file     Path to input GCode file (finishing pass from Snapmaker Luban)
  allowance      Remaining material depth in mm after rough cut (e.g., 1.0)
  output-file    Path for optimized output GCode file

Optional Flags:
  --force, -f              Overwrite output file without confirmation
  --strategy=<value>, -s   Multi-axis move handling strategy (default: safe)
                           Allowed values: safe, all-axes, split, aggressive
  --help, -h               Display this help message
  --version, -v            Display version information

Examples:
  snapmaker-cnc-finisher finishing.cnc 1.0 output.cnc
  snapmaker-cnc-finisher finishing.cnc 1.5 output.cnc --force
  snapmaker-cnc-finisher finishing.cnc 0.5 output.cnc --strategy=aggressive

For more information, visit: https://github.com/chrisns/snapmaker-cnc-finisher
```

### Example 6: Version Display

```bash
snapmaker-cnc-finisher --version
```

**Output**:
```
snapmaker-cnc-finisher version 1.0.0
Built with Go 1.25.3
Platform: darwin/arm64
```

---

## Exit Codes

| Code | Meaning | Description |
|------|---------|-------------|
| `0` | Success | Optimization completed successfully |
| `1` | Invalid Arguments | Missing arguments, invalid allowance, or invalid strategy |
| `2` | Input File Error | Input file not found, not readable, or malformed GCode |
| `3` | Output File Error | Output file not writable or user declined overwrite prompt |
| `4` | Processing Error | Unexpected error during GCode parsing or filtering |
| `5` | Resource Error | Out of memory or disk space during processing |

---

## Standard Output (stdout)

### Progress Updates

**Format**: Single-line status (overwritten using `\r` carriage return)

**Example**:
```
Processing: 45,230 / 100,000 lines (45.2%) | Removed: 12,450 | Elapsed: 3.2s | ETA: 3.8s
```

**Frequency**: Every 10,000 lines OR every 2 seconds (whichever is more frequent) per SC-005.

### Final Summary

**Format**: Multi-line report displayed after completion

**Example**:
```
✓ Optimization complete!

Input:  finishing.cnc (1.2 MB, 100,000 lines)
Output: output.cnc (850 KB, 72,450 lines)

Results:
  Lines removed:       27,550 (27.6%)
  File size reduced:   350 KB (29.2%)
  Estimated time saved: 12.5 minutes

Z-axis reference: Detected from header metadata (min_z/max_z)
Strategy used: safe

Processing time: 6.4 seconds
```

---

## Standard Error (stderr)

### Warnings

**Format**: `WARNING: <message>`

**Examples**:
```
WARNING: Input file has no .cnc extension. Proceeding anyway.
WARNING: Snapmaker header missing or malformed. Using fallback Z-reference (material surface).
WARNING: Feed rate not specified for line 420. Using default 1000 mm/min.
```

### Errors

**Format**: `ERROR: <message>`

**Examples**:
```
ERROR: Input file 'nonexistent.cnc' not found.
ERROR: Invalid allowance value '-1.0'. Must be >= 0.
ERROR: Invalid strategy 'invalid'. Must be one of: safe, all-axes, split, aggressive.
ERROR: Cannot write to output file 'readonly.cnc'. Permission denied.
```

### User Prompts (Interactive Mode)

**Example** (when output file exists and `--force` not provided):
```
Output file 'output.cnc' already exists. Overwrite? (y/N):
```

**Behavior**:
- `y` or `Y` → Proceed with overwrite
- `n`, `N`, or Enter → Abort with exit code 3
- Invalid input → Re-prompt

---

## Environment Variables

**None**. Tool does not read environment variables.

---

## Configuration Files

**None**. Tool does not read configuration files (per Scope Boundaries in spec.md).

---

## Input File Format Contract

### Expected Format: Snapmaker Luban GCode

**Header Section** (optional but recommended):
```
;Header Start
;header_type: cnc
;file_total_lines: 100000
;estimated_time(s): 3600
;min_x(mm): 0.0
;max_x(mm): 200.0
;min_y(mm): 0.0
;max_y(mm): 150.0
;min_z(mm): -10.0
;max_z(mm): 5.0
;min_b(deg): 0.0
;max_b(deg): 360.0
;Header End
```

**Body Section**:
- Line-based format (one command per line)
- Commands: `G0`, `G1`, `M3`, `M5`, etc.
- Parameters: `X`, `Y`, `Z`, `B`, `F`, `S`, etc.
- Comments: Lines starting with `;` or inline comments after commands
- Example line: `G1 X10.5 Y20.3 Z-1.2 F1500 ;cutting move`

**Validation**:
- Tool MUST handle files with or without Snapmaker header (FR-002)
- Tool MUST issue warning if header missing/malformed but proceed if parseable
- Tool MUST reject completely unparseable files (exit code 2)

---

## Output File Format Contract

### Guaranteed Format

**Header Preservation**:
- Original Snapmaker header MUST be preserved verbatim (FR-007)
- Header statistics (e.g., `file_total_lines`, `estimated_time`) MAY become outdated (tool does not update them)

**Body**:
- All retained lines written in original order
- Original formatting preserved (spacing, comments, line structure)
- Filtered lines omitted entirely (no placeholders or comments indicating removal)

**Example Comparison**:

**Input**:
```
G1 X10 Y10 Z-0.3 F1500  ; Shallow cut
G1 X20 Y10 Z-1.5 F1500  ; Deep cut
```

**Output** (assuming allowance = 1.0mm):
```
G1 X20 Y10 Z-1.5 F1500  ; Deep cut
```

---

## Error Messages Contract

### Error Message Format

**Pattern**: `ERROR: <clear description> [<actionable hint>]`

### Error Message Catalog

| Trigger | Error Message | Exit Code |
|---------|---------------|-----------|
| Missing arguments | `ERROR: Expected 3 arguments. Usage: snapmaker-cnc-finisher <input-file> <allowance> <output-file>` | 1 |
| Non-numeric allowance | `ERROR: Invalid allowance value 'abc'. Must be a numeric value >= 0.` | 1 |
| Negative allowance | `ERROR: Invalid allowance value '-1.0'. Must be >= 0.` | 1 |
| Input file not found | `ERROR: Input file 'missing.cnc' not found.` | 2 |
| Input file not readable | `ERROR: Cannot read input file 'readonly.cnc'. Permission denied.` | 2 |
| Invalid strategy | `ERROR: Invalid strategy 'wrong'. Must be one of: safe, all-axes, split, aggressive.` | 1 |
| Output not writable | `ERROR: Cannot write to output file 'path/output.cnc'. Directory does not exist or permission denied.` | 3 |
| Out of memory | `ERROR: Out of memory while processing file. Try processing on a system with more RAM.` | 5 |
| Disk full | `ERROR: Out of disk space while writing output file. Free up disk space and try again.` | 5 |

---

## Performance Contract

Per success criteria in spec.md:

- **SC-001**: Process 100,000-line files in < 10 seconds
- **SC-005**: Progress updates every 10,000 lines or 2 seconds (whichever more frequent)
- **SC-006**: Handle up to 10 million lines without crashing or running out of memory

These are **guaranteed behaviors** that the tool MUST meet.

---

## Compatibility Contract

**Platforms**:
- macOS (Intel x86_64, ARM64 M1/M2/M3)
- Windows (x86_64)
- Linux (x86_64)

**Dependencies**: None (static binary)

**GCode Compatibility**:
- Primary: Snapmaker Luban output format
- Secondary: Generic GCode (will attempt to parse but may issue warnings)

---

## Versioning

**Semantic Versioning**: `MAJOR.MINOR.PATCH`

**Breaking Changes** (MAJOR increment):
- Changing positional argument order
- Removing flags
- Changing exit code meanings
- Changing output format in incompatible ways

**Non-Breaking Changes** (MINOR increment):
- Adding new flags
- Adding new strategy values
- Improving error messages
- Performance improvements

**Patches** (PATCH increment):
- Bug fixes
- Documentation updates

---

## Testing Contract

**Contract Tests** (per Constitution Principle III):

Test suite MUST verify:
1. All argument combinations produce expected exit codes
2. Invalid inputs trigger appropriate error messages
3. Progress updates occur at documented frequency
4. Output file format matches input format structure
5. Help and version flags work correctly

**Example Test Cases**:
```go
func TestCLIContract(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        expectExit int
        expectOut  string
    }{
        {
            name:       "Missing arguments",
            args:       []string{},
            expectExit: 1,
            expectOut:  "ERROR: Expected 3 arguments",
        },
        {
            name:       "Invalid allowance",
            args:       []string{"input.cnc", "abc", "output.cnc"},
            expectExit: 1,
            expectOut:  "ERROR: Invalid allowance value 'abc'",
        },
        {
            name:       "Help flag",
            args:       []string{"--help"},
            expectExit: 0,
            expectOut:  "Usage: snapmaker-cnc-finisher",
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute CLI and verify exit code + output
        })
    }
}
```

---

**Generated By**: Claude Code (Sonnet 4.5)
**References**: [spec.md](../spec.md), [data-model.md](../data-model.md)
