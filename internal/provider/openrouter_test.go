package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenRouterProvider_Generate_Success(t *testing.T) {
	// Create a mock OpenRouter server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// baseURL is "http://server/api/v1" and we append "/chat/completions"
		expectedPath := "/api/v1/chat/completions"
		if r.URL.Path != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		// Return success response
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "user",
						"content": "Test response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
				"cost":              0.00025,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL + "/api/v1",
	}

	// Generate
	response, usage, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if response != "Test response" {
		t.Errorf("expected 'Test response', got: %s", response)
	}
	if usage.PromptTokens != 100 {
		t.Errorf("expected 100 prompt tokens, got: %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("expected 50 completion tokens, got: %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got: %d", usage.TotalTokens)
	}
	if usage.EstimatedCost != 0.00025 {
		t.Errorf("expected 0.00025 cost, got: %f", usage.EstimatedCost)
	}
}

func TestOpenRouterProvider_Generate_AuthError(t *testing.T) {
	// Create a mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid API key"}}`))
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "invalid-key",
		baseURL: server.URL,
	}

	_, _, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if err.Error() != "openrouter authentication failed.\ncheck OPENROUTER_API_KEY or visit openrouter.ai/keys to regenerate" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenRouterProvider_Generate_RateLimit(t *testing.T) {
	attemptCount := 0

	// Create a mock server that returns 429 first, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "user",
						"content": "Test response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
				"cost":              0.00025,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	response, _, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err != nil {
		t.Fatalf("expected no error after retry, got: %v", err)
	}
	if response != "Test response" {
		t.Errorf("expected 'Test response', got: %s", response)
	}
	if attemptCount != 2 {
		t.Errorf("expected 2 attempts, got: %d", attemptCount)
	}
}

func TestOpenRouterProvider_Generate_BadRequest(t *testing.T) {
	// Create a mock server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid request"}}`))
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	_, _, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err == nil {
		t.Fatal("expected error for 400")
	}
	// Bad request should fail immediately (no retry)
}

func TestOpenRouterProvider_Generate_EmptyChoices(t *testing.T) {
	// Create a mock server that returns 200 with empty choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 0,
				"total_tokens":      100,
				"cost":              0.0001,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	_, _, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if err.Error() != "no response from model" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenRouterProvider_Generate_ContextCancelled(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := p.Generate(ctx, "test-model", "test prompt")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	// Should return quickly due to context cancellation
}

func TestOpenRouterProvider_Generate_ServerError(t *testing.T) {
	attemptCount := 0

	// Create a mock server that returns 500 first, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "user",
						"content": "Test response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
				"cost":              0.00025,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenRouterProvider{
		client:  &http.Client{Timeout: 5 * time.Minute},
		apiKey:  "test-key",
		baseURL: server.URL,
	}

	_, _, err := p.Generate(context.Background(), "test-model", "test prompt")
	if err != nil {
		t.Fatalf("expected no error after retry, got: %v", err)
	}
	if attemptCount != 2 {
		t.Errorf("expected 2 attempts, got: %d", attemptCount)
	}
}
