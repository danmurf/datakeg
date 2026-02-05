package ollama

import (
	"context"
	"os"
	"testing"
)

// TestNewClient tests client creation with various Ollama host configurations.
func TestNewClient(t *testing.T) {
	// Skip if Ollama is not configured
	if os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("OLLAMA_HOST not set, skipping live test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

// TestGenerate tests the Generate method with a live Ollama instance.
func TestGenerate(t *testing.T) {
	// Skip if Ollama is not configured
	if os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("OLLAMA_HOST not set, skipping live test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	response, err := client.Generate(ctx, "tinyllama", "Say hello in exactly one word.")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if response == "" {
		t.Error("Generate() returned empty response")
	}
}

// TestGenerateCancellation tests that context cancellation works properly.
func TestGenerateCancellation(t *testing.T) {
	// Skip if Ollama is not configured
	if os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("OLLAMA_HOST not set, skipping live test")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.Generate(ctx, "tinyllama", "Generate a very long response...")
	// Should either return immediately or fail with context cancelled error
	if err != nil && err != context.Canceled {
		t.Errorf("Generate() unexpected error = %v", err)
	}
}
