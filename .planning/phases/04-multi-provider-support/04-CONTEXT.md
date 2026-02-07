# Phase 4: Multi-Provider Support - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Abstract the LLM provider layer so users can generate training data via Ollama (existing default) or OpenRouter (new), with secure API key management via environment variables. Includes cost estimation for paid providers and a list-providers discovery subcommand.

</domain>

<decisions>
## Implementation Decisions

### Provider Selection UX
- Default provider is Ollama (backward compatible, no breaking changes)
- `--provider` flag accepts: `ollama`, `openrouter` (no aliases)
- CLI flags only — no config file for provider/model preferences
- New `datakeg list-providers` subcommand shows available providers, whether each is configured (key set vs not), and available models per provider (query Ollama for local models, OpenRouter for available models)

### API Key Management
- Environment variables only — no config file, no keychain
- OpenRouter key: `OPENROUTER_API_KEY` (matches OpenRouter's own convention)
- Ollama continues using `OLLAMA_HOST` env var (existing behavior)
- Missing key → clear error with setup instructions: "OPENROUTER_API_KEY not set. Export it or visit openrouter.ai/keys to generate one."
- `list-providers` shows configured/not configured status per provider (no key values revealed)

### Model Selection
- Single `--model` flag works for both providers (Ollama model name or OpenRouter model ID)
- OpenRouter requires `--model` — no default model (models change fast, user picks explicitly)
- Ollama retains existing default model behavior
- Claude's Discretion: whether to validate model against API or accept free-form strings

### Error & Fallback Behavior
- Retry with exponential backoff on API failures (rate limits, server errors) — 3 retries, then fail
- No provider fallback — if the chosen provider fails mid-run, the run fails (per-document files from --skip-merge provide partial results)
- No upfront key validation — fail on first API call, not on startup
- Authentication errors include detailed fix steps: "Invalid API key. Check OPENROUTER_API_KEY or visit openrouter.ai/keys to regenerate."
- Error messages follow existing UX-06 pattern: "[What failed]\n[Actionable guidance]"

### Cost / Token Estimation
- Pre-run estimate shown for paid providers (approximate based on document sizes and pair counts)
- Confirmation prompt before paid runs: "Estimated cost: $X.XX. Actual cost may differ. You are responsible for all API charges. Continue? [y/N]"
- `--dry-run` flag prints estimate and exits (for scripting/CI)
- `--yes` / `-y` flag auto-confirms (skips confirmation prompt for CI/scripts)
- Post-run summary always shown for paid providers: actual tokens used and estimated cost (from API usage data)
- Ollama (local) runs skip cost estimation and confirmation entirely

### Claude's Discretion
- Provider interface design and abstraction pattern
- OpenRouter API client implementation details
- Token estimation algorithm
- Exact retry backoff timing
- Model validation approach (free-form vs API-validated)

</decisions>

<specifics>
## Specific Ideas

- Cost confirmation wording should be straightforward, not friendly — "You are responsible for all API charges"
- `list-providers` should query live data (Ollama local models, OpenRouter available models) when providers are configured
- Per-run cost summary uses actual token counts from API responses, not estimates

</specifics>

<deferred>
## Deferred Ideas

- Cost tracking across runs / spending limits — future phase or backlog
- Config file for persistent preferences — explicitly decided against for now, revisit if needed

</deferred>

---

*Phase: 04-multi-provider-support*
*Context gathered: 2026-02-07*
