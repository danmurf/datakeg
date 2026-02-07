# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2025-02-05)

**Core value:** Transform documentation into high-quality, deduplicated training data with minimal manual effort.
**Current focus:** Phase 5: Chat Format Support

## Current Position

Phase: 5 of 7 (Chat Format Support)
Plan: 02 of 02 in current phase
Status: **Phase Complete**
Last activity: 2026-02-07 — Completed Plan 05-02, Phase 5 finished

Progress: [█████████░] 71% (Phase 5 complete: 5/7 phases done)

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
- Phase 5: Added FormatType enum with "completion" and "chat" values
- Phase 5: Chat templates instruct LLM for context-free user messages and style-matching assistant responses
- Phase 5: Generator selects correct template based on format type and split type
- Phase 5: Generator parses chat-format LLM responses (user/assistant JSON) into Pair structs
- Phase 5: ChatMessage struct with Messages []Message for OpenAI-compatible output
- Phase 5: --format flag defaults to "completion" for backward compatibility
- Phase 5: --system-message adds optional system message at position 0 in chat messages
- Phase 5: Merge is format-agnostic (raw line concatenation) for both formats

### Pending Todos

None.

### Roadmap Evolution

- Phase 4 added: Multi-Provider Support (OpenRouter integration, provider abstraction, API key management)
- Phase 5 added: Chat Format Support (chat-style messages/roles JSONL, new templates, --format flag)
- Phase 6 added: Reasoning Format Support (chain-of-thought training data for reasoning models)
- Phase 7 added: Claude Provider Support (Anthropic API provider via Claude subscription)
- Phase 8 added: Format Conversion (convert generated JSONL to model-specific training formats via templates)

### Blockers/Concerns

None

## Session Continuity

Last session: 2026-02-07 (phase 8 added)
Stopped at: Added Phase 8 to roadmap, not yet planned
Resume file: None

---
*State initialized: 2025-02-05*
*Last updated: 2026-02-07 (Phase 8 added)*
