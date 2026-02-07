package cost

import (
	"testing"

	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/provider"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"5 chars", "hello", 2}, // ceil(5/4) = 2
		{"1000 chars", string(make([]byte, 1000)), 250},
		{"4 chars", "test", 1}, // ceil(4/4) = 1
		{"1 char", "a", 1},     // ceil(1/4) = 1
		{"4000 chars", string(make([]byte, 4000)), 1000},
		{"1 byte UTF-8 char", "é", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.input)
			if result != tt.expected {
				t.Errorf("EstimateTokens(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTracker_Add(t *testing.T) {
	tracker := &Tracker{}

	// Add nil (should not panic, should be skipped)
	tracker.Add(nil)
	if tracker.TotalTokens != 0 {
		t.Errorf("expected 0 tokens after nil add, got: %d", tracker.TotalTokens)
	}

	// Add first usage
	usage1 := &provider.UsageMetadata{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		EstimatedCost:    0.00025,
	}
	tracker.Add(usage1)
	if tracker.TotalPromptTokens != 100 {
		t.Errorf("expected 100 prompt tokens, got: %d", tracker.TotalPromptTokens)
	}
	if tracker.TotalCompletionTokens != 50 {
		t.Errorf("expected 50 completion tokens, got: %d", tracker.TotalCompletionTokens)
	}
	if tracker.TotalTokens != 150 {
		t.Errorf("expected 150 total tokens, got: %d", tracker.TotalTokens)
	}
	if tracker.TotalCost != 0.00025 {
		t.Errorf("expected 0.00025 cost, got: %f", tracker.TotalCost)
	}

	// Add second usage
	usage2 := &provider.UsageMetadata{
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
		EstimatedCost:    0.00050,
	}
	tracker.Add(usage2)
	if tracker.TotalPromptTokens != 300 {
		t.Errorf("expected 300 prompt tokens after second add, got: %d", tracker.TotalPromptTokens)
	}
	if tracker.TotalCompletionTokens != 150 {
		t.Errorf("expected 150 completion tokens after second add, got: %d", tracker.TotalCompletionTokens)
	}
	if tracker.TotalTokens != 450 {
		t.Errorf("expected 450 total tokens after second add, got: %d", tracker.TotalTokens)
	}
	if tracker.TotalCost != 0.00075 {
		t.Errorf("expected 0.00075 cost after second add, got: %f", tracker.TotalCost)
	}
}

func TestTracker_Summary(t *testing.T) {
	tracker := &Tracker{
		TotalPromptTokens:     1000,
		TotalCompletionTokens: 500,
		TotalTokens:           1500,
		TotalCost:             0.0015,
	}

	summary := tracker.Summary()
	expected := "Tokens used: 1000 prompt + 500 completion = 1500 total. Estimated cost: $0.0015"
	if summary != expected {
		t.Errorf("Summary() = %q, want %q", summary, expected)
	}
}

func TestEstimateRunCost(t *testing.T) {
	// Create a single document with 1000 chars
	docs := []processor.Document{
		{Content: string(make([]byte, 1000)), Name: "test.md"},
	}

	// Estimate cost: pairs = ceil(1000/1000 * 10) = 10 pairs
	// Prompt: 1000 chars + 500 template = 1500 chars -> ceil(1500/4) = 375 tokens
	// Completion: 10 pairs * 3 splits * 200 chars = 6000 chars -> ceil(6000/4) = 1500 tokens
	// Total prompt tokens = 375, completion = 1500
	// Cost: (375/1M * $1.0) + (1500/1M * $2.0) = $0.000375 + $0.003 = $0.003375
	cost := EstimateRunCost(docs, 10.0, 1.0, 2.0)

	// Allow for some variation in the estimate
	if cost < 0.002 || cost > 0.005 {
		t.Errorf("EstimateRunCost = %f, expected between 0.002 and 0.005", cost)
	}
}
