# Tasks: GCode Finishing Pass Optimizer

**Input**: Design documents from `/specs/001-gcode-finishing-optimizer/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), contracts/ (complete)

**Tests**: This feature uses Test-Driven Development (TDD) per Constitution Principle III (NON-NEGOTIABLE). All test tasks are MANDATORY and must be completed BEFORE implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Per plan.md, this is a single binary CLI tool:
- Source: `internal/` (private packages), `cmd/gcode-optimizer/` (main entry point)
- Tests: `tests/unit/`, `tests/integration/`
- Root: `go.mod`, `README.md`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project directory structure: `cmd/gcode-optimizer/`, `internal/{parser,optimizer,writer,progress}/`, `tests/{unit,integration,integration/fixtures}/`
- [X] T002 Initialize Go module in go.mod with `module github.com/chrisns/snapmaker-cnc-finisher`
- [X] T003 [P] Add dependency `github.com/256dpi/gcode v0.3.0` to go.mod
- [X] T004 [P] Create README.md with basic project description and installation instructions
- [X] T005 [P] Create .gitignore with Go-specific entries (vendor/, *.exe, *.test, coverage.out)
- [X] T006 [P] Create test fixtures directory and add freya-subset.cnc (first 1000 lines of freya.cnc for fast tests)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 Implement HeaderMetadata struct in internal/parser/parser.go with all fields from data-model.md
- [X] T008 Implement ModalState struct in internal/parser/parser.go with X, Y, Z, B, F fields
- [X] T009 [P] Implement MoveClassification enum (Shallow, Deep, CrossingEnter, CrossingLeave, NonCutting) in internal/optimizer/optimizer.go
- [X] T010 [P] Implement OptimizationStrategy enum (Conservative, Aggressive) in internal/optimizer/optimizer.go
- [X] T011 [P] Implement IntersectionPoint struct in internal/optimizer/optimizer.go
- [X] T012 [P] Implement OptimizationResult struct in internal/progress/progress.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic GCode Optimization (Priority: P1) 🎯 MVP

**Goal**: Optimize finishing pass GCode by removing redundant shallow cutting operations, supporting both conservative and aggressive strategies, handling threshold-crossing moves correctly.

**Independent Test**: Provide freya-subset.cnc and 1.0mm allowance, verify output has fewer lines and correct move filtering/splitting based on strategy.

### Tests for User Story 1 (TDD - MANDATORY) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T013 [P] [US1] Unit test for header parsing in tests/unit/parser_test.go - verify HeaderMetadata extraction from Snapmaker Luban header
- [X] T014 [P] [US1] Unit test for modal state initialization in tests/unit/parser_test.go - verify Z initialized from max_z, X/Y/B default to 0
- [X] T015 [P] [US1] Unit test for modal state updates in tests/unit/parser_test.go - verify coordinates update only when specified, others persist
- [X] T016 [P] [US1] Unit test for ScanMinZ in tests/unit/parser_test.go - verify deepest Z value found in G1 commands
- [X] T017 [P] [US1] Unit test for ClassifyMove in tests/unit/optimizer_test.go - table-driven tests for Shallow, Deep, CrossingEnter, CrossingLeave classifications
- [X] T018 [P] [US1] Unit test for CalculateIntersection in tests/unit/move_test.go - verify parametric interpolation accuracy, edge cases (horizontal move, division by zero)
- [X] T019 [P] [US1] Unit test for SplitMove (aggressive) in tests/unit/move_test.go - verify correct G1 commands generated, feed rate preserved, 3-4 decimal precision
- [X] T020 [P] [US1] Unit test for ShouldPreserve (conservative vs aggressive) in tests/unit/optimizer_test.go - verify strategy-specific behavior
- [X] T021 [P] [US1] Integration test for end-to-end optimization (aggressive strategy) in tests/integration/cli_test.go - verify file with crossing moves produces correct output
- [X] T022 [P] [US1] Integration test for end-to-end optimization (conservative strategy) in tests/integration/cli_test.go - verify crossing moves preserved entirely
- [X] T023 [P] [US1] Integration test for 3-axis vs 4-axis detection in tests/integration/cli_test.go - verify is_rotate header flag handling

### Implementation for User Story 1

- [X] T024 [P] [US1] Implement ParseFile function in internal/parser/parser.go to parse GCode file from io.Reader using gcode.ParseFile
- [X] T025 [P] [US1] Implement parseHeader function in internal/parser/parser.go to extract HeaderMetadata from comment lines (;key: value format)
- [X] T026 [US1] Implement Parser.ResetState in internal/parser/parser.go to initialize ModalState from HeaderMetadata (Z from max_z, others 0)
- [X] T027 [US1] Implement Parser.UpdateState in internal/parser/parser.go to update ModalState from gcode.Line based on X/Y/Z/B/F codes present
- [X] T028 [US1] Implement Parser.ScanMinZ in internal/parser/parser.go to find minimum Z value across all G1 commands in file
- [X] T029 [P] [US1] Implement NewOptimizer function in internal/optimizer/optimizer.go with minZ, allowance, strategy parameters
- [X] T030 [US1] Implement Optimizer.ClassifyMove in internal/optimizer/optimizer.go using startZ, endZ comparison to threshold
- [X] T031 [US1] Implement Optimizer.CalculateIntersection in internal/optimizer/optimizer.go using parametric linear interpolation formula from data-model.md
- [X] T032 [US1] Implement Optimizer.ShouldPreserve in internal/optimizer/optimizer.go returning true/false based on classification and strategy
- [X] T033 [US1] Implement Optimizer.SplitMove in internal/optimizer/optimizer.go to generate two gcode.Line structs at intersection point with correct coordinates and feed rate
- [X] T034 [P] [US1] Implement Writer.WriteFile in internal/writer/writer.go to output gcode.File using gcode.WriteFile, preserving header
- [X] T035 [US1] Implement main CLI entry point in cmd/gcode-optimizer/main.go with flag parsing (input, allowance, output, --force, --strategy)
- [X] T036 [US1] Implement input validation in cmd/gcode-optimizer/main.go - file existence, allowance numeric/non-negative, strategy enum check
- [X] T037 [US1] Implement optimization pipeline in cmd/gcode-optimizer/main.go: parse → scan min_z → calculate threshold → classify/filter moves → write output
- [X] T038 [US1] Add output file overwrite confirmation prompt in cmd/gcode-optimizer/main.go (skip if --force flag present)
- [X] T039 [US1] Add console message displaying detected min_z and calculated threshold per FR-004

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently - basic optimization works with both strategies

---

## Phase 4: User Story 2 - Progress Monitoring (Priority: P2)

**Goal**: Provide real-time progress updates during optimization showing lines processed, ETA, and final statistics including time savings.

**Independent Test**: Run tool on large test file (freya-subset.cnc duplicated 10x), observe console progress updates every 2 seconds OR 10k lines, verify final statistics display.

### Tests for User Story 2 (TDD - MANDATORY) ⚠️

- [X] T040 [P] [US2] Unit test for ProgressReporter.Update in tests/unit/progress_test.go - verify update frequency (2s OR 10k lines criteria)
- [X] T041 [P] [US2] Unit test for ETA calculation in tests/unit/progress_test.go - verify formula: (elapsed / processed) × (total - processed)
- [X] T042 [P] [US2] Unit test for OptimizationResult calculations in tests/unit/progress_test.go - verify reduction %, time savings, lines/sec
- [X] T043 [P] [US2] Integration test for progress display in tests/integration/cli_test.go - verify console updates during processing of large file

### Implementation for User Story 2

- [X] T044 [P] [US2] Implement ProgressReporter struct in internal/progress/progress.go with totalLines, processedLines, startTime, lastUpdate fields
- [X] T045 [US2] Implement Reporter.Update in internal/progress/progress.go to display progress if 2s elapsed OR 10k lines processed since last update
- [X] T046 [US2] Implement ETA calculation in Reporter.Update using formula from research.md
- [X] T047 [US2] Implement Reporter.Finish in internal/progress/progress.go to display final progress line
- [X] T048 [P] [US2] Implement ResultFormatter.Format in internal/progress/progress.go to generate formatted statistics string per data-model.md display format
- [X] T049 [US2] Implement ResultFormatter.Display in internal/progress/progress.go to print formatted result to stdout
- [X] T050 [US2] Integrate ProgressReporter into optimization pipeline in cmd/gcode-optimizer/main.go - initialize with totalLines from header
- [X] T051 [US2] Call Reporter.Update after processing each line in cmd/gcode-optimizer/main.go
- [X] T052 [US2] Calculate OptimizationResult statistics in cmd/gcode-optimizer/main.go: lines removed/preserved/split, file sizes, reduction %, time savings
- [X] T053 [US2] Implement time savings calculation in cmd/gcode-optimizer/main.go: sum (distance / feed_rate) for each removed G1 move
- [X] T054 [US2] Display final results using ResultFormatter.Display after optimization completes

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - optimization works with visible progress

---

## Phase 5: User Story 3 - Error Handling (Priority: P3)

**Goal**: Provide clear, actionable error messages for invalid inputs (missing files, bad allowance, invalid strategy, unparseable GCode).

**Independent Test**: Run tool with various invalid inputs, verify error messages match specification and tool exits with non-zero code.

### Tests for User Story 3 (TDD - MANDATORY) ⚠️

- [X] T055 [P] [US3] Integration test for non-existent input file in tests/integration/cli_test.go - verify error message and exit code
- [X] T056 [P] [US3] Integration test for invalid allowance (negative) in tests/integration/cli_test.go - verify error message format
- [X] T057 [P] [US3] Integration test for invalid allowance (non-numeric) in tests/integration/cli_test.go - verify error message
- [X] T058 [P] [US3] Integration test for invalid --strategy flag in tests/integration/cli_test.go - verify error lists valid options
- [X] T059 [P] [US3] Integration test for missing Snapmaker header in tests/integration/cli_test.go - verify warning issued but processing continues
- [X] T060 [P] [US3] Integration test for unparseable file in tests/integration/cli_test.go - verify parse error message

### Implementation for User Story 3

- [X] T061 [US3] Implement file existence check in cmd/gcode-optimizer/main.go before parsing - error: "Error: Input file not found: <path>"
- [X] T062 [US3] Implement allowance validation in cmd/gcode-optimizer/main.go using strconv.ParseFloat, check >= 0 - error: "Error: Allowance must be a non-negative number, got: <value>"
- [X] T063 [US3] Implement strategy validation in cmd/gcode-optimizer/main.go - error: "Invalid strategy '<value>'. Valid options are: conservative, aggressive" (exact format per spec clarification #4)
- [X] T064 [US3] Add header validation warning in internal/parser/parser.go - check for ";tool_head" containing "CNC", add to warnings list if missing
- [X] T065 [US3] Implement Parser.Warnings method in internal/parser/parser.go to return accumulated warning messages
- [X] T066 [US3] Display parser warnings to console in cmd/gcode-optimizer/main.go after successful parse
- [X] T067 [US3] Add error handling for gcode.ParseFile failures in internal/parser/parser.go - wrap error: "failed to parse GCode file: %w"
- [X] T068 [US3] Add error handling for file I/O failures in internal/writer/writer.go - wrap errors with context
- [X] T069 [US3] Ensure all error messages to stderr use fmt.Fprintf(os.Stderr, ...) in cmd/gcode-optimizer/main.go
- [X] T070 [US3] Ensure tool exits with code 1 on all errors in cmd/gcode-optimizer/main.go using os.Exit(1)

**Checkpoint**: All user stories should now be independently functional - optimization with progress and comprehensive error handling

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, CI/CD, and documentation

- [X] T071 [P] Create GitHub Actions workflow .github/workflows/ci.yml for lint, test (matrix: macos-latest, ubuntu-latest, windows-latest)
- [X] T072 [P] Create GitHub Actions workflow .github/workflows/release.yml for multi-arch builds (darwin/amd64, darwin/arm64, windows/amd64, linux/amd64)
- [X] T073 [P] Add build configuration in Makefile or build script with CGO_ENABLED=0, -ldflags="-s -w", -trimpath flags
- [X] T074 [P] Update README.md with usage examples from quickstart.md, installation instructions, troubleshooting
- [X] T075 [P] Add --version flag support in cmd/gcode-optimizer/main.go to display version number
- [X] T076 [P] Add --help flag support in cmd/gcode-optimizer/main.go with usage, examples, flag descriptions
- [X] T077 [P] Run gofmt on all .go files in project
- [X] T078 [P] Run go vet ./... and fix any issues
- [X] T079 [P] Run go test -cover ./... and verify >= 80% coverage per constitution
- [X] T080 [P] Add additional unit tests if coverage < 80% in tests/unit/
- [X] T081 [P] Create test fixture simple-3axis.cnc in tests/integration/fixtures/ for basic 3-axis test case
- [X] T082 [P] Create test fixture 4axis-rotary.cnc in tests/integration/fixtures/ with B rotation for 4-axis test
- [X] T083 [P] Create test fixture threshold-crossing.cnc in tests/integration/fixtures/ with moves crossing threshold
- [X] T084 Manual smoke test on macOS: build binary, run on freya.cnc, verify output
- [X] T085 Manual smoke test on Linux: build binary, run on freya.cnc, verify output
- [X] T086 Manual smoke test on Windows: build binary, run on freya.cnc, verify output
- [X] T087 [P] Performance test with 10M line file in tests/integration/performance_test.go - verify memory usage <2GB, processing completes without crash, meets SC-006
- [X] T088 Create CHANGELOG.md for v1.0.0 release documenting initial features
- [X] T089 Create LICENSE file (MIT license per plan.md)
- [X] T090 Tag v1.0.0 release in git after all tests pass
- [X] T091 Verify GitHub Actions release workflow creates binaries for all 4 target platforms
- [X] T092 Create GitHub release with changelog and binary attachments
- [X] T093 Update repository description and topics on GitHub

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion (T001-T006) - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion (T007-T012)
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete (T013-T070)

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1 (integrates with optimization pipeline but doesn't modify core logic)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2 (adds validation layer, doesn't modify optimization or progress)

### Within Each User Story (TDD Workflow)

1. **Tests FIRST** - Write all test tasks for the story, verify they FAIL
2. **Foundational types** - Implement structs/enums (can be parallel)
3. **Core functions** - Implement parsing, classification, optimization logic (sequential due to dependencies)
4. **Integration** - Wire into main CLI pipeline
5. **Verify tests PASS** - All tests for the story should now pass

### Parallel Opportunities

**Setup Phase (T001-T006)**:
- T003, T004, T005, T006 can all run in parallel after T001-T002

**Foundational Phase (T007-T012)**:
- T009, T010, T011, T012 can all run in parallel after T007-T008

**User Story 1 Tests (T013-T023)**:
- All test tasks can be launched in parallel (different test files/functions)

**User Story 1 Implementation**:
- T024, T025 (parser functions) can run in parallel
- T029 (optimizer creation) can run in parallel with parser tasks
- T034 (writer) can run in parallel with optimizer tasks

**User Story 2 Tests (T040-T043)**:
- All test tasks can run in parallel

**User Story 2 Implementation**:
- T044, T048 can run in parallel (different structs)

**User Story 3 Tests (T055-T060)**:
- All test tasks can run in parallel

**Polish Phase (T071-T092)**:
- T071, T072, T073 (CI/CD) can run in parallel
- T074, T075, T076 (documentation/help) can run in parallel
- T077, T078, T079 (linting/testing) can run in parallel
- T081, T082, T083 (test fixtures) can run in parallel
- T084, T085, T086 (smoke tests) can run in parallel on different platforms

**Cross-Story Parallelization**:
- Once Foundational completes, ALL user stories (US1, US2, US3) can start in parallel with different developers

---

## Parallel Example: User Story 1 (TDD)

### Step 1: Launch all tests in parallel (verify they FAIL)

```bash
# Run all User Story 1 unit tests together:
Task T013: "Unit test for header parsing in tests/unit/parser_test.go"
Task T014: "Unit test for modal state initialization in tests/unit/parser_test.go"
Task T015: "Unit test for modal state updates in tests/unit/parser_test.go"
Task T016: "Unit test for ScanMinZ in tests/unit/parser_test.go"
Task T017: "Unit test for ClassifyMove in tests/unit/optimizer_test.go"
Task T018: "Unit test for CalculateIntersection in tests/unit/move_test.go"
Task T019: "Unit test for SplitMove in tests/unit/move_test.go"
Task T020: "Unit test for ShouldPreserve in tests/unit/optimizer_test.go"

