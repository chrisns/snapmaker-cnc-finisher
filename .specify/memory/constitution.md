<!--
SYNC IMPACT REPORT
==================
Version Change: 1.0.0 → 1.1.0
Rationale: Added new Principle VI for exhaustive research methodology during planning phases

Modified Principles:
- No existing principles modified

Added Sections:
- NEW: VI. Exhaustive Research During Planning

Removed Sections: None

Templates Status:
✅ .specify/templates/plan-template.md - requires update for Phase 0 research requirements
✅ .specify/templates/spec-template.md - reviewed, compatible (no changes needed)
✅ .specify/templates/tasks-template.md - reviewed, compatible (no changes needed)

Follow-up TODOs:
- Update plan-template.md to add explicit Context7/WebSearch/Perplexity usage requirements in Phase 0

Migration Notes:
- New principle establishes mandatory research methodology for all planning phases
- Requires exhaustive use of Context7 for library/framework documentation
- Requires heavy use of WebSearch and Perplexity for current best practices and solutions
- Applies to all future features; existing features should be retrofitted during next planning iteration
-->

# snapmaker-cnc-finisher Constitution

## Core Principles

### I. Cross-Platform Portability

All code MUST be portable across macOS (Intel & ARM), Windows, and Linux without platform-specific workarounds or conditional compilation paths that compromise maintainability.

**Rationale**: The tool serves users across diverse operating systems. Using Go's standard library and avoiding OS-specific APIs ensures consistent behavior and reduces maintenance burden.

**Rules**:
- Use Go standard library abstractions (`filepath`, `os`, `io`) instead of platform-specific calls
- File paths MUST use `filepath.Join()` and respect `filepath.Separator`
- Line endings MUST be handled via `bufio.Scanner` or explicit normalization
- Test on all three target platforms (macOS Intel, macOS ARM, Windows, Linux) before release

### II. Static Binary Distribution

The project MUST compile to a single, statically-linked binary with zero external runtime dependencies.

**Rationale**: Users expect to download a single executable that "just works" without installing runtimes, shared libraries, or managing dependencies. Static binaries maximize portability and simplify distribution.

**Rules**:
- No dynamic linking (CGO_ENABLED=0 required)
- No external runtime dependencies (interpreters, frameworks, system libraries)
- All assets MUST be embedded via `embed` package if needed
- Binary size should be monitored but portability > size optimization

### III. Comprehensive Testing (NON-NEGOTIABLE)

Test-Driven Development is mandatory: tests MUST be written first, validated to fail, then implementation proceeds to make them pass.

**Rationale**: Go's testing ecosystem (`go test`, table-driven tests, benchmarks) makes TDD natural and effective. Catching bugs early reduces long-term maintenance costs and ensures reliability across platforms.

**Rules**:
- Red-Green-Refactor cycle strictly enforced: write test → verify failure → implement → verify pass → refactor
- Unit tests required for all exported functions and types
- Integration tests required for CLI workflows and file I/O operations
- Contract tests required for external interfaces (file format parsing, API contracts if applicable)
- Tests MUST pass on all target platforms before merging
- Use table-driven tests for comprehensive input coverage
- Minimum 80% code coverage; uncovered code requires justification

### IV. Open Source & Release Management

The project follows open-source best practices with public development on GitHub and automated multi-architecture releases.

**Rationale**: Transparency builds trust, automated releases reduce manual errors, and multi-arch support ensures all users have first-class binaries.

**Rules**:
- Repository hosted at `github.com/chrisns/snapmaker-cnc-finisher`
- GitHub Actions MUST automate: linting, testing (all platforms), building (all architectures), releasing
- Releases MUST include binaries for:
  - macOS Intel (darwin/amd64)
  - macOS ARM (darwin/arm64)
  - Windows (windows/amd64)
  - Linux (linux/amd64)
- Semantic versioning (MAJOR.MINOR.PATCH) required
- Changelog MUST accompany each release (generated or manual)
- All PRs MUST pass CI checks before merge

### V. Go Best Practices

Code MUST follow Go community standards and idiomatic patterns.

**Rationale**: Go has strong conventions that improve readability, maintainability, and interoperability with the broader ecosystem.

**Rules**:
- `gofmt` and `go vet` MUST pass without warnings
- Follow Effective Go guidelines
- Use standard project layout (cmd/, internal/, pkg/ where appropriate)
- Dependency management via Go modules (`go.mod`)
- Avoid premature optimization; profile before optimizing
- Error handling: return errors, don't panic (except in truly unrecoverable situations)
- Minimal external dependencies; prefer standard library

### VI. Exhaustive Research During Planning

During all planning phases (Phase 0 and beyond), AI assistants MUST conduct exhaustive research using Context7, WebSearch, and Perplexity to ensure decisions are informed by current best practices, up-to-date documentation, and proven solutions.

