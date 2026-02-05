# Feature Research

**Domain:** LLM Fine-tuning Dataset Generation Tools
**Researched:** 2026-02-05
**Confidence:** MEDIUM

**Note:** This research is based on training knowledge (Jan 2025 cutoff) without WebSearch verification. Findings marked LOW confidence where ecosystem patterns may have shifted.

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Multiple input formats | Users have docs in various formats (MD, TXT, PDF, HTML) | MEDIUM | MD/TXT are LOW complexity; PDF/HTML require parsing libraries |
| JSONL output | Industry standard for LLM training data | LOW | Simple line-delimited JSON with prompt/completion pairs |
| Train/valid/test splits | Standard ML practice for evaluating model performance | LOW | Basic percentage-based splitting with file outputs |
| Deduplication | Duplicate examples degrade model training quality | MEDIUM | Exact match is LOW; semantic dedup is HIGH |
| Progress indication | Long-running jobs need visibility | LOW | Simple progress bar or file count tracking |
| Configurable parameters | Users need control over splits, counts, formatting | LOW | CLI flags or config file |
| Batch processing | Single documents aren't useful; need folder/directory support | LOW | Iterate over files in directory |
| Error handling | Files fail to process; LLM calls timeout; need graceful degradation | MEDIUM | Per-file error reporting without stopping entire run |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Smart pair generation | Use LLM to generate diverse Q&A pairs from docs | MEDIUM | Requires good prompting; quality varies by model |
| Exclusion-based generation | Prevent overlap between train/valid/test by passing exclusions to LLM | HIGH | Must track pairs across splits; increases prompt size |
| Template customization | Users can define their own prompt templates for different use cases | LOW | Load templates from files or embed in binary |
| Fill mechanism | Auto-generate more pairs when dedup reduces counts below target | MEDIUM | Requires detecting shortfall and making additional LLM calls |
| Quality filtering | Filter out low-quality pairs (too short, too similar, malformed) | HIGH | Requires heuristics or additional LLM evaluation |
| Domain-specific templates | Pre-built templates for code, APIs, FAQs, tutorials | MEDIUM | Requires testing templates per domain |
| Pair count based on doc length | Automatically scale pair generation to document size | LOW | Simple calculation: doc_length × pairs_per_1k |
| Embedded templates | No external file dependencies; everything in binary | LOW | Use Go embed package |
| Local LLM support | No API costs; works offline | MEDIUM | Integration with Ollama, llama.cpp, etc. |
| Multiple LLM providers | Support OpenAI, Anthropic, local models | MEDIUM | Abstraction layer for different APIs |
| Resume/checkpoint | Continue interrupted runs without reprocessing | HIGH | Requires state tracking, temp file management |
| Parallel processing | Process multiple documents concurrently | MEDIUM | Goroutines with rate limiting |
| Semantic deduplication | Detect similar pairs using embeddings, not just exact match | HIGH | Requires embedding model, similarity threshold tuning |
| Output format options | Support Alpaca, ShareGPT, ChatML, custom formats | MEDIUM | Template-based output formatting |
| Validation before output | Check JSONL validity, pair quality before writing | LOW | JSON parsing + basic heuristics |
| Cost estimation | Predict LLM API costs before running | LOW | Calculate tokens × rate (harder with local models) |
| Incremental updates | Add new docs without regenerating entire dataset | HIGH | Requires tracking what's been processed |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Auto-tune all parameters | "Just make it work without config" | Optimal settings vary wildly by domain; auto-tuning adds complexity and often guesses wrong | Provide good defaults with clear documentation on when to adjust |
| GUI/Web interface for v1 | "CLI is hard for non-technical users" | Adds significant complexity; delays core functionality; most users in this space are technical | Ship CLI first; add UI only if validated user demand |
| Real-time generation preview | "I want to see pairs as they're generated" | Slows down generation; creates UI complexity; most users care about final output | Show progress, log to file, inspect after completion |
| Automatic dataset balancing | "Balance categories/topics automatically" | Requires topic detection, adds complexity, often wrong; users have specific balance needs | Let users control splits explicitly; provide tools to inspect distribution |
| Built-in model training | "Generate datasets AND train the model" | Scope creep; training is complex with many frameworks; different use cases | Focus on dataset generation; users choose their training tool |
| Automatic prompt engineering | "Optimize prompts for best pairs" | Expensive (requires evaluation), slow, domain-specific; "best" is subjective | Provide good default templates; allow customization |
| Cloud storage integration | "Save directly to S3/GCS" | Adds dependencies, auth complexity; most users have local workflows | Write to local disk; users can upload via their own tools |
| Version control for datasets | "Track dataset changes over time" | Users already have git; reinventing VCS adds complexity | Generate timestamped files; users version with git |