# Run all User Story 1 integration tests together:
Task T021: "Integration test for end-to-end optimization (aggressive) in tests/integration/cli_test.go"
Task T022: "Integration test for end-to-end optimization (conservative) in tests/integration/cli_test.go"
Task T023: "Integration test for 3-axis vs 4-axis detection in tests/integration/cli_test.go"

# Expected: All tests should FAIL (not yet implemented)
```

### Step 2: Implement foundational structures in parallel

```bash
Task T024: "Implement ParseFile in internal/parser/parser.go"
Task T025: "Implement parseHeader in internal/parser/parser.go"
Task T029: "Implement NewOptimizer in internal/optimizer/optimizer.go"
Task T034: "Implement Writer.WriteFile in internal/writer/writer.go"
```

### Step 3: Implement core logic sequentially (dependencies)

```bash
# Sequential due to dependencies:
Task T026 → T027 → T028 (parser functions depend on each other)
Task T030 → T031 → T032 → T033 (optimizer functions depend on each other)
Task T035 → T036 → T037 → T038 → T039 (CLI integration pipeline)
```

### Step 4: Verify all tests PASS

```bash
go test ./tests/unit/... -v
go test ./tests/integration/... -v
# Expected: All User Story 1 tests should now PASS
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: Foundational (T007-T012) - CRITICAL
3. Complete Phase 3: User Story 1 (T013-T039)
   - Write tests first (T013-T023), verify FAIL
   - Implement (T024-T039), verify tests PASS
