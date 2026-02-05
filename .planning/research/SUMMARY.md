# Project Research Summary

**Project:** datakeg
**Domain:** Go CLI for LLM fine-tuning dataset generation
**Researched:** 2026-02-05
**Confidence:** MEDIUM

## Executive Summary

datakeg is a Go CLI tool that transforms raw documentation (markdown/text) into high-quality fine-tuning datasets using local LLM generation. The research reveals this is a unique positioning in the current ecosystem: most tools either provide training frameworks that consume JSONL data, or are manual processes where humans write prompt/completion pairs. datakeg bridges the gap by automating the generation of diverse, deduplicated question-answer pairs from documentation using local LLMs (Ollama), with intelligent exclusion logic to prevent train/validation/test overlap.

The recommended approach leverages Go's excellent standard library with minimal dependencies (Cobra for CLI, progressbar for UX), using a pipeline orchestrator pattern to coordinate reading, processing, deduplication, and output phases. The architecture emphasizes separation of concerns (CLI layer, orchestration layer, processing layer, integration layer) with worker pools for concurrent LLM processing. Critical to success is implementing semantic deduplication and exclusion-based generation from the start—not as afterthoughts—since retrofitting these after initial implementation multiplies complexity.

The primary risk is generating circular/self-referential training data where the LLM reinforces its own patterns rather than learning from documentation. This is mitigated by treating source documents as the single source of truth, validating completions against source material, and implementing quality filtering. Secondary risks include train/test leakage from inadequate deduplication and prompt template overfitting from using single static templates. Both are addressable through proper architecture choices in Phase 1.

## Key Findings

### Recommended Stack

Go 1.23+ with a moderate dependency approach: leverage stdlib for heavy lifting (HTTP, JSON, file I/O, embedding) while using quality dependencies where they significantly improve DX/UX. The Cobra CLI framework is recommended over stdlib flag package due to multiple complex flags and subcommand extensibility. For progress reporting, progressbar/v3 provides professional UX for long-running operations. Ollama integration should try the official Go SDK first (if it exists) but fallback to stdlib http.Client + encoding/json since Ollama's REST API is straightforward.

**Core technologies:**
- Go 1.23+: Current stable with embed support, excellent for CLI distribution
- Cobra v1.8+: Industry standard for Go CLIs (used by kubectl, gh, hugo), handles complex flag scenarios
- progressbar/v3: Clean progress bars with ETA for long-running LLM generation
- Ollama (Go SDK or stdlib HTTP): Local LLM inference, no API costs, works offline
- embed (stdlib): Self-contained binary with prompt templates, zero runtime dependencies

**Critical version note:** Research based on training data through Jan 2025. All library versions require verification before implementation. Particularly verify if github.com/ollama/ollama-go exists and is maintained.

### Expected Features

The MVP feature set is carefully scoped to validate the core value proposition: automated generation of diverse, deduplicated training pairs from raw documentation. This distinguishes datakeg from simple chunking scripts or manual pair creation workflows.

**Must have (table stakes):**
- MD/TXT input parsing with batch processing
- JSONL output (industry standard for LLM training)
- Train/validation/test splits (standard ML practice)
- Exact match deduplication (prevents duplicate examples)
- Progress indication (visibility for long-running jobs)
- Error handling (LLM calls fail; need graceful degradation)

**Should have (competitive advantage):**
- Smart pair generation using LLM (core value proposition)
- Exclusion-based generation (prevents overlap better than post-hoc dedup)
- Fill mechanism (maintains target counts after dedup)
- Template system with embed (customization without external dependencies)
- Configurable parameters via CLI flags

**Defer (v2+):**
- Additional input formats (PDF, HTML, DOCX) — wait for user demand per format
- Multiple LLM providers (OpenAI, Anthropic) — Ollama covers initial use case
- Semantic deduplication — high complexity, exact match may suffice initially
- Parallel processing — optimize only if single-threaded proves too slow
- Resume/checkpoint — add if jobs become long enough to require it

**Anti-features to avoid:**
- GUI/web interface for v1 (scope creep, delays core value)
- Auto-tune all parameters (settings vary wildly by domain)
- Built-in model training (out of scope; users choose their training framework)
- Cloud storage integration (adds auth complexity; users upload via their tools)

### Architecture Approach

