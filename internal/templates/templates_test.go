package templates

import (
	"strings"
	"testing"
)

// TestExecuteTemplate_train tests the training template execution.
func TestExecuteTemplate_train(t *testing.T) {
	data := PromptData{
		DocumentContent: "This is a test document about Go programming.",
		PairCount:       3,
		DocumentName:    "test.md",
	}

	result, err := ExecuteTemplate("train.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify template substitutions
	if !strings.Contains(result, "This is a test document about Go programming.") {
		t.Error("ExecuteTemplate() missing DocumentContent")
	}
	if !strings.Contains(result, "3") {
		t.Error("ExecuteTemplate() missing PairCount")
	}
	if !strings.Contains(result, "training data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_valid tests the validation template execution.
func TestExecuteTemplate_valid(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test content for validation.",
		PairCount:       5,
		DocumentName:    "valid.md",
	}

	result, err := ExecuteTemplate("valid.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify template substitutions
	if !strings.Contains(result, "Test content for validation.") {
		t.Error("ExecuteTemplate() missing DocumentContent")
	}
	if !strings.Contains(result, "5") {
		t.Error("ExecuteTemplate() missing PairCount")
	}
	if !strings.Contains(result, "validation data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_test tests the test template execution.
func TestExecuteTemplate_test(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test content for evaluation.",
		PairCount:       2,
		DocumentName:    "test_doc.md",
	}

	result, err := ExecuteTemplate("test.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify template substitutions
	if !strings.Contains(result, "Test content for evaluation.") {
		t.Error("ExecuteTemplate() missing DocumentContent")
	}
	if !strings.Contains(result, "2") {
		t.Error("ExecuteTemplate() missing PairCount")
	}
	if !strings.Contains(result, "test data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_invalidName tests error handling for invalid template names.
func TestExecuteTemplate_invalidName(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test",
		PairCount:       1,
	}

	_, err := ExecuteTemplate("nonexistent.tmpl", data)
	if err == nil {
		t.Error("ExecuteTemplate() should fail for invalid template")
	}
}
