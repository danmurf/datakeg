package cost

import (
	"fmt"
	"math"

	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/provider"
)

// EstimateTokens approximates token count using 4:1 character-to-token ratio.
// This is an approximation suitable for English text. Actual counts may vary by 20-50%
// for non-English, code, or special characters.
func EstimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 4.0))
}

// EstimateRunCost calculates estimated API cost based on document sizes and pair counts.
// Uses rough estimates: prompt = document chars + template overhead, completion = ~200 chars per pair.
// This is an approximation for pre-run estimation purposes.
func EstimateRunCost(documents []processor.Document, pairsPer1KChars float64, promptPricePerMToken float64, completionPricePerMToken float64) float64 {
	var totalPromptTokens int
	var totalCompletionTokens int

	for _, doc := range documents {
		// Estimate prompt tokens: document content + template overhead (~500 chars per call)
		docChars := len(doc.Content)
		promptChars := docChars + 500 // Template overhead
		promptTokens := EstimateTokens(string(make([]byte, promptChars)))
		totalPromptTokens += promptTokens

		// Estimate pairs for this document
		pairs := int(math.Ceil(float64(docChars) / 1000 * pairsPer1KChars))
		if pairs < 1 && docChars > 0 {
			pairs = 1
		}

		// Estimate completion tokens: ~200 chars per pair, 3 splits
		completionChars := pairs * 3 * 200
		completionTokens := EstimateTokens(string(make([]byte, completionChars)))
		totalCompletionTokens += completionTokens
	}

	// Calculate cost
	promptCost := (float64(totalPromptTokens) / 1_000_000) * promptPricePerMToken
	completionCost := (float64(totalCompletionTokens) / 1_000_000) * completionPricePerMToken

	return promptCost + completionCost
}

// Tracker accumulates token usage and calculates cost across a run.
type Tracker struct {
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalTokens           int
	TotalCost             float64
}

// Add adds usage metadata to the running totals.
// Safe to call with nil usage (skipped for local providers like Ollama).
func (t *Tracker) Add(usage *provider.UsageMetadata) {
	if usage == nil {
		return
	}
	t.TotalPromptTokens += usage.PromptTokens
	t.TotalCompletionTokens += usage.CompletionTokens
	t.TotalTokens += usage.TotalTokens
	t.TotalCost += usage.EstimatedCost
}

// Summary returns a formatted summary of the usage.
func (t *Tracker) Summary() string {
	return fmt.Sprintf("Tokens used: %d prompt + %d completion = %d total. Estimated cost: $%.4f",
		t.TotalPromptTokens, t.TotalCompletionTokens, t.TotalTokens, t.TotalCost)
}
