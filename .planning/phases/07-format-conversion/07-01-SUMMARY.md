---
phase: 07-format-conversion
plan: "01"
type: execute
subsystem: converter
tags: ["conversion", "templates", "jsonl", "training-data"]
completed: 2026-02-07
duration: 5 min
---

# Phase 7 Plan 1: Conversion Templates and Converter Package Summary

## Overview

Implemented the conversion engine for transforming generated JSONL files into model-specific training formats. Added 4 built-in conversion templates and a converter package with streaming line-by-line processing.

## Key Files Created

| File | Purpose |
|------|---------|
| `internal/templates/conversions/mistral-instruct.tmpl` | Mistral Instruct format with `[INST]` tags |
| `internal/templates/conversions/llama3-instruct.tmpl` | Llama 3 Instruct with header tags |
| `internal/templates/conversions/chatml.tmpl` | ChatML format with `<|im_start|>` tags |
| `internal/templates/conversions/deepseek-r1.tmpl` | DeepSeek-R1 integrated reasoning format |
| `internal/converter/converter.go` | Core conversion logic |
| `internal/converter/converter_test.go` | Comprehensive tests |
| `internal/templates/templates.go` | Added conversion template loaders |

## Tech Stack Added

- **Library**: `text/template` - Go standard library for template execution
- **Library**: `bufio.Scanner` - Streaming line-by-line file reading
- **Library**: `encoding/json` - JSON parsing for format detection
- **Pattern**: `//go:embed` - Embedded templates in binary

## Conversion Templates

| Template | Source Format | Output Format | Description |
|----------|---------------|---------------|-------------|
| mistral-instruct | completion | Mistral Instruct | `{"text": "<s>[INST] {{.Prompt}} [/INST] {{.Completion}}</s>"}` |
| llama3-instruct | chat | Llama 3 | `{"text": "<|begin_of_text|>{{range .Messages}}..."}` |
| chatml | chat | ChatML | `{"text": "{{range .Messages}}<|im_start|>{{.Role}}..."}` |
| deepseek-r1 | reasoning | Integrated | `{"prompt": "{{.Question}}", "completion": "{{.Reasoning}}\n\n{{.Answer}}"}` |

## Converter Functions

| Function | Purpose |
|----------|---------|
| `ConvertJSONL(input, output, tmpl, format)` | Streams JSONL line-by-line through template |
| `DetectFormat(jsonLine)` | Detects format from JSON field structure |
| `DetectFormatFromFile(path)` | Reads first line and detects format |
| `ValidateTemplate(tmpl, format)` | Pre-validates template produces valid JSON |

## Template Functions

| Function | Purpose |
|----------|---------|
| `LoadConversionTemplate(name)` | Loads built-in template from embedded FS |
| `LoadCustomConversionTemplate(path)` | Loads user-provided template file |
| `ListConversionTemplates()` | Returns available template names |
| `jsonEscape(s)` | Escapes special JSON characters |

## Tests Added

- Format detection (completion, chat, reasoning, unknown)
- Format detection from file (including empty/whitespace-only)
- Template validation (valid templates, invalid field references)
- JSONL conversion (completion, chat formats)
- Edge cases: empty files, invalid JSON, special characters
- Built-in template loading and listing

## Verification

- `make lint` passes ✓
- `make test` passes ✓
- `go build ./...` compiles ✓
- All 12 converter tests pass ✓

## Decisions Made

1. **Template loading**: Used `ReadFile` + `Parse` instead of `ParseFS` for embedded templates to ensure proper template name handling
2. **jsonEscape function**: Added to templates package for consistent JSON escaping across all conversion templates
3. **Streaming conversion**: Used bufio.Scanner with 1MB buffer for line-by-line processing, handles large JSONL files without memory issues

## Dependencies

- **Requires**: Phase 6 (Reasoning Format) - Uses `writer.ReasoningPair`, `writer.ChatMessage`, `writer.TrainingPair` types
- **Requires**: `internal/templates` package for template embedding

## Future Considerations

- Add `--skip-errors` flag for lenient conversion mode
- Support batch conversion with glob patterns
- Add template versioning for format evolution
