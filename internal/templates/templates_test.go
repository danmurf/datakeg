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
	if !strings.Contains(result, "TEST data") {
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

// TestExecuteTemplate_excludePairs tests that ExcludePairs are conditionally rendered.
func TestExecuteTemplate_excludePairs(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test document content.",
		PairCount:       2,
		ExcludePairs: []ExcludePair{
			{Prompt: "Q1", Completion: "A1"},
			{Prompt: "Q2", Completion: "A2"},
		},
	}

	result, err := ExecuteTemplate("train.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify exclusion section is rendered
	if !strings.Contains(result, "Do NOT generate questions that are semantically similar to any of the following") {
		t.Error("ExecuteTemplate() missing exclusion instruction")
	}
	if !strings.Contains(result, "Q1") {
		t.Error("ExecuteTemplate() missing first exclude pair prompt")
	}
	if !strings.Contains(result, "A1") {
		t.Error("ExecuteTemplate() missing first exclude pair completion")
	}
	if !strings.Contains(result, "Q2") {
		t.Error("ExecuteTemplate() missing second exclude pair prompt")
	}
	if !strings.Contains(result, "A2") {
		t.Error("ExecuteTemplate() missing second exclude pair completion")
	}
}

// TestExecuteTemplate_noExcludePairs tests that no exclusion section renders when ExcludePairs is nil.
func TestExecuteTemplate_noExcludePairs(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test document content.",
		PairCount:       2,
		ExcludePairs:    nil,
	}

	result, err := ExecuteTemplate("valid.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify exclusion section is NOT rendered
	if strings.Contains(result, "Do NOT generate questions that are semantically similar to any of the following") {
		t.Error("ExecuteTemplate() should not render exclusion section when ExcludePairs is nil")
	}
	if strings.Contains(result, "Prompt:") {
		t.Error("ExecuteTemplate() should not render exclude pair prompts when ExcludePairs is nil")
	}
}

// TestExecuteTemplate_emptyExcludePairs tests that no exclusion section renders when ExcludePairs is empty.
func TestExecuteTemplate_emptyExcludePairs(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test document content.",
		PairCount:       2,
		ExcludePairs:    []ExcludePair{},
	}

	result, err := ExecuteTemplate("test.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify exclusion section is NOT rendered
	if strings.Contains(result, "Do NOT generate questions that are semantically similar to any of the following") {
		t.Error("ExecuteTemplate() should not render exclusion section when ExcludePairs is empty")
	}
}

// TestExecuteTemplate_chatTrain tests the chat training template execution.
func TestExecuteTemplate_chatTrain(t *testing.T) {
	data := PromptData{
		DocumentContent: "This is a test document about Go programming.",
		PairCount:       3,
		DocumentName:    "test.md",
	}

	result, err := ExecuteTemplate("chat_train.tmpl", data)
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
	if !strings.Contains(result, "single-turn") {
		t.Error("ExecuteTemplate() missing single-turn instruction")
	}
	if !strings.Contains(result, "chat training data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_chatValid tests the chat validation template execution.
func TestExecuteTemplate_chatValid(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test content for validation.",
		PairCount:       5,
		DocumentName:    "valid.md",
	}

	result, err := ExecuteTemplate("chat_valid.tmpl", data)
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
	if !strings.Contains(result, "validation chat data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_chatTest tests the chat test template execution.
func TestExecuteTemplate_chatTest(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test content for evaluation.",
		PairCount:       2,
		DocumentName:    "test_doc.md",
	}

	result, err := ExecuteTemplate("chat_test.tmpl", data)
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
	if !strings.Contains(result, "TEST chat data") {
		t.Error("ExecuteTemplate() wrong template used")
	}
}

// TestExecuteTemplate_chatExcludePairs tests that exclusion section renders correctly in chat template.
func TestExecuteTemplate_chatExcludePairs(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test document content.",
		PairCount:       2,
		ExcludePairs: []ExcludePair{
			{Prompt: "Q1", Completion: "A1"},
			{Prompt: "Q2", Completion: "A2"},
		},
	}

	result, err := ExecuteTemplate("chat_train.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify exclusion section is rendered
	if !strings.Contains(result, "Do NOT generate questions semantically similar to these previously generated pairs") {
		t.Error("ExecuteTemplate() missing exclusion instruction")
	}
	if !strings.Contains(result, "Q1") {
		t.Error("ExecuteTemplate() missing first exclude pair prompt")
	}
	if !strings.Contains(result, "A1") {
		t.Error("ExecuteTemplate() missing first exclude pair completion")
	}
	if !strings.Contains(result, "Q2") {
		t.Error("ExecuteTemplate() missing second exclude pair prompt")
	}
	if !strings.Contains(result, "A2") {
		t.Error("ExecuteTemplate() missing second exclude pair completion")
	}
}

// TestExecuteTemplate_chatNoExcludePairs tests that no exclusion section renders when ExcludePairs is nil in chat template.
func TestExecuteTemplate_chatNoExcludePairs(t *testing.T) {
	data := PromptData{
		DocumentContent: "Test document content.",
		PairCount:       2,
		ExcludePairs:    nil,
	}

	result, err := ExecuteTemplate("chat_valid.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	// Verify exclusion section is NOT rendered
	if strings.Contains(result, "Do NOT generate questions semantically similar to these previously generated pairs") {
		t.Error("ExecuteTemplate() should not render exclusion section when ExcludePairs is nil")
	}
}
