# Phase 4: Multi-Provider Support - Research

**Researched:** 2026-02-07
**Domain:** Multi-provider LLM API abstraction in Go
**Confidence:** HIGH

## Summary

This phase abstracts the LLM provider layer to support both Ollama (local) and OpenRouter (cloud API), enabling users to choose their preferred model provider. The standard approach in Go is to define small, focused interfaces in the consumer package (not the provider), use the Strategy pattern for interchangeable implementations, and avoid premature abstraction. OpenRouter provides a unified API for 500+ models with OpenAI-compatible endpoints, token usage tracking in responses, and built-in retry/fallback capabilities. Cost estimation requires either tiktoken-go for accurate token counting or a 4:1 character-to-token approximation. The cenkalti/backoff/v5 library is the de facto standard for exponential backoff with jitter. User confirmation for paid operations uses simple Y/N prompts with --yes flag support.

**Primary recommendation:** Define a small Provider interface in the generator package with a single Generate method, implement it for both Ollama and OpenRouter, use character-based estimation (1 token ~= 4 chars) for pre-run cost calculation, and implement retry logic with cenkalti/backoff/v5.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| net/http | stdlib | HTTP client for OpenRouter API | Standard library, production-ready timeouts and context support |
| encoding/json | stdlib | JSON marshaling for API requests/responses | Standard library, sufficient for API integration |
| cenkalti/backoff/v5 | v5.x | Exponential backoff with jitter | De facto standard, Google's algorithm port, used by HashiCorp |
| spf13/cobra | v1.10.2+ | CLI framework (already used) | Already in project, supports persistent flags |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| AlecAivazis/survey/v2 | v2.x | Interactive CLI prompts | User confirmation for paid runs (Y/N prompt) |
| pkoukk/tiktoken-go | latest | Token counting (optional) | Accurate token estimation (OpenAI models only) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| cenkalti/backoff | Custom retry logic | backoff provides battle-tested jitter, cap, and reset logic |
| survey/v2 | fmt.Scanln for Y/N input | survey provides better UX but fmt.Scanln is zero-dependency |
| tiktoken-go | Character-based approximation (4 chars = 1 token) | tiktoken is accurate but only works for OpenAI models; approximation works for all providers |

**Installation:**
```bash
go get github.com/cenkalti/backoff/v5
go get github.com/AlecAivazis/survey/v2
# tiktoken-go optional - only if choosing accurate token counting
go get github.com/pkoukk/tiktoken-go
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── provider/           # New package for provider abstraction
│   ├── provider.go     # Interface definition (consumer-side)
│   ├── ollama.go       # Ollama implementation (wraps existing client)
│   ├── openrouter.go   # OpenRouter implementation
│   └── factory.go      # Provider factory based on --provider flag
├── generator/          # Existing generator package
│   └── generator.go    # Updated to use Provider interface
├── ollama/             # Existing Ollama client (no changes needed)
│   └── client.go
└── cost/               # New package for cost estimation
    ├── estimator.go    # Pre-run cost estimation
    └── tracker.go      # Post-run actual cost tracking
```

### Pattern 1: Strategy Pattern with Consumer-Side Interface
**What:** Define interface where it's used (generator), not where it's implemented (provider package)
**When to use:** When you need interchangeable implementations (Ollama vs OpenRouter)
**Example:**
```go
// internal/provider/provider.go
package provider

import "context"

// Provider abstracts LLM generation across different providers.
// Interface is small (1 method) to avoid over-abstraction.
type Provider interface {
    // Generate sends prompt to LLM and returns response.
    // Returns usage metadata for cost tracking (nil for local providers).
    Generate(ctx context.Context, model, prompt string) (response string, usage *UsageMetadata, err error)
}

// UsageMetadata tracks token usage and cost (nil for free/local providers).
type UsageMetadata struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    EstimatedCost    float64 // In USD
}
```

