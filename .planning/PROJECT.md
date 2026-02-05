# datakeg

## What This Is

A Go CLI tool that converts raw documentation (Markdown, plain text) into LLM fine-tuning datasets. Point it at a folder of docs, and it generates JSONL files containing prompt/completion pairs suitable for training, validation, and testing.

## Core Value

Transform documentation into high-quality, deduplicated training data with minimal manual effort.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] CLI accepts source folder (docs) and output folder (training data) as arguments
- [ ] Processes .md and .txt files from source folder
- [ ] Calls LLM to generate prompt/completion pairs from each document
- [ ] Calculates pair count based on document length (configurable pairs per 1k chars)
- [ ] Splits pairs by configurable percentages (valid%, test%, train is remainder)
- [ ] Generates three JSONL files per document: train.jsonl, valid.jsonl, test.jsonl
- [ ] Uses separate prompt templates for train, valid, and test generation
- [ ] Passes exclusion pairs to LLM to prevent overlap (train→valid→test)
- [ ] Post-generation deduplication removes any remaining duplicates
- [ ] Fill mechanism generates additional pairs when dedup reduces count
- [ ] Merges all per-document files into final master train.jsonl, valid.jsonl, test.jsonl
- [ ] Supports Ollama as LLM provider with configurable model (default: gpt-oss:20b)
- [ ] Prompt templates use variables: {{document_content}}, {{num_pairs}}, {{document_name}}, {{exclude_pairs}}
- [ ] Prompt templates stored in folder and embedded into binary at build time
- [ ] Progress reporting shows current file, split type, and pair generation status
- [ ] All configuration has sensible defaults, overridable via CLI flags

### Out of Scope

- GUI or web interface — CLI only for v1
- Multiple LLM provider support beyond Ollama — extensible later but not v1
- Semantic deduplication — exact match only for v1
- Resume/checkpoint for interrupted runs — start fresh for v1
- Parallel document processing — sequential for v1 simplicity

## Context

**Output format:** Standard JSONL with `{"prompt": "...", "completion": "..."}` per line.

**File naming:** train.jsonl, valid.jsonl, test.jsonl (per-doc files prefixed with doc name).

**Generation flow per document:**
1. Calculate total pairs from doc length × pairs_per_1k
2. Determine counts: train = total - valid - test, based on percentages
3. Generate train pairs (no exclusions)
4. Generate valid pairs (exclude train pairs)
5. Generate test pairs (exclude train + valid pairs)
6. Deduplicate across all three sets
7. If any set is short, fill with additional LLM calls (passing all existing pairs as exclusions)
8. Write per-doc JSONL files

**Merge step:** After all documents processed, concatenate respective files into master set.

## Constraints

- **Language**: Go — for easy distribution as single binary
- **Initial LLM**: Ollama with gpt-oss:20b — local inference, no API costs
- **Templates**: Must be embedded in binary (no external file dependencies at runtime)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Post-generation dedup over inline | Passing all pairs to LLM would slow generation significantly | — Pending |
| Single {{exclude_pairs}} variable | Simpler than separate variables per context; populated appropriately | — Pending |
| Percentages for valid/test config | More intuitive than specifying train% (which is just the remainder) | — Pending |

---
*Last updated: 2025-02-05 after initialization*
