---
phase: 04-multi-provider-support
plan: 01
wave: 1
subsystem: provider-abstraction
tags: [provider, ollama, abstraction, cli]
---

# Phase 4 Plan 1: Provider Interface and Ollama Wrapper - Summary

**Committed:** db33c70
**Duration:** ~5 minutes
**Tasks:** 2/2 completed

## Objective

Create the Provider interface abstraction, wrap the existing Ollama client as the first implementation, refactor the Generator to use the interface instead of a concrete `*ollama.Client`, and wire the `--provider` flag through the CLI pipeline.

## What Was Delivered

### Provider Package (`internal/provider/`)

| File | Purpose |
|------|---------|
| `provider.go` | Defines `Provider` interface with `Generate(ctx, model, prompt) (string, *UsageMetadata, error)` method. Also defines `ProviderType` constants (`ollama`, `openrouter`) and `UsageMetadata` struct for cost tracking. |
| `ollama.go` | `OllamaProvider` struct wrapping existing `ollama.Client`. Implements `Provider` interface, returns `nil` for `UsageMetadata` (local provider). |
| `factory.go` | `NewProvider(providerType)` factory function. Returns appropriate provider or error for unknown types. OpenRouter returns "not yet implemented" placeholder. |

### Generator Refactoring (`internal/generator/generator.go`)

- Changed `Generator.client *ollama.Client` to `Generator.provider provider.Provider`
- Updated `NewGenerator` signature: `NewGenerator(p provider.Provider, config Config)`
- Changed all `g.client.Generate()` calls to `g.provider.Generate()`
- Ignored usage return (`_`) for now - will be accumulated in Plan 04-02
- Removed `ollama` import, added `provider` import

### CLI Changes (`cmd/datakeg/`)

| File | Changes |
|------|---------|
| `main.go` | Added `--provider` persistent flag (default: "ollama"). Updated help text. Wired `flagProvider` through to `ExecuteGeneratePipeline`. |
| `commands/generate.go` | Added `providerType` parameter to `ExecuteGeneratePipeline`. Uses `provider.NewProvider()` instead of `ollama.NewClient()`. |

## Verification Results

- `make lint` passes
- `make test` passes (all existing tests green)
- `go build ./cmd/datakeg/...` succeeds
- `./datakeg generate --help` shows `--provider` flag with default "ollama"
- `./datakeg --help` shows provider flag at root level

## Key Files Modified/Created

| Status | Path | Description |
|--------|------|-------------|
| Created | `internal/provider/provider.go` | Provider interface and types |
| Created | `internal/provider/ollama.go` | OllamaProvider implementation |
| Created | `internal/provider/factory.go` | Provider factory |
| Modified | `internal/generator/generator.go` | Uses Provider interface |
| Modified | `cmd/datakeg/main.go` | Added --provider flag |
| Modified | `cmd/datakeg/commands/generate.go` | Uses provider factory |

## Decisions Made

- Provider interface minimal (1 method) to avoid over-abstraction
- UsageMetadata returns nil for Ollama (local/free provider)
- Provider type as string constants for CLI compatibility
- Error messages follow UX-06 pattern with actionable guidance

## Next Steps

Proceed to **Plan 04-02**: Implement OpenRouter provider with retry logic, cost estimation, and --yes/--dry-run flags.

---

*Summary created: 2026-02-07*
