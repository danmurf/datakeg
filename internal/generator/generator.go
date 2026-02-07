package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/provider"
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
		PairsPer1KChars: 10.0,
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
	provider provider.Provider
	config   Config
}

// NewGenerator creates a new Generator with the given provider and config.
func NewGenerator(p provider.Provider, config Config) *Generator {
	if config.Model == "" {
		config.Model = DefaultConfig().Model
	}
	return &Generator{
		provider: p,
		config:   config,
	}
}

// GetConfig returns the generator's configuration.
func (g *Generator) GetConfig() Config {
	return g.config
}

// Generate creates prompt/completion pairs from a document for a specific split type.
// It accepts excludePairs to avoid regenerating similar pairs.
// Returns pairs, accumulated usage metadata, and any error.
func (g *Generator) Generate(ctx context.Context, doc *processor.Document, split SplitType, excludePairs []Pair) ([]Pair, *provider.UsageMetadata, error) {
	// Calculate pair count using float64 to avoid integer truncation
	totalPairs := g.CalculatePairs(doc.Content)
	pairCounts := g.CalculateSplitCounts(totalPairs)

	var count int
	switch split {
	case SplitTrain:
		count = pairCounts.Train
	case SplitValid:
		count = pairCounts.Valid
	case SplitTest:
		count = pairCounts.Test
	default:
		return nil, nil, fmt.Errorf("unknown split type: %s", split)
	}

	if count <= 0 {
		return nil, nil, nil
	}

	// Get the correct template for this split type
	templateName := g.getTemplateName(split)

	// Convert generator.Pair to templates.ExcludePair for template rendering
	var excludeTemplatePairs []templates.ExcludePair
	for _, p := range excludePairs {
		excludeTemplatePairs = append(excludeTemplatePairs, templates.ExcludePair{
			Prompt:     p.Prompt,
			Completion: p.Completion,
		})
	}

	// Execute template to get the prompt
	promptData := templates.PromptData{
		DocumentContent: doc.Content,
		PairCount:       count,
		DocumentName:    doc.Name,
		ExcludePairs:    excludeTemplatePairs,
	}

	prompt, err := templates.ExecuteTemplate(templateName, promptData)
	if err != nil {
		return nil, nil, fmt.Errorf("execute template: %w", err)
	}

	// Call provider to generate pairs
	response, usage, err := g.provider.Generate(ctx, g.config.Model, prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("generate pairs: %w", err)
	}

	// Track accumulated usage
	var totalUsage provider.UsageMetadata
	if usage != nil {
		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.CompletionTokens += usage.CompletionTokens
		totalUsage.TotalTokens += usage.TotalTokens
		totalUsage.EstimatedCost += usage.EstimatedCost
	}

	// Parse the response into pairs
	rawPairs := g.parseResponse(response)

	// Validate all pairs - discard invalid ones
	var validPairs []Pair
	for _, p := range rawPairs {
		if validatePair(p) {
			validPairs = append(validPairs, p)
		}
	}

	// Deduplicate among themselves
	uniquePairs := deduplicatePairs(validPairs)

	// Also deduplicate against excludePairs
	uniquePairs = deduplicateAgainstExclusions(uniquePairs, excludePairs)

	// If we have enough pairs, return them
	if len(uniquePairs) >= count {
		return uniquePairs[:count], &totalUsage, nil
	}

	// Backfill loop - retry up to 3 times to reach target count
	const maxBackfillAttempts = 3
	allPairs := uniquePairs

	for attempt := 0; attempt < maxBackfillAttempts && len(allPairs) < count; attempt++ {
		gap := count - len(allPairs)

		// Combine original excludePairs with allPairs we've collected so far
		newExcludePairs := append(excludePairs, allPairs...)

		// Convert to template format
		var newExcludeTemplatePairs []templates.ExcludePair
		for _, p := range newExcludePairs {
			newExcludeTemplatePairs = append(newExcludeTemplatePairs, templates.ExcludePair{
				Prompt:     p.Prompt,
				Completion: p.Completion,
			})
		}

		// Execute template with new exclusions and gap as PairCount
		backfillPromptData := templates.PromptData{
			DocumentContent: doc.Content,
			PairCount:       gap,
			DocumentName:    doc.Name,
			ExcludePairs:    newExcludeTemplatePairs,
		}

		backfillPrompt, err := templates.ExecuteTemplate(templateName, backfillPromptData)
		if err != nil {
			// Log error but continue with what we have
			fmt.Fprintf(os.Stderr, "Warning: template execution failed for backfill attempt %d: %v\n", attempt+1, err)
			continue
		}

		// Call provider for backfill
		backfillResponse, backfillUsage, err := g.provider.Generate(ctx, g.config.Model, backfillPrompt)
		if err != nil {
			// Log error but continue with what we have
			fmt.Fprintf(os.Stderr, "Warning: generation failed for backfill attempt %d: %v\n", attempt+1, err)
			continue
		}

		// Accumulate backfill usage
		if backfillUsage != nil {
			totalUsage.PromptTokens += backfillUsage.PromptTokens
			totalUsage.CompletionTokens += backfillUsage.CompletionTokens
			totalUsage.TotalTokens += backfillUsage.TotalTokens
			totalUsage.EstimatedCost += backfillUsage.EstimatedCost
		}

		// Parse backfill response
		backfillPairs := g.parseResponse(backfillResponse)

		// Validate backfill pairs
		var validBackfillPairs []Pair
		for _, p := range backfillPairs {
			if validatePair(p) {
				validBackfillPairs = append(validBackfillPairs, p)
			}
		}

		// Deduplicate backfill pairs among themselves
		uniqueBackfillPairs := deduplicatePairs(validBackfillPairs)

		// Deduplicate against both original excludePairs and all pairs we've collected
		uniqueBackfillPairs = deduplicateAgainstExclusions(uniqueBackfillPairs, append(excludePairs, allPairs...))

		// Append genuinely new pairs to allPairs
		for _, p := range uniqueBackfillPairs {
			if len(allPairs) < count {
				allPairs = append(allPairs, p)
			}
		}
	}

	// Log warning if we couldn't reach target count
	if len(allPairs) < count {
		fmt.Fprintf(os.Stderr, "Warning: could not reach target count for %s split after %d attempts (got %d, wanted %d)\n", split, maxBackfillAttempts, len(allPairs), count)
	}

	// Return what we have (may be less than count if max attempts reached)
	if len(allPairs) > count {
		return allPairs[:count], &totalUsage, nil
	}
	return allPairs, &totalUsage, nil
}