4. **STOP and VALIDATE**: Test US1 independently with freya.cnc
5. Deploy/demo if ready (basic optimization works)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP! ✅)
3. Add User Story 2 → Test independently → Deploy/Demo (with progress! ✅)
4. Add User Story 3 → Test independently → Deploy/Demo (with error handling! ✅)
5. Add Polish → Final release v1.0.0 🎉

### Parallel Team Strategy

With multiple developers:

1. **Week 1**: Team completes Setup + Foundational together (T001-T012)
2. **Week 2-3**: Once Foundational is done:
   - Developer A: User Story 1 (T013-T039)
   - Developer B: User Story 2 (T040-T054) - can start early, integrates later
   - Developer C: User Story 3 (T055-T070) - can start early, integrates later
3. **Week 4**: Integration and Polish (T071-T092)
4. Stories complete and integrate independently without conflicts

---

## Test Coverage Requirements (Constitution Principle III)

Per the project constitution, comprehensive testing is NON-NEGOTIABLE:

- **TDD Workflow**: Red → Green → Refactor strictly enforced
- **Unit Tests**: All exported functions in internal/* packages
- **Integration Tests**: CLI workflows, file I/O, end-to-end optimization
- **Contract Tests**: Internal package interfaces (parser, optimizer, writer, progress)
- **Platform Tests**: CI must pass on macOS, Linux, Windows before merge
- **Coverage**: Minimum 80% code coverage; uncovered code requires justification

**Test Task Counts**:
- User Story 1: 11 test tasks (T013-T023)
- User Story 2: 4 test tasks (T040-T043)
- User Story 3: 6 test tasks (T055-T060)
- Total: 21 test tasks (23% of all implementation tasks)

---

## Notes

- **[P] tasks**: Different files, no dependencies - can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- **TDD workflow**: Tests MUST fail before implementation, then pass after
- **Constitution compliance**: Every task aligns with project principles (cross-platform, static binary, Go best practices, comprehensive testing)
- **File paths**: All absolute paths from repository root for clarity
- **Commit strategy**: Commit after each task or logical group (e.g., all tests for a story)
- **Stop at any checkpoint**: Validate story independently before proceeding
- **Avoid**: Vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Summary

**Total Tasks**: 93 (including setup, foundational, 3 user stories, polish)

**Tasks per Phase**:
- Phase 1 (Setup): 6 tasks
- Phase 2 (Foundational): 6 tasks
- Phase 3 (US1 - Basic Optimization): 27 tasks (11 tests + 16 implementation)
- Phase 4 (US2 - Progress Monitoring): 15 tasks (4 tests + 11 implementation)
- Phase 5 (US3 - Error Handling): 16 tasks (6 tests + 10 implementation)
- Phase 6 (Polish): 23 tasks

**Parallel Opportunities**:
- 43 tasks marked [P] (46% can run in parallel given proper sequencing)
- All 3 user stories can proceed in parallel after Foundational phase
- All tests within a user story can run in parallel

**Independent Test Criteria**:
- US1: Run on freya-subset.cnc with 1mm allowance, verify line reduction and correct move filtering
- US2: Run on large test file, observe progress updates every 2s or 10k lines
- US3: Test with invalid inputs, verify error messages match spec

**Suggested MVP Scope**: User Story 1 only (39 tasks total including setup/foundational) = Basic working optimizer
