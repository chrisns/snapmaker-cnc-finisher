# Implementation Tasks: GCode Finishing Pass Optimizer

**Feature Branch**: `001-gcode-finishing-optimizer`
**Generated**: 2025-10-26
**Status**: Ready for Implementation

---

## Overview

This task list implements the GCode Finishing Pass Optimizer using **Test-Driven Development (TDD)** per Constitution Principle III. Tasks are organized by user story to enable independent, incremental delivery.

**Total Tasks**: 95
**User Stories**: 3 (P1, P2, P3)
**Estimated MVP Scope**: User Story 1 only (47 tasks)

---

## Implementation Strategy

### TDD Workflow (Mandatory)

Per Constitution Principle III, follow **Red-Green-Refactor** strictly:

1. **RED**: Write test first, verify it fails
2. **GREEN**: Implement minimal code to make test pass
3. **REFACTOR**: Clean up code while keeping tests green

Every exported function/type requires:
- Unit test (table-driven where appropriate)
- Minimum 80% code coverage
- Tests MUST pass on all platforms (macOS, Windows, Linux)

### Delivery Strategy

**MVP First**: Complete User Story 1 (P1) for initial release:
- Core optimization logic
- Basic CLI interface
- File I/O with streaming
- Contract tests for CLI

**Incremental**: Add User Story 2 (P2), then User Story 3 (P3) in subsequent releases.

**Parallel Opportunities**: Tasks marked `[P]` can run concurrently (independent files).

---

## Phase 1: Project Setup

**Goal**: Initialize Go project structure, dependencies, and CI/CD.

### Tasks

- [x] T001 Initialize Go module with `go mod init github.com/chrisns/snapmaker-cnc-finisher`
- [x] T002 Create project directory structure (cmd/, internal/, tests/, .github/workflows/)
- [x] T003 Add dependency `github.com/256dpi/gcode` via `go get github.com/256dpi/gcode`
- [x] T004 [P] Create `.gitignore` for Go (binaries, IDE files, test coverage)
- [x] T005 [P] Create `LICENSE` file (per Constitution Principle IV - open source)
- [x] T006 [P] Create initial `README.md` with project description and quick start
- [x] T007 [P] Create GitHub Actions CI workflow at `.github/workflows/ci.yml` (lint, test matrix: macOS/Windows/Linux)
- [x] T008 [P] Create GitHub Actions release workflow at `.github/workflows/release.yml` (multi-arch builds per Constitution Principle IV)

**Completion Criteria**: `go build` succeeds, CI pipeline runs successfully, project structure matches plan.md.

---

## Phase 2: Foundational Layer

**Goal**: Implement blocking prerequisites used by all user stories (GCode parsing, file I/O primitives).

### Tasks

- [x] T009 **TEST**: Create test for GCode command parsing in `tests/unit/gcode/command_test.go` (parse "G1 X10.5 Y20.3 Z-1.2 F1500")
- [x] T010 Implement GCode command struct in `internal/gcode/command.go` (Letter, Value, Comment fields)
- [x] T011 **TEST**: Create test for GCode line parsing in `tests/unit/gcode/parser_test.go` (table-driven: comments, G0, G1, M-codes)
- [x] T012 Implement GCode line parser in `internal/gcode/parser.go` using `github.com/256dpi/gcode.ParseLine`
- [x] T013 **TEST**: Create test for file header metadata extraction in `tests/unit/gcode/metadata_test.go` (MinZ/MaxZ, axis config)
- [x] T014 Implement header metadata extraction in `internal/gcode/metadata.go` (scan first 50 lines for `;MIN_Z`, detect B-axis)
- [x] T014b **TEST**: Create test for Z-axis reference fallback chain in `tests/unit/gcode/metadata_test.go` (table-driven: valid header → use metadata; missing min_z/max_z → fallback to machine origin; no metadata → fallback to surface convention Z=0. Verify console alert indicates which method was used)
- [x] T015 **TEST**: Create test for buffered file reading in `tests/unit/gcode/file_test.go` (bufio.Scanner streaming)
- [x] T016 Implement file reading with streaming in `internal/gcode/file.go` (open, scan lines, handle errors)
- [x] T017 **TEST**: Create test for buffered file writing in `tests/unit/gcode/file_test.go` (bufio.Writer)
- [x] T018 Implement file writing with buffering in `internal/gcode/file.go` (create, write lines, flush). Flush strategy: flush every 1000 lines OR on completion. Use deferred flush to ensure cleanup on error.