### Pattern 2: Factory Pattern for Provider Selection
**What:** Create provider instance based on CLI flag
**When to use:** When provider selection happens at runtime based on user input
**Example:**
```go
// internal/provider/factory.go
package provider

import "fmt"

type ProviderType string

const (
    ProviderOllama     ProviderType = "ollama"
    ProviderOpenRouter ProviderType = "openrouter"
)

func NewProvider(providerType ProviderType) (Provider, error) {
    switch providerType {
    case ProviderOllama:
        return NewOllamaProvider()
    case ProviderOpenRouter:
        return NewOpenRouterProvider()
    default:
        return nil, fmt.Errorf("unknown provider: %s", providerType)
    }
}
```

### Pattern 3: OpenRouter API Client with Context and Timeout
**What:** HTTP client with per-request context and client-level timeout
**When to use:** All external API calls (OpenRouter)
**Example:**
```go
// internal/provider/openrouter.go
package provider

import (
    "context"
    "net/http"
    "time"
)

type OpenRouterProvider struct {
    client  *http.Client
    apiKey  string
    baseURL string
}

func NewOpenRouterProvider() (*OpenRouterProvider, error) {
    apiKey := os.Getenv("OPENROUTER_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("OPENROUTER_API_KEY not set. Export it or visit openrouter.ai/keys to generate one")
    }

    return &OpenRouterProvider{
        client: &http.Client{
            Timeout: 5 * time.Minute, // Client-level timeout
        },
        apiKey:  apiKey,
        baseURL: "https://openrouter.ai/api/v1",
    }, nil
}

func (p *OpenRouterProvider) Generate(ctx context.Context, model, prompt string) (string, *UsageMetadata, error) {
    // Request-specific context timeout (from caller) overrides client timeout
    req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", body)
    if err != nil {
        return "", nil, err
    }

    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("Content-Type", "application/json")

    // Implementation continues...
}
```

### Pattern 4: Exponential Backoff with Jitter
**What:** Retry failed API requests with exponential backoff and jitter
**When to use:** Rate limits (429), server errors (5xx), network errors
**Example:**
```go
// Using cenkalti/backoff/v5
import "github.com/cenkalti/backoff/v5"

func (p *OpenRouterProvider) generateWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error

    operation := func() error {
        resp, err = p.client.Do(req)
        if err != nil {
            return err // Retry on network errors
        }

        // Retry on rate limit or server errors
        if resp.StatusCode == 429 || resp.StatusCode >= 500 {
            resp.Body.Close()
            return fmt.Errorf("retryable error: %d", resp.StatusCode)
        }

        return nil // Success
    }

    // Exponential backoff: 1s, 2s, 4s (max 3 retries)
    exponentialBackoff := backoff.NewExponentialBackOff()
    exponentialBackoff.MaxElapsedTime = 30 * time.Second

    err = backoff.Retry(operation, backoff.WithContext(exponentialBackoff, ctx))
    return resp, err
}
```

### Pattern 5: List Providers Command
**What:** New subcommand to show available providers and their configuration status
**When to use:** User wants to discover which providers are available and configured
**Example:**
```go
// cmd/datakeg/commands/list_providers.go
package commands

func listProvidersCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "list-providers",
        Short: "List available LLM providers and their configuration status",
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Println("Available providers:")
            fmt.Println()

            // Check Ollama
            _, ollamaErr := provider.NewOllamaProvider()
            status := "✓ Configured"
            if ollamaErr != nil {
                status = "✗ Not configured (OLLAMA_HOST not set or Ollama not running)"
            }
            fmt.Printf("  ollama: %s\n", status)

            // Check OpenRouter
            _, openrouterErr := provider.NewOpenRouterProvider()
            status = "✓ Configured"
            if openrouterErr != nil {
                status = "✗ Not configured (OPENROUTER_API_KEY not set)"
            }
            fmt.Printf("  openrouter: %s\n", status)

            return nil
        },
    }
}
```

