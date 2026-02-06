package writer

import (
	"os"
	"strings"
	"testing"
)

func TestWriteJSONL(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-*.jsonl")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFilename := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFilename); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()

	// Write test pairs
	pairs := []TrainingPair{
		{Prompt: "What is Go?", Completion: "Go is a programming language."},
		{Prompt: "What is Cobra?", Completion: "Cobra is a CLI framework."},
	}

	if err := WriteJSONL(tmpFilename, pairs); err != nil {
		t.Fatalf("WriteJSONL failed: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		if len(line) == 0 {
			t.Errorf("line %d is empty", i)
			continue
		}
		// Basic check: lines should contain the expected keys
		if !strings.Contains(line, "prompt") || !strings.Contains(line, "completion") {
			t.Errorf("line %d missing expected keys: %s", i, line)
		}
	}
}

func TestWriteJSONLAppend(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-append-*.jsonl")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFilename := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFilename); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()

	// Write first batch
	pairs1 := []TrainingPair{
		{Prompt: "Q1", Completion: "A1"},
	}
	if err := WriteJSONL(tmpFilename, pairs1); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Append second batch
	pairs2 := []TrainingPair{
		{Prompt: "Q2", Completion: "A2"},
	}
	if err := WriteJSONLAppend(tmpFilename, pairs2); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Verify total lines
	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after append, got %d", len(lines))
	}
}
