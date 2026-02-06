---
phase: 03-production-ready
plan: "01"
subsystem: cli
tags: [cobra, cli, jsonl, output-format]

# Dependency graph
requires:
  - phase: 02-smart-generation
    provides: Pipeline wiring and split percentage configuration
provides:
  - "--skip-merge" flag for conditional master file generation
  - Per-document JSONL file output during processing
  - Separate merge step controllable via CLI flag
affects: Phase 03-02 (merge subcommand implementation)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Per-document file output pattern for granular output
    - Conditional merge step based on CLI flag

key-files:
  created: []
  modified:
    - cmd/datakeg/main.go (flag addition and parameter passing)
    - cmd/datakeg/commands/generate.go (per-document files and conditional merge)

key-decisions:
  - "Decided to write per-document files during processing (not after) to enable --skip-merge to skip merge entirely"
  - "Used sanitizeDocName helper to create consistent per-document filenames following {docname}_{split}.jsonl pattern"

patterns-established:
  - "Per-document file output: Each document gets its own _train.jsonl, _valid.jsonl, _test.jsonl files"
  - "Conditional merge: Master files only written when --skip-merge=false (default)"

# Metrics
duration: ~3 min
completed: 2026-02-06
---

# Phase 3 Plan 1: Per-Document Output with --skip-merge Flag

**Added --skip-merge flag to generate command with per-document JSONL file output during processing**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-02-06T22:58:42Z
- **Completed:** 2026-02-06T23:01:XXZ
- **Tasks:** 3/3
- **Files modified:** 2

## Accomplishments

- Added --skip-merge boolean flag to generate command CLI
- Implemented per-document JSONL file output during document processing
- Created sanitizeDocName helper for consistent filename generation
- Added conditional merge step that skips master file writing when --skip-merge=true

## Task Commits

Each task was committed atomically:

1. **Task 1: Add --skip-merge flag to CLI** - `37afb22` (feat)
2. **Task 2: Modify pipeline to write per-document files** - `37afb22` (part of same commit)
3. **Task 3: Update writer to support per-document paths** - `37afb22` (part of same commit)

**Plan metadata:** `37afb22` (docs: complete plan)

## Files Created/Modified

- `cmd/datakeg/main.go` - Added flagSkipMerge variable, flag registration, and parameter passing to pipeline
- `cmd/datakeg/commands/generate.go` - Added skipMerge parameter, per-document file writing with sanitizeDocName helper, conditional merge logic

## Decisions Made

- Per-document files are written during processing (not after) to enable true skip-merge functionality
- Per-document filenames follow pattern: `{docname}_{split}.jsonl` (e.g., `doc1_train.jsonl`)
- Document names are sanitized: extension removed, spaces replaced with underscores
- Master files (train.jsonl, valid.jsonl, test.jsonl) are only written when --skip-merge=false (default)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - implementation followed plan exactly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Per-document file output complete
- --skip-merge flag wired to pipeline
- Ready for Phase 03-02 merge subcommand implementation (which will use per-document files as input)

---

*Phase: 03-production-ready*
*Completed: 2026-02-06*
