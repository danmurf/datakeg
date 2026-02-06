# Requirements: datakeg

**Defined:** 2025-02-05
**Core Value:** Transform documentation into high-quality, deduplicated training data with minimal manual effort.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Input/Output

- [ ] **IO-01**: CLI reads .md files from source folder
- [ ] **IO-02**: CLI reads .txt files from source folder
- [ ] **IO-03**: CLI outputs JSONL format with `{"prompt": "...", "completion": "..."}` per line
- [ ] **IO-04**: CLI processes all documents in source folder (batch processing)
- [ ] **IO-05**: CLI generates per-document output files ({docname}_train.jsonl, {docname}_valid.jsonl, {docname}_test.jsonl)
- [ ] **IO-06**: CLI merges per-document files into master train.jsonl, valid.jsonl, test.jsonl

### CLI Commands

- [ ] **CLI-01**: `generate <source> <output>` subcommand runs full pipeline (generate + merge)
- [ ] **CLI-02**: `generate --skip-merge` flag outputs per-doc files only, skips merge step
- [ ] **CLI-03**: `merge <output>` subcommand merges existing per-doc files into master files
- [ ] **CLI-04**: Source folder argument specifies input documentation directory
- [ ] **CLI-05**: Output folder argument specifies training data output directory

### LLM Integration

- [ ] **LLM-01**: CLI connects to Ollama for LLM inference
- [ ] **LLM-02**: Model is configurable via flag (default: gpt-oss:20b)
- [ ] **LLM-03**: Prompt templates are embedded in binary at build time
- [ ] **LLM-04**: Separate prompt templates exist for train, valid, and test generation
- [ ] **LLM-05**: Templates support variables: {{document_content}}, {{num_pairs}}, {{document_name}}, {{exclude_pairs}}

### Dataset Generation

- [ ] **GEN-01**: CLI calculates total pair count based on document length (configurable pairs per 1k chars)
- [ ] **GEN-02**: CLI splits pairs by percentages (configurable valid%, test%, train is remainder)
- [ ] **GEN-03**: CLI makes separate LLM calls for train, valid, and test pairs
- [ ] **GEN-04**: Valid generation receives train pairs via {{exclude_pairs}} to prevent overlap
- [ ] **GEN-05**: Test generation receives train + valid pairs via {{exclude_pairs}} to prevent overlap
- [ ] **GEN-06**: Fill mechanism detects when dedup reduces count below target
- [ ] **GEN-07**: Fill mechanism makes additional LLM calls with all existing pairs as {{exclude_pairs}}
- [ ] **GEN-08**: Fill mechanism loops until target pair counts are met

### Data Quality

- [ ] **QUAL-01**: CLI performs exact-match deduplication within each split
- [ ] **QUAL-02**: CLI performs exact-match deduplication across splits (valid vs train, test vs train+valid)
- [ ] **QUAL-03**: Train/valid/test splits are configurable via flags
- [ ] **QUAL-04**: Default split is 60% train, 20% valid, 20% test

### User Experience

- [ ] **UX-01**: CLI displays progress showing current file being processed
- [ ] **UX-02**: CLI displays progress showing current split type (train/valid/test)
- [ ] **UX-03**: CLI displays progress showing pair generation status
- [ ] **UX-04**: All configuration has sensible defaults
- [ ] **UX-05**: All configuration is overridable via CLI flags
- [ ] **UX-06**: CLI provides helpful error messages on failure

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Extended Input Formats

- **IO-07**: CLI reads .pdf files from source folder
- **IO-08**: CLI reads .html files from source folder
- **IO-09**: CLI reads .docx files from source folder

### Extended LLM Support

- **LLM-06**: CLI supports OpenAI API as LLM provider
- **LLM-07**: CLI supports Anthropic API as LLM provider
- **LLM-08**: Template customization via external files (override embedded templates)

### Advanced Quality

- **QUAL-05**: Semantic deduplication using embeddings
- **QUAL-06**: Quality filtering to reject low-quality pairs

### Performance

- **PERF-01**: Parallel document processing
- **PERF-02**: Resume/checkpoint for interrupted runs

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| GUI/Web interface | CLI-first approach; adds complexity without validating core value |
| Built-in model training | Scope creep; users choose their training tool |
| Auto-tuning parameters | Optimal settings vary by domain; adds complexity |
| Cloud storage integration | Users can upload via their own tools |
| Real-time generation preview | Slows down generation; users inspect after completion |
| Cost estimation | Less relevant for local Ollama models |
| Multiple output formats | JSONL is standard; wait for user demand |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| IO-01 | Phase 1 | ✓ Complete | 2026-02-05 |
| IO-02 | Phase 1 | ✓ Complete | 2026-02-05 |
| IO-03 | Phase 1 | ✓ Complete | 2026-02-05 |
| IO-04 | Phase 1 | ✓ Complete | 2026-02-05 |
| IO-05 | Phase 3 | ✓ Complete | 2026-02-06 |
| IO-06 | Phase 3 | ✓ Complete | 2026-02-06 |
| CLI-01 | Phase 1 | ✓ Complete | 2026-02-05 |
| CLI-02 | Phase 3 | ✓ Complete | 2026-02-06 |
| CLI-03 | Phase 3 | ✓ Complete | 2026-02-06 |
| CLI-04 | Phase 1 | ✓ Complete | 2026-02-05 |
| CLI-05 | Phase 1 | ✓ Complete | 2026-02-05 |
| LLM-01 | Phase 1 | ✓ Complete | 2026-02-05 |
| LLM-02 | Phase 1 | ✓ Complete | 2026-02-05 |
| LLM-03 | Phase 3 | ✓ Complete | 2026-02-06 |
| LLM-04 | Phase 3 | ✓ Complete | 2026-02-06 |
| LLM-05 | Phase 3 | ✓ Complete | 2026-02-06 |
| GEN-01 | Phase 1 | ✓ Complete | 2026-02-05 |
| GEN-02 | Phase 1 | ✓ Complete | 2026-02-05 |
| GEN-03 | Phase 1 | ✓ Complete | 2026-02-05 |
| GEN-04 | Phase 2 | ✓ Complete |
| GEN-05 | Phase 2 | ✓ Complete |
| GEN-06 | Phase 2 | ✓ Complete |
| GEN-07 | Phase 2 | ✓ Complete |
| GEN-08 | Phase 2 | ✓ Complete |
| QUAL-01 | Phase 2 | ✓ Complete |
| QUAL-02 | Phase 2 | ✓ Complete |
| QUAL-03 | Phase 2 | ✓ Complete |
| QUAL-04 | Phase 2 | ✓ Complete |
| UX-01 | Phase 3 | ✓ Complete | 2026-02-06 |
| UX-02 | Phase 3 | ✓ Complete | 2026-02-06 |
| UX-03 | Phase 3 | ✓ Complete | 2026-02-06 |
| UX-04 | Phase 1 | ✓ Complete | 2026-02-05 |
| UX-05 | Phase 3 | ✓ Complete | 2026-02-06 |
| UX-06 | Phase 3 | ✓ Complete | 2026-02-06 |

**Coverage:**
- v1 requirements: 34 total
- Mapped to phases: 34
- Unmapped: 0 ✓

---
*Requirements defined: 2025-02-05*
*Last updated: 2026-02-06 (all v1 requirements complete)*
