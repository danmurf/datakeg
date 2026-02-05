# Pitfalls Research

**Domain:** LLM Fine-tuning Dataset Generation Tools
**Researched:** 2026-02-05
**Confidence:** MEDIUM (Based on training knowledge; WebSearch unavailable for verification)

## Critical Pitfalls

### Pitfall 1: Generating Circular/Self-Referential Training Data

**What goes wrong:**
Using an LLM to generate training data from documentation, then fine-tuning the same or similar model on that data creates a circular knowledge loop. The model doesn't learn new information—it reinforces its existing patterns and biases, potentially amplifying errors or hallucinations present in the generated data.

**Why it happens:**
It seems efficient to use the same LLM infrastructure for both generation and fine-tuning. Developers don't realize that synthetic data from model X used to train model X (or similar) leads to model collapse or quality degradation over iterations.

**How to avoid:**
- Use generated data to fine-tune for format/style, not factual knowledge
- Treat source documentation as ground truth; generated prompts should query that truth
- Include source document references in training examples
- Validate that completions don't hallucinate beyond source material
- Consider the generation model as a "prompt engineering assistant" not a knowledge source

**Warning signs:**
- Generated completions contain information not in source documents
- Model becomes more confident but less accurate after fine-tuning
- Validation accuracy decreases while training accuracy increases
- Generated examples all sound similar or formulaic

**Phase to address:**
Phase 1 (Core generation) - Architecture must enforce source document as single source of truth

---

### Pitfall 2: Inadequate Deduplication Leading to Train/Test Leakage

**What goes wrong:**
Near-duplicate examples appear in both training and validation/test sets. This inflates accuracy metrics during evaluation but the model performs poorly on truly unseen data. Simple exact-match deduplication misses semantic duplicates or minor variations.

**Why it happens:**
Developers implement exact string matching for deduplication, which misses:
- Same content with different whitespace/formatting
- Semantic duplicates with different wording
- Prompt variations that lead to identical completions
- Document sections that appear in multiple source files

**How to avoid:**
- Use semantic similarity (embeddings + cosine similarity) for deduplication
- Deduplicate on completion content, not just prompts
- Deduplicate BEFORE splitting into train/valid/test
- Set similarity threshold (0.85-0.95 cosine similarity) to catch near-duplicates
- Hash normalized versions (lowercase, strip whitespace) as first pass
- Consider document-level tracking to avoid cross-contamination

**Warning signs:**
- Validation accuracy suspiciously close to training accuracy
- Model performs much worse on real queries than validation metrics suggest
- Many validation examples "feel familiar" when spot-checking
- Exact or near-exact duplicates found when manually reviewing splits

**Phase to address:**
Phase 1 (Core generation) - Must be implemented before any train/test split logic

---

### Pitfall 3: Prompt Template Overfitting

**What goes wrong:**
All training examples use identical or very similar prompt formats. The fine-tuned model learns to expect exact formatting and fails on natural variations. Users must learn the exact prompt structure the tool used, rather than being able to ask questions naturally.

**Why it happens:**
Using a single hardcoded prompt template is simpler to implement. Developers focus on generating working examples quickly without considering prompt diversity.

**How to avoid:**
- Generate multiple prompt variations per document chunk (3-5 variations minimum)
- Vary prompt structure: questions, commands, fill-in-the-blank, summarization requests
- Include both explicit and implicit context styles
- Randomize formality levels, question types, and framing
- Test generated prompts against a "prompt variation score"
- Allow users to provide custom prompt templates

**Warning signs:**
- All generated prompts start with the same phrase (e.g., "What is...")
- Prompts follow a rigid structure (e.g., always question format)
- Manual inspection shows repetitive phrasing patterns
- Fine-tuned model works with exact template but fails on paraphrases

**Phase to address:**
Phase 2 (Advanced generation) - Requires template system and variation logic

---

### Pitfall 4: Ignoring Document Structure and Context Windows