**Completion Criteria**: All foundational tests pass, file I/O streams large files without loading into memory, header metadata correctly extracted from test fixtures.

---

## Phase 3: User Story 1 - Basic GCode Optimization (P1)

**Goal**: Core value proposition - remove redundant cutting operations to reduce machining time.

**Independent Test**: Process a finishing GCode file with 1.0mm allowance, verify output has fewer G1 commands and preserved G0/M-codes.

### Test Data Setup

- [x] T019 [P] Create test fixture `tests/testdata/finishing_3axis.cnc` (100 lines, mixed shallow/deep G1 commands)
- [x] T020 [P] Create test fixture `tests/testdata/finishing_4axis.cnc` (includes B-axis commands)
- [x] T021 [P] Create test fixture `tests/testdata/all_shallow.cnc` (all G1 Z-values < 1.0mm for max filtering test)
- [x] T022 [P] Create test fixture `tests/testdata/all_deep.cnc` (all G1 Z-values > 1.0mm for min filtering test)

### Core Optimization Logic

- [ ] T023 **TEST**: Create test for Z-depth comparison in `tests/unit/optimizer/filter_test.go` (table-driven: Z=-0.5 vs 1.0mm allowance → filter)
- [ ] T024 Implement depth filtering logic in `internal/optimizer/filter.go` (compare Z-value vs allowance threshold)
- [ ] T025 **TEST**: Create test for G0/M-code preservation in `tests/unit/optimizer/filter_test.go` (rapid moves and machine codes always retained, including table-driven test: "G0 Z-0.2 (shallow rapid) → KEEP (rapid moves always preserved even if shallow)")
- [ ] T026 Implement command type detection in `internal/optimizer/filter.go` (IsRapidMove, IsCuttingMove, IsMachineCode)
- [ ] T027 **TEST**: Create test for strategy enum in `tests/unit/optimizer/strategy_test.go` (parse "safe"/"all-axes"/"split"/"aggressive")
- [ ] T028 Implement FilterStrategy enum in `internal/optimizer/strategy.go` (Safe, AllAxes, Split, Aggressive constants)
- [ ] T029 **TEST**: Create test for multi-axis move filtering in `tests/unit/optimizer/filter_test.go` (safe strategy: preserve if Z deep)
- [ ] T030 Implement multi-axis move filtering in `internal/optimizer/filter.go` (apply strategy to G1 commands with multiple axes)

### Statistics Tracking

- [ ] T031 **TEST**: Create test for statistics tracking in `tests/unit/optimizer/stats_test.go` (accumulate removed lines, calculate percentages)
- [ ] T032 Implement Statistics struct in `internal/optimizer/stats.go` (TotalLines, RemovedLines, BytesIn/Out, EstimatedTimeSaved)
- [ ] T033 **TEST**: Create test for time savings calculation in `tests/unit/optimizer/stats_test.go` (distance ÷ feed rate for removed G1 moves)
- [ ] T034 Implement time savings calculator in `internal/optimizer/stats.go` (Euclidean distance, accumulate time per removed move)

### CLI Interface

- [ ] T035 **TEST**: Create contract test for CLI arguments in `tests/contract/cli_contract_test.go` (3 args required, --force and --strategy flags)
- [ ] T036 Implement CLI argument parser in `internal/cli/args.go` using stdlib `flag` package (parse input/allowance/output, --force, --strategy)
- [ ] T037 **TEST**: Create test for argument validation in `tests/unit/cli/args_test.go` (non-numeric allowance → error, negative allowance → error)
- [ ] T038 Implement argument validation in `internal/cli/args.go` (validate file exists, allowance >= 0, strategy is valid enum)
- [ ] T039 **TEST**: Create test for output formatting in `tests/unit/cli/output_test.go` (final summary format, error message format)
- [ ] T040 Implement console output formatter in `internal/cli/output.go` (PrintSummary, PrintError, exit codes per CLI contract)
- [ ] T040b **TEST**: Create test for feed rate preservation in `tests/unit/optimizer/filter_test.go` (verify F parameter values preserved in kept G1 commands)

### Main Entry Point

- [ ] T041 **TEST**: Create integration test for end-to-end CLI in `tests/integration/cli_test.go` (run binary, verify output file created)
- [ ] T042 Implement main.go in `cmd/snapmaker-cnc-finisher/main.go` (wire together: parse args, read file, filter, write output, print summary)
- [ ] T043 **TEST**: Run contract test suite ensuring all CLI interface requirements met (exit codes, help text, error messages)