A layered architecture with clear separation of concerns enables testing components independently and avoids tight coupling. The pipeline orchestrator pattern coordinates sequential phases (read → process → dedupe → fill → write) while worker pools handle concurrent LLM calls within the process phase. Context-based cancellation throughout enables graceful shutdown with Ctrl-C.

**Major components:**
1. **CLI Parser** (Cobra) — flag parsing, validation, help generation; isolated from core logic
2. **Pipeline Orchestrator** — coordinates execution phases, manages state between them, single point of control
3. **Document Reader** — discovers and reads source files, filters by extension (.md/.txt)
4. **LLM Processor** — calls Ollama API with retry logic, implements worker pool for concurrent processing
5. **Deduplicator + Filler** — hash-based exact match detection, regenerates pairs when counts fall short
6. **Template Manager** — embeds prompt templates in binary using embed.FS, renders with text/template
7. **JSONL Writer** — writes per-document files immediately (progressive enhancement), merges at end

**Key architectural patterns:**
- Worker pool with errgroup for concurrent LLM calls (I/O bound, benefits from parallelism)
- Progressive enhancement output (write per-doc JSONL immediately, provides crash recovery)
- Retry with exponential backoff (handles transient Ollama failures)
- Context-based cancellation (graceful Ctrl-C shutdown)

**Project structure:** cmd/datakeg/main.go as entry point with internal/ packages organized by concern (cli/, pipeline/, document/, llm/, dataset/, output/, templates/). Templates embedded from templates/ directory at compile time.

### Critical Pitfalls

Research identified 8 critical pitfalls, with the top 5 most relevant to initial phases:

1. **Generating circular/self-referential training data** — Using LLM output to train the same model creates knowledge loops that amplify errors. Prevention: treat source docs as single source of truth, validate completions don't hallucinate beyond source material, include source references in examples.

2. **Inadequate deduplication leading to train/test leakage** — Exact string matching misses semantic duplicates. Prevention: implement semantic similarity (embeddings + cosine similarity), deduplicate BEFORE splitting into train/valid/test, normalize before hashing, set 0.85-0.95 similarity threshold.

3. **Prompt template overfitting** — Single hardcoded template causes models to expect exact formatting. Prevention: generate 3-5 prompt variations per chunk, vary structure (questions/commands/fill-in-blank), randomize formality levels. *Note: Phase 2 concern, not Phase 1.*

4. **Ignoring document structure and context windows** — Character-count chunking breaks code blocks, tables, semantic boundaries. Prevention: parse markdown structure, respect semantic boundaries, add overlap between chunks, validate prompt+completion fits context window.

5. **Not handling Ollama API failures and rate limits** — No retry logic means transient failures lose progress. Prevention: exponential backoff retry, save progress incrementally (checkpoint), validate connection before starting, handle specific error codes, add concurrency limits.

**Phase 1 critical:** Pitfalls #1, #2, #4, #5 must be addressed in core generation phase. Retrofitting these after initial implementation multiplies complexity. Pitfall #3 (prompt variation) can be deferred to Phase 2 advanced generation.

## Implications for Roadmap

Based on combined research, suggested 4-phase structure with clear dependencies and rationale:

### Phase 1: Foundation & Basic Generation
**Rationale:** Establishes foundational components that everything else depends on. Cannot test LLM generation until basic CLI, document reading, and Ollama integration work. Deduplication architecture must be correct from the start (retrofitting semantic dedup is HIGH cost per PITFALLS.md recovery strategies).

**Delivers:** Working CLI that reads markdown files, generates training pairs via Ollama, performs basic deduplication, writes JSONL output

