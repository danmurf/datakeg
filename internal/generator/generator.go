package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/danmurf/datakeg/internal/ollama"
	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/templates"
)

// SplitType represents the type of dataset split.
type SplitType string

const (
	SplitTrain SplitType = "train"
	SplitValid SplitType = "valid"
	SplitTest  SplitType = "test"
)

// Config holds configuration for pair generation.
type Config struct {
	PairsPer1KChars float64 // Number of pairs to generate per 1000 characters
	ValidPercent    float64 // Percentage of pairs for validation
	TestPercent     float64 // Percentage of pairs for testing
	Model           string  // Ollama model to use
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		PairsPer1KChars: 2.0,
		ValidPercent:    10.0,
		TestPercent:     10.0,
		Model:           "gpt-oss:20b",
	}
}

// Pair represents a single prompt/completion pair.
type Pair struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// Generator handles the generation of prompt/completion pairs from documents.
type Generator struct {
	client *ollama.Client
	config Config
}

// NewGenerator creates a new Generator with the given Ollama client and config.
func NewGenerator(client *ollama.Client, config Config) *Generator {
	if config.Model == "" {
		config.Model = DefaultConfig().Model
	}
	return &Generator{
		client: client,
		config: config,
	}
}

// Generate creates prompt/completion pairs from a document for a specific split type.
func (g *Generator) Generate(ctx context.Context, doc *processor.Document, split SplitType) ([]Pair, error) {
	// Calculate pair count using float64 to avoid integer truncation
	totalPairs := g.calculatePairs(doc.Content)
	pairCounts := g.calculateSplitCounts(totalPairs)

	var count int
	switch split {
	case SplitTrain:
		count = pairCounts.Train
	case SplitValid:
		count = pairCounts.Valid
	case SplitTest:
		count = pairCounts.Test
	default:
		return nil, fmt.Errorf("unknown split type: %s", split)
	}

	if count <= 0 {
		return nil, fmt.Errorf("pair count is zero or negative for %s split", split)
	}

	// Get the correct template for this split type
	templateName := g.getTemplateName(split)

	// Execute template to get the prompt
	promptData := templates.PromptData{
		DocumentContent: doc.Content,
		PairCount:       count,
		DocumentName:    doc.Name,
	}

	prompt, err := templates.ExecuteTemplate(templateName, promptData)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	// Call Ollama to generate pairs
	response, err := g.client.Generate(ctx, g.config.Model, prompt)
	if err != nil {
		return nil, fmt.Errorf("generate pairs: %w", err)
	}

	// Parse the response into pairs (implementation depends on response format)
	pairs := g.parseResponse(response, count)

	return pairs, nil
}

// calculatePairs calculates the total number of pairs based on document length.
func (g *Generator) calculatePairs(content string) int {
	charCount := float64(len(content))
	pairs := math.Ceil(charCount / 1000 * g.config.PairsPer1KChars)
	return int(pairs)
}

// SplitCounts holds the pair counts for each split.
type SplitCounts struct {
	Train int
	Valid int
	Test  int
}

// calculateSplitCounts distributes pairs across splits using float64 math.
func (g *Generator) calculateSplitCounts(total int) SplitCounts {
	validCount := int(math.Ceil(float64(total) * g.config.ValidPercent / 100))
	testCount := int(math.Ceil(float64(total) * g.config.TestPercent / 100))
	// Train gets the remainder
	trainCount := total - validCount - testCount

	return SplitCounts{
		Train: trainCount,
		Valid: validCount,
		Test:  testCount,
	}
}

// getTemplateName returns the template file name for the given split type.
func (g *Generator) getTemplateName(split SplitType) string {
	switch split {
	case SplitTrain:
		return "train.tmpl"
	case SplitValid:
		return "valid.tmpl"
	case SplitTest:
		return "test.tmpl"
	default:
		return "train.tmpl"
	}
}

func (g *Generator) parseResponse(response string, expectedCount int) []Pair {
	// Try to find and parse JSON array from response
	// The LLM should return: [{"prompt": "...", "completion": "..."}, ...]

	// Clean up the response - find JSON array boundaries
	start := strings.Index("[", response)
	end := strings.LastIndex("]", response)

	if start == -1 || end == -1 || end <= start {
		// No JSON array found, return as single pair
		return []Pair{
			{
				Prompt:     "",
				Completion: response,
			},
		}
	}

	jsonStr := response[start : end+1]

	// Try to parse as array of Pair
	var pairs []Pair
	if err := json.Unmarshal([]byte(jsonStr), &pairs); err != nil {
		// Try to unescape if it's double-encoded
		var raw interface{}
		if err2 := json.Unmarshal([]byte(jsonStr), &raw); err2 == nil {
			// Check if it's a string containing JSON
			if str, ok := raw.(string); ok {
				// Try parsing the string as JSON
				if innerPairs, err3 := parseJSONArrayString(str); err3 == nil {
					return innerPairs
				}
			}
		}

		// Fallback: return the response as a single completion
		return []Pair{
			{
				Prompt:     "",
				Completion: response,
			},
		}
	}

	// If we got fewer pairs than expected, pad with empty pairs
	for len(pairs) < expectedCount {
		pairs = append(pairs, Pair{
			Prompt:     "",
			Completion: "",
		})
	}

	return pairs[:expectedCount]
}

// parseJSONArrayString parses a string that contains a JSON array
func parseJSONArrayString(s string) ([]Pair, error) {
	var pairs []Pair
	if err := json.Unmarshal([]byte(s), &pairs); err != nil {
		return nil, err
	}
	return pairs, nil
}
