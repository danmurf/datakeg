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

func TestWriteChatJSONL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-chat-*.jsonl")
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

	messages := []ChatMessage{
		{
			Messages: []Message{
				{Role: "user", Content: "What is Go?"},
				{Role: "assistant", Content: "Go is a programming language."},
			},
		},
		{
			Messages: []Message{
				{Role: "user", Content: "What is Cobra?"},
				{Role: "assistant", Content: "Cobra is a CLI framework."},
			},
		},
	}

	if err := WriteChatJSONL(tmpFilename, messages); err != nil {
		t.Fatalf("WriteChatJSONL failed: %v", err)
	}

	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if len(line) == 0 {
			t.Errorf("line %d is empty", i)
			continue
		}
		if !strings.Contains(line, "messages") || !strings.Contains(line, "role") || !strings.Contains(line, "user") || !strings.Contains(line, "assistant") {
			t.Errorf("line %d missing expected keys: %s", i, line)
		}
	}
}

func TestWriteChatJSONLAppend(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-chat-append-*.jsonl")
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

	messages1 := []ChatMessage{
		{
			Messages: []Message{
				{Role: "user", Content: "Q1"},
				{Role: "assistant", Content: "A1"},
			},
		},
	}
	if err := WriteChatJSONL(tmpFilename, messages1); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	messages2 := []ChatMessage{
		{
			Messages: []Message{
				{Role: "user", Content: "Q2"},
				{Role: "assistant", Content: "A2"},
			},
		},
	}
	if err := WriteChatJSONLAppend(tmpFilename, messages2); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after append, got %d", len(lines))
	}
}

func TestConvertPairToChatMessage_noSystem(t *testing.T) {
	prompt := "What is Go?"
	completion := "Go is a programming language."

	msg := ConvertPairToChatMessage(prompt, completion, "")

	if len(msg.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msg.Messages))
	}
	if msg.Messages[0].Role != "user" {
		t.Errorf("first message role = %q, want %q", msg.Messages[0].Role, "user")
	}
	if msg.Messages[0].Content != prompt {
		t.Errorf("first message content = %q, want %q", msg.Messages[0].Content, prompt)
	}
	if msg.Messages[1].Role != "assistant" {
		t.Errorf("second message role = %q, want %q", msg.Messages[1].Role, "assistant")
	}
	if msg.Messages[1].Content != completion {
		t.Errorf("second message content = %q, want %q", msg.Messages[1].Content, completion)
	}
}

func TestConvertPairToChatMessage_withSystem(t *testing.T) {
	prompt := "What is Go?"
	completion := "Go is a programming language."
	systemMessage := "You are a helpful assistant."

	msg := ConvertPairToChatMessage(prompt, completion, systemMessage)

	if len(msg.Messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(msg.Messages))
	}
	if msg.Messages[0].Role != "system" {
		t.Errorf("first message role = %q, want %q", msg.Messages[0].Role, "system")
	}
	if msg.Messages[0].Content != systemMessage {
		t.Errorf("first message content = %q, want %q", msg.Messages[0].Content, systemMessage)
	}
	if msg.Messages[1].Role != "user" {
		t.Errorf("second message role = %q, want %q", msg.Messages[1].Role, "user")
	}
	if msg.Messages[1].Content != prompt {
		t.Errorf("second message content = %q, want %q", msg.Messages[1].Content, prompt)
	}
	if msg.Messages[2].Role != "assistant" {
		t.Errorf("third message role = %q, want %q", msg.Messages[2].Role, "assistant")
	}
	if msg.Messages[2].Content != completion {
		t.Errorf("third message content = %q, want %q", msg.Messages[2].Content, completion)
	}
}

func TestWriteReasoningJSONL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-reasoning-*.jsonl")
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

	pairs := []ReasoningPair{
		{Question: "Why X?", Reasoning: "「thinking」Step 1...「/thinking」", Answer: "Because Y."},
		{Question: "How A?", Reasoning: "「thinking」Step 1...「/thinking」", Answer: "By doing B."},
	}

	if err := WriteReasoningJSONL(tmpFilename, pairs); err != nil {
		t.Fatalf("WriteReasoningJSONL failed: %v", err)
	}

	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	for i, line := range lines {
		if len(line) == 0 {
			t.Errorf("line %d is empty", i)
			continue
		}
		if !strings.Contains(line, "question") || !strings.Contains(line, "reasoning") || !strings.Contains(line, "answer") {
			t.Errorf("line %d missing expected keys: %s", i, line)
		}
	}
}

func TestWriteReasoningJSONLAppend(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-reasoning-append-*.jsonl")
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

	pairs1 := []ReasoningPair{
		{Question: "Q1", Reasoning: "R1", Answer: "A1"},
	}
	if err := WriteReasoningJSONL(tmpFilename, pairs1); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	pairs2 := []ReasoningPair{
		{Question: "Q2", Reasoning: "R2", Answer: "A2"},
	}
	if err := WriteReasoningJSONLAppend(tmpFilename, pairs2); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	content, err := os.ReadFile(tmpFilename)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after append, got %d", len(lines))
	}
}

func TestConvertPairToReasoningPair_withThinkTags(t *testing.T) {
	prompt := "Why does X?"
	completion := "「thinking」Step 1: ... Step 2: ... Therefore...「/thinking」\n\nBecause Y."

	pair := ConvertPairToReasoningPair(prompt, completion)

	if pair.Question != prompt {
		t.Errorf("Question = %q, want %q", pair.Question, prompt)
	}
	if pair.Reasoning != "「thinking」Step 1: ... Step 2: ... Therefore...「/thinking」" {
		t.Errorf("Reasoning = %q, want %q", pair.Reasoning, "「thinking」Step 1: ... Step 2: ... Therefore...「/thinking」")
	}
	if pair.Answer != "Because Y." {
		t.Errorf("Answer = %q, want %q", pair.Answer, "Because Y.")
	}
}

func TestConvertPairToReasoningPair_withoutThinkTags(t *testing.T) {
	prompt := "Why does X?"
	completion := "Because Y."

	pair := ConvertPairToReasoningPair(prompt, completion)

	if pair.Question != prompt {
		t.Errorf("Question = %q, want %q", pair.Question, prompt)
	}
	if pair.Reasoning != "" {
		t.Errorf("Reasoning = %q, want empty string", pair.Reasoning)
	}
	if pair.Answer != "Because Y." {
		t.Errorf("Answer = %q, want %q", pair.Answer, "Because Y.")
	}
}

func TestConvertPairToIntegratedReasoning(t *testing.T) {
	prompt := "Why does X?"
	completion := "「thinking」...「/thinking」\n\nBecause Y."

	pair := ConvertPairToIntegratedReasoning(prompt, completion)

	if pair.Prompt != prompt {
		t.Errorf("Prompt = %q, want %q", pair.Prompt, prompt)
	}
	if pair.Completion != completion {
		t.Errorf("Completion = %q, want %q", pair.Completion, completion)
	}
}
