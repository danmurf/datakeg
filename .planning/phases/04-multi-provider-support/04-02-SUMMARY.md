---
phase: 04-multi-provider-support
plan: 02
wave: 2
subsystem: openrouter-provider
tags: [openrouter, cost-estimation, cli-flags, provider]
---

# Phase 4 Plan 2: OpenRouter Provider Implementation - Summary

**Committed:** 6bfe158
**Duration:** ~8 minutes
**Tasks:** 2/2 completed

## Objective

Implement the OpenRouter provider with exponential backoff retry logic, add cost estimation and confirmation flow for paid providers, and wire --yes, --dry-run flags into the CLI pipeline.

## What Was Delivered

### OpenRouter Provider (`internal/provider/openrouter.go`)

| Feature | Implementation |
|---------|----------------|
| HTTP Client | 5-minute timeout, Bearer auth, HTTP-Referer header |
| Request/Response | OpenAI-compatible `/chat/completions` endpoint |
| Usage Metadata | Token counts and cost from API response |
| Retry Logic | `cenkalti/backoff/v5` with exponential backoff |
| Retry Config | 30s max, 1s/2s/4s intervals, respects context cancellation |
| Retry-able Codes | 429 (rate limit), 5xx (server errors) |
| Non-Retryable | 401 (auth), 400 (bad request), other 4xx |

### Cost Estimation (`internal/cost/estimator.go`)

| Function | Purpose |
|----------|---------|
| `EstimateTokens(text)` | 4:1 character-to-token approximation |
| `EstimateRunCost(documents, pairsPer1K, promptPrice, completionPrice)` | Pre-run cost estimate |
| `Tracker` struct | Accumulates usage across run, nil-safe for Ollama |
| `Tracker.Add(usage)` | Adds usage from each provider call |
| `Tracker.Summary()` | Formatted: "Tokens used: X + Y = Z. Cost: $X.XXXX" |

### CLI Changes

| Flag | Purpose |
|------|---------|
| `--provider` | Select LLM provider (ollama, openrouter), default: ollama |
| `--yes, -y` | Skip confirmation prompts (for CI/scripts) |
| `--dry-run` | Print cost estimate and exit without generating |

### Generator Changes (`internal/generator/generator.go`)

- `Generate()` signature: `([]Pair, *UsageMetadata, error)` → `([]Pair, *provider.UsageMetadata, error)`
- Usage accumulated from initial call + all backfill attempts
- Usage returned alongside pairs for CLI tracking

## Verification Results

- `make lint` passes
- `make test` passes (all existing tests green)
- `go build ./cmd/datakeg/...` succeeds
- `./datakeg generate --help` shows --provider, --yes, --dry-run flags
- `./datakeg generate --provider openrouter --model "" <source> <output>` returns: "openrouter requires an explicit model selection"
- `./datakeg generate --provider openrouter <source> <output>` (without --model) fails with guidance

## Key Files Modified/Created

| Status | Path | Description |
|--------|------|-------------|
| Created | `internal/provider/openrouter.go` | Full OpenRouter provider implementation |
| Created | `internal/cost/estimator.go` | Token estimation and cost tracking |
| Modified | `internal/provider/factory.go` | OpenRouter provider now created |
| Modified | `internal/generator/generator.go` | Returns UsageMetadata from Generate() |
| Modified | `cmd/datakeg/main.go` | Added --yes, --dry-run flags |
| Modified | `cmd/datakeg/commands/generate.go` | Cost flow, confirmation, post-run summary |

## User Workflow

### Paid Provider (OpenRouter) Flow

```
$ datakeg generate --provider openrouter --model meta-llama/llama-3.1-70b-instruct ./docs ./output
Starting pipeline...
Loading documents from ./docs...
Loaded 5 documents
Estimated cost: $0.0234
Actual cost may differ. You are responsible for all API charges.

Continue? [y/N] y
Creating openrouter provider...
...
Pipeline complete!
Total pairs: train=150, valid=25, test=25

Tokens used: 4500 prompt + 1200 completion = 5700 total. Estimated cost: $0.0234
```

### Dry Run (Cost Estimate Only)

```
$ datakeg generate --provider openrouter --model meta-llama/llama-3.1-70b-instruct --dry-run ./docs ./output
...
Estimated cost: $0.0234
Dry run - exiting without generation.
```

### CI/Script Mode (Auto-Confirm)

```
$ datakeg generate --provider openrouter --model meta-llama/llama-3.1-70b-instruct --yes ./docs ./output
...
Estimated cost: $0.0234
...
```

## Decisions Made

- Retry strategy: Exponential backoff with jitter (prevents thundering herd)
- Cost estimate: 4:1 char-to-token approximation (documented as approximate)
- Confirmation: Y/N prompt with default "N" (user must explicitly confirm)
- Post-run summary: Uses actual token counts from API, not estimates

## Dependencies Added

- `github.com/cenkalti/backoff/v5 v5.0.3` - Exponential backoff with context support

## Next Steps

Proceed to **Plan 04-03**: Implement `list-providers` subcommand and write tests for provider and cost packages.

---

*Summary created: 2026-02-07*
