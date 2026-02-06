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

// GetConfig returns the generator's configuration.
func (g *Generator) GetConfig() Config {
	return g.config
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
// Minimum 1 pair for any non-empty document.
func (g *Generator) calculatePairs(content string) int {
	charCount := float64(len(content))
	pairs := math.Ceil(charCount / 1000 * g.config.PairsPer1KChars)

	// Ensure minimum 1 pair for any non-empty document
	if pairs < 1 && charCount > 0 {
		return 1
	}
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
	if total <= 0 {
		return SplitCounts{Train: 0, Valid: 0, Test: 0}
	}

	// For small numbers of pairs, ensure we don't get negative counts
	if total == 1 {
		// With 1 pair, assign to train only
		return SplitCounts{Train: 1, Valid: 0, Test: 0}
	}
	if total == 2 {
		// With 2 pairs, give 1 to train, 1 to valid (test gets 0)
		return SplitCounts{Train: 1, Valid: 1, Test: 0}
	}

	validCount := int(math.Ceil(float64(total) * g.config.ValidPercent / 100))
	testCount := int(math.Ceil(float64(total) * g.config.TestPercent / 100))
	// Train gets the remainder
	trainCount := total - validCount - testCount

	// Ensure minimum 1 pair for each split when possible
	if trainCount < 1 && total >= 3 {
		trainCount = 1
		// Redistribute the extra pair
		remaining := total - 1
		validCount = int(math.Ceil(float64(remaining) * g.config.ValidPercent / 100))
		testCount = remaining - validCount
	}
	if validCount < 1 && total >= 2 {
		validCount = 1
	}
	if testCount < 1 && total >= 3 {
		testCount = 1
	}

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
	var pairs []Pair

	// Try to parse response only if it's not empty
	if len(strings.TrimSpace(response)) > 0 {
		// Try to find and parse JSON array from response
		// The LLM should return: [{"prompt": "...", "completion": "..."}, ...]

		// Clean up the response - find JSON array boundaries
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")

		if start != -1 && end != -1 && end > start {
			jsonStr := response[start : end+1]

			// Try to parse as array of Pair
			if err := json.Unmarshal([]byte(jsonStr), &pairs); err != nil {
				// Try to unescape if it's double-encoded
				var raw interface{}
				if err2 := json.Unmarshal([]byte(jsonStr), &raw); err2 == nil {
					// Check if it's a string containing JSON
					if str, ok := raw.(string); ok {
						// Try parsing the string as JSON
						if innerPairs, err3 := parseJSONArrayString(str); err3 == nil {
							pairs = innerPairs
						}
					}
				}
			}
		}
	}

	// Always ensure we return exactly expectedCount pairs
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

// validatePair returns true if the pair has non-empty, non-whitespace prompt and completion.
func validatePair(p Pair) bool {
	return strings.TrimSpace(p.Prompt) != "" && strings.TrimSpace(p.Completion) != ""
}

// deduplicatePairs removes exact duplicate pairs. A pair is a duplicate if both
// prompt AND completion match a previously seen pair (case-sensitive).
// Uses map[string]struct{} as a hash set with "prompt|||completion" as key.
func deduplicatePairs(pairs []Pair) []Pair {
	seen := make(map[string]struct{})
	var result []Pair
	for _, p := range pairs {
		key := fmt.Sprintf("%s|||%s", p.Prompt, p.Completion)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, p)
		}
	}
	return result
}
