# Quickstart Guide: GCode Finishing Pass Optimizer

**Version**: 1.0.0 (planned)
**Audience**: CNC operators using Snapmaker Luban
**Prerequisites**: None (standalone binary)

## What is This Tool?

The GCode Finishing Pass Optimizer reduces machining time by removing redundant shallow cutting operations from finishing passes. When you've already done a rough cut with an allowance (e.g., 1mm of material left), the finishing pass GCode often contains complete toolpaths including depths already handled by the rough cut. This tool intelligently removes those redundant "air cutting" moves while preserving the final finishing layer.

**Real-world example**: A finishing pass with 8.2 million lines optimized to 5.1 million lines (38% reduction), saving ~142 minutes of machining time.

---

## Installation

### Option 1: Download Pre-built Binary (Recommended)

1. Visit the [Releases page](https://github.com/chrisns/snapmaker-cnc-finisher/releases)
2. Download the binary for your platform:
   - **macOS Intel**: `gcode-optimizer-darwin-amd64`
   - **macOS ARM (M1/M2)**: `gcode-optimizer-darwin-arm64`
   - **Windows**: `gcode-optimizer-windows-amd64.exe`
   - **Linux**: `gcode-optimizer-linux-amd64`
3. Make it executable (macOS/Linux):
   ```bash
   chmod +x gcode-optimizer-*
   ```
4. Optional: Move to a directory in your PATH:
   ```bash
   sudo mv gcode-optimizer-* /usr/local/bin/gcode-optimizer
   ```

### Option 2: Install from Source

Requires Go 1.21 or later:

```bash
go install github.com/chrisns/snapmaker-cnc-finisher/cmd/gcode-optimizer@latest
```

---

## Basic Usage

### Syntax

```bash
gcode-optimizer [flags] <input.cnc> <allowance> <output.cnc>
```

**Arguments**:
- `<input.cnc>`: Path to your finishing pass GCode file (produced by Snapmaker Luban)
- `<allowance>`: Material thickness left after rough cut (in millimeters, e.g., `1.0`)
- `<output.cnc>`: Path for the optimized output file

**Flags** (optional):
- `--force`: Overwrite output file without confirmation prompt
- `--strategy=<conservative|aggressive>`: Optimization strategy (default: `aggressive`)

---

## Examples

### Example 1: Basic Optimization (1mm Allowance)

You've completed a rough cut leaving 1mm of material. Optimize the finishing pass:

```bash
gcode-optimizer finishing.cnc 1.0 finishing-optimized.cnc
```

**Output**:
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

### Example 2: Conservative Strategy (Safer, Less Optimization)

If you prefer maximum safety and don't mind slightly less time savings:

```bash
gcode-optimizer --strategy=conservative finishing.cnc 1.0 finishing-safe.cnc
```

**Difference from aggressive**:
- **Aggressive** (default): Splits moves that cross the threshold at the exact intersection point
- **Conservative**: Preserves entire moves that cross the threshold (safer, but includes some redundant cutting)

**When to use conservative**:
- First time using the tool on a critical part
- Uncertain about rough cut uniformity
- Prefer to err on the side of caution

### Example 3: Overwrite Existing File

Skip the confirmation prompt when overwriting:

```bash
gcode-optimizer --force finishing.cnc 1.0 finishing-optimized.cnc
```

**Use case**: Automation scripts, batch processing

### Example 4: Different Allowance Values

Adjust allowance based on your rough cut settings:

```bash
# 0.5mm allowance (aggressive rough cut)
gcode-optimizer finishing.cnc 0.5 finishing-optimized.cnc

# 2.0mm allowance (conservative rough cut)
gcode-optimizer finishing.cnc 2.0 finishing-optimized.cnc
```

**Effect**: Smaller allowance = less optimization (fewer moves removed). Larger allowance = more optimization (more moves removed).

---

## Typical Workflow

### Step 1: Generate Rough and Finishing Passes in Luban

1. Open your model in Snapmaker Luban
2. Set up toolpaths:
   - **Rough pass**: Configure with allowance (e.g., 1.0mm)
   - **Finishing pass**: Generate without allowance (complete final surface)
3. Export both GCode files:
   - `roughing.cnc`
   - `finishing.cnc`

### Step 2: Run Rough Cut on CNC

Load and execute `roughing.cnc` on your Snapmaker. This removes the bulk of material but leaves the specified allowance.

### Step 3: Optimize Finishing Pass

```bash
gcode-optimizer finishing.cnc 1.0 finishing-optimized.cnc
```

**Important**: Use the same allowance value you configured in Luban's rough pass.

### Step 4: Run Optimized Finishing Pass

Load `finishing-optimized.cnc` on your Snapmaker and execute. This will:
- Skip the shallow cuts already handled by the rough pass
- Execute only the final finishing layer
- Save significant machining time

---

## Understanding Output Statistics

### Depth Analysis

```
Min Z: -12.037mm
Threshold: -11.037mm (1.0mm allowance)
```

- **Min Z**: Deepest cut in the finishing pass
- **Threshold**: Calculated as `min_z + allowance` — only moves at or below this depth are preserved

### Processing Summary

```
Lines removed: 3,127,845 (38.0%)
Moves split: 1,234 (aggressive strategy)
```

- **Lines removed**: Shallow moves entirely above the threshold (already cut by rough pass)
- **Moves split**: Moves that cross the threshold (only deep portion preserved)

### Time Savings Estimate

```
Estimated time savings: 142.5 minutes
```

Calculated by summing the machining time of all removed G1 moves:
```
time = (move_distance / feed_rate) for each removed move
```

**Note**: This is an estimate based on commanded feed rates. Actual savings may vary due to acceleration/deceleration, but provides a good ballpark figure.

---

## Troubleshooting

### Error: "Input file not found"

**Cause**: File path is incorrect or file doesn't exist

**Solution**:
```bash
# Check file exists
ls -lh finishing.cnc

# Use absolute path if relative path fails
gcode-optimizer /full/path/to/finishing.cnc 1.0 output.cnc
```

### Error: "Invalid allowance value"

**Cause**: Allowance argument is not a valid number or is negative

**Solution**:
```bash
# Correct: use decimal number
gcode-optimizer finishing.cnc 1.0 output.cnc

# Wrong: negative or non-numeric
gcode-optimizer finishing.cnc -1.0 output.cnc  # ❌
gcode-optimizer finishing.cnc abc output.cnc    # ❌
```

### Error: "Invalid strategy"

**Cause**: `--strategy` flag has unsupported value

**Solution**:
```bash
# Valid options
gcode-optimizer --strategy=aggressive finishing.cnc 1.0 output.cnc
gcode-optimizer --strategy=conservative finishing.cnc 1.0 output.cnc

# Wrong
gcode-optimizer --strategy=medium finishing.cnc 1.0 output.cnc  # ❌
```

### Warning: "Missing or malformed Snapmaker header"

**Cause**: Input file doesn't have standard Snapmaker Luban header

**Effect**: Tool will proceed with processing if file is still valid GCode

**Solution**:
- If output looks correct, no action needed
- If you encounter issues, verify the input file was actually produced by Snapmaker Luban

### Warning: "Threshold above maximum Z"

**Cause**: Allowance is too large (threshold exceeds the highest point in the file)

**Effect**: Most/all moves will be removed (likely not what you want)

**Solution**: Reduce the allowance value to match your actual rough cut settings

---

## Advanced Usage

### Checking Tool Version

```bash
gcode-optimizer --version
```

### Viewing Help

```bash
gcode-optimizer --help
```

**Output**:
```
Usage: gcode-optimizer [flags] <input.cnc> <allowance> <output.cnc>

Optimize Snapmaker Luban finishing pass GCode by removing redundant shallow cuts.

Arguments:
  input.cnc    Path to finishing pass GCode file
  allowance    Material thickness left after rough cut (mm)
  output.cnc   Path for optimized output file

Flags:
  --force            Overwrite output without confirmation
  --strategy=<str>   Optimization strategy: conservative|aggressive (default: aggressive)
  --version          Show version information
  --help             Show this help message

Examples:
  gcode-optimizer finishing.cnc 1.0 finishing-opt.cnc
  gcode-optimizer --strategy=conservative finishing.cnc 1.0 output.cnc
  gcode-optimizer --force finishing.cnc 1.0 output.cnc

For more information: https://github.com/chrisns/snapmaker-cnc-finisher
```

### Batch Processing (Script Example)

Process multiple files with the same allowance:

```bash
#!/bin/bash
ALLOWANCE=1.0

for file in finishing-*.cnc; do
    output="${file%.cnc}-optimized.cnc"
    echo "Optimizing $file → $output"
    gcode-optimizer --force "$file" "$ALLOWANCE" "$output"
done
```

---

## Safety Recommendations

1. **Test on Non-Critical Parts First**: Run your first optimized file on scrap material to verify correctness

2. **Verify Allowance Matches Rough Cut**: Use the exact same allowance value you configured in Luban's rough pass

3. **Inspect Optimized File**: Open the optimized GCode in a text editor or viewer to spot-check that the structure looks reasonable

4. **Start with Conservative Strategy**: For critical parts, use `--strategy=conservative` for the safest optimization

5. **Monitor First Run**: Watch the first few minutes of the optimized finishing pass to ensure the tool enters the material at the expected depth

---

## What Gets Preserved vs. Removed

### Always Preserved ✅
- **Header and metadata**: Snapmaker Luban header information
- **G0 rapid moves**: Even if shallow (rapid positioning doesn't cut material)
- **M-codes**: Machine commands (spindle control, etc.)
- **Comments**: All inline and line comments
- **G1 moves at/below threshold**: Cutting moves in the finishing zone

### Removed ❌
- **Shallow G1 moves**: Cutting moves entirely above the threshold (already handled by rough cut)

### Conditionally Modified ⚙️
- **Threshold-crossing G1 moves**:
  - **Conservative strategy**: Entire move preserved
  - **Aggressive strategy**: Split at threshold intersection, only deep portion preserved

---

## Expected Results

### Typical Reduction Ranges

Based on test data (spec SC-008):
- **File size reduction**: 15-40% (typical for finishing passes with 0.5-2.0mm allowances)
- **Machining time savings**: 20%+ (spec SC-002 minimum target)

**Factors affecting reduction**:
- **Allowance size**: Larger allowance = more shallow moves = greater reduction
- **Part geometry**: Flat surfaces optimize more than complex contours
- **Toolpath strategy**: Different Luban toolpath patterns yield different results

### Surface Quality

**No change expected**: The optimized file should produce identical final surface quality as the original finishing pass (spec SC-003). The tool removes only redundant shallow cuts, not any material-removing operations in the finishing zone.

---

## Support and Feedback

- **Issues/Bugs**: https://github.com/chrisns/snapmaker-cnc-finisher/issues
- **Discussions**: https://github.com/chrisns/snapmaker-cnc-finisher/discussions
- **Documentation**: https://github.com/chrisns/snapmaker-cnc-finisher/wiki

---

## License

MIT License - see [LICENSE](https://github.com/chrisns/snapmaker-cnc-finisher/blob/main/LICENSE) file for details.
