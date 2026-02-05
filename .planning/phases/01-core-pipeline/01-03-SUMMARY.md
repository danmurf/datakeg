---
phase: 01-core-pipeline
plan: "01-03"
subsystem: processing
tags: [go, file-processing, pair-generation, templates]

dependency_graph:
  requires:
    - "01-02-templates" - Uses embedded templates for prompt generation
    - "01-02-ollama" - Calls Ollama client for LLM generation
  provides:
    - "Document loading for .md and .txt files"
    - "Pair generation with configurable split ratios"
  affects:
    - "01-04-writer" - Depends on loaded documents and generated pairs

tech_stack:
  added: []
  patterns:
    - "filepath.Walk for recursive directory traversal"
    - "float64-based calculations with math.Ceil to avoid integer truncation"
    - "Split-type dispatching for template selection"

key_files:
  created:
    - "internal/processor/files.go"
    - "internal/generator/generator.go"
  modified: []

decisions: []

metrics:
  duration: "~2 min"
  completed: "2026-02-05"
---

# Phase 1 Plan 3: File Processor + Pair Generator Summary

## Objective

Created the document processing and pair generation logic to load files and generate prompt/completion pairs using the Ollama client and embedded templates from plan 01-02.

## What Was Delivered

### 1. File Processor (`internal/processor/files.go`)

**`LoadDocuments` function:**
- Recursively scans directory tree using `filepath.Walk`
- Filters for `.md` and `.txt` file extensions only
- Returns `[]Document` with Name, Path, and Content fields
- Skips directories, processes only matching files

**Document struct:**
```go
type Document struct {
    Name    string // Filename without extension
    Path    string // Full path to the file
    Content string // File contents
}
```

### 2. Pair Generator (`internal/generator/generator.go`)

**`Generator` struct:**
- Embeds Ollama client for LLM calls
- Holds configuration for pairs-per-1k-chars and split percentages
- `Generate(ctx, doc, splitType)` method creates pairs for a specific split

**Pair calculation using float64:**
```go
func (g *Generator) calculatePairs(content string) int {
    charCount := float64(len(content))
    pairs := math.Ceil(charCount / 1000 * g.config.PairsPer1KChars)
    return int(pairs)
}
```

**Split distribution:**
- Uses float64 percentages to calculate valid/test counts
- Train gets remainder: `total - valid - test`
- All math uses `math.Ceil` to avoid integer truncation

**Template selection per split type:**
| Split Type | Template |
|------------|----------|
| train | train.tmpl |
| valid | valid.tmpl |
| test | test.tmpl |

**Default configuration:**
- Pairs per 1K chars: 2.0
- Valid percentage: 10%
- Test percentage: 10%
- Model: gpt-oss:20b

## Verification Results

| Check | Status |
|-------|--------|
| `go build ./internal/processor` | ✓ Pass |
| `go build ./internal/generator` | ✓ Pass |
| `go build ./internal/...` | ✓ Pass |

## Deviations from Plan

**None** - Plan executed exactly as written.

## Dependencies

No new dependencies added.

## Next Steps

Plan 01-04 can now use these components:
- Import `internal/processor` to load documents from source directory
- Import `internal/generator` to create prompt/completion pairs
- Pass Generator to writer for JSONL output
