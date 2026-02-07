package usage

import (
	"math"

	"github.com/danmurf/datakeg/internal/provider"
)

// EstimateTokens approximates token count using 4:1 character-to-token ratio.
// This is an approximation suitable for English text. Actual counts may vary by 20-50%
// for non-English, code, or special characters.
func EstimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 4.0))
}

// Tracker accumulates token usage across a run.
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
