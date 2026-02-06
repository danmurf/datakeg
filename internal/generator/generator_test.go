package generator

import (
	"testing"
)

func TestGenerator_calculatePairs(t *testing.T) {
	tests := []struct {
		name            string
		pairsPer1K      float64
		content         string
		want            int
	}{
		{
			name:       "empty content",
			pairsPer1K: 2.0,
			content:    "",
			want:       0,
		},
		{
			name:       "small content below threshold",
			pairsPer1K: 2.0,
			content:    "Small text",
			want:       1,
		},
		{
			name:       "exactly 1000 chars",
			pairsPer1K: 2.0,
			content:    string(make([]byte, 1000)),
			want:       2,
		},
		{
			name:       "exactly 500 chars",
			pairsPer1K: 2.0,
			content:    string(make([]byte, 500)),
			want:       1,
		},
		{
			name:       "1001 chars rounds up",
			pairsPer1K: 2.0,
			content:    string(make([]byte, 1001)),
			want:       3,
		},
		{
			name:       "large content 5000 chars",
			pairsPer1K: 2.0,
			content:    string(make([]byte, 5000)),
			want:       10,
		},
		{
			name:       "different pair rate 5 per 1k",
			pairsPer1K: 5.0,
			content:    string(make([]byte, 1000)),
			want:       5,
		},
		{
			name:       "fractional pair rate 0.5 per 1k",
			pairsPer1K: 0.5,
			content:    string(make([]byte, 1000)),
			want:       1,
		},
		{
			name:       "very small pair rate with large content",
			pairsPer1K: 0.1,
			content:    string(make([]byte, 5000)),
			want:       1,
		},
		{
			name:       "non-ascii content same length",
			pairsPer1K: 2.0,
			content:    "日本語のテキスト", // Each character is multiple bytes
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				PairsPer1KChars: tt.pairsPer1K,
				ValidPercent:    10.0,
				TestPercent:     10.0,
				Model:           "test-model",
			}
			g := NewGenerator(nil, config)
			got := g.calculatePairs(tt.content)
			if got != tt.want {
				t.Errorf("calculatePairs() = %v, want %v (content length: %d)", got, tt.want, len(tt.content))
			}
		})
	}
}

func TestGenerator_calculateSplitCounts(t *testing.T) {
	tests := []struct {
		name         string
		total        int
		validPercent float64
		testPercent  float64
		want         SplitCounts
	}{
		{
			name:         "zero pairs",
			total:        0,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 0, Valid: 0, Test: 0},
		},
		{
			name:         "one pair goes to train",
			total:        1,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 1, Valid: 0, Test: 0},
		},
		{
			name:         "two pairs split between train and valid",
			total:        2,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 1, Valid: 1, Test: 0},
		},
		{
			name:         "three pairs minimum distribution",
			total:        3,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 1, Valid: 1, Test: 1},
		},
		{
			name:         "ten pairs standard split",
			total:        10,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 8, Valid: 1, Test: 1},
		},
		{
			name:         "one hundred pairs standard split",
			total:        100,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 80, Valid: 10, Test: 10},
		},
		{
			name:         "twenty pairs standard split",
			total:        20,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 16, Valid: 2, Test: 2},
		},
		{
			name:         "different percentages 20-20",
			total:        10,
			validPercent: 20.0,
			testPercent:  20.0,
			want:         SplitCounts{Train: 6, Valid: 2, Test: 2},
		},
		{
			name:         "different percentages 15-15",
			total:        100,
			validPercent: 15.0,
			testPercent:  15.0,
			want:         SplitCounts{Train: 70, Valid: 15, Test: 15},
		},
		{
			name:         "uneven split with rounding",
			total:        7,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 5, Valid: 1, Test: 1},
		},
		{
			name:         "large count",
			total:        1000,
			validPercent: 10.0,
			testPercent:  10.0,
			want:         SplitCounts{Train: 800, Valid: 100, Test: 100},
		},
		{
			name:         "minimal valid percent",
			total:        100,
			validPercent: 5.0,
			testPercent:  5.0,
			want:         SplitCounts{Train: 90, Valid: 5, Test: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				PairsPer1KChars: 2.0,
				ValidPercent:    tt.validPercent,
				TestPercent:     tt.testPercent,
				Model:           "test-model",
			}
			g := NewGenerator(nil, config)
			got := g.calculateSplitCounts(tt.total)
			if got != tt.want {
				t.Errorf("calculateSplitCounts() = %+v, want %+v", got, tt.want)
			}
			// Verify totals add up (unless total is 0)
			if tt.total > 0 {
				sum := got.Train + got.Valid + got.Test
				if sum != tt.total {
					t.Errorf("split counts don't sum to total: %d + %d + %d = %d, want %d",
						got.Train, got.Valid, got.Test, sum, tt.total)
				}
			}
		})
	}
}