### Validation

- [ ] T044 Run `go test ./... -race -cover` and verify 80%+ coverage
- [ ] T045 Run `gofmt` and `go vet` ensuring no warnings (per Constitution Principle V)
- [ ] T046 Build static binary with `CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath` and test on all platforms via CI
- [ ] T047 **Acceptance Test**: Process `tests/testdata/finishing_3axis.cnc` with 1.0mm allowance, verify SC-001 (< 10 seconds), SC-002 (20%+ reduction)

**User Story 1 Completion Criteria**:
- ✅ Can process GCode files and produce optimized output
- ✅ Removes G1 commands at shallow depths (< allowance)
- ✅ Preserves G0 (rapid moves), M-codes, comments, header
- ✅ Supports 3-axis and 4-axis GCode
- ✅ Implements all 4 multi-axis strategies (safe/all-axes/split/aggressive)
- ✅ CLI accepts 3 arguments + 2 flags (--force, --strategy)
- ✅ Displays final statistics (lines removed, time saved, file size reduction)
- ✅ All tests pass on macOS, Windows, Linux
- ✅ Code coverage >= 80%
- ✅ Static binary builds successfully

---

## Phase 4: User Story 2 - Progress Monitoring (P2)

**Goal**: Real-time progress updates for large file processing.

**Independent Test**: Process a large GCode file and observe console shows progress every 10k lines or 2 seconds.

### Test Data Setup

- [ ] T048 [P] Create fixture generator script `tests/scripts/generate_large_gcode.go` (generates 10M+ line GCode file with realistic commands)
- [ ] T049 [P] Run generator to create test fixture `tests/testdata/large_file.cnc` (10M+ lines for progress testing)

### Progress Reporting

- [ ] T050 **TEST**: Create test for progress update timing in `tests/unit/cli/progress_test.go` (updates every 10k lines or 2s)
- [ ] T051 Implement progress tracker in `internal/cli/progress.go` (track elapsed time, lines processed, calculate % complete)
- [ ] T052 **TEST**: Create test for progress display format in `tests/unit/cli/progress_test.go` (single-line overwrite with `\r`)
- [ ] T053 Implement progress display in `internal/cli/progress.go` (format: "Processing: 45,230 / 100,000 lines (45.2%) | Removed: 12,450 | Elapsed: 3.2s | ETA: 3.8s")
- [ ] T054 **TEST**: Create integration test for progress updates in `tests/integration/cli_test.go` (capture stdout during large file processing)
- [ ] T055 Integrate progress reporting into main.go (call progress update every 10k lines or 2s during file processing loop)

### Validation

- [ ] T057 **Acceptance Test**: Process `tests/testdata/large_file.cnc` (10M lines), verify SC-005 (progress updates every 10k lines or 2s), SC-006 (handles without crashing)

**User Story 2 Completion Criteria**:
- ✅ Console displays real-time progress during processing
- ✅ Progress updates at minimum every 10k lines or 2 seconds
- ✅ Final summary shows total lines removed and file size reduction
- ✅ Large files (10M+ lines) process without memory issues
- ✅ All tests pass

---

## Phase 5: User Story 3 - Error Handling (P3)

**Goal**: Clear, actionable error messages for invalid inputs.

**Independent Test**: Provide various invalid inputs and verify appropriate error messages are displayed.

### Test Data Setup

- [ ] T058 [P] Create test fixture `tests/testdata/malformed_header.cnc` (missing Snapmaker header)
- [ ] T059 [P] Create test fixture `tests/testdata/unparseable.cnc` (binary file or corrupted GCode)
- [ ] T060 [P] Create test fixture `tests/testdata/no_feed_rate.cnc` (G1 commands missing F parameter)

### Error Validation

- [ ] T061 **TEST**: Create test for input file validation in `tests/unit/cli/args_test.go` (non-existent file → error message)
- [ ] T062 Implement input file existence check in `internal/cli/args.go` (return clear error if file not found)
- [ ] T063 **TEST**: Create test for output file validation in `tests/unit/cli/args_test.go` (unwritable directory → error message)
- [ ] T064 Implement output file writability check in `internal/cli/args.go` (verify parent dir exists and is writable)
- [ ] T065 **TEST**: Create test for GCode parsing errors in `tests/unit/gcode/parser_test.go` (malformed line → skip with warning)
- [ ] T066 Implement GCode parsing error handling in `internal/gcode/parser.go` (log warning for unparseable lines, continue processing)
- [ ] T067 **TEST**: Create test for header validation in `tests/unit/gcode/metadata_test.go` (missing header → warning but proceed)
- [ ] T068 Implement header validation in `internal/gcode/metadata.go` (issue console warning if Snapmaker header missing/malformed)
- [ ] T069 **TEST**: Create test for feed rate fallback in `tests/unit/optimizer/stats_test.go` (missing F → use 1000 mm/min default with warning)
- [ ] T070 Implement feed rate fallback logic in `internal/optimizer/stats.go` (track last known feed rate, use default if never specified)

