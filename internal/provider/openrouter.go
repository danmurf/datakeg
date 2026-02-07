package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// OpenRouterProvider implements the Provider interface for OpenRouter's cloud API.
type OpenRouterProvider struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewOpenRouterProvider creates a new OpenRouter provider.
func NewOpenRouterProvider() (*OpenRouterProvider, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not set. Export it or visit openrouter.ai/keys to generate one")
	}

	return &OpenRouterProvider{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
	}, nil
}

// openRouterRequest represents the request body for OpenRouter chat completions.
type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

// openRouterMessage represents a message in the chat completion request.
type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse represents the response from OpenRouter chat completions.
type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
	Usage   openRouterUsage    `json:"usage"`
}

// openRouterChoice represents a choice in the chat completion response.
type openRouterChoice struct {
	Message      openRouterMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

// openRouterUsage represents token usage in the response.
type openRouterUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost,omitempty"`
}

// Generate sends a prompt to OpenRouter and returns the response with usage metadata.
func (p *OpenRouterProvider) Generate(ctx context.Context, model, prompt string) (string, *UsageMetadata, error) {
	reqBody := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := p.doRequestWithRetry(ctx, body)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	var openrouterResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openrouterResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
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

// doRequestWithRetry performs the HTTP request with exponential backoff retry.
func (p *OpenRouterProvider) doRequestWithRetry(ctx context.Context, body []byte) (*http.Response, error) {
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.InitialInterval = 1 * time.Second
	exponentialBackoff.Multiplier = 2.0
	exponentialBackoff.MaxInterval = 10 * time.Second

	var resp *http.Response
	var err error

	operation := func() (*http.Response, error) {
		req, reqErr := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
		if reqErr != nil {
			return nil, reqErr
		}

		req.Header.Set("Authorization", "Bearer "+p.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HTTP-Referer", "https://github.com/danmurf/datakeg")

		resp, err = p.client.Do(req)
		if err != nil {
			return nil, err
		}

		// Retry on rate limit or server errors
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("retryable error: %d", resp.StatusCode)
		}

		return resp, nil
	}

	resp, err = backoff.Retry(ctx, operation,
		backoff.WithMaxElapsedTime(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// Check for non-retryable errors
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp)
	}

	return resp, nil
}

// handleErrorResponse handles non-200 HTTP responses.
func (p *OpenRouterProvider) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("openrouter authentication failed.\ncheck OPENROUTER_API_KEY or visit openrouter.ai/keys to regenerate")
	case http.StatusBadRequest:
		return fmt.Errorf("openrouter bad request: %s", string(body))
	default:
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return fmt.Errorf("openrouter error %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("openrouter error: %d %s", resp.StatusCode, string(body))
	}
}
