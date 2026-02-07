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
- [ ] **Phase 4: Multi-Provider Support** - Provider abstraction, OpenRouter integration, API key management
- [ ] **Phase 5: Chat Format Support** - Chat-style training data with messages/roles JSONL output
- [ ] **Phase 6: Reasoning Format Support** - Chain-of-thought training data for reasoning models
- [ ] **Phase 7: Format Conversion** - Convert generated JSONL to model-specific training formats via templates
- [ ] **Phase 8: Claude Provider Support** - Anthropic API provider for generating data via Claude subscription

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

### Phase 4: Multi-Provider Support
**Goal**: Users can choose between Ollama and OpenRouter (with any model) for pair generation, with secure API key storage and cost transparency

**Depends on**: Phase 3

**Requirements**: TBD during planning

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg generate --provider openrouter --model <model> <source> <output>` and pairs are generated via OpenRouter API
  2. User runs `datakeg generate --provider ollama --model <model>` to select any Ollama model (not just the default)
  3. User can securely store and retrieve API keys per provider (not passed as plain CLI flags)
  4. Generator is provider-agnostic — uses a common interface, not coupled to Ollama directly
  5. Existing Ollama workflow continues to work as default with no breaking changes

**Plans:** 3 plans

Plans:
- [ ] 04-01-PLAN.md — Provider interface, Ollama wrapper, Generator refactor, --provider flag
- [ ] 04-02-PLAN.md — OpenRouter provider with retry, cost estimation, confirmation, --yes, --dry-run
- [ ] 04-03-PLAN.md — list-providers subcommand and comprehensive tests

### Phase 5: Chat Format Support
**Goal**: Users can generate chat-style training data (messages with role/content) in addition to the existing prompt/completion format

**Depends on**: Phase 4

**Requirements**: TBD during planning

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg generate --format chat <source> <output>` and receives JSONL files with `{"messages":[{"role":"user","content":"..."},{"role":"assistant","content":"..."}]}` format
  2. Existing `--format completion` (default) continues to produce current prompt/completion JSONL unchanged
  3. Chat-specific prompt templates instruct the LLM to generate multi-turn conversations from documentation
  4. Deduplication and exclusion logic works correctly with the chat message format
  5. Per-document files, merge, and all existing features work with both formats

**Plans:** 2 plans

Plans:
- [ ] 05-01-PLAN.md — FormatType enum, chat templates, format-aware generator (template selection + chat response parsing)
- [ ] 05-02-PLAN.md — CLI flags (--format, --system-message), chat JSONL writer, pipeline wiring (generate + merge)

### Phase 6: Reasoning Format Support
**Goal**: Users can generate chain-of-thought training data for reasoning models, with step-by-step reasoning traces in the output

**Depends on**: Phase 5

**Requirements**: TBD during planning

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg generate --format reasoning <source> <output>` and receives JSONL with structured reasoning traces (e.g. thinking/steps/answer fields)
  2. Reasoning-specific prompt templates instruct the LLM to produce step-by-step problem solving from documentation
  3. Output format is compatible with common reasoning model fine-tuning pipelines
  4. Deduplication and exclusion logic works correctly with the reasoning format
  5. All existing features (per-doc files, merge, skip-merge, provider selection) work with reasoning format

**Plans:** 2 plans

Plans:
- [ ] 06-01-PLAN.md — FormatReasoning enum, reasoning templates, reasoning response parser, ReasoningFormat variants
- [ ] 06-02-PLAN.md — CLI flags (--reasoning-format), reasoning JSONL writer, pipeline wiring

### Phase 7: Format Conversion
**Goal**: Users can convert generated JSONL output files into model-specific training formats (e.g., Mistral Instruct) using customizable templates

**Depends on**: Phase 5

**Requirements**: TBD during planning

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg convert --template mistral-instruct <output>` and generated JSONL files are converted to the target format
  2. Conversion templates define how completion, chat, and reasoning data maps to the target format (e.g., `<|user|>\n...\n<|assistant|>\n...`)
  3. Users can create custom conversion templates for any model-specific format
  4. Built-in templates ship embedded in the binary for common formats (Mistral Instruct, etc.)
  5. Conversion works with all source formats (completion, chat, reasoning) where the template supports them

**Plans:** 2 plans

Plans:
- [ ] 07-01-PLAN.md — Converter package, conversion templates, template loader with jsonEscape
- [ ] 07-02-PLAN.md — CLI convert command, flag registration, end-to-end wiring

### Phase 8: Claude Provider Support
**Goal**: Users can generate training data using their Anthropic API key / Claude subscription, leveraging the provider abstraction from Phase 4

**Depends on**: Phase 4

**Requirements**: TBD during planning

**Success Criteria** (what must be TRUE):
  1. User runs `datakeg generate --provider claude --model claude-sonnet-4-5-20250929 <source> <output>` and pairs are generated via the Anthropic API
  2. Anthropic API key is stored securely via the same key management system from Phase 4
  3. Claude provider implements the same provider interface as Ollama and OpenRouter — no special-casing in the generator
  4. User can select any available Claude model (Haiku, Sonnet, Opus, etc.)
  5. All output formats (completion, chat, reasoning) work with the Claude provider

**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 8 to break down)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Core Pipeline | 3/3 | ✓ Complete | 2026-02-05 |
| 2. Smart Generation | 3/3 | ✓ Complete | 2026-02-06 |
| 3. Production Ready | 3/3 | ✓ Complete | 2026-02-06 |
| 4. Multi-Provider Support | 0/3 | Not Started | — |
| 5. Chat Format Support | 0/? | Not Started | — |
| 6. Reasoning Format Support | 0/2 | Not Started | — |
| 7. Format Conversion | 0/2 | Not Started | — |
| 8. Claude Provider Support | 0/? | Not Started | — |

---
*Roadmap created: 2025-02-05*
*Last updated: 2026-02-07 (Phase 7 planned)*
