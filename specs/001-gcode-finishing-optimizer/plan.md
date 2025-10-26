# Implementation Plan: GCode Finishing Pass Optimizer

**Branch**: `001-gcode-finishing-optimizer-1` | **Date**: 2025-10-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-gcode-finishing-optimizer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a command-line tool in Go that optimizes Snapmaker Luban GCode finishing passes by removing redundant shallow cutting operations already handled by roughing passes. The tool accepts an input GCode file, an allowance value (remaining material thickness), and output path, then produces an optimized GCode file that skips air-cutting moves while preserving the final finishing layer. Supports both 3-axis and 4-axis CNC configurations with two optimization strategies (conservative: preserve crossing moves, aggressive: split moves at threshold).

**Primary Technical Approach**: Parse GCode using `github.com/256dpi/gcode` library, track modal state throughout file processing, calculate depth threshold from minimum Z value + allowance, filter/split G1 cutting moves based on depth relative to threshold, preserve all non-cutting commands and file structure.

## Technical Context

**Language/Version**: Go 1.21
**Primary Dependencies**: `github.com/256dpi/gcode` (v0.3.0) for GCode parsing, stdlib only otherwise (`flag`, `bufio`, `os`, `filepath`)
**Storage**: File I/O (input/output GCode text files)
**Testing**: `go test` with table-driven tests, minimum 80% code coverage
**Target Platform**: Cross-platform (Linux, macOS Intel/ARM, Windows) - static binary distribution
**Project Type**: Single binary CLI tool
**Performance Goals**: Process 100,000-line files in <10 seconds, handle up to 10 million lines without memory issues
**Constraints**: <200ms startup time, streaming/memory-conscious processing for large files, preserve exact GCode format compatibility
**Scale/Scope**: Single-purpose optimization tool, 3-4 core packages, ~1500-2000 LOC estimated

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Cross-Platform Portability ✅
- **Status**: PASS
- **Compliance**: Using Go stdlib abstractions (`filepath`, `os`, `bufio`), no platform-specific calls
- **Evidence**: Feature spec requires macOS, Windows, Linux support; design uses portable file I/O patterns

### Principle II: Static Binary Distribution ✅
- **Status**: PASS
- **Compliance**: Single Go binary with CGO_ENABLED=0, no external runtime dependencies
- **Evidence**: Constitution mandates static linking; gcode library is pure Go, no CGO requirements

### Principle III: Comprehensive Testing (NON-NEGOTIABLE) ✅
- **Status**: PASS
- **Compliance**: TDD workflow required, table-driven tests for move filtering logic, 80% coverage minimum
- **Evidence**: Feature spec defines testable acceptance criteria; constitution requires red-green-refactor cycle

### Principle IV: Open Source & Release Management ✅
- **Status**: PASS
- **Compliance**: Repository at github.com/chrisns/snapmaker-cnc-finisher, GitHub Actions for multi-arch builds
- **Evidence**: Constitution specifies release automation for darwin/amd64, darwin/arm64, windows/amd64, linux/amd64

### Principle V: Go Best Practices ✅
- **Status**: PASS
- **Compliance**: Standard project layout (cmd/, internal/), Go modules, minimal dependencies, gofmt/go vet enforced
- **Evidence**: Constitution mandates idiomatic Go; design follows Effective Go guidelines

### Principle VI: Exhaustive Research During Planning ✅
- **Status**: PASS
- **Compliance**: Conducted WebSearch for Go testing best practices, CLI flag patterns, library documentation
- **Evidence**: Research completed for gcode library API, table-driven testing patterns, Go 1.21/1.22 updates

**Overall Gate Status**: ✅ PASS - All principles satisfied, no violations requiring justification

## Project Structure

### Documentation (this feature)

```text
specs/001-gcode-finishing-optimizer/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── gcode-optimizer/
    └── main.go          # CLI entry point, flag parsing, orchestration

internal/
├── parser/
│   └── parser.go        # GCode file parsing, modal state tracking
├── optimizer/
│   ├── optimizer.go     # Core optimization logic, threshold calculation
│   ├── strategy.go      # Conservative/aggressive strategies
│   └── move.go          # Move splitting, intersection calculations
├── writer/
│   └── writer.go        # GCode output writing, format preservation
└── progress/
    └── progress.go      # Progress reporting, ETA calculation, statistics

tests/
├── integration/
│   ├── cli_test.go      # End-to-end CLI workflow tests
│   └── fixtures/        # Sample GCode files (freya.cnc subset, test cases)
└── unit/
    ├── parser_test.go   # Modal state, header parsing tests
    ├── optimizer_test.go # Threshold calculation, move filtering tests
    └── move_test.go     # Parametric interpolation, split accuracy tests

go.mod                   # Go module definition
go.sum                   # Dependency checksums
README.md                # User documentation, installation, usage
```

**Structure Decision**: Single project layout (Option 1) selected because this is a standalone CLI tool with no web/mobile components. The `cmd/` directory contains the main entry point, `internal/` holds implementation packages (not exposed as public API), and `tests/` separates integration from unit tests. This follows Go best practices for small-to-medium CLI tools and aligns with constitution's simplicity principle.

## Complexity Tracking

*No violations requiring justification - all constitution checks passed.*