### Pattern 6: Cost Confirmation Prompt
**What:** Show estimated cost and require Y/N confirmation before paid API runs
**When to use:** When provider is paid (OpenRouter) and --yes flag not set
**Example:**
```go
// Using AlecAivazis/survey/v2
import "github.com/AlecAivazis/survey/v2"

func confirmPaidRun(estimatedCost float64, skipConfirm bool) error {
    if skipConfirm {
        return nil // --yes flag set, skip prompt
    }

    fmt.Printf("\nEstimated cost: $%.4f\n", estimatedCost)
    fmt.Println("Actual cost may differ. You are responsible for all API charges.")
    fmt.Println()

    confirm := false
    prompt := &survey.Confirm{
        Message: "Continue?",
        Default: false,
    }

    if err := survey.AskOne(prompt, &confirm); err != nil {
        return err
    }

    if !confirm {
        return fmt.Errorf("operation cancelled by user")
    }

    return nil
}
```

### Anti-Patterns to Avoid
- **Oversized interfaces:** Don't add methods like ListModels, ValidateModel, GetPricing to Provider interface - keep it focused on Generate only
- **Defining interface in provider package:** Put Provider interface in generator package (consumer), not in provider package (implementer)
- **Premature abstraction:** Don't create interfaces for every struct - only abstract when you have multiple implementations
- **Using interfaces for testing only:** Don't create Provider interface just for mocking - create it because you genuinely have multiple implementations (Ollama and OpenRouter)
- **Default HTTP client in production:** Never use http.DefaultClient - always set timeouts and context

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Exponential backoff | Custom sleep doubling loop | cenkalti/backoff/v5 | Handles jitter (prevents thundering herd), max elapsed time, context cancellation, and clock reset properly |
| Token counting | Character counting with fixed ratio | pkoukk/tiktoken-go (optional) or 4:1 approximation | Accurate tokenization requires BPE algorithm; custom counting will be wrong for non-English, code, special chars |
| OpenRouter streaming | Manual SSE parsing | Standard response.Body read with bufio.Scanner or tmaxmax/go-sse | SSE has edge cases (data: prefix, [DONE] marker, error events mid-stream) |
| Y/N confirmation | Custom fmt.Scanln loops | AlecAivazis/survey/v2 Confirm | Handles invalid input, default values, Ctrl+C gracefully |
| API error handling | Custom HTTP status switch | Structured error types with metadata | Need to preserve upstream errors, rate limit info, retry-ability signal |

**Key insight:** Retry logic is deceptively complex - you need jitter to avoid thundering herd, max elapsed time to prevent infinite loops, context cancellation for timeouts, and proper clock management. cenkalti/backoff has 1000+ stars and is used by HashiCorp because it handles all edge cases correctly.

## Common Pitfalls

### Pitfall 1: Using Default HTTP Client Timeout
**What goes wrong:** http.DefaultClient has no timeout (0 = infinite), causing hangs on network issues or slow APIs
**Why it happens:** http.DefaultClient is convenient, documentation doesn't emphasize the risk
**How to avoid:** Always create http.Client with explicit Timeout field (3-5 minutes for LLM APIs)
**Warning signs:** Application hangs indefinitely when OpenRouter is slow or unreachable

### Pitfall 2: Token Estimation Without Tiktoken Only Works for English
**What goes wrong:** 4:1 character-to-token ratio is inaccurate for non-English text, code, or special characters (can be off by 50%+)
**Why it happens:** Approximation is based on English prose statistics; BPE tokenization behaves differently for other content
**How to avoid:** Document that pre-run estimates are approximations; use actual token counts from API response for post-run summaries
**Warning signs:** Estimated cost is $5, actual cost is $10 (user upset about inaccuracy)

### Pitfall 3: Not Checking API Key on First API Call
**What goes wrong:** User decision: "No upfront key validation - fail on first API call, not on startup"
**Why it happens:** This is intentional to avoid latency on startup
**How to avoid:** Provide clear authentication error messages: "Invalid API key. Check OPENROUTER_API_KEY or visit openrouter.ai/keys to regenerate"
**Warning signs:** Cryptic HTTP 401 errors without actionable guidance