**Includes:**
- CLI framework with Cobra (flag parsing, validation)
- Document reader (MD/TXT input, batch processing)
- Ollama client with retry logic (addresses Pitfall #5)
- Basic template system with embed (single template per split type)
- LLM processor (single-threaded initially)
- Hash-based exact match deduplication (addresses Pitfall #2 partially)
- Train/valid/test splitting with adaptive ratios
- JSONL output writer
- Progress reporting (basic console output)

**Avoids:**
- Circular training data (Pitfall #1) — validate completions against source
- Document structure issues (Pitfall #4) — parse markdown properly
- Ollama failures (Pitfall #5) — retry logic from start

**Research needs:** LOW — well-documented patterns for Go CLI, Ollama API is straightforward HTTP

### Phase 2: Smart Generation & Quality
**Rationale:** Once basic generation works, enhance with features that differentiate datakeg from simple chunking scripts. Exclusion-based generation is the key competitive advantage per FEATURES.md analysis. Quality filtering prevents garbage data from polluting training sets.

**Delivers:** Intelligent pair generation with diversity, exclusion logic to prevent overlap, quality filtering

**Includes:**
- Exclusion-based generation (pass previously generated pairs to LLM to avoid duplicates)
- Fill mechanism (regenerate when dedup reduces counts below target)
- Semantic deduplication (embeddings + cosine similarity, addresses Pitfall #2 fully)
- Template variation system (addresses Pitfall #3)
- Quality filtering (length checks, format validation, source grounding)
- Metadata tracking (template version, generation timestamp, addresses Pitfall #8)

**Avoids:**
- Train/test leakage (Pitfall #2) — semantic dedup catches near-duplicates
- Prompt overfitting (Pitfall #3) — template variations increase diversity
- Low quality pollution (Pitfall #5 from research) — automated filtering

**Research needs:** MEDIUM — embedding model selection for semantic dedup, quality metrics definition, exclusion prompt design

### Phase 3: Concurrency & Performance
**Rationale:** Once generation quality is solid, optimize for performance. Worker pools for concurrent LLM calls provide significant speedup (I/O bound operations). Only optimize after measuring real bottlenecks per ARCHITECTURE.md scaling considerations.

**Delivers:** Fast generation through concurrent processing, handles large document sets efficiently

**Includes:**
- Worker pool for parallel LLM processing (errgroup pattern)
- Configurable concurrency (--workers flag, respects Ollama capacity)
- Rate limiting (avoid overwhelming Ollama)
- Improved progress reporting (progressbar/v3 with ETA)
- Memory-efficient streaming (don't load all docs in memory)

**Architecture dependency:** Must coordinate exclusion-based generation across concurrent workers (complexity noted in FEATURES.md dependency analysis)

**Research needs:** LOW — worker pool is standard Go pattern, well-documented

### Phase 4: Polish & Distribution
**Rationale:** Final UX improvements and packaging for distribution. Makes tool production-ready and accessible.

**Delivers:** Professional UX, easy installation, comprehensive documentation

**Includes:**
- Enhanced error messages with fix suggestions
- Summary statistics after generation (counts, quality metrics)
- Dry-run mode (--dry-run to preview without generating)
- Limit flag (--limit N for testing with small samples)
- Binary distribution setup (goreleaser for multi-platform builds)
- Documentation (README, usage examples, troubleshooting)

**Research needs:** NONE — standard packaging patterns

### Phase Ordering Rationale

- **Phase 1 before 2:** Cannot implement exclusion logic or quality filtering without basic generation working. Deduplication architecture must be correct from start (HIGH recovery cost per PITFALLS.md).

- **Phase 2 before 3:** Must validate generation quality before optimizing performance. Concurrent processing complicates exclusion-based generation per FEATURES.md conflict note—need working sequential version first.

- **Phase 3 before 4:** Performance optimization shapes UX (progress reporting needs accurate metrics). Polish phase depends on knowing actual bottlenecks.

- **Grouping rationale:** Each phase delivers standalone value and validates assumptions before adding complexity. Phases align with natural testing boundaries (unit test → integration test → performance test → UX validation).

### Research Flags

**Phases needing deeper research during planning:**
- **Phase 2:** Embedding model selection (which model for semantic dedup?), quality metrics definition (what makes a "good" pair?), exclusion prompt engineering (how to pass exclusions effectively?)
- **Phase 2:** Semantic similarity threshold tuning (0.85? 0.90? 0.95? depends on embedding model)

**Phases with standard patterns (skip research-phase):**
- **Phase 1:** Go CLI setup, Cobra framework, file I/O, HTTP clients — well-documented, established patterns
- **Phase 3:** Worker pools with errgroup — standard Go concurrency pattern, thoroughly documented
- **Phase 4:** Binary distribution with goreleaser — standard tooling, good documentation

**During Phase 2 planning:** Invoke `/gsd:research-phase` to investigate embedding model options, quality filtering metrics, and exclusion prompt strategies. This is the highest uncertainty area.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM | Cobra/Viper are well-established, but cannot verify Ollama Go SDK status without web access. Stdlib fallback is HIGH confidence. |
| Features | HIGH | Feature prioritization based on clear competitive analysis and MVP principles. Exclusion-based generation as differentiator is well-justified. |
| Architecture | HIGH | Pipeline orchestrator and worker pool patterns are standard Go practices, well-documented in ecosystem. Component separation enables testability. |
| Pitfalls | HIGH | Pitfalls based on established ML/LLM dataset generation knowledge. Train/test leakage and circular data are well-known issues with documented solutions. |

**Overall confidence:** MEDIUM-HIGH

Research is comprehensive and internally consistent. Stack recommendations require version verification (MEDIUM confidence on specific versions, HIGH confidence on patterns). Feature prioritization and architectural patterns are solid (HIGH confidence). Pitfall identification is thorough and maps clearly to prevention strategies (HIGH confidence).

### Gaps to Address

Research limitations due to unavailable web search/fetch tools:

1. **Ollama Go SDK verification:** Cannot confirm if github.com/ollama/ollama-go or github.com/ollama/ollama exists and is maintained. **Action:** Verify during Phase 1 kickoff. If SDK unavailable, use stdlib http.Client (HIGH confidence fallback).

2. **Current library versions:** Cobra v1.8+, Viper v1.18+, progressbar/v3 v3.14+ are from training data through Jan 2025. **Action:** Check latest releases before go get, adjust for any API changes.

3. **Embedding model for semantic dedup:** Research doesn't specify which embedding model to use (sentence-transformers? OpenAI embeddings? local model?). **Action:** Research during Phase 2 planning via `/gsd:research-phase`.

4. **Ollama context window limits:** Research mentions validating context windows but doesn't specify Ollama's limits per model. **Action:** Check Ollama documentation during Phase 1; implement dynamic validation based on selected model.

5. **Quality filtering metrics:** Pitfall research identifies need for quality filtering but doesn't define specific metrics or thresholds. **Action:** Research during Phase 2 planning; consider length bounds, format checks, source grounding algorithms.

6. **Exclusion prompt engineering:** Research identifies exclusion-based generation as key differentiator but doesn't provide prompt templates. **Action:** Research during Phase 2 planning; test various prompt formats for passing exclusions.

**None of these gaps block Phase 1 implementation.** Ollama SDK verification is needed immediately but has clear fallback. Other gaps are Phase 2 concerns addressable through focused research during that phase's planning.

## Sources

### Primary (HIGH confidence - architectural patterns)
- Go standard library patterns (context, errgroup, embed, filepath, bufio, encoding/json)
- Go CLI best practices (cmd/ structure, internal/ packages, flag parsing)
- Pipeline and worker pool patterns from Go ecosystem
- ML train/validation/test split conventions
- JSONL format conventions for LLM training data

### Secondary (MEDIUM confidence - ecosystem knowledge)
- Cobra/Viper CLI frameworks (training data through Jan 2025, requires version verification)
- Ollama API patterns (REST API structure is clear, SDK status unverified)
- LLM fine-tuning dataset generation patterns (axolotl, alpaca-lora ecosystems)
- Synthetic data generation best practices
- Document processing and chunking strategies for LLM context windows

### Tertiary (LOW confidence - needs validation)
- Specific library versions (Cobra v1.8+, Viper v1.18+, progressbar/v3 v3.14+)
- Ollama Go SDK existence and quality
- Current state of dataset generation tools (2026 tools may have emerged since training cutoff)
- Embedding model recommendations for deduplication (need current benchmarks)

### Recommended Verification
Before Phase 1 implementation:
- [ ] Verify Ollama Go SDK at github.com/ollama/ollama or github.com/ollama/ollama-go
- [ ] Check Ollama REST API documentation for current endpoint structure
- [ ] Verify latest Cobra release (github.com/spf13/cobra/releases)
- [ ] Check progressbar/v3 latest release (github.com/schollz/progressbar/releases)
- [ ] Confirm Go 1.23+ embed syntax hasn't changed

Before Phase 2 implementation:
- [ ] Research current embedding models for semantic similarity
- [ ] Survey Hugging Face datasets for JSONL format patterns
- [ ] Search GitHub for "LLM dataset generation" tools emerged since Jan 2025
- [ ] Investigate quality filtering metrics from recent papers

---
*Research completed: 2026-02-05*
*Ready for roadmap: yes*
