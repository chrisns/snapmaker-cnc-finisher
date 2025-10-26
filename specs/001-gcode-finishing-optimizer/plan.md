# Implementation Plan: GCode Finishing Pass Optimizer

**Branch**: `001-gcode-finishing-optimizer` | **Date**: 2025-10-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-gcode-finishing-optimizer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

A Go CLI tool that optimizes Snapmaker Luban finishing pass GCode files by removing redundant cutting operations that occur at depths already handled by a rough cut. The tool accepts an input file path, allowance threshold (e.g., 1.0mm), and output file path, then generates an optimized GCode file with reduced machining time while preserving final surface quality. Target outcomes include 20%+ machining time reduction, sub-10-second processing for 100k line files, and support for both 3-axis and 4-axis CNC configurations.

## Technical Context

**Language/Version**: Go 1.25.3 (latest stable - October 2025)
**Primary Dependencies**: `github.com/256dpi/gcode` (GCode parsing), stdlib only otherwise (flag, bufio, os, filepath)
**Storage**: File I/O with `bufio.Scanner` streaming (reading input GCode), `bufio.Writer` (writing optimized output)
**Testing**: Go testing framework (`go test`), table-driven tests, benchmarks, race detector
**Target Platform**: Cross-platform CLI (macOS Intel/ARM, Windows, Linux) - static binary distribution (CGO_ENABLED=0)
**Project Type**: Single project (CLI tool)
**Performance Goals**: Process 100k-line GCode files in <10 seconds, support up to 10M lines without memory issues
**Constraints**: Memory-conscious streaming for large files, <200MB memory footprint, static binary with zero runtime dependencies
**Scale/Scope**: Single-purpose CLI tool, ~5-10 core packages, comprehensive test coverage (80%+), multi-platform CI/CD via GitHub Actions matrix

**Research Notes**: All unknowns resolved in [research.md](./research.md). Key decisions: (1) stdlib `flag` chosen over Cobra per minimal dependency principle, (2) `github.com/256dpi/gcode` library for battle-tested parsing, (3) Go 1.25.3 (latest stable as of October 2025), (4) `bufio.Scanner` streaming for memory efficiency.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Cross-Platform Portability
✅ **PASS** - Feature spec explicitly requires cross-platform CLI tool. All file I/O will use Go standard library (`filepath`, `os`, `bufio`). No platform-specific APIs planned.

### Principle II: Static Binary Distribution
✅ **PASS** - Feature aligns perfectly with static binary requirement. Single executable with zero runtime dependencies. CGO_ENABLED=0 enforced.

### Principle III: Comprehensive Testing (TDD)
✅ **PASS** - Spec defines testable success criteria (SC-001 through SC-008). TDD workflow mandatory per constitution. Table-driven tests planned for GCode parsing, depth filtering, edge cases. Target: 80%+ coverage.

### Principle IV: Open Source & Release Management
✅ **PASS** - Repository: `github.com/chrisns/snapmaker-cnc-finisher`. Multi-arch releases planned (macOS Intel/ARM, Windows, Linux). GitHub Actions CI/CD required for all platforms.

### Principle V: Go Best Practices
✅ **PASS** - Project will follow standard Go layout (cmd/, internal/, pkg/), use Go modules, minimize dependencies, follow Effective Go guidelines. `gofmt` and `go vet` enforced via CI.

### Principle VI: Exhaustive Research During Planning
✅ **PASS** - Phase 0 research completed with exhaustive use of Context7 (spf13/cobra docs), WebSearch (5 queries: Go versions, GCode libraries, CLI best practices, file streaming, CI/CD patterns), and Perplexity (3 queries: CLI framework tradeoffs, GCode parsing strategies, Go 1.25.3 features). All decisions documented with citations in [research.md](./research.md).

**Post-Design Gate Result**: ✅ **PASS** - No violations detected. All design decisions align with constitution principles:
- Data model uses pure Go types (Principle I: cross-platform)
- CLI interface contract enforces static binary (Principle II)
- Comprehensive test plan defined (Principle III: TDD workflow)
- GitHub Actions matrix CI specified (Principle IV: automation)
- Stdlib `flag` + single dependency `github.com/256dpi/gcode` (Principle V: minimal deps)
- Research citations embedded throughout planning artifacts (Principle VI: exhaustive research)

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
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
└── snapmaker-cnc-finisher/     # Main CLI entry point
    └── main.go

internal/
├── gcode/                       # GCode parsing and manipulation
│   ├── parser.go               # GCode line parser
│   ├── command.go              # GCode command types (G0, G1, M-codes)
│   ├── file.go                 # File reading/writing with streaming
│   └── metadata.go             # Header metadata extraction (axis config, Z reference)
├── optimizer/                   # Optimization logic
│   ├── filter.go               # Depth-based filtering
│   ├── strategy.go             # Multi-axis move strategies (safe/all-axes/split/aggressive)
│   └── stats.go                # Statistics tracking (lines removed, time savings)
└── cli/                         # CLI argument parsing and UI
    ├── args.go                 # Command-line argument parser
    ├── progress.go             # Progress reporting
    └── output.go               # Console output formatting

pkg/
└── (empty initially - public APIs if needed for extensions)

tests/
├── testdata/                   # Sample GCode files for testing
│   ├── finishing_3axis.cnc
│   ├── finishing_4axis.cnc
│   ├── malformed_header.cnc
│   └── large_file.cnc
├── unit/                       # Unit tests (mirror internal/ structure)
│   ├── gcode/
│   ├── optimizer/
│   └── cli/
├── integration/                # End-to-end CLI tests
│   └── cli_test.go
└── contract/                   # GCode format contract tests
    └── snapmaker_format_test.go

.github/
└── workflows/
    ├── ci.yml                  # Lint, test (multi-platform), build
    └── release.yml             # Multi-arch binary releases

go.mod
go.sum
README.md
LICENSE
```

**Structure Decision**: Using **Option 1: Single project** layout with Go standard project structure. CLI tool fits perfectly into `cmd/` + `internal/` + `pkg/` pattern per Principle V (Go Best Practices). `internal/` ensures encapsulation of implementation details, `cmd/` provides clean entry point, and `tests/` mirrors Go community conventions for comprehensive testing per Principle III.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. Constitution check passed all principles.
