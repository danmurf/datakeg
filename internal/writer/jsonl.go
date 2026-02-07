package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TrainingPair represents a single prompt/completion pair for training.
type TrainingPair struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatMessage represents a chat conversation with multiple messages.
type ChatMessage struct {
	Messages []Message `json:"messages"`
}

// ReasoningPair represents a reasoning training example with separate fields.
// Compatible with DeepSeek-R1 and similar reasoning model fine-tuning pipelines.
type ReasoningPair struct {
	Question  string `json:"question"`
	Reasoning string `json:"reasoning"`
	Answer    string `json:"answer"`
}

// WriteJSONL writes training pairs to a JSONL file.
// Each line in the output is a valid JSON object.
func WriteJSONL(filename string, pairs []TrainingPair) error {
	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	// Create JSON encoder - Encoder.Encode automatically adds newlines
	encoder := json.NewEncoder(file)

	for i, pair := range pairs {
		if err := encoder.Encode(pair); err != nil {
			return fmt.Errorf("encode pair %d: %w", i, err)
		}
	}

	// Ensure all buffered data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync file %s: %w", filename, err)
	}

	return nil
}

// WriteJSONLAppend appends training pairs to an existing JSONL file.
func WriteJSONLAppend(filename string, pairs []TrainingPair) error {
	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)

	for _, pair := range pairs {
		if err := encoder.Encode(pair); err != nil {
			return fmt.Errorf("encode pair: %w", err)
		}
	}

	return nil
}

// WriteChatJSONL writes chat messages to a JSONL file.
func WriteChatJSONL(filename string, messages []ChatMessage) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)

	for i, msg := range messages {
		if err := encoder.Encode(msg); err != nil {
			return fmt.Errorf("encode message %d: %w", i, err)
		}
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync file %s: %w", filename, err)
	}

	return nil
}

// WriteChatJSONLAppend appends chat messages to an existing JSONL file.
func WriteChatJSONLAppend(filename string, messages []ChatMessage) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)

	for _, msg := range messages {
		if err := encoder.Encode(msg); err != nil {
			return fmt.Errorf("encode message: %w", err)
		}
	}

	return nil
}

// ConvertPairToChatMessage converts a generator.Pair to a ChatMessage.
// If systemMessage is non-empty, it is inserted as the first message with role "system".
func ConvertPairToChatMessage(prompt, completion, systemMessage string) ChatMessage {
	var messages []Message

	if systemMessage != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: systemMessage,
		})
	}

	messages = append(messages,
		Message{Role: "user", Content: prompt},
		Message{Role: "assistant", Content: completion},
	)

	return ChatMessage{Messages: messages}
}

// WriteReasoningJSONL writes reasoning pairs to a JSONL file.
func WriteReasoningJSONL(filename string, pairs []ReasoningPair) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)

	for i, pair := range pairs {
		if err := encoder.Encode(pair); err != nil {
			return fmt.Errorf("encode pair %d: %w", i, err)
		}
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync file %s: %w", filename, err)
	}

	return nil
}

// WriteReasoningJSONLAppend appends reasoning pairs to an existing JSONL file.
func WriteReasoningJSONLAppend(filename string, pairs []ReasoningPair) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", filename, closeErr)
		}
	}()

	encoder := json.NewEncoder(file)

	for _, pair := range pairs {
		if err := encoder.Encode(pair); err != nil {
			return fmt.Errorf("encode pair: %w", err)
		}
	}

	return nil
}

// ConvertPairToReasoningPair converts a generator.Pair to a ReasoningPair.
// The completion is expected to contain the combined reasoning+answer (joined by "\n\n").
// It splits on the thinking tags to extract reasoning and answer separately.
func ConvertPairToReasoningPair(prompt, completion string) ReasoningPair {
	reasoning := ""
	answer := completion

	if strings.Contains(completion, "「thinking」") {
		idx := strings.Index(completion, "「thinking」")
		reasoningStart := idx + len("「thinking」")

		// Find the closing tag
		if strings.Contains(completion[reasoningStart:], "「/thinking」") {
			closeIdx := strings.Index(completion[reasoningStart:], "「/thinking」")
			reasoning = completion[idx : reasoningStart+closeIdx+len("「/thinking」")]
			answer = strings.TrimSpace(completion[reasoningStart+closeIdx+len("「/thinking」"):])
		}
	}

	return ReasoningPair{
		Question:  prompt,
		Reasoning: reasoning,
		Answer:    answer,
	}
}

// ConvertPairToIntegratedReasoning converts a generator.Pair to a TrainingPair
// for the integrated reasoning format (prompt/completion with inline think tags).
func ConvertPairToIntegratedReasoning(prompt, completion string) TrainingPair {
	return TrainingPair{
		Prompt:     prompt,
		Completion: completion,
	}
}
