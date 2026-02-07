// Package converter provides utilities for converting JSONL files between different formats
// using Go templates for transformation rules.
package converter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/writer"
)

// DetectFormat detects the format of a JSONL line by examining its field structure.
// Returns the detected format or an error if the format cannot be determined.
func DetectFormat(jsonLine []byte) (generator.FormatType, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonLine, &raw); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Check for chat format: {"messages": [...]}
	if messages, ok := raw["messages"].([]interface{}); ok && len(messages) > 0 {
		return generator.FormatChat, nil
	}

	// Check for reasoning format: {"question": "...", "reasoning": "...", "answer": "..."}
	if _, hasQuestion := raw["question"]; hasQuestion {
		if _, hasReasoning := raw["reasoning"]; hasReasoning {
			if _, hasAnswer := raw["answer"]; hasAnswer {
				return generator.FormatReasoning, nil
			}
		}
	}

	// Check for completion format: {"prompt": "...", "completion": "..."}
	if _, hasPrompt := raw["prompt"]; hasPrompt {
		if _, hasCompletion := raw["completion"]; hasCompletion {
			return generator.FormatCompletion, nil
		}
	}

	return "", fmt.Errorf("unknown format: detected fields %v", getMapKeys(raw))
}

// DetectFormatFromFile reads the first non-empty line from a JSONL file and detects its format.
// Returns the detected format or an error if the format cannot be determined.
func DetectFormatFromFile(filePath string) (generator.FormatType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 1MB initial buffer, 10MB max for long lines

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		return DetectFormat([]byte(line))
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return "", fmt.Errorf("read file: %w", scanErr)
	}

	return "", fmt.Errorf("file is empty: no JSON lines found")
}

// ValidateTemplate validates that a template produces valid JSON output for the given format.
// It executes the template with sample data and verifies the output is valid JSON.
func ValidateTemplate(tmpl *template.Template, format generator.FormatType) error {
	// Create sample data based on format
	var testData interface{}
	switch format {
	case generator.FormatCompletion:
		testData = writer.TrainingPair{Prompt: "test prompt", Completion: "test completion"}
	case generator.FormatChat:
		testData = writer.ChatMessage{
			Messages: []writer.Message{
				{Role: "user", Content: "test user message"},
				{Role: "assistant", Content: "test assistant response"},
			},
		}
	case generator.FormatReasoning:
		testData = writer.ReasoningPair{
			Question:  "test question",
			Reasoning: "test reasoning",
			Answer:    "test answer",
		}
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	// Execute template with sample data
	var buf strings.Builder
	if err := tmpl.Execute(&buf, testData); err != nil {
		return fmt.Errorf("template execution failed with %s format: %w", format, err)
	}

	// Verify output is valid JSON
	var result interface{}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		return fmt.Errorf("template output is not valid JSON for %s format: %w\nOutput: %s", format, err, buf.String())
	}

	return nil
}

// ConvertJSONL reads a JSONL file line by line, transforms each line using the provided template,
// and writes the transformed output to a new JSONL file.
// Returns the number of lines converted or an error.
func ConvertJSONL(inputPath, outputPath string, tmpl *template.Template, sourceFormat generator.FormatType) (int, error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return 0, fmt.Errorf("open input file: %w", err)
	}
	defer func() {
		if closeErr := inputFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Use scanner for line-by-line reading
	scanner := bufio.NewScanner(inputFile)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 1MB initial buffer, 10MB max

	lineNum := 0
	linesConverted := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lineNum++

		// Unmarshal based on source format
		var data interface{}
		switch sourceFormat {
		case generator.FormatCompletion:
			var pair writer.TrainingPair
			if err := json.Unmarshal([]byte(line), &pair); err != nil {
				return linesConverted, fmt.Errorf("line %d: failed to parse %s format: %w", lineNum, sourceFormat, err)
			}
			data = pair
		case generator.FormatChat:
			var msg writer.ChatMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				return linesConverted, fmt.Errorf("line %d: failed to parse %s format: %w", lineNum, sourceFormat, err)
			}
			data = msg
		case generator.FormatReasoning:
			var rp writer.ReasoningPair
			if err := json.Unmarshal([]byte(line), &rp); err != nil {
				return linesConverted, fmt.Errorf("line %d: failed to parse %s format: %w", lineNum, sourceFormat, err)
			}
			data = rp
		default:
			return linesConverted, fmt.Errorf("line %d: unknown format: %s", lineNum, sourceFormat)
		}

		// Execute template
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return linesConverted, fmt.Errorf("line %d: template execution failed: %w", lineNum, err)
		}

		// Write template output as JSON line
		// The template output should already be valid JSON
		if _, err := outputFile.WriteString(buf.String()); err != nil {
			return linesConverted, fmt.Errorf("line %d: write output: %w", lineNum, err)
		}
		if _, err := outputFile.WriteString("\n"); err != nil {
			return linesConverted, fmt.Errorf("line %d: write newline: %w", lineNum, err)
		}

		linesConverted++
	}

	if err := scanner.Err(); err != nil {
		return linesConverted, fmt.Errorf("read input file: %w", err)
	}

	// Sync output file
	if err := outputFile.Sync(); err != nil {
		return linesConverted, fmt.Errorf("sync output file: %w", err)
	}

	return linesConverted, nil
}

// getMapKeys returns the keys of a map as a slice of strings.
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
