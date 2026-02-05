// Package ollama provides a wrapper around the official Ollama Go client
// for generating text from prompts using local Ollama models.
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
)

// Client wraps the official Ollama API client with simplified generation methods.
type Client struct {
	*api.Client
}

// NewClient creates a new Ollama client from the environment.
// It reads the OLLAMA_HOST environment variable to determine the endpoint.
// Returns an error if the client cannot be created (e.g., OLLAMA_HOST not set).
func NewClient() (*Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("create ollama client: %w", err)
	}
	return &Client{Client: client}, nil
}

// Generate sends a prompt to the specified Ollama model and returns the generated response.
// It uses streaming to accumulate the response incrementally, which provides real-time
// feedback and reduces memory usage for long responses.
//
// The context parameter controls cancellation - passing a cancelled context
// will stop the generation as soon as possible.
func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	formatJSON, _ := json.Marshal("json")

	req := &api.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: new(bool),
		Format: formatJSON,
	}
	*req.Stream = true

	var response strings.Builder

	err := c.Client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response.WriteString(resp.Response)

		// Check for context cancellation during streaming
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	if err != nil {
		return "", fmt.Errorf("generate response: %w", err)
	}

	return response.String(), nil
}
