---
phase: 07-format-conversion
plan: "02"
type: execute
subsystem: cli
tags: ["cli", "conversion", "command", "cobra"]
completed: 2026-02-07
duration: 5 min
---

# Phase 7 Plan 2: CLI Convert Command Summary

## Overview

Wired the converter engine into the CLI with a `datakeg convert` command. Users can now convert JSONL files to model-specific formats using built-in or custom templates.

## Key Files Modified

| File | Changes |
|------|---------|
| `cmd/datakeg/main.go` | Added `convertCmd` with flags, `runConvert` handler |
| `cmd/datakeg/commands/convert.go` | New file: `ExecuteConvertPipeline`, `ListBuiltinTemplates` |

## New CLI Command

```bash
datakeg convert <input.jsonl> <output.jsonl>
```

### Flags

| Flag | Short | Purpose |
|------|-------|---------|
| `--template` | `-t` | Built-in template name (mistral-instruct, llama3-instruct, chatml, deepseek-r1) |
| `--custom-template` | | Path to custom template file |
| `--source-format` | | Source format (completion, chat, reasoning). Auto-detected if omitted |
| `--list-templates` | `-l` | List available built-in templates |

### Examples

```bash
# Convert with built-in template
datakeg convert --template mistral-instruct train.jsonl train_mistral.jsonl

# Convert with custom template
datakeg convert --custom-template my-format.tmpl train.jsonl train_custom.jsonl

# List available templates
datakeg convert --list-templates

# Specify source format explicitly
datakeg convert --template mistral-instruct --source-format completion train.jsonl out.jsonl
```

## Command Behavior

1. **Template Loading**: Either `--template` or `--custom-template` is required
2. **Format Detection**: Source format auto-detected from JSONL structure unless `--source-format` specified
3. **Validation**: Template validated against source format before conversion
4. **Conversion**: Streams JSONL line-by-line through template
5. **Output**: Reports line count and confirms completion

## List Templates Output

```
Built-in conversion templates:

  chatml               (chat      ) - ChatML format with <|im_start|> tags
  deepseek-r1         (reasoning ) - DeepSeek-R1 integrated reasoning format
  llama3-instruct     (chat      ) - Llama 3 Instruct with header tags
  mistral-instruct    (completion) - Mistral Instruct format with [INST] tags

Usage: datakeg convert --template <name> <input.jsonl> <output.jsonl>
Custom: datakeg convert --custom-template <file.tmpl> <input.jsonl> <output.jsonl>
```

## Functions Added

| Function | Purpose |
|----------|---------|
| `ExecuteConvertPipeline(input, output, template, customTemplate, sourceFormat)` | Orchestrates conversion pipeline |
| `ListBuiltinTemplates()` | Prints available templates with descriptions |

## Error Handling

- Missing template flag: Clear error with `--list-templates` suggestion
- Missing file args: Helpful usage message
- Template not found: Suggests `--list-templates`
- Format mismatch: Clear validation error
- Invalid input: Line number in error message

## Verification

- `make lint` passes ✓
- `make test` passes ✓
- `go build ./...` compiles ✓
- `datakeg convert --help` shows correct usage ✓
- `datakeg convert --list-templates` lists 4 templates ✓
- Conversion tested with all 4 templates ✓

## Conversion Test Results Mistral Instruct (completion →

### Mistral)
Input: `{"prompt":"What is 2+2?","completion":"The answer is 4."}`
Output: `{"text":"<s>[INST] What is 2+2? [/INST] The answer is 4.</s>"}`

### Llama 3 Instruct (chat → Llama 3)
Input: `{"messages":[{"role":"user","content":"Hello"}]}`
Output: `{"text":"<|begin_of_text|><|start_header_id|>user<|end_header_id|>\n\nHello<|eot_id|>"}`

### DeepSeek-R1 (reasoning → integrated)
Input: `{"question":"What is 2+2?","reasoning":"Math question","answer":"4"}`
Output: `{"prompt":"What is 2+2?","completion":"Math question\n\n4"}`

## Dependencies

- **Requires**: Plan 07-01 (converter package, conversion templates)
- **Uses**: `internal/converter` for format detection, validation, conversion
- **Uses**: `internal/templates` for template loading

## Future Enhancements

- `--skip-errors` flag for lenient conversion
- Batch conversion with glob patterns
- Progress indicator for large files
- `--dry-run` to preview first conversion