func TestGenerator_getTemplateName(t *testing.T) {
	tests := []struct {
		name      string
		splitType SplitType
		want      string
	}{
		{
			name:      "train split",
			splitType: SplitTrain,
			want:      "train.tmpl",
		},
		{
			name:      "valid split",
			splitType: SplitValid,
			want:      "valid.tmpl",
		},
		{
			name:      "test split",
			splitType: SplitTest,
			want:      "test.tmpl",
		},
		{
			name:      "unknown split defaults to train",
			splitType: SplitType("unknown"),
			want:      "train.tmpl",
		},
		{
			name:      "empty split defaults to train",
			splitType: SplitType(""),
			want:      "train.tmpl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			got := g.getTemplateName(tt.splitType)
			if got != tt.want {
				t.Errorf("getTemplateName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_parseResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectedCount int
		want          []Pair
	}{
		{
			name:          "valid single pair",
			response:      `[{"prompt": "What is Go?", "completion": "A programming language."}]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "What is Go?", Completion: "A programming language."},
			},
		},
		{
			name: "valid multiple pairs",
			response: `[
				{"prompt": "Q1", "completion": "A1"},
				{"prompt": "Q2", "completion": "A2"}
			]`,
			expectedCount: 2,
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "Q2", Completion: "A2"},
			},
		},
		{
			name:          "response with extra text before JSON",
			response:      `Here are the pairs:\n[{"prompt": "Q", "completion": "A"}]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:          "response with extra text after JSON",
			response:      `[{"prompt": "Q", "completion": "A"}]\nThese are good pairs.`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:          "empty response",
			response:      "",
			expectedCount: 1,
			want: []Pair{
				{Prompt: "", Completion: ""},
			},
		},
		{
			name:          "whitespace only",
			response:      "   \n\t  ",
			expectedCount: 1,
			want: []Pair{
				{Prompt: "", Completion: ""},
			},
		},
		{
			name:          "no JSON array",
			response:      "Just some text without JSON",
			expectedCount: 1,
			want: []Pair{
				{Prompt: "", Completion: ""},
			},
		},
		{
			name:          "malformed JSON",
			response:      `[{"prompt": "Q", "completion": }]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "", Completion: ""},
			},
		},
		{
			name:          "fewer pairs than expected pads with empty",
			response:      `[{"prompt": "Q1", "completion": "A1"}]`,
			expectedCount: 3,
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "", Completion: ""},
				{Prompt: "", Completion: ""},
			},
		},
		{
			name: "more pairs than expected truncates",
			response: `[
				{"prompt": "Q1", "completion": "A1"},
				{"prompt": "Q2", "completion": "A2"},
				{"prompt": "Q3", "completion": "A3"}
			]`,
			expectedCount: 2,
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "Q2", Completion: "A2"},
			},
		},
		{
			name:          "empty array",
			response:      `[]`,
			expectedCount: 2,
			want: []Pair{
				{Prompt: "", Completion: ""},
				{Prompt: "", Completion: ""},
			},
		},
		{
			name:          "JSON with escaped quotes",
			response:      `[{"prompt": "What's \"Go\"?", "completion": "It's a language."}]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: `What's "Go"?`, Completion: "It's a language."},
			},
		},
		{
			name:          "nested JSON should extract outer",
			response:      `[{"prompt": "Q", "completion": "A"}]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:          "multiline content in pairs",
			response:      `[{"prompt": "Line1\nLine2", "completion": "Answer1\nAnswer2"}]`,
			expectedCount: 1,
			want: []Pair{
				{Prompt: "Line1\nLine2", Completion: "Answer1\nAnswer2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			got := g.parseResponse(tt.response, tt.expectedCount)

			if len(got) != len(tt.want) {
				t.Errorf("parseResponse() returned %d pairs, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Prompt != tt.want[i].Prompt {
					t.Errorf("pair[%d].Prompt = %q, want %q", i, got[i].Prompt, tt.want[i].Prompt)
				}
				if got[i].Completion != tt.want[i].Completion {
					t.Errorf("pair[%d].Completion = %q, want %q", i, got[i].Completion, tt.want[i].Completion)
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name  string
		got   interface{}
		want  interface{}
	}{
		{"PairsPer1KChars", config.PairsPer1KChars, 2.0},
		{"ValidPercent", config.ValidPercent, 10.0},
		{"TestPercent", config.TestPercent, 10.0},
		{"Model", config.Model, "gpt-oss:20b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("DefaultConfig().%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantModel   string
	}{
		{
			name: "with specified model",
			config: Config{
				PairsPer1KChars: 2.0,
				ValidPercent:    10.0,
				TestPercent:     10.0,
				Model:           "custom-model",
			},
			wantModel: "custom-model",
		},
		{
			name: "empty model uses default",
			config: Config{
				PairsPer1KChars: 2.0,
				ValidPercent:    10.0,
				TestPercent:     10.0,
				Model:           "",
			},
			wantModel: "gpt-oss:20b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, tt.config)
			if g == nil {
				t.Fatal("NewGenerator() returned nil")
			}
			if g.config.Model != tt.wantModel {
				t.Errorf("NewGenerator().config.Model = %v, want %v", g.config.Model, tt.wantModel)
			}
		})
	}
}

func TestGenerator_GetConfig(t *testing.T) {
	config := Config{
		PairsPer1KChars: 3.5,
		ValidPercent:    15.0,
		TestPercent:     20.0,
		Model:           "test-model",
	}

	g := NewGenerator(nil, config)
	got := g.GetConfig()

	tests := []struct {
		name  string
		got   interface{}
		want  interface{}
	}{
		{"PairsPer1KChars", got.PairsPer1KChars, 3.5},
		{"ValidPercent", got.ValidPercent, 15.0},
		{"TestPercent", got.TestPercent, 20.0},
		{"Model", got.Model, "test-model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("GetConfig().%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