**What goes wrong:**
Chunking documents without respecting semantic boundaries (sections, paragraphs, code blocks). Chunks cut off mid-sentence or split related information. Context window limits are ignored, leading to truncated examples or incomplete information in completions.

**Why it happens:**
Simple line-count or character-count chunking is easy to implement. Developers don't account for markdown structure, code fences, lists, or tables. Context window limits aren't validated until examples fail during fine-tuning.

**How to avoid:**
- Parse markdown structure to find semantic boundaries
- Keep related content together (code blocks, lists, tables)
- Add overlap between chunks to preserve context
- Validate that prompt + completion fits within model's context window
- Add chunk metadata (section title, hierarchy level)
- Respect language-specific boundaries (code function boundaries, paragraph breaks)

**Warning signs:**
- Completions reference information not in the prompt
- Code examples are incomplete or syntactically broken
- Lists or tables are split across examples
- Tokenization errors or truncation warnings during fine-tuning
- Generated examples feel disjointed or lack context

**Phase to address:**
Phase 1 (Core generation) - Document parsing must respect structure from the start

---

### Pitfall 5: No Quality Filtering or Validation

**What goes wrong:**
Low-quality generated examples pollute the training set. Examples with hallucinations, incorrect information, poor formatting, or nonsensical completions degrade model performance. "Garbage in, garbage out" at scale.

**Why it happens:**
Validating LLM outputs is harder than generating them. Developers assume the generation model is reliable or lack metrics to identify bad examples. Manual review doesn't scale to thousands of examples.

**How to avoid:**
- Implement automated quality checks:
  - Completion length (too short = incomplete, too long = rambling)
  - Source grounding (completion only uses information from prompt/context)
  - Format validation (proper markdown, code syntax if applicable)
  - Semantic coherence (prompt and completion are related)
- Use a separate model to score example quality
- Sample random examples for manual review (spot checks)
- Track generation model temperature/sampling params per example
- Log and report rejected examples for analysis

**Warning signs:**
- Wide variance in completion quality when manually reviewing
- Some completions are generic filler text
- Completions contradict source documentation
- Code examples don't compile or run
- Answers to questions don't actually answer the question

**Phase to address:**
Phase 2 (Advanced generation) - Add quality filters after basic generation works

---

### Pitfall 6: Fixed Train/Valid/Test Split Ratios Without Considering Dataset Size

**What goes wrong:**
Using standard 80/10/10 or 70/15/15 splits regardless of total dataset size. With small datasets (< 1000 examples), validation/test sets are too small for meaningful evaluation. With large datasets (> 100k examples), validation sets are unnecessarily large, wasting potential training data.

**Why it happens:**
Following machine learning conventions without adapting to dataset characteristics. Not considering that statistical significance requires minimum sample sizes, not fixed percentages.

**How to avoid:**
- Use absolute minimum sizes for validation/test (e.g., 500 examples each)
- Adjust ratios based on total size:
  - < 1000 examples: 70/20/10 or even 80/10/10
  - 1000-10000: 80/10/10
  - > 10000: 90/5/5 or even 95/2.5/2.5
- Ensure validation set is large enough for reliable metrics (>= 500 examples)
- Consider k-fold cross-validation for small datasets
- Document split strategy and rationale

**Warning signs:**
- Validation set has < 100 examples (too small for reliable metrics)
- Test set has < 50 examples (too small for final evaluation)
- With 50k examples, using 5k for validation (wasteful)
- High variance in validation metrics between runs
- Unable to detect overfitting due to small validation set

**Phase to address:**
Phase 1 (Core generation) - Split logic should be intelligent from the start

---

### Pitfall 7: Not Handling Ollama API Failures and Rate Limits

**What goes wrong:**
Generation runs crash or hang when Ollama is unavailable, overloaded, or returns errors. Long-running generation jobs fail completely without partial results. No retry logic means transient failures lose progress.

**Why it happens:**
Happy-path coding assumes Ollama always responds successfully. Developers don't test failure modes (connection refused, timeout, rate limit, out of memory). Error handling is added as an afterthought.