**Rationale**: Technology evolves rapidly. Relying solely on training data or cached knowledge risks implementing outdated patterns, deprecated APIs, or reinventing solutions that already exist. Exhaustive research ensures decisions reflect the current state of the ecosystem, maximizes code quality, and prevents technical debt from uninformed choices.

**Rules**:
- **Context7 MUST be used exhaustively** for all library and framework documentation during planning:
  - Retrieve official docs for ALL dependencies being considered or used
  - Verify API signatures, recommended patterns, and compatibility constraints
  - Check for deprecated features and migration guides before making architectural decisions
  - Confirm idiomatic usage patterns specific to library versions in use
- **WebSearch MUST be heavily used** to gather:
  - Current best practices (e.g., "Go error handling patterns 2025", "FastAPI performance optimization 2025")
  - Known issues and gotchas with libraries/frameworks under consideration
  - Comparison articles and decision frameworks for architectural choices
  - Recent blog posts, conference talks, and community discussions on relevant topics
- **Perplexity MUST be heavily used** for:
  - Synthesizing conflicting information from multiple sources
  - Exploring emerging patterns and recent ecosystem changes
  - Validating assumptions about tool/library capabilities
  - Discovering alternative approaches not immediately obvious
- **Research must be documented** in Phase 0 outputs:
  - Include links/citations to Context7 docs consulted
  - Summarize key findings from WebSearch and Perplexity queries
  - Justify technology choices with evidence from research, not just preference
- **Continuous research throughout planning**:
  - Don't limit research to initial exploration; revisit tools when new questions arise
  - If implementation reveals gaps, return to research mode before proceeding
  - Update research.md when new information changes architectural decisions

## Build & Release Requirements

### CI/CD Pipeline

GitHub Actions workflows MUST:
1. **Lint**: Run `gofmt`, `go vet`, and optionally `golangci-lint`
2. **Test**: Execute `go test ./...` on matrix: [macOS-latest, ubuntu-latest, windows-latest]
3. **Build**: Cross-compile for all target architectures using `GOOS` and `GOARCH`
4. **Release**: On tagged commits, create GitHub Release with all binaries attached

### Build Configurations

| Platform       | GOOS    | GOARCH | Binary Name              |
|----------------|---------|--------|--------------------------|
| macOS Intel    | darwin  | amd64  | snapmaker-cnc-finisher   |
| macOS ARM      | darwin  | arm64  | snapmaker-cnc-finisher   |
| Windows        | windows | amd64  | snapmaker-cnc-finisher.exe |
| Linux          | linux   | amd64  | snapmaker-cnc-finisher   |

**Static linking enforced via**: `CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath`

### Release Checklist

Before creating a release tag:
1. All tests pass on all platforms
2. Version bumped in code (if applicable)
3. Changelog updated or auto-generated
4. No open blockers or critical bugs
5. Manual smoke test on at least one platform

## Development Workflow

### Code Review

- All changes via pull requests (no direct commits to `main`)
- At least one approval required
- CI MUST pass (all platforms, all tests)
- Code review checklist:
  - [ ] Constitution compliance verified
  - [ ] Tests added/updated (TDD followed)
  - [ ] Cross-platform compatibility considered
  - [ ] Error handling appropriate
  - [ ] Documentation updated if needed

### Testing Gates

**Pre-commit** (local):
- `go fmt ./...`
- `go vet ./...`
- `go test ./...`

**Pre-merge** (CI):
- Linting passes
- Tests pass on macOS, Linux, Windows
- Builds succeed for all target architectures
- Code coverage ≥80% or explained

### Branching Strategy

- `main`: stable, always releasable
- Feature branches: `feature/short-description`
- Bugfix branches: `fix/short-description`
- Release tags: `vMAJOR.MINOR.PATCH` (e.g., `v1.2.3`)

## Governance

### Amendment Process

1. Proposed changes MUST be documented in a PR to this constitution
2. Rationale for change MUST be provided
3. Impact on existing code MUST be assessed
4. Migration plan required if breaking changes
5. Approval from maintainer(s) required
6. Version bump according to:
   - **MAJOR**: Principle removals, incompatible policy changes
   - **MINOR**: New principles added, expanded requirements
   - **PATCH**: Clarifications, typos, non-semantic fixes

### Compliance Review

- All PRs MUST be checked against these principles
- Violations require explicit justification or rejection
- Complexity that contradicts simplicity principles MUST be justified and documented
- Use `.specify/templates/plan-template.md` and other templates for structured design artifacts

### Enforcement

This constitution supersedes all other practices and preferences. When in doubt:
1. Refer to these principles
2. Prioritize: Correctness > Performance > Convenience
3. Test rigorously (all platforms)
4. Document deviations explicitly

**Version**: 1.1.0 | **Ratified**: 2025-10-26 | **Last Amended**: 2025-10-26