### Pitfall 4: Retrying Non-Retryable Errors
**What goes wrong:** Retrying 400 (bad request) or 401 (invalid key) wastes time and may trigger rate limits
**Why it happens:** Blanket retry logic doesn't distinguish permanent from transient errors
**How to avoid:** Only retry 429 (rate limit), 5xx (server errors), and network errors; fail immediately on 4xx (except 429)
**Warning signs:** CLI hangs for 30 seconds on invalid API key before failing

### Pitfall 5: Forgetting Jitter in Exponential Backoff
**What goes wrong:** All clients retry simultaneously on provider outage, creating thundering herd that overloads provider when it recovers
**Why it happens:** Naive implementation doubles delay without randomization
**How to avoid:** Use cenkalti/backoff which includes jitter by default
**Warning signs:** Provider comes back online, immediately gets overloaded by synchronized retries

### Pitfall 6: Mid-Stream Errors in SSE Responses
**What goes wrong:** OpenRouter returns HTTP 200 with headers, then sends error event mid-stream; naive code treats 200 as success and panics on malformed JSON
**Why it happens:** SSE headers are sent before content generation starts; errors discovered during generation
**How to avoid:** Parse SSE events; check for finish_reason: "error" and error object in stream chunks
**Warning signs:** Panic: "invalid character 'e' looking for beginning of value" when parsing streaming response

### Pitfall 7: Not Respecting Context Cancellation in Retry Loops
**What goes wrong:** User hits Ctrl+C, but retry loop continues for 30 seconds until max elapsed time
**Why it happens:** Retry logic doesn't check ctx.Done() between attempts
**How to avoid:** Use backoff.WithContext(backoff, ctx) to wire context into retry logic
**Warning signs:** Ctrl+C doesn't stop the command; user has to kill the process

### Pitfall 8: Defining Provider Interface Too Early
**What goes wrong:** Interface has wrong methods (e.g., GetConfig, ListModels, Close) that don't apply to all providers
**Why it happens:** Premature abstraction before understanding what both providers need
**How to avoid:** Implement both Ollama and OpenRouter clients first, then extract minimal interface from common methods
**Warning signs:** OpenRouter provider has no-op methods to satisfy oversized interface

### Pitfall 9: Cost Calculation from Tokens Without Pricing Data
**What goes wrong:** Can't calculate cost from token count without knowing model's per-token pricing
**Why it happens:** OpenRouter's pricing varies by model; need to fetch pricing from /models endpoint or use response metadata
**How to avoid:** Use cost field from OpenRouter response (already calculated upstream); only estimate pre-run cost for confirmation
**Warning signs:** Post-run cost summary shows $0.00 despite using paid model

## Code Examples

Verified patterns from official sources and Go standard practices:

### OpenRouter API Request
```go
// Based on: https://openrouter.ai/docs/api/reference/overview
package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

type OpenRouterRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OpenRouterResponse struct {
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

type Choice struct {
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

type Usage struct {
    PromptTokens     int     `json:"prompt_tokens"`
    CompletionTokens int     `json:"completion_tokens"`
    TotalTokens      int     `json:"total_tokens"`
    Cost             float64 `json:"cost,omitempty"` // OpenRouter-specific
}

func (p *OpenRouterProvider) Generate(ctx context.Context, model, prompt string) (string, *UsageMetadata, error) {
    reqBody := OpenRouterRequest{
        Model: model,
        Messages: []Message{
            {Role: "user", Content: prompt},
        },
    }

    body, err := json.Marshal(reqBody)
    if err != nil {
        return "", nil, err
    }

    req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return "", nil, err
    }

    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := p.client.Do(req)
    if err != nil {
        return "", nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var errResp struct {
            Error struct {
                Code    int    `json:"code"`
                Message string `json:"message"`
            } `json:"error"`
        }
        json.NewDecoder(resp.Body).Decode(&errResp)
        return "", nil, fmt.Errorf("API error %d: %s", errResp.Error.Code, errResp.Error.Message)
    }

    var openrouterResp OpenRouterResponse
    if err := json.NewDecoder(resp.Body).Decode(&openrouterResp); err != nil {
        return "", nil, err
    }

    if len(openrouterResp.Choices) == 0 {
        return "", nil, fmt.Errorf("no response from model")
    }

    usage := &UsageMetadata{
        PromptTokens:     openrouterResp.Usage.PromptTokens,
        CompletionTokens: openrouterResp.Usage.CompletionTokens,
        TotalTokens:      openrouterResp.Usage.TotalTokens,
        EstimatedCost:    openrouterResp.Usage.Cost,
    }

    return openrouterResp.Choices[0].Message.Content, usage, nil
}
```

