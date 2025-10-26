# GCode Finishing Pass Optimizer

A command-line tool that optimizes Snapmaker Luban CNC finishing pass GCode files by removing redundant shallow cutting operations already handled by rough cuts.

## Features

- **Time Savings**: Reduces machining time by 20%+ on typical rough+finish workflows
- **Smart Optimization**: Removes air-cutting moves while preserving the final finishing layer
- **Two Strategies**: Conservative (safer) or Aggressive (maximum optimization) modes
- **Progress Tracking**: Real-time progress updates with ETA and statistics
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Installation

### Option 1: Download Pre-built Binary

Visit the [Releases page](https://github.com/chrisns/snapmaker-cnc-finisher/releases) and download the binary for your platform.

### Option 2: Install from Source

Requires Go 1.21 or later:

```bash
go install github.com/chrisns/snapmaker-cnc-finisher/cmd/gcode-optimizer@latest
```

## Quick Start

### Basic Usage

```bash
gcode-optimizer finishing.cnc 1.0 finishing-optimized.cnc
```

**Arguments**:
- `finishing.cnc`: Your finishing pass GCode file (from Snapmaker Luban)
- `1.0`: Allowance value in mm (material left after rough cut)
- `finishing-optimized.cnc`: Output file path

### Flags

- `--force`: Overwrite output file without confirmation
- `--strategy=<conservative|aggressive>`: Optimization strategy (default: aggressive)

### Example with Flags

```bash
gcode-optimizer --strategy=aggressive --force finishing.cnc 1.0 output.cnc
```

## How It Works

1. **Rough Cut**: You run a rough cut with an allowance (e.g., 1mm of material left)
2. **Optimization**: This tool analyzes your finishing pass GCode and removes moves that only cut material already removed by the rough pass
3. **Finishing Pass**: Run the optimized file to save significant machining time

The tool calculates a depth threshold based on the deepest cut in your finishing pass plus the allowance value. Only moves at or below this threshold are preserved.

## Optimization Strategies

- **Aggressive** (default): Splits moves that cross the threshold at the exact intersection point for maximum time savings
- **Conservative**: Preserves entire moves that cross the threshold for added safety

## Requirements

- Finishing pass GCode files must be produced by Snapmaker Luban
- Rough cut must have been completed with consistent allowance
- Z-axis represents depth (standard for CNC machining)

## Project Status

This project is under active development. See the full specification and implementation plan in the `/specs` directory.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or pull request on GitHub.

## Links

- **Documentation**: See `/specs/001-gcode-finishing-optimizer/quickstart.md` for detailed usage guide
- **Issues**: https://github.com/chrisns/snapmaker-cnc-finisher/issues
- **Repository**: https://github.com/chrisns/snapmaker-cnc-finisher
