---
phase: 03-production-ready
plan: "03"
subsystem: cli
tags: [error-handling, ux, cli, merge, generate]

# Dependency graph
requires:
  - phase: 03-01
    provides: Per-document file output with --skip-merge flag
  - phase: 03-02
    provides: Merge subcommand for consolidating per-document files
provides:
  - Improved error messages with actionable guidance
  - Progress messages for merge operation
  - Comprehensive Phase 3 verification

affects: [cli, user-experience, documentation]

# Tech tracking
tech-stack:
  added: []
  patterns: [Actionable error message pattern: summary + guidance + optional technical details]

key-files:
  created: []
  modified:
    - cmd/datakeg/commands/generate.go
    - cmd/datakeg/commands/merge.go

key-decisions:
  - "Error messages follow pattern: short summary + actionable guidance"

patterns-established:
  - "Error message pattern: [What failed] + [How to fix it]"

# Metrics
duration: ~8 min
completed: 2026-02-06
---

# Phase 3 Plan 3: Error Message Improvements Summary

**Professional error handling with actionable guidance across generate and merge commands**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-02-06T23:31:58Z
- **Completed:** 2026-02-06T23:40:22Z
- **Tasks:** 3/3
- **Files modified:** 2

## Accomplishments

- Improved all generate command error messages with actionable user guidance
- Enhanced merge command with progress messages and helpful error handling
- Verified all Phase 3 success criteria across generate, merge, templates, and error handling
- Binary builds successfully and all commands function as expected

## Task Commits

Each task was committed atomically:

1. **Task 1: Audit and improve generate command errors** - `ada1466` (feat)
2. **Task 2: Improve merge command error handling** - `f1d57df` (feat)
3. **Task 3: Verify Phase 3 success criteria** - Verification complete (no code changes)

**Plan metadata:** (this summary commit)

## Files Created/Modified

- `cmd/datakeg/commands/generate.go` - Improved error messages with actionable guidance
- `cmd/datakeg/commands/merge.go` - Added progress messages and helpful error handling

## Decisions Made

- Error message pattern: "[What failed]\n[Actionable guidance on how to fix it]"
- All file write errors include disk space and permissions checks
- Ollama connection errors include troubleshooting steps (ollama serve, ollama pull)
- Merge errors guide users to run `datakeg generate --skip-merge` first

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All Phase 3 success criteria verified complete
- Error handling (UX-06) implemented with actionable guidance
- Per-document output (IO-05) working with --skip-merge flag
- Merge subcommand working correctly
- Embedded templates (LLM-03, LLM-04, LLM-05) verified in binary
- Progress reporting (UX-01, UX-02, UX-03) implemented
- CLI flag overrides (UX-05) verified functional

---

*Phase: 03-production-ready*
*Completed: 2026-02-06*