### List OpenRouter Models
```go
// Based on: https://openrouter.ai/docs/api/api-reference/models/get-models
package provider

func (p *OpenRouterProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Data []struct {
            ID          string `json:"id"`
            Name        string `json:"name"`
            Description string `json:"description"`
            Pricing     struct {
                Prompt     string `json:"prompt"`
                Completion string `json:"completion"`
            } `json:"pricing"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    models := make([]ModelInfo, len(response.Data))
    for i, m := range response.Data {
        models[i] = ModelInfo{
            ID:          m.ID,
            Name:        m.Name,
            Description: m.Description,
        }
    }

    return models, nil
}
```

### Character-Based Token Estimation
```go
// Based on: https://help.openai.com/en/articles/4936856-what-are-tokens-and-how-to-count-them
package cost

// EstimateTokens approximates token count using 4:1 character-to-token ratio.
// This is an approximation suitable for English text. Actual counts may vary by 20-50%
// for non-English, code, or special characters.
func EstimateTokens(text string) int {
    return int(math.Ceil(float64(len(text)) / 4.0))
}

// EstimateCost calculates estimated API cost based on character count and pricing.
// Returns cost in USD.
func EstimateCost(promptChars, expectedCompletionChars int, promptPricePerMToken, completionPricePerMToken float64) float64 {
    promptTokens := EstimateTokens(string(make([]byte, promptChars)))
    completionTokens := EstimateTokens(string(make([]byte, expectedCompletionChars)))

    promptCost := (float64(promptTokens) / 1_000_000) * promptPricePerMToken
    completionCost := (float64(completionTokens) / 1_000_000) * completionPricePerMToken

    return promptCost + completionCost
}
```

### Cobra --yes Flag Pattern
```go
// Based on: https://cobra.dev/docs/how-to-guides/working-with-flags/
package commands

func generateCommand() *cobra.Command {
    var skipConfirm bool

    cmd := &cobra.Command{
        Use:   "generate",
        Short: "Generate training data",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Get provider from --provider flag
            providerType := cmd.Flag("provider").Value.String()

            // Show cost estimate and confirm if paid provider
            if providerType == "openrouter" && !skipConfirm {
                estimatedCost := calculateEstimatedCost(documents)
                if err := confirmPaidRun(estimatedCost); err != nil {
                    return err
                }
            }

            // Continue with generation...
            return nil
        },
    }

    cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompts (for CI/scripts)")

    return cmd
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hard-coded Ollama client | Provider interface with factory | Phase 4 (2026) | Enables multiple LLM providers |
| No cost visibility | Pre-run estimate + post-run actual cost | Phase 4 (2026) | Users know costs before committing |
| Immediate failure on API errors | Exponential backoff with 3 retries | Phase 4 (2026) | Resilient to transient failures |
| No model discovery | list-providers command | Phase 4 (2026) | Users can explore available options |

**Deprecated/outdated:**
- Direct ollama.Client usage in generator: Will be wrapped in provider.OllamaProvider for interface compliance
- Hardcoded "gpt-oss:20b" default: Remains default for Ollama, but OpenRouter requires explicit --model flag

## Open Questions

Things that couldn't be fully resolved:

1. **OpenRouter Retry-After header support**
   - What we know: OpenRouter returns X-RateLimit headers on 429 errors
   - What's unclear: Whether OpenRouter includes standard Retry-After header
   - Recommendation: Use exponential backoff (cenkalti/backoff) regardless; if Retry-After is present, respect it