## Feature Dependencies

```
[JSONL output]
    (foundational — everything builds on this)

[Input parsing (MD/TXT)]
    └──requires──> [Batch processing]
                       └──requires──> [Progress indication]

[LLM integration]
    └──requires──> [Template system]
                       └──enables──> [Smart pair generation]
                                         └──enables──> [Exclusion-based generation]
                                                           └──enables──> [Fill mechanism]

[Train/valid/test splits]
    └──requires──> [Deduplication]
                       └──enables──> [Fill mechanism]

[Template customization]
    ──enhances──> [Domain-specific templates]

[Local LLM support] ──conflicts──> [Cost estimation]
    (Cost estimation assumes API pricing; local models are "free")
```

### Dependency Notes

- **JSONL output is foundational:** All other features assume this output format exists
- **LLM integration enables smart generation:** Without LLM, you're just chunking docs (low value)
- **Exclusion-based generation requires template system:** Must pass exclusions via template variables
- **Fill mechanism requires dedup:** Can't know you're short on pairs until after dedup
- **Parallel processing conflicts with exclusion-based generation:** Hard to coordinate exclusions across concurrent processes

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] MD/TXT input parsing — Essential input formats
- [x] JSONL output — Standard format
- [x] Ollama LLM integration — Local inference, no API costs
- [x] Train/valid/test splits — Table stakes ML practice
- [x] Smart pair generation — Core value proposition
- [x] Exclusion-based generation — Quality differentiator
- [x] Deduplication (exact match) — Prevents training on duplicates
- [x] Fill mechanism — Ensures target pair counts
- [x] Template system with embed — Customization without external deps
- [x] Batch processing — Process folders, not single files
- [x] Progress indication — Visibility into long-running jobs
- [x] CLI configuration — Sensible defaults, overridable flags

### Add After Validation (v1.x)

Features to add once core is working and users provide feedback.

- [ ] Additional input formats (PDF, HTML, DOCX) — Wait for user demand by format
- [ ] Quality filtering — Add if users report low-quality pairs
- [ ] Validation before output — Add if users encounter malformed JSONL
- [ ] Multiple output formats (Alpaca, ShareGPT) — Wait for user requests per format
- [ ] Parallel processing — Add if single-threaded is too slow for real workloads
- [ ] Domain-specific templates — Build based on actual user domains

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Multiple LLM providers (OpenAI, Anthropic) — Ollama covers initial use case
- [ ] Semantic deduplication — High complexity; exact match may suffice
- [ ] Resume/checkpoint — High complexity; users can rerun quickly enough
- [ ] Incremental updates — Wait for users with large, evolving doc sets
- [ ] Cost estimation — Less relevant for local models

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| JSONL output | HIGH | LOW | P1 |
| MD/TXT parsing | HIGH | LOW | P1 |
| Train/valid/test splits | HIGH | LOW | P1 |
| Ollama integration | HIGH | MEDIUM | P1 |
| Smart pair generation | HIGH | MEDIUM | P1 |
| Exclusion-based generation | HIGH | HIGH | P1 |
| Deduplication (exact) | HIGH | MEDIUM | P1 |
| Fill mechanism | MEDIUM | MEDIUM | P1 |
| Template system | MEDIUM | LOW | P1 |
| Batch processing | HIGH | LOW | P1 |
| Progress indication | MEDIUM | LOW | P1 |
| CLI configuration | HIGH | LOW | P1 |
| PDF/HTML parsing | MEDIUM | MEDIUM | P2 |
| Quality filtering | MEDIUM | HIGH | P2 |
| Parallel processing | MEDIUM | MEDIUM | P2 |
| Domain templates | MEDIUM | MEDIUM | P2 |
| Multiple LLM providers | MEDIUM | MEDIUM | P2 |
| Semantic deduplication | LOW | HIGH | P3 |
| Resume/checkpoint | LOW | HIGH | P3 |
| Output format options | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch (validates core value)
- P2: Should have, add when users request (improves usability)
- P3: Nice to have, future consideration (low ROI for now)

