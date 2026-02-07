# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2025-02-05)

**Core value:** Transform documentation into high-quality, deduplicated training data with minimal manual effort.
**Current focus:** Phase 4: Multi-Provider Support

## Current Position

Phase: 4 of 7 (Multi-Provider Support)
Plan: 0 of ? in current phase
Status: Phase added, not yet planned
Last activity: 2026-02-07 — Added Phase 7: Claude Provider Support

Progress: [████░░░░░░] 43% (Phases 4-7 not started)

## Performance Metrics

**Velocity:**
- Total plans completed: 8
- Average duration: ~3-5 min
- Total execution time: ~0.4 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation | 3 | 4 | ~2 min |
| 02-smart-generation | 3 | 3 | ~4 min |
| 03-production-ready | 3 | 3 | ~5 min |

**Final Status:**
- All 3 phases complete (9/9 plans)
- Core functionality: document processing, Ollama integration, training data generation
- Production features: per-document output (IO-05), --skip-merge flag, merge subcommand
- Error handling (UX-06): Professional error messages with actionable guidance
- User workflow: generate --skip-merge → inspect per-document files → merge into master files

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Fixed Ollama client by removing JSON schema format that was causing empty responses
- Fixed pair count calculation for small documents to avoid negative split counts
- Modified pipeline to skip splits with 0 pairs required
- Used strings.TrimSpace for whitespace validation in validatePair
- Case-sensitive deduplication matching with "prompt|||completion" key format
- Added ExcludePair struct to templates package to avoid circular import with generator.Pair
- parseResponse() no longer pads with empty pairs - returns whatever was parsed
- deduplicateAgainstExclusions only filters against exclusions (not internal deduplication)
- Backfill loop accumulates all collected pairs as new exclusions for each retry iteration
- Per-document files written during processing (not after) for true skip-merge functionality
- Per-document filenames follow {docname}_{split}.jsonl pattern with sanitized names
- Merge command combines per-document files using *_{split}.jsonl pattern matching
- Error message pattern: "[What failed]\n[Actionable guidance on how to fix it]"
- All file write errors include disk space and permissions checks
- Ollama connection errors include troubleshooting steps (ollama serve, ollama pull)
- Merge errors guide users to run `datakeg generate --skip-merge` first

### Pending Todos

None.

### Roadmap Evolution

- Phase 4 added: Multi-Provider Support (OpenRouter integration, provider abstraction, API key management)
- Phase 5 added: Chat Format Support (chat-style messages/roles JSONL, new templates, --format flag)
- Phase 6 added: Reasoning Format Support (chain-of-thought training data for reasoning models)
- Phase 7 added: Claude Provider Support (Anthropic API provider via Claude subscription)

### Blockers/Concerns

None

## Session Continuity

Last session: 2026-02-07 (phase 7 added)
Stopped at: Added Phase 7 to roadmap, not yet planned
Resume file: None

---
*State initialized: 2025-02-05*
*Last updated: 2026-02-07 (Phase 7 added)*