**How to avoid:**
- Implement exponential backoff retry logic for Ollama calls
- Save progress incrementally (checkpoint after each successful generation)
- Validate Ollama connection before starting generation
- Add timeout configuration for LLM calls
- Handle specific Ollama error codes appropriately
- Allow resuming generation from checkpoint on failure
- Add concurrency limits to avoid overwhelming Ollama

**Warning signs:**
- Tool hangs indefinitely when Ollama is slow
- Crashes without saving any results on errors
- "Connection refused" errors are not caught
- No way to recover partial results from failed runs
- Users must regenerate everything after any failure

**Phase to address:**
Phase 1 (Core generation) - Error handling must be present from the start

---

### Pitfall 8: Embedding Prompts Without Version/Provenance Tracking

**What goes wrong:**
Prompt templates are embedded in the binary with no version tracking. When templates change, there's no way to know which version generated which datasets. Reproducibility is impossible. Can't debug why datasets generated at different times behave differently.

**Why it happens:**
Embedded templates are convenient for distribution but easy to forget to version. Developers update templates during development without tracking changes. No metadata is saved with generated datasets.

**How to avoid:**
- Version embedded prompt templates (semantic versioning)
- Save template version in dataset metadata/JSONL comments
- Include generation timestamp and tool version in output
- Hash prompt templates and store hash with results
- Add `--template-version` flag to show embedded versions
- Consider external template files for easier versioning (with default embedded fallback)
- Log all generation parameters (model, temperature, template version) with dataset

**Warning signs:**
- Can't reproduce a dataset generated last week
- Unknown which template version generated a given dataset
- Template changes break existing workflows without warning
- No audit trail for dataset generation
- Debugging issues requires guessing which template was used

**Phase to address:**
Phase 1 (Core generation) - Metadata tracking should be built in from the start

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Exact string deduplication only | Easy to implement, fast | Misses semantic duplicates, inflates metrics | Never - use hash + semantic from start |
| Single hardcoded prompt template | Simple, quick to ship | Severe overfitting, poor model generalization | MVP only, replace in Phase 2 |
| Fixed 80/10/10 split regardless of size | Standard practice, no logic needed | Wastes data or creates unreliable metrics | Never - implement adaptive splitting |
| No checkpointing during generation | Simpler code, fewer files | Lost progress on any failure | Very small datasets (< 100 examples) |
| Ignoring Ollama errors | Happy-path works quickly | Production failures, poor UX | Never - handle from the start |
| Character-count chunking only | Trivial to implement | Breaks structure, poor quality | Throwaway experiments only |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Ollama API | Assuming localhost:11434 always available | Check connection, configurable endpoint, graceful failure |
| Ollama API | Not handling context window limits | Validate prompt+completion token count before calling |
| Ollama API | Ignoring model-specific capabilities | Verify model supports required features (e.g., function calling) |
| File I/O | Loading entire file into memory | Stream large files, chunked processing |
| JSONL output | Not validating JSON per line | Validate each line independently before writing |
| Markdown parsing | Treating markdown as plain text | Use proper parser to respect structure (code blocks, tables) |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Loading all documents into memory | OOM errors, slow startup | Stream processing, lazy loading | > 100MB of source docs |
| Sequential LLM calls | Very slow generation | Concurrent requests (with rate limiting) | > 100 examples to generate |
| No caching of embeddings | Slow deduplication, repeated work | Cache embeddings per chunk | > 1000 examples |
| Regenerating from scratch on failure | Long recovery time, frustration | Checkpoint and resume | > 30 min generation time |
| Computing edit distance on all pairs | O(n²) deduplication time | Hash-based first pass + similarity for candidates | > 10k examples |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| No input validation on file paths | Path traversal, reading sensitive files | Validate paths stay within source directory |
| Embedding API keys in binary | Key exposure in version control | Use environment variables or config files |
| No sanitization of LLM outputs | Injection attacks if dataset used in production | Sanitize/escape generated content |
| Running Ollama with elevated privileges | Unnecessary attack surface | Run with minimal required permissions |
| Including sensitive data in training examples | Data leakage through fine-tuned model | Filter/redact sensitive info before generation |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No progress reporting | User has no idea how long it will take | Show progress bar, ETA, examples/sec |
| Unclear error messages | User doesn't know how to fix issues | Specific errors with suggested fixes |
| All-or-nothing generation | Can't test with small samples | Add `--limit` flag to generate N examples |
| No dry-run mode | Can't preview before committing | Add `--dry-run` to show what would be generated |
| Requiring exact flag syntax | Frustrating to remember | Sensible defaults, common aliases |
| No summary of what was generated | User must inspect files to understand output | Print summary stats (counts, splits, quality metrics) |

