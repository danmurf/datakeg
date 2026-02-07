package generator

import (
	"testing"
)

func TestGenerator_calculatePairs(t *testing.T) {
	tests := []struct {
		name       string
		pairsPer1K float64
		content    string
		want       int
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
			got := g.CalculatePairs(tt.content)
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
		{
			name:         "non-integer split rounds correctly 201 at 20-20",
			total:        201,
			validPercent: 20.0,
			testPercent:  20.0,
			want:         SplitCounts{Train: 121, Valid: 40, Test: 40},
		},
		{
			name:         "non-integer split rounds correctly 42 at 20-20",
			total:        42,
			validPercent: 20.0,
			testPercent:  20.0,
			want:         SplitCounts{Train: 26, Valid: 8, Test: 8},
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
			got := g.CalculateSplitCounts(tt.total)
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
			got := g.getTemplateName(FormatCompletion, tt.splitType)
			if got != tt.want {
				t.Errorf("getTemplateName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_parseResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []Pair
	}{
		{
			name:     "valid single pair",
			response: `[{"prompt": "What is Go?", "completion": "A programming language."}]`,
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
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "Q2", Completion: "A2"},
			},
		},
		{
			name:     "response with extra text before JSON",
			response: `Here are the pairs:\n[{"prompt": "Q", "completion": "A"}]`,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:     "response with extra text after JSON",
			response: `[{"prompt": "Q", "completion": "A"}]\nThese are good pairs.`,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:     "empty response",
			response: "",
			want:     []Pair{},
		},
		{
			name:     "whitespace only",
			response: "   \n\t  ",
			want:     []Pair{},
		},
		{
			name:     "no JSON array",
			response: "Just some text without JSON",
			want:     []Pair{},
		},
		{
			name:     "malformed JSON",
			response: `[{"prompt": "Q", "completion": }]`,
			want:     []Pair{},
		},
		{
			name: "more pairs returned than before - no truncation",
			response: `[
				{"prompt": "Q1", "completion": "A1"},
				{"prompt": "Q2", "completion": "A2"},
				{"prompt": "Q3", "completion": "A3"}
			]`,
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "Q2", Completion: "A2"},
				{Prompt: "Q3", Completion: "A3"},
			},
		},
		{
			name:     "empty array",
			response: `[]`,
			want:     []Pair{},
		},
		{
			name:     "JSON with escaped quotes",
			response: `[{"prompt": "What's \"Go\"?", "completion": "It's a language."}]`,
			want: []Pair{
				{Prompt: `What's "Go"?`, Completion: "It's a language."},
			},
		},
		{
			name:     "nested JSON should extract outer",
			response: `[{"prompt": "Q", "completion": "A"}]`,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:     "multiline content in pairs",
			response: `[{"prompt": "Line1\nLine2", "completion": "Answer1\nAnswer2"}]`,
			want: []Pair{
				{Prompt: "Line1\nLine2", Completion: "Answer1\nAnswer2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			got := g.parseResponse(tt.response)

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

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    FormatType
		wantErr bool
	}{
		{
			name:    "valid completion format",
			input:   "completion",
			want:    FormatCompletion,
			wantErr: false,
		},
		{
			name:    "valid chat format",
			input:   "chat",
			want:    FormatChat,
			wantErr: false,
		},
		{
			name:    "valid reasoning format",
			input:   "reasoning",
			want:    FormatReasoning,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "json",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "unknown format",
			input:   "text",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTemplateName_FormatAware(t *testing.T) {
	tests := []struct {
		name      string
		format    FormatType
		splitType SplitType
		want      string
	}{
		{
			name:      "completion train",
			format:    FormatCompletion,
			splitType: SplitTrain,
			want:      "train.tmpl",
		},
		{
			name:      "completion valid",
			format:    FormatCompletion,
			splitType: SplitValid,
			want:      "valid.tmpl",
		},
		{
			name:      "completion test",
			format:    FormatCompletion,
			splitType: SplitTest,
			want:      "test.tmpl",
		},
		{
			name:      "chat train",
			format:    FormatChat,
			splitType: SplitTrain,
			want:      "chat_train.tmpl",
		},
		{
			name:      "chat valid",
			format:    FormatChat,
			splitType: SplitValid,
			want:      "chat_valid.tmpl",
		},
		{
			name:      "chat test",
			format:    FormatChat,
			splitType: SplitTest,
			want:      "chat_test.tmpl",
		},
		{
			name:      "reasoning train",
			format:    FormatReasoning,
			splitType: SplitTrain,
			want:      "reasoning_train.tmpl",
		},
		{
			name:      "reasoning valid",
			format:    FormatReasoning,
			splitType: SplitValid,
			want:      "reasoning_valid.tmpl",
		},
		{
			name:      "reasoning test",
			format:    FormatReasoning,
			splitType: SplitTest,
			want:      "reasoning_test.tmpl",
		},
		{
			name:      "unknown split defaults to reasoning_train",
			format:    FormatReasoning,
			splitType: SplitType("unknown"),
			want:      "reasoning_train.tmpl",
		},
		{
			name:      "unknown format defaults to completion train",
			format:    FormatType("unknown"),
			splitType: SplitTrain,
			want:      "train.tmpl",
		},
		{
			name:      "unknown split defaults to chat_train",
			format:    FormatChat,
			splitType: SplitType("unknown"),
			want:      "chat_train.tmpl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			g.config.Format = tt.format
			got := g.getTemplateName(tt.format, tt.splitType)
			if got != tt.want {
				t.Errorf("getTemplateName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_parseChatResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []Pair
	}{
		{
			name:     "valid single chat pair",
			response: `[{"user": "What is Go?", "assistant": "A programming language."}]`,
			want: []Pair{
				{Prompt: "What is Go?", Completion: "A programming language."},
			},
		},
		{
			name: "valid multiple chat pairs",
			response: `[
				{"user": "Q1", "assistant": "A1"},
				{"user": "Q2", "assistant": "A2"}
			]`,
			want: []Pair{
				{Prompt: "Q1", Completion: "A1"},
				{Prompt: "Q2", Completion: "A2"},
			},
		},
		{
			name:     "response with extra text before JSON",
			response: `Here are the pairs:\n[{"user": "Q", "assistant": "A"}]`,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:     "response with extra text after JSON",
			response: `[{"user": "Q", "assistant": "A"}]\nThese are good pairs.`,
			want: []Pair{
				{Prompt: "Q", Completion: "A"},
			},
		},
		{
			name:     "empty response",
			response: "",
			want:     []Pair{},
		},
		{
			name:     "whitespace only",
			response: "   \n\t  ",
			want:     []Pair{},
		},
		{
			name:     "no JSON array",
			response: "Just some text without JSON",
			want:     []Pair{},
		},
		{
			name:     "malformed JSON",
			response: `[{"user": "Q", "assistant": }]`,
			want:     []Pair{},
		},
		{
			name:     "empty array",
			response: `[]`,
			want:     []Pair{},
		},
		{
			name:     "JSON with escaped quotes",
			response: `[{"user": "What's \"Go\"?", "assistant": "It's a language."}]`,
			want: []Pair{
				{Prompt: `What's "Go"?`, Completion: "It's a language."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			got := g.parseChatResponse(tt.response)

			if len(got) != len(tt.want) {
				t.Errorf("parseChatResponse() returned %d pairs, want %d", len(got), len(tt.want))
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

func TestParseReasoningFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ReasoningFormat
		wantErr bool
	}{
		{
			name:    "valid separate format",
			input:   "separate",
			want:    ReasoningFormatSeparate,
			wantErr: false,
		},
		{
			name:    "valid integrated format",
			input:   "integrated",
			want:    ReasoningFormatIntegrated,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "unknown format",
			input:   "merged",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReasoningFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReasoningFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseReasoningFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerator_parseReasoningResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []Pair
	}{
		{
			name:     "valid single reasoning pair",
			response: `[{"question": "Why does X?", "reasoning": "「thinking」Step 1: ... Therefore...「/thinking」", "answer": "Because Y."}]`,
			want: []Pair{
				{Prompt: "Why does X?", Completion: "「thinking」Step 1: ... Therefore...「/thinking」\n\nBecause Y."},
			},
		},
		{
			name: "valid multiple reasoning pairs",
			response: `[
				{"question": "Why X?", "reasoning": "「thinking」Step 1...「/thinking」", "answer": "Because Y."},
				{"question": "How A?", "reasoning": "「thinking」Step 1...「/thinking」", "answer": "By doing B."}
			]`,
			want: []Pair{
				{Prompt: "Why X?", Completion: "「thinking」Step 1...「/thinking」\n\nBecause Y."},
				{Prompt: "How A?", Completion: "「thinking」Step 1...「/thinking」\n\nBy doing B."},
			},
		},
		{
			name:     "response with extra text before JSON",
			response: `Here are the pairs:\n[{"question": "Q", "reasoning": "「thinking」R「/thinking」", "answer": "A"}]`,
			want: []Pair{
				{Prompt: "Q", Completion: "「thinking」R「/thinking」\n\nA"},
			},
		},
		{
			name:     "response with extra text after JSON",
			response: `[{"question": "Q", "reasoning": "「thinking」R「/thinking」", "answer": "A"}]\nThese are good pairs.`,
			want: []Pair{
				{Prompt: "Q", Completion: "「thinking」R「/thinking」\n\nA"},
			},
		},
		{
			name:     "empty response",
			response: "",
			want:     []Pair{},
		},
		{
			name:     "whitespace only",
			response: "   \n\t  ",
			want:     []Pair{},
		},
		{
			name:     "no JSON array",
			response: "Just some text without JSON",
			want:     []Pair{},
		},
		{
			name:     "malformed JSON",
			response: `[{"question": "Q", "reasoning": "R", "answer": }]`,
			want:     []Pair{},
		},
		{
			name:     "empty array",
			response: `[]`,
			want:     []Pair{},
		},
		{
			name:     "reasoning without think tags",
			response: `[{"question": "Why X?", "reasoning": "Step 1: ... Therefore...", "answer": "Because Y."}]`,
			want: []Pair{
				{Prompt: "Why X?", Completion: "Step 1: ... Therefore...\n\nBecause Y."},
			},
		},
		{
			name:     "empty reasoning field",
			response: `[{"question": "Why X?", "reasoning": "", "answer": "Because Y."}]`,
			want: []Pair{
				{Prompt: "Why X?", Completion: "\n\nBecause Y."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(nil, DefaultConfig())
			got := g.parseReasoningResponse(tt.response)

			if len(got) != len(tt.want) {
				t.Errorf("parseReasoningResponse() returned %d pairs, want %d", len(got), len(tt.want))
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
		name string
		got  interface{}
		want interface{}
	}{
		{"PairsPer1KChars", config.PairsPer1KChars, 10.0},
		{"ValidPercent", config.ValidPercent, 10.0},
		{"TestPercent", config.TestPercent, 10.0},
		{"Model", config.Model, "gpt-oss:20b"},
		{"Format", config.Format, FormatCompletion},
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
		name      string
		config    Config
		wantModel string
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
		name string
		got  interface{}
		want interface{}
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

func TestValidatePair(t *testing.T) {
	tests := []struct {
		name string
		pair Pair
		want bool
	}{
		{
			name: "valid pair with non-empty prompt and completion",
			pair: Pair{Prompt: "What is Go?", Completion: "A programming language."},
			want: true,
		},
		{
			name: "empty prompt",
			pair: Pair{Prompt: "", Completion: "A programming language."},
			want: false,
		},
		{
			name: "empty completion",
			pair: Pair{Prompt: "What is Go?", Completion: ""},
			want: false,
		},
		{
			name: "whitespace-only prompt",
			pair: Pair{Prompt: "   \n\t", Completion: "A programming language."},
			want: false,
		},
		{
			name: "whitespace-only completion",
			pair: Pair{Prompt: "What is Go?", Completion: "   \n\t"},
			want: false,
		},
		{
			name: "both empty",
			pair: Pair{Prompt: "", Completion: ""},
			want: false,
		},
		{
			name: "prompt with leading/trailing whitespace but content",
			pair: Pair{Prompt: "  What is Go?  ", Completion: "  A programming language.  "},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validatePair(tt.pair)
			if got != tt.want {
				t.Errorf("validatePair(%+v) = %v, want %v", tt.pair, got, tt.want)
			}
		})
	}
}

func TestDeduplicatePairs(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []Pair
		expected []Pair
	}{
		{
			name:     "no duplicates - all preserved, order maintained",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
		},
		{
			name:     "exact duplicate - second removed",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A1"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}},
		},
		{
			name:     "same prompt, different completion - only first kept",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A2"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}},
		},
		{
			name:     "different prompt, same completion - both kept",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A1"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A1"}},
		},
		{
			name:     "multiple duplicates - only first of each kept",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}, {Prompt: "Q1", Completion: "A1"}, {Prompt: "Q3", Completion: "A3"}, {Prompt: "Q2", Completion: "A2"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}, {Prompt: "Q3", Completion: "A3"}},
		},
		{
			name:     "empty input - empty output",
			pairs:    []Pair{},
			expected: []Pair{},
		},
		{
			name:     "single pair - single pair returned",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}},
		},
		{
			name:     "three identical pairs - one pair returned",
			pairs:    []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A1"}},
			expected: []Pair{{Prompt: "Q1", Completion: "A1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicatePairs(tt.pairs)
			if len(got) != len(tt.expected) {
				t.Errorf("deduplicatePairs() returned %d pairs, want %d", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i].Prompt != tt.expected[i].Prompt || got[i].Completion != tt.expected[i].Completion {
					t.Errorf("deduplicatePairs()[%d] = %+v, want %+v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestDeduplicateAgainstExclusions(t *testing.T) {
	tests := []struct {
		name          string
		pairs         []Pair
		excludePairs  []Pair
		expectedPairs []Pair
	}{
		{
			name:          "no exclusions - all pairs kept",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
			excludePairs:  []Pair{},
			expectedPairs: []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
		},
		{
			name:          "nil exclusions - all pairs kept",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}},
			excludePairs:  nil,
			expectedPairs: []Pair{{Prompt: "Q1", Completion: "A1"}},
		},
		{
			name:          "exact match excluded - pair removed",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}},
			expectedPairs: []Pair{{Prompt: "Q2", Completion: "A2"}},
		},
		{
			name:          "multiple exclusions - multiple pairs removed",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}, {Prompt: "Q3", Completion: "A3"}},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q3", Completion: "A3"}},
			expectedPairs: []Pair{{Prompt: "Q2", Completion: "A2"}},
		},
		{
			name:          "same prompt different completion - excluded by prompt match",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A2"}},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}},
			expectedPairs: []Pair{},
		},
		{
			name:          "same completion different prompt - only matching prompt excluded",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A1"}},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}},
			expectedPairs: []Pair{{Prompt: "Q2", Completion: "A1"}},
		},
		{
			name:          "all pairs excluded - empty result",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q2", Completion: "A2"}},
			expectedPairs: []Pair{},
		},
		{
			name:          "empty pairs - empty result",
			pairs:         []Pair{},
			excludePairs:  []Pair{{Prompt: "Q1", Completion: "A1"}},
			expectedPairs: []Pair{},
		},
		{
			name:          "duplicate pairs - duplicates preserved, no exclusions",
			pairs:         []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A1"}},
			excludePairs:  []Pair{},
			expectedPairs: []Pair{{Prompt: "Q1", Completion: "A1"}, {Prompt: "Q1", Completion: "A1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateAgainstExclusions(tt.pairs, tt.excludePairs)
			if len(got) != len(tt.expectedPairs) {
				t.Errorf("deduplicateAgainstExclusions() returned %d pairs, want %d", len(got), len(tt.expectedPairs))
				return
			}
			for i := range got {
				if got[i].Prompt != tt.expectedPairs[i].Prompt || got[i].Completion != tt.expectedPairs[i].Completion {
					t.Errorf("deduplicateAgainstExclusions()[%d] = %+v, want %+v", i, got[i], tt.expectedPairs[i])
				}
			}
		})
	}
}
