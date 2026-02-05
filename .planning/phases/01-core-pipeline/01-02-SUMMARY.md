---
phase: 01-core-pipeline
plan: "01-02"
subsystem: llm-integration
tags: [go, ollama, templates, embedded, client-wrapper]

dependency_graph:
  requires: []
  provides:
    - "Ollama client for LLM text generation"
    - "Embedded prompt templates for train/valid/test splits"
  affects:
    - "01-03-generator" - Depends on these packages for generation

tech_stack:
  added:
    - "github.com/ollama/ollama/api v0.15.4"
  patterns:
    - "Streaming Ollama API client with context cancellation"
    - "Embedded filesystem templates with text/template"

key_files:
  created:
    - "internal/ollama/client.go"
    - "internal/ollama/client_test.go"
    - "internal/templates/templates.go"
    - "internal/templates/templates_test.go"
    - "internal/templates/prompts/train.tmpl"
    - "internal/templates/prompts/valid.tmpl"
    - "internal/templates/prompts/test.tmpl"
  modified:
    - "go.mod"
    - "go.sum"

decisions: []

metrics:
  duration: "2 min"
  completed: "2026-02-05"
---

# Phase 1 Plan 2: Ollama Client + Embedded Templates Summary

## Objective

Created the LLM communication layer and prompt template system to enable the generator (next plan) to call Ollama with properly formatted prompts.

## What Was Delivered

### 1. Ollama Client Wrapper (`internal/ollama/client.go`)

**Provides:**
- `Client` struct embedding the official `*api.Client`
- `NewClient()` function that creates client from `OLLAMA_HOST` environment variable
- `Generate(ctx, model, prompt)` method with streaming response support

**Key features:**
- Streaming responses accumulated in `strings.Builder` for memory efficiency
- Context cancellation support (Ctrl+C stops generation)
- Clean error wrapping with contextual messages

### 2. Embedded Prompt Templates (`internal/templates/`)

**Three template files for dataset splits:**

| File | Purpose |
|------|---------|
| `train.tmpl` | Training data generation - focuses on diverse QA pairs |
| `valid.tmpl` | Validation data generation - tests comprehension |
| `test.tmpl` | Test data generation - challenging evaluation pairs |

**Template variables:**
- `{{.DocumentContent}}` - Substituted with the document text
- `{{.PairCount}}` - Substituted with target number of pairs

**Implementation:**
- Templates embedded using `//go:embed prompts/*.tmpl` directive
- `ExecuteTemplate(name string, data PromptData)` function loads and executes templates
- `PromptData` struct holds DocumentContent, PairCount, and DocumentName

## Verification Results

| Check | Status |
|-------|--------|
| `go build ./internal/ollama` | ✓ Pass |
| `go build ./internal/templates` | ✓ Pass |
| `go test ./internal/ollama/...` | ✓ 3 tests skip (no Ollama) |
| `go test ./internal/templates/...` | ✓ 4 tests pass |

## Deviations from Plan

**None** - Plan executed exactly as written.

## Dependencies Added

```
github.com/ollama/ollama v0.15.4
├── github.com/bahlo/generic-list-go v0.2.0
├── github.com/buger/jsonparser v1.1.1
├── github.com/google/uuid v1.6.0
├── github.com/mailru/easyjson v0.7.7
├── github.com/wk8/go-ordered-map/v2 v2.1.8
├── golang.org/x/crypto v0.43.0
├── golang.org/x/sys v0.37.0
└── gopkg.in/yaml.v3 v3.0.1
```

## Next Steps

Plan 01-03 can now use these components:
- Import `internal/ollama` for LLM calls
- Import `internal/templates` for prompt template execution
