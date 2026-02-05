---
phase: 1
plan: 1
subsystem: cli
tags:
  - cobra
  - jsonl
  - cli
  - training-data
---

# Phase 1 Plan 1: CLI Scaffold with Cobra, JSONL Writer Summary

**One-liner:** CLI scaffold with Cobra framework and JSONL writer for training pair output

## Objective

Establish the CLI skeleton and JSONL output capability for the datakeg tool.

## Tasks Completed

| # | Name | Commit | Files |
|---|------|--------|-------|
| 1 | CLI scaffold with Cobra | 9372cf7 | `cmd/datakeg/main.go` |
| 2 | JSONL writer | c8b2dbe | `internal/writer/jsonl.go`, `internal/writer/jsonl_test.go` |

## Key Files Created

### CLI Scaffold (`cmd/datakeg/main.go`)

- Root command with description and usage info
- `generate` subcommand accepting `<source>` and `<output>` positional arguments
- 5 configuration flags:
  - `--model` / `-m`: Ollama model (default: gpt-oss:20b)
  - `--train-pct`: Training set percentage (default: 0.6)
  - `--valid-pct`: Validation set percentage (default: 0.2)
  - `--test-pct`: Test set percentage (default: 0.2)
  - `--pairs-per-1k`: Target pairs per 1000 characters (default: 1.0)
- `--timeout` / `-t`: Operation timeout in minutes (default: 30)

### JSONL Writer (`internal/writer/jsonl.go`)

- `TrainingPair` struct with `prompt` and `completion` JSON fields
- `WriteJSONL(filename string, pairs []TrainingPair)`: Creates new JSONL file
- `WriteJSONLAppend(filename string, pairs []TrainingPair)`: Appends to existing file
- Uses `json.Encoder` for proper JSONL formatting (one JSON object per line)

## Dependency Graph

| Relationship | Target |
|--------------|--------|
| requires | None (foundational) |
| provides | CLI framework, JSONL output capability |
| affects | All subsequent plans (01-02, 01-03, etc.) |

## Tech Stack

### Added Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| github.com/spf13/cobra | v1.10.2 | CLI framework |
| github.com/spf13/pflag | v1.0.9 | POSIX-compliant flag parsing |

### Established Patterns

- Cobra command structure with `RunE` for error handling
- json.Encoder for streaming JSONL output
- Deferred file.Close() with Sync() for data integrity

## Decisions Made

No architectural decisions required during this plan. Implementation followed established patterns from research.

## Deviations from Plan

**None** - Plan executed exactly as written.

## Authentication Gates

None - This plan established foundational CLI and I/O capabilities without external service dependencies.

## Verification

### CLI Binary

```bash
$ datakeg --help
datakeg transforms raw documentation into LLM training datasets.

Usage:
  datakeg [command]

Flags:
  -m, --model string   Ollama model to use (default "gpt-oss:20b")
```

### Generate Command

```bash
$ datakeg generate --help
Usage:
  datakeg generate <source> <output> [flags]

Flags:
      --pairs-per-1k float   Target pairs per 1000 characters (default 1)
      --test-pct float       Test set percentage (0.0-1.0) (default 0.2)
  -t, --timeout int          Operation timeout in minutes (default 30)
      --train-pct float      Training set percentage (0.0-1.0) (default 0.6)
      --valid-pct float       Validation set percentage (0.0-1.0) (default 0.2)

Global Flags:
  -m, --model string   Ollama model to use (default "gpt-oss:20b")
```

### JSONL Output Format

```jsonl
{"prompt":"What is Go?","completion":"Go is a programming language."}
{"prompt":"What is Cobra?","completion":"Cobra is a CLI framework."}
```

## Success Criteria Status

| Criterion | Status | Evidence |
|-----------|--------|----------|
| CLI binary builds and runs | ✅ PASS | `go build -o datakeg ./cmd/datakeg` succeeds |
| Generate command accepts source and output args | ✅ PASS | `Args: cobra.ExactArgs(2)` enforced |
| All 5 config flags present | ✅ PASS | model, train-pct, valid-pct, test-pct, pairs-per-1k all present |
| JSONL writer produces valid JSONL format | ✅ PASS | json.Encoder produces valid per-line JSON |

## Metrics

| Metric | Value |
|--------|-------|
| Duration | ~2 minutes |
| Tasks Completed | 2/2 |
| Files Created | 3 |
| Lines Added | ~224 |

## Next Steps

This plan establishes foundational infrastructure. Plan 01-02 (prompt templates) and plan 01-03 (Ollama integration) will build upon this CLI scaffold.

---
*SUMMARY generated: 2026-02-05*
*Plan: 01-01 of Phase 1 (Core Pipeline)*