2. **Model validation strategy**
   - What we know: User decided this is "Claude's Discretion" - can validate against API or accept free-form
   - What's unclear: User preference for strict vs permissive validation
   - Recommendation: Start permissive (no validation) - fail on first API call with clear error. Add validation later if users request it.

3. **Token estimation accuracy requirements**
   - What we know: 4:1 approximation works for English; tiktoken-go is accurate but OpenAI-only
   - What's unclear: Whether accuracy is critical enough to warrant tiktoken-go dependency
   - Recommendation: Start with 4:1 approximation, document it's approximate. Add tiktoken-go if users report inaccurate estimates.

4. **Dry-run output format**
   - What we know: --dry-run prints estimate and exits
   - What's unclear: Should it print JSON for scripting or human-readable text?
   - Recommendation: Human-readable by default (matches existing CLI UX); add --format=json if needed later

5. **OpenRouter streaming vs non-streaming**
   - What we know: Ollama uses streaming for real-time feedback; OpenRouter supports both
   - What's unclear: Whether to implement streaming for OpenRouter or use simpler non-streaming
   - Recommendation: Start non-streaming (simpler, no SSE parsing). Add streaming if users want progress for long completions.

## Sources

### Primary (HIGH confidence)
- [OpenRouter API Reference](https://openrouter.ai/docs/api/reference/overview) - Endpoints, authentication, request/response schema
- [OpenRouter Quickstart Guide](https://openrouter.ai/docs/quickstart) - Basic integration patterns
- [OpenRouter Usage Accounting](https://openrouter.ai/docs/guides/guides/usage-accounting) - Token usage and cost tracking
- [OpenRouter Models API](https://openrouter.ai/docs/api/api-reference/models/get-models) - List models endpoint
- [OpenRouter Error Handling](https://openrouter.ai/docs/api/reference/errors-and-debugging) - Error codes and mid-stream errors
- [cenkalti/backoff](https://github.com/cenkalti/backoff) - Exponential backoff library (Google's algorithm)
- [pkg.go.dev/net/http](https://pkg.go.dev/net/http) - Go standard library HTTP client
- [Cobra Flags Documentation](https://cobra.dev/docs/how-to-guides/working-with-flags/) - CLI flag patterns

### Secondary (MEDIUM confidence)
- [Cloudflare: The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) - HTTP client timeout best practices
- [OneUptime: How to Implement Retry Logic in Go](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view) - Retry patterns
- [AlecAivazis/survey](https://github.com/AlecAivazis/survey) - CLI prompt library
- [7 Common Interface Mistakes in Go](https://medium.com/@andreiboar/7-common-interface-mistakes-in-go-1d3f8e58be60) - Interface design anti-patterns
- [Interface pollution (#5) - 100 Go Mistakes](https://100go.co/5-interface-pollution/) - Over-abstraction pitfalls
- [OpenRouter Error Guide](https://help.janitorai.com/en/article/openrouter-error-guide-10ear52/) - Common integration issues
- [OpenAI: What are tokens and how to count them](https://help.openai.com/en/articles/4936856-what-are-tokens-and-how-to-count-them) - Token counting approximation (4:1 ratio)

### Tertiary (LOW confidence)
- [Design Patterns in Go](https://refactoring.guru/design-patterns/go) - General patterns (not multi-provider specific)
- [pkoukk/tiktoken-go](https://github.com/pkoukk/tiktoken-go) - Token counting library (optional, unverified for this use case)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - cenkalti/backoff is widely used, net/http is stdlib, OpenRouter API is well-documented
- Architecture: HIGH - Provider interface pattern is standard Go practice, verified with multiple sources
- Pitfalls: MEDIUM - Based on general Go HTTP client knowledge and OpenRouter docs; some edge cases may exist
- Cost estimation: MEDIUM - 4:1 ratio is documented approximation; accuracy varies by content type

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (30 days - OpenRouter API is stable, Go patterns are mature)
