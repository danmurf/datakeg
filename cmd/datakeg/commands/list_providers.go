package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/danmurf/datakeg/internal/provider"
)

// ExecuteListProviders shows available providers and their configuration status.
func ExecuteListProviders() error {
	fmt.Println("Available providers:")
	fmt.Println()

	// Check Ollama status
	_, ollamaErr := provider.NewOllamaProvider()
	if ollamaErr != nil {
		fmt.Println("  ollama: Not configured (OLLAMA_HOST not set or Ollama not running)")
	} else {
		fmt.Println("  ollama: Configured")
		// Try to list local models
		models, err := listOllamaModels()
		if err != nil {
			fmt.Printf("    Could not fetch models: %v\n", err)
		} else if len(models) > 0 {
			fmt.Printf("    Models: %s\n", models)
		}
	}

	// Check OpenRouter status
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openrouterKey == "" {
		fmt.Println("  openrouter: Not configured (OPENROUTER_API_KEY not set)")
	} else {
		fmt.Println("  openrouter: Configured")
		// Try to list OpenRouter models
		models, err := listOpenRouterModels()
		if err != nil {
			fmt.Printf("    Could not fetch models: %v\n", err)
		} else if len(models) > 0 {
			fmt.Printf("    Models (sample): %s\n", models)
			fmt.Println("    See https://openrouter.ai/models for full list")
		}
	}

	return nil
}

// listOllamaModels attempts to list available Ollama models.
func listOllamaModels() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try using Ollama's API to list models
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// listOpenRouterModels fetches available models from OpenRouter.
func listOpenRouterModels() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return first 10 models as a sample
	sampleSize := 10
	if len(result.Data) < sampleSize {
		sampleSize = len(result.Data)
	}

	models := make([]string, sampleSize)
	for i := 0; i < sampleSize; i++ {
		models[i] = result.Data[i].ID
	}

	return models, nil
}