## "Looks Done But Isn't" Checklist

- [ ] **Deduplication:** Often missing semantic similarity check — verify by checking cosine similarity of random example pairs
- [ ] **Train/test split:** Often missing validation that no duplicates across splits — verify by hashing all examples per split
- [ ] **Prompt variation:** Often missing diversity metrics — verify by computing prompt similarity within dataset
- [ ] **Error handling:** Often missing Ollama connection failures — verify by stopping Ollama during generation
- [ ] **Context window:** Often missing token count validation — verify with examples that should exceed limits
- [ ] **Quality filtering:** Often missing automated checks — verify by intentionally generating bad examples
- [ ] **Progress reporting:** Often missing time estimates — verify by watching output during long runs
- [ ] **Checkpoint/resume:** Often missing incremental saves — verify by killing process mid-run and resuming
- [ ] **Metadata tracking:** Often missing generation parameters — verify that output includes version/timestamp
- [ ] **Document structure:** Often missing markdown parsing — verify with complex docs (code blocks, tables)

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Circular training data | HIGH | Regenerate with different approach; can't fix existing data |
| Train/test leakage | MEDIUM | Re-split with proper deduplication; keep examples but redo splits |
| Prompt template overfitting | HIGH | Regenerate with varied prompts; can't fix model already trained |
| Broken document chunking | MEDIUM | Fix parser, regenerate affected examples; source docs still valid |
| No quality filtering | LOW-MEDIUM | Run filtering pass on existing dataset; most examples salvageable |
| Fixed split ratios | LOW | Re-split existing examples with better ratios; no regeneration needed |
| Ollama failures | NONE | Retry from checkpoint; no data loss |
| No version tracking | NONE | Start tracking now; can't recover past versions but can prevent future issues |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Circular/self-referential data | Phase 1: Core generation | Manual review of sample completions vs source docs |
| Train/test leakage | Phase 1: Core generation | Compute similarity scores across splits |
| Prompt template overfitting | Phase 2: Advanced generation | Measure prompt diversity with similarity metrics |
| Ignoring document structure | Phase 1: Core generation | Review chunked output for broken code/tables |
| No quality filtering | Phase 2: Advanced generation | Sample random examples, measure quality scores |
| Fixed split ratios | Phase 1: Core generation | Verify validation set size is appropriate |
| Ollama API failures | Phase 1: Core generation | Integration test with Ollama unavailable |
| No version tracking | Phase 1: Core generation | Check that output includes metadata |

## Sources

**Confidence Note:** This research is based on training knowledge (January 2025 cutoff). WebSearch was unavailable for verification. These are well-established patterns in ML/LLM dataset generation but should be validated against current best practices (2026) when possible.

Domain knowledge sources (from training):
- LLM fine-tuning best practices literature
- Synthetic data generation research
- ML dataset quality frameworks
- Common patterns from dataset generation tools (Argilla, LabelStudio, custom tooling)
- Known issues in train/test contamination
- Document processing and chunking strategies

Recommended validation:
- Check current Ollama documentation for API best practices
- Review recent papers on synthetic training data quality
- Consult latest LLM fine-tuning guides from model providers
- Verify current embedding model recommendations for deduplication

---
*Pitfalls research for: LLM Fine-tuning Dataset Generation Tools*
*Researched: 2026-02-05*
*Confidence: MEDIUM (training knowledge without current source verification)*
