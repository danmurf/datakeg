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

// ProviderType represents the type of LLM provider.
type ProviderType string

const (
	// ProviderOllama represents the local Ollama provider.
	ProviderOllama ProviderType = "ollama"
	// ProviderOpenRouter represents the OpenRouter cloud API provider.
	ProviderOpenRouter ProviderType = "openrouter"
)