## Competitor Feature Analysis

**Note:** LOW confidence — ecosystem may have evolved. Based on training knowledge of synthetic data and dataset generation tools.

| Feature | axolotl/alpaca-lora ecosystem | OpenAI fine-tuning workflow | datakeg approach |
|---------|------------------------------|----------------------------|------------------|
| Input format | Expects pre-made JSONL | Expects pre-made JSONL | Generates JSONL from raw docs |
| Pair generation | Manual or external scripts | Manual creation | LLM-generated with exclusions |
| Train/valid/test | Manual splits | Separate uploads | Automatic splits with dedup |
| LLM provider | N/A (data tools, not generators) | OpenAI API | Ollama (local) initially |
| Template system | Sometimes via scripts | N/A | Embedded, customizable |
| Deduplication | Manual or separate tools | Not provided | Built-in exact match |

**Key differentiation:** Most tools in this space are either:
1. **Training frameworks** (axolotl, alpaca-lora) that consume JSONL but don't generate it
2. **Manual processes** where humans write prompt/completion pairs
3. **Simple chunking scripts** that split docs but don't generate Q&A pairs

**datakeg differentiator:** Automated generation of diverse, deduplicated Q&A pairs from raw documentation using LLM intelligence, with built-in splits and exclusion logic.

## Rationale for MVP Choices

### Why Smart Pair Generation is P1
- **Core value proposition:** Without LLM-generated pairs, this is just a chunking script
- **Market gap:** Most tools require manual pair creation; automation is the differentiator
- **Validation dependency:** Can't validate value without the core feature

### Why Exclusion-Based Generation is P1
- **Quality differentiator:** Prevents train/valid/test overlap better than post-hoc dedup alone
- **Unique approach:** Not common in ecosystem based on training knowledge
- **Enables Fill mechanism:** Core to maintaining target pair counts after dedup

### Why Parallel Processing is P2
- **Performance unknown:** Don't know if sequential is too slow until we test real workloads
- **Complexity cost:** Coordinating exclusions across parallel processes is complex
- **Premature optimization:** Ship sequential first, parallelize if users complain

### Why Semantic Dedup is P3
- **Exact match may suffice:** Users may not need semantic dedup if exclusion logic works well
- **High complexity:** Requires embedding model, similarity tuning, significant compute
- **Unclear ROI:** Wait for user feedback on whether exact match is insufficient

### Why Resume/Checkpoint is P3
- **Fast enough without it:** If processing is quick enough, users can just rerun
- **High complexity:** State management, partial file handling, cleanup
- **Unknown demand:** Users may not run jobs large enough to need resume

## Sources

**Confidence note:** Research based on training knowledge (Jan 2025 cutoff) of:
- LLM fine-tuning ecosystem (axolotl, alpaca-lora, OpenAI fine-tuning)
- Synthetic data generation patterns
- JSONL dataset conventions
- ML train/valid/test best practices

**Verification needed:**
- Current tools in "documentation to training data" space (2026 tools may have emerged)
- Latest JSONL format conventions
- Ollama API stability and features

**Recommended verification:**
- Survey Hugging Face datasets for JSONL format patterns
- Check Ollama documentation for current API capabilities
- Search GitHub for "synthetic training data" or "LLM dataset generation" tools

---
*Feature research for: LLM fine-tuning dataset generation tools*
*Researched: 2026-02-05*
