---
phase: 02-smart-generation
plan: 02
subsystem: testing
tags: [templates, exclusion, validation, deduplication, backfill]

# Dependency graph
requires:
  - phase: 02-01
    provides: "validatePair and deduplicatePairs functions with test coverage"
provides:
  - ExcludePairs support in templates and Generate function
  - Full generation pipeline with validation, deduplication, exclusion filtering, and backfill
affects: [02-03, 03-generation-cli]

# Tech tracking
tech-stack:
  added: []
  patterns: [Conditional Go template rendering, Retry loop with accumulated exclusions]

key-files:
  created: []
  modified:
    - internal/templates/templates.go
    - internal/templates/templates_test.go
    - internal/templates/prompts/train.tmpl
    - internal/templates/prompts/valid.tmpl
    - internal/templates/prompts/test.tmpl
    - internal/generator/generator.go
    - internal/generator/generator_test.go

key-decisions:
  - "Added ExcludePair struct to templates package to avoid circular import with generator.Pair"
  - "parseResponse() no longer pads with empty pairs - returns whatever was parsed"
  - "deduplicateAgainstExclusions only filters against exclusions, not internal deduplication"

patterns-established:
  - "Templates conditionally render content using {{if .Field}}"
  - "Backfill retry accumulates all collected pairs as new exclusions"

# Metrics
duration: 7 min
completed: 2026-02-06
---

# Phase 2: Smart Generation - Plan 2 Summary

**Exclusion support added to templates with full Generate() refactoring: conditional rendering, validation, deduplication, exclusion filtering, and 3x retry backfill**

## Performance

- **Duration:** 7 min
- **Started:** 2026-02-06T22:29:06Z
- **Completed:** 2026-02-06T22:36:21Z
- **Tasks:** 2/2 complete
- **Files modified:** 7

## Accomplishments

- Added ExcludePairs field to PromptData struct with ExcludePair struct definition
- Updated all three templates (train.tmpl, valid.tmpl, test.tmpl) with conditional exclusion sections
- Refactored Generate() to accept excludePairs parameter and pass to template
- Generate() now validates pairs, deduplicates internally, and filters against exclusions
- Implemented backfill loop with up to 3 retries to reach target pair count
- Removed padding behavior from parseResponse() - no longer returns empty pairs
- Added comprehensive tests for ExcludePairs template rendering and deduplication

## Task Commits

1. **Task 1: ExcludePairs in templates** - `bce3e0d` (feat)
2. **Task 2: Generate refactor** - `1fd6878` (feat)

**Plan metadata:** `1fd6878` (docs: complete plan)

## Files Created/Modified

- `internal/templates/templates.go` - Added ExcludePair struct and ExcludePairs field to PromptData
- `internal/templates/templates_test.go` - Added tests for ExcludePairs conditional rendering
- `internal/templates/prompts/train.tmpl` - Added conditional exclusion section
- `internal/templates/prompts/valid.tmpl` - Added conditional exclusion section
- `internal/templates/prompts/test.tmpl` - Added conditional exclusion section
- `internal/generator/generator.go` - Refactored Generate() with exclusion, validation, dedup, backfill
- `internal/generator/generator_test.go` - Updated parseResponse tests, added deduplicateAgainstExclusions tests

## Decisions Made

- Used separate ExcludePair struct in templates package to avoid circular import with generator.Pair
- deduplicateAgainstExclusions filters only against exclusion pairs (not internal deduplication) since deduplicatePairs is called first
- Backfill loop accumulates all collected pairs as new exclusions for each retry iteration
- parseResponse() now returns whatever pairs were parsed without padding (truncates to expectedCount)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Templates ready with exclusion support
- Generator ready with validation, deduplication, exclusion filtering, and backfill
- cmd/datakeg Generate() signature mismatch will be fixed in Plan 02-03

---
*Phase: 02-smart-generation*
*Completed: 2026-02-06*
