# Snapmaker CNC Finisher

A Go CLI tool that optimizes Snapmaker Luban finishing pass GCode files by removing redundant cutting operations that occur at depths already handled by a rough cut.

## Features

- **Time Savings**: Reduce machining time by 20%+ by eliminating air-cutting operations
- **Multi-Axis Support**: Works with both 3-axis and 4-axis CNC configurations
- **Smart Z-Axis Detection**: Auto-detects Z-axis reference from GCode headers with intelligent fallbacks
- **Configurable Strategies**: Multiple multi-axis move handling strategies (safe, all-axes, split, aggressive)
- **Real-time Progress**: Console progress updates for large files
- **Cross-Platform**: Static binary works on macOS (Intel/ARM), Windows, and Linux

## Installation

### Option 1: Download Pre-Built Binary (Recommended)

Visit the [Releases page](https://github.com/chrisns/snapmaker-cnc-finisher/releases) and download the binary for your platform:

- **macOS Intel**: `snapmaker-cnc-finisher-darwin-amd64`
- **macOS ARM (M1/M2/M3/M4)**: `snapmaker-cnc-finisher-darwin-arm64`
- **Windows**: `snapmaker-cnc-finisher-windows-amd64.exe`
- **Linux**: `snapmaker-cnc-finisher-linux-amd64`

#### Make Executable (macOS/Linux)

```bash
chmod +x snapmaker-cnc-finisher-*
```

#### Optional: Add to PATH

**macOS/Linux**:
```bash
sudo mv snapmaker-cnc-finisher-* /usr/local/bin/snapmaker-cnc-finisher
```

**Windows**:
1. Move the executable to a directory in your PATH (e.g., `C:\Windows\System32`)
2. Or create a new directory and add it to your PATH environment variable

### Option 2: Build from Source

**Requirements**: Go 1.25.3 or later

```bash
# Clone the repository
git clone https://github.com/chrisns/snapmaker-cnc-finisher.git
cd snapmaker-cnc-finisher

# Build
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o snapmaker-cnc-finisher ./cmd/snapmaker-cnc-finisher

# Or install directly
go install github.com/chrisns/snapmaker-cnc-finisher/cmd/snapmaker-cnc-finisher@latest
```

## Quick Start

```bash
# Basic usage
snapmaker-cnc-finisher finishing.cnc 1.0 optimized.cnc

# With force overwrite flag
snapmaker-cnc-finisher finishing.cnc 1.0 optimized.cnc --force

# With custom multi-axis strategy
snapmaker-cnc-finisher finishing.cnc 1.0 optimized.cnc --strategy=aggressive

# Display help
snapmaker-cnc-finisher --help

# Show version
snapmaker-cnc-finisher --version
```

### Arguments

1. **Input file**: Path to the finishing GCode file (produced by Snapmaker Luban)
2. **Allowance**: Remaining material depth after rough cut in mm (e.g., 1.0)
3. **Output file**: Path for the optimized GCode file

### Optional Flags

- `--force`: Overwrite output file without prompting
- `--strategy`: Multi-axis move handling strategy
  - `safe` (default): Preserve entire move if Z exceeds threshold
  - `all-axes`: Preserve only if all axes indicate finishing work
  - `split`: Attempt to split into single-axis commands
  - `aggressive`: Remove entire move if Z is shallow

## How It Works

1. Analyzes the finishing pass GCode to identify Z-axis reference point
2. Compares Z-depth of each cutting operation (G1) against the allowance threshold
3. Removes operations that occur at depths already handled by the rough cut
4. Preserves all non-cutting commands (G0 rapid moves, M-codes, comments, headers)
5. Writes optimized GCode that maintains identical final surface quality

## Requirements

- Finishing GCode files produced by Snapmaker Luban
- Rough cut must have been completed with consistent allowance
- Go 1.21+ (for building from source)

## Performance

- Processes 100k-line GCode files in under 10 seconds
- Supports files up to 10 million lines
- Memory footprint under 200MB

## Development

```bash
# Clone the repository
git clone https://github.com/chrisns/snapmaker-cnc-finisher.git
cd snapmaker-cnc-finisher

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Build
CGO_ENABLED=0 go build -o snapmaker-cnc-finisher ./cmd/snapmaker-cnc-finisher
```

## Contributing

Contributions are welcome! This project follows Test-Driven Development (TDD) and Go best practices.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/chrisns/snapmaker-cnc-finisher.git
cd snapmaker-cnc-finisher

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with race detector
go test ./... -race

# Run linter
go vet ./...

# Format code
go fmt ./...
```

### Before Submitting a PR

Please ensure:
- [ ] All tests pass on all platforms (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No linter warnings (`go vet ./...`)
- [ ] Test coverage is maintained at 80%+ (`go test ./... -cover`)
- [ ] Documentation is updated if needed
- [ ] Commit messages follow conventional commit format

### Testing Requirements

- Write tests before implementation (TDD/Red-Green-Refactor)
- Use table-driven tests where appropriate
- Ensure tests pass on macOS, Windows, and Linux
- Target minimum 80% code coverage

### Project Structure

```
cmd/
└── snapmaker-cnc-finisher/    # Main CLI entry point
internal/
├── gcode/                      # GCode parsing
├── optimizer/                  # Optimization logic
└── cli/                        # CLI interface
tests/
├── unit/                       # Unit tests
├── integration/                # Integration tests
├── contract/                   # Contract tests
└── testdata/                   # Test fixtures
```

### Reporting Issues

When reporting issues, please include:
- Tool version (`snapmaker-cnc-finisher --version`)
- Operating system and architecture
- Full error message or unexpected output
- Sample GCode file (if relevant and possible)
- Steps to reproduce

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- Uses [github.com/256dpi/gcode](https://github.com/256dpi/gcode) for GCode parsing
- Built for the Snapmaker CNC community
