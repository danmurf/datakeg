package provider

import (
	"context"
	"fmt"

	"github.com/danmurf/datakeg/internal/ollama"
)

// OllamaProvider wraps the existing Ollama client and implements the Provider interface.
type OllamaProvider struct {
	client *ollama.Client
}

// NewOllamaProvider creates a new Ollama provider by wrapping the existing ollama.Client.
func NewOllamaProvider() (*OllamaProvider, error) {
	client, err := ollama.NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create Ollama client: %w", err)
	}
	return &OllamaProvider{client: client}, nil
}

// Generate sends a prompt to the Ollama model and returns the response.
// Returns nil for UsageMetadata since Ollama is a local/free provider.
func (p *OllamaProvider) Generate(ctx context.Context, model, prompt string) (string, *UsageMetadata, error) {
	response, err := p.client.Generate(ctx, model, prompt)
	if err != nil {
		return "", nil, fmt.Errorf("ollama generation failed: %w\nmake sure Ollama is running (ollama serve) and the model is available (ollama pull %s)", err, model)
	}
	return response, nil, nil
}