### File Overwrite Handling

- [ ] T071 **TEST**: Create test for file overwrite prompt in `tests/unit/cli/args_test.go` (existing file without --force → prompt user)
- [ ] T072 Implement file overwrite confirmation in `internal/cli/args.go` (prompt user "Overwrite? (y/N):", respect --force flag)
- [ ] T073 **TEST**: Create integration test for overwrite behavior in `tests/integration/cli_test.go` (--force bypasses prompt)

### Error Message Contract

- [ ] T074 **TEST**: Create contract test for error messages in `tests/contract/cli_contract_test.go` (verify all error message formats match CLI contract)
- [ ] T075 Implement error message formatter in `internal/cli/output.go` (format: "ERROR: <description> [<hint>]", use correct exit codes)

### Validation

- [ ] T076 **Acceptance Test**: Test all error scenarios from spec.md Edge Cases, verify SC-007 (95% have clear messages)

**User Story 3 Completion Criteria**:
- ✅ Non-existent input file → clear error message (exit code 2)
- ✅ Invalid allowance → clear error message (exit code 1)
- ✅ Malformed GCode → warning but continues processing
- ✅ Missing feed rate → uses default (1000 mm/min) with warning
- ✅ Output file exists → prompts for confirmation (or bypasses with --force)
- ✅ Unparseable file → error message (exit code 2)
- ✅ All error messages actionable and helpful
- ✅ All tests pass

---

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Complete remaining requirements (help text, version info, documentation).

### CLI Polish

- [ ] T076 **TEST**: Create test for `--help` flag in `tests/contract/cli_contract_test.go` (displays usage, exits with code 0)
- [ ] T077 Implement `--help` flag handler in `internal/cli/args.go` (print help text from CLI contract, exit)
- [ ] T078 **TEST**: Create test for `--version` flag in `tests/contract/cli_contract_test.go` (displays version info, exits with code 0)
- [ ] T079 Implement `--version` flag handler in `internal/cli/args.go` (print version, Go version, platform, exit)

### Documentation

- [ ] T080 [P] Update README.md with installation instructions, usage examples, contribution guidelines
- [ ] T081 [P] Add CHANGELOG.md for v1.0.0 release notes
- [ ] T082 [P] Verify quickstart.md examples match actual CLI behavior

### Performance & Benchmarking

- [ ] T083 **TEST**: Create benchmark test in `tests/unit/optimizer/filter_bench_test.go` (benchmark filtering 100k lines)
- [ ] T084 Run benchmarks, verify SC-001 (< 10s for 100k lines), optimize if needed
- [ ] T085 **TEST**: Create benchmark for memory usage in `tests/unit/gcode/file_bench_test.go` (verify < 200MB for 10M lines)
- [ ] T086 Profile memory usage with `go test -memprofile`, verify SC-006 compliance

### Code Quality

- [ ] T087 Run `golangci-lint run --enable=godot,godox` (enforce GoDoc comments on exported symbols, flag TODO/FIXME)
- [ ] T088 Review `golangci-lint` output from T087, fix any missing GoDoc comments on exported functions
- [ ] T089 Run full test suite with race detector on all platforms via CI (`go test -race ./...`)
- [ ] T090 Generate code coverage report, ensure >= 80% per Constitution Principle III

### Release Preparation

- [ ] T091 Tag v1.0.0 release, trigger GitHub Actions release workflow
- [ ] T092 Verify multi-arch binaries build successfully (darwin/amd64, darwin/arm64, windows/amd64, linux/amd64)
- [ ] T093 Test downloaded binaries on each platform (smoke test)
- [ ] T094 Publish release notes with CHANGELOG, link to binaries

**Phase 6 Completion Criteria**:
- ✅ Help and version flags work correctly
- ✅ README contains complete documentation
- ✅ All benchmarks meet performance targets
- ✅ Code coverage >= 80%
- ✅ Multi-arch release builds successfully
- ✅ All success criteria (SC-001 through SC-008) validated

