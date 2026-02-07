package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/templates"
)

// createTempJSONL creates a temporary JSONL file with the given lines.
func createTempJSONL(t *testing.T, lines []string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.jsonl")

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			t.Fatalf("Failed to write line: %v", err)
		}
	}

	return filePath
}

// createTestTemplate creates a template with the jsonEscape function registered.
func createTestTemplate(t *testing.T, tmplContent string) *template.Template {
	t.Helper()

	tmpl, err := template.New("test").
		Funcs(template.FuncMap{"jsonEscape": jsonEscape}).
		Parse(tmplContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	return tmpl
}

// jsonEscape is the same function as in templates package - duplicated here for testing
func jsonEscape(s string) string {
	result := strings.ReplaceAll(s, "\\", "\\\\")
	result = strings.ReplaceAll(result, "\"", "\\\"")
	result = strings.ReplaceAll(result, "\n", "\\n")
	result = strings.ReplaceAll(result, "\r", "\\r")
	result = strings.ReplaceAll(result, "\t", "\\t")
	return result
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    generator.FormatType
		wantErr bool
	}{
		{
			name:    "completion format",
			input:   `{"prompt":"test prompt","completion":"test completion"}`,
			want:    generator.FormatCompletion,
			wantErr: false,
		},
		{
			name:    "chat format",
			input:   `{"messages":[{"role":"user","content":"test"}]}`,
			want:    generator.FormatChat,
			wantErr: false,
		},
		{
			name:    "reasoning format",
			input:   `{"question":"test","reasoning":"test reasoning","answer":"test answer"}`,
			want:    generator.FormatReasoning,
			wantErr: false,
		},
		{
			name:    "unknown format",
			input:   `{"unknown":"field"}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   `not valid json`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFormat([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectFormatFromFile(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    generator.FormatType
		wantErr bool
	}{
		{
			name:    "completion format",
			lines:   []string{`{"prompt":"x","completion":"y"}`},
			want:    generator.FormatCompletion,
			wantErr: false,
		},
		{
			name:    "chat format",
			lines:   []string{`{"messages":[{"role":"user","content":"x"}]}`},
			want:    generator.FormatChat,
			wantErr: false,
		},
		{
			name:    "reasoning format",
			lines:   []string{`{"question":"x","reasoning":"y","answer":"z"}`},
			want:    generator.FormatReasoning,
			wantErr: false,
		},
		{
			name:    "empty file",
			lines:   []string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only file",
			lines:   []string{"", "  ", "\n"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempJSONL(t, tt.lines)
			got, err := DetectFormatFromFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormatFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormatFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    *template.Template
		format  generator.FormatType
		wantErr bool
	}{
		{
			name:    "valid completion template",
			tmpl:    createTestTemplate(t, `{"prompt":"{{.Prompt}}","completion":"{{.Completion}}"}`),
			format:  generator.FormatCompletion,
			wantErr: false,
		},
		{
			name:    "valid chat template",
			tmpl:    createTestTemplate(t, `{"role":"user","content":"test"}`),
			format:  generator.FormatChat,
			wantErr: false,
		},
		{
			name:    "valid reasoning template",
			tmpl:    createTestTemplate(t, `{"question":"{{.Question}}","reasoning":"{{.Reasoning}}","answer":"{{.Answer}}"}`),
			format:  generator.FormatReasoning,
			wantErr: false,
		},
		{
			name:    "template with invalid field reference",
			tmpl:    createTestTemplate(t, `{"nonexistent":"{{.NonExistent}}"}`),
			format:  generator.FormatCompletion,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.tmpl, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvertJSONL(t *testing.T) {
	// Create a simple completion template
	tmpl := createTestTemplate(t, `{"text":"Q: {{.Prompt | jsonEscape}} A: {{.Completion | jsonEscape}}"}`)

	// Create test input
	inputLines := []string{
		`{"prompt":"What is 2+2?","completion":"The answer is 4."}`,
		`{"prompt":"What is the capital of France?","completion":"Paris."}`,
		`{"prompt":"What color is the sky?","completion":"Blue."}`,
	}
	inputPath := createTempJSONL(t, inputLines)

	// Create temp output path
	tmpDir := filepath.Dir(inputPath)
	outputPath := filepath.Join(tmpDir, "output.jsonl")

	// Convert
	linesConverted, err := ConvertJSONL(inputPath, outputPath, tmpl, generator.FormatCompletion)
	if err != nil {
		t.Fatalf("ConvertJSONL() error = %v", err)
	}

	if linesConverted != 3 {
		t.Errorf("ConvertJSONL() linesConverted = %v, want 3", linesConverted)
	}

	// Verify output
	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(string(outputContent)), "\n")
	if len(outputLines) != 3 {
		t.Errorf("Output has %d lines, want 3", len(outputLines))
	}

	// Each line should be valid JSON
	for i, line := range outputLines {
		if len(line) == 0 {
			t.Errorf("Line %d is empty", i)
			continue
		}
		// Just check it starts with {"text":...} since it's template output
		if !strings.HasPrefix(line, `{"text":"`) {
			t.Errorf("Line %d doesn't match expected format: %s", i, line)
		}
	}
}

func TestConvertJSONL_ChatFormat(t *testing.T) {
	// Create a simple chat template that just outputs a fixed string
	tmpl := createTestTemplate(t, `{"converted":"true"}`)

	// Create test input with chat format
	inputLines := []string{
		`{"messages":[{"role":"user","content":"Hello"},{"role":"assistant","content":"Hi there!"}]}`,
	}
	inputPath := createTempJSONL(t, inputLines)

	tmpDir := filepath.Dir(inputPath)
	outputPath := filepath.Join(tmpDir, "output.jsonl")

	linesConverted, err := ConvertJSONL(inputPath, outputPath, tmpl, generator.FormatChat)
	if err != nil {
		t.Fatalf("ConvertJSONL() error = %v", err)
	}

	if linesConverted != 1 {
		t.Errorf("ConvertJSONL() linesConverted = %v, want 1", linesConverted)
	}
}

func TestConvertJSONL_EmptyFile(t *testing.T) {
	tmpl := createTestTemplate(t, `{"text":"{{.Prompt}}"}`)

	inputLines := []string{}
	inputPath := createTempJSONL(t, inputLines)

	tmpDir := filepath.Dir(inputPath)
	outputPath := filepath.Join(tmpDir, "output.jsonl")

	linesConverted, err := ConvertJSONL(inputPath, outputPath, tmpl, generator.FormatCompletion)
	if err != nil {
		t.Errorf("ConvertJSONL() error on empty file = %v", err)
	}

	if linesConverted != 0 {
		t.Errorf("ConvertJSONL() linesConverted = %v, want 0", linesConverted)
	}
}

func TestConvertJSONL_InvalidJSON(t *testing.T) {
	tmpl := createTestTemplate(t, `{"text":"{{.Prompt}}"}`)

	inputLines := []string{
		`{"prompt":"valid"}`,
		`not valid json`,
	}
	inputPath := createTempJSONL(t, inputLines)

	tmpDir := filepath.Dir(inputPath)
	outputPath := filepath.Join(tmpDir, "output.jsonl")

	_, err := ConvertJSONL(inputPath, outputPath, tmpl, generator.FormatCompletion)
	if err == nil {
		t.Error("ConvertJSONL() should return error for invalid JSON")
	}
}

func TestLoadConversionTemplate(t *testing.T) {
	// Test loading built-in templates
	templateNames := []string{"mistral-instruct", "llama3-instruct", "chatml", "deepseek-r1"}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			tmpl, err := templates.LoadConversionTemplate(name)
			if err != nil {
				t.Errorf("LoadConversionTemplate(%s) error = %v", name, err)
				return
			}
			if tmpl == nil {
				t.Errorf("LoadConversionTemplate(%s) returned nil", name)
			}
		})
	}

	// Test non-existent template
	_, err := templates.LoadConversionTemplate("nonexistent")
	if err == nil {
		t.Error("LoadConversionTemplate(nonexistent) should return error")
	}
}

func TestListConversionTemplates(t *testing.T) {
	templateList, err := templates.ListConversionTemplates()
	if err != nil {
		t.Errorf("ListConversionTemplates() error = %v", err)
		return
	}

	// Should have at least our 4 built-in templates
	if len(templateList) < 4 {
		t.Errorf("ListConversionTemplates() returned %d templates, want at least 4", len(templateList))
	}

	// Check for expected templates
	expected := map[string]bool{
		"mistral-instruct": false,
		"llama3-instruct":  false,
		"chatml":           false,
		"deepseek-r1":      false,
	}

	for _, name := range templateList {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected template %s not found", name)
		}
	}
}