// CalculatePairs calculates the total number of pairs based on document length.
// Minimum 1 pair for any non-empty document.
func (g *Generator) CalculatePairs(content string) int {
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

// CalculateSplitCounts distributes pairs across splits using float64 math.
// Valid and test are rounded to nearest integer; train gets the remainder
// to ensure the total is preserved exactly.
func (g *Generator) CalculateSplitCounts(total int) SplitCounts {
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

	validCount := int(math.Round(float64(total) * g.config.ValidPercent / 100))
	testCount := int(math.Round(float64(total) * g.config.TestPercent / 100))

	// Ensure minimum 1 for valid and test when total >= 3
	if validCount < 1 {
		validCount = 1
	}
	if testCount < 1 {
		testCount = 1
	}

	// Train gets the remainder to ensure total is preserved exactly
	trainCount := total - validCount - testCount

	// Ensure minimum 1 for train (reduce valid/test proportionally if needed)
	if trainCount < 1 {
		trainCount = 1
		remaining := total - 1
		validRatio := g.config.ValidPercent / (g.config.ValidPercent + g.config.TestPercent)
		validCount = int(math.Round(float64(remaining) * validRatio))
		testCount = remaining - validCount
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

func (g *Generator) parseResponse(response string) []Pair {
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

	return pairs
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

// deduplicatePairs removes exact duplicate pairs and also removes pairs with duplicate prompts.
// A pair is a duplicate if both prompt AND completion match a previously seen pair (case-sensitive),
// OR if the prompt matches exactly (regardless of completion).
// Uses map[string]struct{} as hash sets with "prompt|||completion" and "prompt" as keys.
func deduplicatePairs(pairs []Pair) []Pair {
	seen := make(map[string]struct{})
	seenPrompts := make(map[string]struct{})
	var result []Pair
	for _, p := range pairs {
		key := fmt.Sprintf("%s|||%s", p.Prompt, p.Completion)
		promptKey := p.Prompt
		_, keyExists := seen[key]
		_, promptExists := seenPrompts[promptKey]
		if !keyExists && !promptExists {
			seen[key] = struct{}{}
			seenPrompts[promptKey] = struct{}{}
			result = append(result, p)
		}
	}
	return result
}

// deduplicateAgainstExclusions removes pairs that match entries in excludePairs.
// A pair matches if both prompt AND completion are identical (case-sensitive),
// OR if the prompt matches exactly (regardless of completion).
func deduplicateAgainstExclusions(pairs []Pair, excludePairs []Pair) []Pair {
	excluded := make(map[string]struct{})
	excludedPrompts := make(map[string]struct{})
	for _, ep := range excludePairs {
		key := fmt.Sprintf("%s|||%s", ep.Prompt, ep.Completion)
		excluded[key] = struct{}{}
		excludedPrompts[ep.Prompt] = struct{}{}
	}

	var result []Pair
	for _, p := range pairs {
		key := fmt.Sprintf("%s|||%s", p.Prompt, p.Completion)
		_, keyExists := excluded[key]
		_, promptExists := excludedPrompts[p.Prompt]
		if !keyExists && !promptExists {
			result = append(result, p)
		}
	}
	return result
}