---

## Task Dependency Graph

### User Story Completion Order

```
Phase 1 (Setup)
     ↓
Phase 2 (Foundational)
     ↓
     ├─→ User Story 1 (P1) - Core Optimization ← **MVP RELEASE**
     │        ↓
     ├─→ User Story 2 (P2) - Progress Monitoring
     │        ↓
     └─→ User Story 3 (P3) - Error Handling
              ↓
         Phase 6 (Polish) ← **v1.0.0 RELEASE**
```

**Key Dependencies**:
- **User Story 1** MUST complete before User Story 2 (progress requires core processing loop)
- **User Story 1** MUST complete before User Story 3 (error handling enhances core functionality)
- **User Story 2 and 3** can run in parallel (independent concerns)

### Within-Story Dependencies

**User Story 1 (P1)**:
```
Test Data (T019-T022) [PARALLEL]
     ↓
Optimization Logic (T023-T030) [SEQUENTIAL: filter → strategy → multi-axis]
     ↓
Statistics (T031-T034) [SEQUENTIAL: struct → calculator]
     ↓
CLI (T035-T040) [SEQUENTIAL: args → validation → output]
     ↓
Main (T041-T047) [SEQUENTIAL: integration → validation → acceptance]
```

**User Story 2 (P2)**:
```
Test Data (T048-T049) [PARALLEL: generator + run]
     ↓
Progress Logic (T050-T053) [SEQUENTIAL: tracker → display]
     ↓
Integration (T054-T056) [SEQUENTIAL: integrate → test → acceptance]
```

**User Story 3 (P3)**:
```
Test Data (T058-T060) [PARALLEL]
     ↓
Error Validation (T061-T070) [MOSTLY PARALLEL]
     ↓
Overwrite Handling (T071-T073) [SEQUENTIAL]
     ↓
Error Contract (T074-T076) [SEQUENTIAL]
```

---

## Parallel Execution Examples

### Phase 1: Setup (All Parallel)
Run concurrently:
- T004 (gitignore)
- T005 (LICENSE)
- T006 (README)
- T007 (CI workflow)
- T008 (Release workflow)

### Phase 2: Foundational (Test-Implement Pairs)
**Iteration 1**: T009 (test) → T010 (impl)
**Iteration 2**: T011 (test) → T012 (impl)
**Iteration 3**: T013 (test) → T014 (impl)
... continue pattern

### User Story 1: Test Data (All Parallel)
Run concurrently:
- T019 (finishing_3axis.cnc)
- T020 (finishing_4axis.cnc)
- T021 (all_shallow.cnc)
- T022 (all_deep.cnc)

### User Story 2: Test Data (Parallel)
Run concurrently:
- T048 (generate_large_gcode.go script)
- T049 (run generator for large_file.cnc)

### User Story 3: Test Data (All Parallel)
Run concurrently:
- T058 (malformed_header.cnc)
- T059 (unparseable.cnc)
- T060 (no_feed_rate.cnc)

---

## Success Metrics Validation

Each user story phase MUST validate its success criteria:

### User Story 1 Validation
- **SC-001**: Processing time < 10s for 100k lines (T047)
- **SC-002**: 20%+ time reduction (T047)
- **SC-003**: Identical surface quality (manual test cut)
- **SC-004**: 3-axis and 4-axis support (T047 with fixtures)

### User Story 2 Validation
- **SC-005**: Progress updates every 10k lines / 2s (T056)
- **SC-006**: Handles 10M lines without crash (T056)

### User Story 3 Validation
- **SC-007**: 95% clear error messages (T076)

### Cross-Cutting Validation
- **SC-008**: 15-40% file size reduction (T047)
- **Code Coverage**: >= 80% (T091)
- **Constitution Compliance**: All 6 principles (validated throughout)

---

## MVP Scope Definition

**Minimum Viable Product** = User Story 1 only (47 tasks: T001-T047)

**MVP Delivers**:
- ✅ Core optimization functionality
- ✅ CLI interface (3 args, 2 flags)
- ✅ Multi-axis strategy support
- ✅ Statistics reporting
- ✅ Cross-platform static binary
- ✅ 80%+ test coverage

**Deferred to Post-MVP**:
- Real-time progress updates (User Story 2)
- Enhanced error handling (User Story 3)
- Help/version flags (Phase 6)
- Comprehensive documentation (Phase 6)

**Rationale**: User Story 1 provides immediate value (20% time savings), is independently testable, and validates core technical approach. User Stories 2-3 enhance UX but aren't blockers for initial adoption.

