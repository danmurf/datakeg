package provider

import "fmt"

// NewProvider creates a provider instance based on the specified type.
func NewProvider(providerType ProviderType) (Provider, error) {
	switch providerType {
	case ProviderOllama:
		return NewOllamaProvider()
	case ProviderOpenRouter:
		return nil, fmt.Errorf("openrouter provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider: %s. Available providers: ollama, openrouter", providerType)
	}
}
