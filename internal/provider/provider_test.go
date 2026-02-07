package provider

import (
	"os"
	"testing"
)

func TestNewProvider_Ollama(t *testing.T) {
	// This test requires Ollama to be running
	// Skip if OLLAMA_HOST is not set
	if os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("OLLAMA_HOST not set, skipping live test")
	}

	p, err := NewProvider(ProviderOllama)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestNewProvider_OpenRouter(t *testing.T) {
	// Test with API key set
	t.Setenv("OPENROUTER_API_KEY", "test-key")
	p, err := NewProvider(ProviderOpenRouter)
	if err != nil {
		t.Fatalf("expected no error with test key, got: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}

	// Test without API key
	t.Setenv("OPENROUTER_API_KEY", "")
	_, err = NewProvider(ProviderOpenRouter)
	if err == nil {
		t.Fatal("expected error when API key not set")
	}
	if err.Error() != "OPENROUTER_API_KEY not set. Export it or visit openrouter.ai/keys to generate one" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	_, err := NewProvider("invalid")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if err.Error() != "unknown provider: invalid. Available providers: ollama, openrouter" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestProviderType_Constants(t *testing.T) {
	if ProviderOllama != "ollama" {
		t.Errorf("expected ProviderOllama to be 'ollama', got: %s", ProviderOllama)
	}
	if ProviderOpenRouter != "openrouter" {
		t.Errorf("expected ProviderOpenRouter to be 'openrouter', got: %s", ProviderOpenRouter)
	}
}