---

## Testing Strategy Summary

**Test Distribution** (95 tasks):
- **Test Tasks**: 27 (28% of total)
- **Implementation Tasks**: 65 (68% of total)
- **Setup/Validation Tasks**: 3 (3% of total)

**Test Types**:
- **Unit Tests**: 20 tasks (table-driven, isolated)
- **Contract Tests**: 4 tasks (CLI interface validation)
- **Integration Tests**: 3 tasks (end-to-end CLI execution)
- **Acceptance Tests**: 3 tasks (user story validation)
- **Benchmarks**: 2 tasks (performance validation)

**TDD Workflow Enforcement**:
- Every implementation file has corresponding test file
- Tests written BEFORE implementation (strict Red-Green-Refactor)
- Tests run on all platforms via GitHub Actions matrix
- Coverage target: >= 80% (Constitution Principle III)

---

## Platform Testing Requirements

Per Constitution Principle I, all tasks MUST pass on:
- **macOS Intel** (darwin/amd64)
- **macOS ARM** (darwin/arm64)
- **Windows** (windows/amd64)
- **Linux** (linux/amd64)

**CI Configuration** (T007):
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.25.3']
```

**Pre-Merge Gate**: All tests MUST pass on all platforms before merging to main.

---

## File Path Reference

Quick lookup of file paths by task:

### Go Modules & Config
- T001: `go.mod`
- T004: `.gitignore`
- T005: `LICENSE`
- T006: `README.md`
- T007: `.github/workflows/ci.yml`
- T008: `.github/workflows/release.yml`

### Core Implementation (`internal/`)
- T010, T026: `internal/gcode/command.go`
- T012, T066: `internal/gcode/parser.go`
- T014, T068: `internal/gcode/metadata.go`
- T016, T018: `internal/gcode/file.go`
- T024, T026, T030: `internal/optimizer/filter.go`
- T028: `internal/optimizer/strategy.go`
- T032, T034, T070: `internal/optimizer/stats.go`
- T036, T038, T062, T064, T072, T078, T080: `internal/cli/args.go`
- T040, T075: `internal/cli/output.go`
- T051, T053: `internal/cli/progress.go`

### Tests (`tests/`)
- T009: `tests/unit/gcode/command_test.go`
- T011, T065: `tests/unit/gcode/parser_test.go`
- T013, T067: `tests/unit/gcode/metadata_test.go`
- T015, T017, T086: `tests/unit/gcode/file_test.go`
- T023, T025, T029: `tests/unit/optimizer/filter_test.go`
- T027: `tests/unit/optimizer/strategy_test.go`
- T031, T033, T069: `tests/unit/optimizer/stats_test.go`
- T035, T074: `tests/contract/cli_contract_test.go`
- T037, T061, T063, T071: `tests/unit/cli/args_test.go`
- T039: `tests/unit/cli/output_test.go`
- T050, T052: `tests/unit/cli/progress_test.go`
- T041, T054, T073: `tests/integration/cli_test.go`
- T084: `tests/unit/optimizer/filter_bench_test.go`

### Test Fixtures (`tests/testdata/`)
- T019: `tests/testdata/finishing_3axis.cnc`
- T020: `tests/testdata/finishing_4axis.cnc`
- T021: `tests/testdata/all_shallow.cnc`
- T022: `tests/testdata/all_deep.cnc`
- T048: `tests/scripts/generate_large_gcode.go` (generator script)
- T049: `tests/testdata/large_file.cnc` (generated by T048)
- T058: `tests/testdata/malformed_header.cnc`
- T059: `tests/testdata/unparseable.cnc`
- T060: `tests/testdata/no_feed_rate.cnc`

### Main Entry Point
- T042: `cmd/snapmaker-cnc-finisher/main.go`

---

## Next Steps

1. **Review this task list** with stakeholders
2. **Begin User Story 1 implementation** (T001-T047 for MVP)
3. **Follow TDD strictly**: Red → Green → Refactor
4. **Run CI on every commit**: Ensure cross-platform compatibility
5. **Release MVP** after T047 completes and all acceptance tests pass
6. **Iterate**: Add User Story 2, then User Story 3 in subsequent releases

---

**Generated By**: Claude Code (Sonnet 4.5)
**References**: [spec.md](./spec.md), [plan.md](./plan.md), [data-model.md](./data-model.md), [contracts/cli-interface.md](./contracts/cli-interface.md)
