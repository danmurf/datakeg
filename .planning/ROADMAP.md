# Roadmap: datakeg

## Overview

datakeg transforms raw documentation into LLM training datasets through a three-phase journey. Phase 1 establishes the core pipeline (read docs → generate pairs via Ollama → write JSONL). Phase 2 adds intelligent generation with exclusion logic and deduplication for quality. Phase 3 completes production readiness with embedded templates, per-document files, and professional UX.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Core Pipeline** - End-to-end MVP from markdown to JSONL training data
- [ ] **Phase 2: Smart Generation** - Exclusion logic, deduplication, and quality controls
- [ ] **Phase 3: Production Ready** - Embedded templates, per-doc files, progress reporting

## Phase Details

### Phase 1: Core Pipeline
**Goal**: Users can generate basic train/valid/test datasets from markdown files using Ollama

**Depends on**: Nothing (first phase)

**Requirements**: IO-01, IO-02, IO-03, IO-04, CLI-01, CLI-04, CLI-05, LLM-01, LLM-02, GEN-01, GEN-02, GEN-03, UX-04

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg generate <source> <output>` and receives three JSONL files (train.jsonl, valid.jsonl, test.jsonl)
  2. User points at folder with .md and .txt files and all documents are processed
  3. Generated pairs follow correct JSONL format with prompt and completion keys
  4. User can configure Ollama model via flag and see it used for generation
  5. Splits match configured percentages (default 60/20/20 train/valid/test)

**Plans**: TBD

Plans:
- [ ] TBD during planning

### Phase 2: Smart Generation
**Goal**: Generated datasets have no duplicates and respect exclusion boundaries between splits

**Depends on**: Phase 1

**Requirements**: GEN-04, GEN-05, GEN-06, GEN-07, GEN-08, QUAL-01, QUAL-02, QUAL-03, QUAL-04

**Success Criteria** (what must be TRUE):
  1. User inspects output files and finds no duplicate pairs within or across train/valid/test splits
  2. Valid set contains no pairs that appear in train set (exclusion working)
  3. Test set contains no pairs that appear in train or valid sets (exclusion working)
  4. When deduplication reduces counts below target, additional pairs are generated automatically
  5. User can configure split percentages via flags and final counts match targets

**Plans:** 3 plans

Plans:
- [ ] 02-01-PLAN.md — TDD: pair validation and deduplication functions
- [ ] 02-02-PLAN.md — Template exclusion support and Generate() refactor with backfill
- [ ] 02-03-PLAN.md — Pipeline wiring: cross-split exclusion and split percentage flags

### Phase 3: Production Ready
**Goal**: Tool is distribution-ready with embedded templates, advanced CLI features, and professional UX

**Depends on**: Phase 2

**Requirements**: IO-05, IO-06, CLI-02, CLI-03, LLM-03, LLM-04, LLM-05, UX-01, UX-02, UX-03, UX-05, UX-06

**Success Criteria** (what must be TRUE):
  1. Binary contains embedded prompt templates and runs without external file dependencies
  2. User sees progress showing current file, split type (train/valid/test), and pair generation status
  3. User runs `datakeg generate --skip-merge` and gets per-document JSONL files without merge step
  4. User runs `datakeg merge <output>` to combine existing per-doc files into master files
  5. User encounters errors and receives helpful messages with actionable guidance
  6. User can override any default configuration via CLI flags

**Plans:** 3 plans

Plans:
- [x] 03-01-PLAN.md — Per-document output files + --skip-merge flag
- [x] 03-02-PLAN.md — Merge subcommand implementation
- [x] 03-03-PLAN.md — UX improvements and error handling

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Core Pipeline | 3/3 | ✓ Complete | 2026-02-05 |
| 2. Smart Generation | 3/3 | ✓ Complete | 2026-02-06 |
| 3. Production Ready | 3/3 | ✓ Complete | 2026-02-06 |

---
*Roadmap created: 2025-02-05*
*Last updated: 2026-02-06 (Phase 3 complete)*
