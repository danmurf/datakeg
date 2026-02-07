# Phase 5: Chat Format Support - Research

**Researched:** 2026-02-07
**Domain:** Chat-style training data generation, OpenAI-compatible messages format, Go CLI format flags
**Confidence:** HIGH

## Summary

Phase 5 adds chat format support to datakeg's existing completion-based training data generation. The standard approach is OpenAI-compatible messages format with role-based JSON structures (`{"messages": [...]}`), widely adopted across the LLM ecosystem. Chat format uses single-turn conversations (one user message, one assistant response) with optional system messages. Implementation requires a format flag, new templates for chat-style prompts, a chat-specific output structure, and deduplication adapted for messages.

The key technical decisions are straightforward: Go's type-safe string enum pattern for format selection, strategy pattern for output formatters, and cobra's flag validation for CLI integration. Templates must instruct the LLM to generate context-free user messages and assistant responses that mirror the source document's voice. System messages can be auto-generated from document context using recent research approaches (SysGen-style annotation) or provided by users via flag.

**Primary recommendation:** Use OpenAI-compatible messages format with single-turn conversations, implement format as a string enum with validation, create separate chat templates per split, and adapt deduplication to hash on user message content rather than prompt field.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Conversation structure:**
- Single-turn only: one user message + one assistant response per entry
- No multi-turn conversations
- System message is optional, enabled via `--system-message` flag
- Claude's Discretion: whether system message content is user-provided or auto-generated from document context

**Template instructions:**
- Mix of question styles: factual, clarifying, how-to, and conceptual questions for diversity
- User messages are context-free (no reference to "the docs" or "the guide")
- Assistant responses must mirror the voice, tone, dialect, and conversational style of the source document — as if the document's author is speaking
- Claude's Discretion: whether to explicitly instruct the LLM to match document style or let it infer naturally
- Separate templates per split (train, valid, test) — same as completion format

**Output shape:**
- OpenAI-compatible messages format: `{"messages": [{"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}]}`
- When system message included: `{"messages": [{"role": "system", "content": "..."}, {"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}]}`
- Messages only — no metadata fields (no source doc name, no timestamps)
- Claude's Discretion: system message position in array (standard is first)
- Per-document files use same naming pattern as completion format (docname_train.jsonl)

**Format flag behavior:**
- One format per run: `--format chat` or `--format completion`, not both
- Claude's Discretion: default format when `--format` not specified (likely completion for backward compatibility)
- Claude's Discretion: merge subcommand format detection (auto-detect vs require flag)
- Claude's Discretion: behavior when format mismatch with existing files in output directory

### Claude's Discretion

- System message implementation (user-provided vs auto-generated)
- Template wording for style matching
- Default format value
- Merge format detection approach
- Format coexistence/overwrite behavior
- Deduplication adaptation for messages format

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope

</user_constraints>

## Standard Stack

### Core Libraries (Already in Use)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| encoding/json | stdlib | JSON encoding/decoding | Go standard library, zero dependencies, supports custom marshaling via interfaces |
| github.com/spf13/cobra | latest | CLI framework | Industry standard for Go CLIs, already in use, excellent flag handling |
| text/template | stdlib | Template execution | Already used for prompt templates, supports Go template syntax |

### Supporting Libraries

No additional libraries required. The codebase already has all necessary dependencies for implementing chat format.

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| encoding/json | github.com/goccy/go-json | Faster encoding but adds dependency; unnecessary for datakeg's use case |
| cobra | stdlib flag package | Less features, no subcommand support; cobra already integrated |
| Manual enum validation | github.com/thediveo/enumflag | Auto-validation library but adds complexity; simple switch validation sufficient |

**Installation:**
No additional dependencies required.

## Architecture Patterns

### Recommended Project Structure

```
internal/
├── generator/        # Add FormatType enum, update Generator interface
├── writer/          # Add ChatMessage struct, WriteChatJSONL function
├── templates/       # Add chat_train.tmpl, chat_valid.tmpl, chat_test.tmpl
└── formatter/       # NEW: Strategy pattern for format-specific logic (OPTIONAL)
```

### Pattern 1: Type-Safe Format Enum

**What:** Define FormatType as a string-based type with const values, similar to existing SplitType pattern.

**When to use:** For format selection flag with compile-time type safety.

**Example:**
```go
// In internal/generator/generator.go or new internal/format/format.go
type FormatType string

const (
    FormatCompletion FormatType = "completion"
    FormatChat       FormatType = "chat"
)

// Validate format string from CLI flag
func ParseFormat(s string) (FormatType, error) {
    switch s {
    case string(FormatCompletion):
        return FormatCompletion, nil
    case string(FormatChat):
        return FormatChat, nil
    default:
        return "", fmt.Errorf("invalid format: %s (must be 'completion' or 'chat')", s)
    }
}
```

### Pattern 2: Strategy Pattern for Output Writers

**What:** Define a Writer interface with format-specific implementations.

**When to use:** When output logic diverges significantly between formats (completion vs chat).

**Example:**
```go
// In internal/writer/writer.go
type Writer interface {
    Write(filename string, pairs []generator.Pair) error
}

// CompletionWriter writes prompt/completion format
type CompletionWriter struct{}

func (w *CompletionWriter) Write(filename string, pairs []generator.Pair) error {
    trainingPairs := convertToTrainingPairs(pairs)
    return WriteJSONL(filename, trainingPairs)
}

// ChatWriter writes messages format
type ChatWriter struct {
    includeSystemMessage bool
    systemMessageContent string
}

func (w *ChatWriter) Write(filename string, pairs []generator.Pair) error {
    chatMessages := convertToChatMessages(pairs, w.includeSystemMessage, w.systemMessageContent)
    return WriteChatJSONL(filename, chatMessages)
}
```

### Pattern 3: Template Selection by Format

**What:** Select template filename based on format type and split type.

**When to use:** Loading the correct template for chat vs completion generation.

**Example:**
```go
// In internal/generator/generator.go
func (g *Generator) getTemplateName(format FormatType, split SplitType) string {
    switch format {
    case FormatCompletion:
        return fmt.Sprintf("%s.tmpl", split) // "train.tmpl"
    case FormatChat:
        return fmt.Sprintf("chat_%s.tmpl", split) // "chat_train.tmpl"
    default:
        return "train.tmpl" // Fallback
    }
}
```

### Pattern 4: Messages Deduplication Key

**What:** Adapt deduplication to hash on user message content instead of prompt field.

**When to use:** Deduplicating chat-format pairs where content is nested in messages array.

**Example:**
```go
// Extract user message content for deduplication key
func extractUserMessage(msg ChatMessage) string {
    for _, m := range msg.Messages {
        if m.Role == "user" {
            return m.Content
        }
    }
    return ""
}

// Deduplication for chat format
func deduplicateChatMessages(messages []ChatMessage) []ChatMessage {
    seen := make(map[string]struct{})
    var result []ChatMessage
    for _, msg := range messages {
        userContent := extractUserMessage(msg)
        if _, exists := seen[userContent]; !exists {
            seen[userContent] = struct{}{}
            result = append(result, msg)
        }
    }
    return result
}
```

### Anti-Patterns to Avoid

- **Format detection from file content:** Don't auto-detect format by parsing existing JSONL files. Require explicit flag to avoid ambiguity and accidental format mixing.
- **Shared templates for both formats:** Don't try to use the same template for both completion and chat formats. The prompt structure and output expectations are fundamentally different.
- **Hardcoded format strings:** Don't use raw strings like "chat" throughout the codebase. Use the FormatType enum for type safety.
- **Mixing formats in output directory:** Don't allow both completion and chat files in the same output directory without clear warnings or errors.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Format validation | Custom string validation logic | Switch statement with error return | Simple, clear, follows existing SplitType pattern |
| JSON encoding | Manual JSON string building | encoding/json with struct tags | Type-safe, handles escaping, standard Go approach |
| Template rendering | String concatenation or manual formatting | text/template (already in use) | Handles injection safety, separation of concerns |
| CLI flag parsing | Manual os.Args parsing | cobra (already in use) | Handles validation, subcommands, help text |

**Key insight:** The codebase already has all the right abstractions. Chat format is an extension, not a replacement. Reuse existing patterns (SplitType enum, template system, writer package) rather than introducing new frameworks.

## Common Pitfalls

### Pitfall 1: Deduplication Key Mismatch

**What goes wrong:** Using the same deduplication key (`prompt|||completion`) for chat format causes all messages to be considered unique because the fields don't exist in the chat structure.

**Why it happens:** Deduplication logic currently uses `Pair.Prompt` and `Pair.Completion` fields directly. Chat format stores content in nested `messages` array.

**How to avoid:** Extract user message content from `messages` array and use that as the deduplication key. Consider creating a common interface or converter function.

**Warning signs:** Tests show every chat message being written even when identical user messages exist across documents.

### Pitfall 2: Template Instruction Ambiguity

**What goes wrong:** LLM generates user messages that reference "the document" or "this guide", breaking the context-free requirement.

**Why it happens:** Template doesn't explicitly forbid document references, so LLM defaults to natural phrasing like "What does the document say about X?"

**How to avoid:** Add explicit template instruction: "User messages must be context-free. Do NOT reference 'the document', 'the guide', 'this text', or similar phrases. Questions should stand alone."

**Warning signs:** Manual review of generated chat pairs shows questions like "What does the guide explain about X?"

### Pitfall 3: System Message Position Inconsistency

**What goes wrong:** System message appears at different positions in the messages array (sometimes first, sometimes after user message), breaking OpenAI format compliance.

**Why it happens:** Template doesn't enforce position or code doesn't validate message order during parsing.

**How to avoid:** Always insert system message at index 0 when constructing messages array. Validate parsed responses to ensure system message (if present) is first.

**Warning signs:** Training platforms reject JSONL files with malformed message order.

### Pitfall 4: Format Flag Default Value Breaking Backward Compatibility

**What goes wrong:** Changing default format from completion to chat breaks existing scripts and workflows that rely on implicit completion format.

**Why it happens:** Developer assumes chat should be default since it's "newer" or "better."

**How to avoid:** Keep `completion` as default format for backward compatibility. Require explicit `--format chat` flag to opt into new behavior.

**Warning signs:** Users report unexpected output format changes after upgrading datakeg version.

### Pitfall 5: Missing Voice/Style Matching in Responses

**What goes wrong:** Assistant responses are generic and formal, not matching the casual/technical/conversational style of the source document.

**Why it happens:** Template doesn't instruct LLM to analyze and mirror document style.

**How to avoid:** Add explicit template instruction: "Assistant responses must mirror the voice, tone, and style of the source document. If the document is casual, use casual language. If technical, use technical language. Match the author's conversational patterns."

**Warning signs:** Generated responses feel inconsistent with source material when manually reviewed.

## Code Examples

### Chat Message Structure (OpenAI-compatible)

```go
// In internal/writer/chat.go
type Message struct {
    Role    string `json:"role"`    // "system", "user", or "assistant"
    Content string `json:"content"` // Message text
}

type ChatMessage struct {
    Messages []Message `json:"messages"`
}

// WriteChatJSONL writes chat messages to JSONL file
func WriteChatJSONL(filename string, messages []ChatMessage) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("create file %s: %w", filename, err)
    }
    defer file.Close()

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
```

### Format Flag with Cobra

```go
// In cmd/datakeg/main.go
var flagFormat string

func init() {
    generateCmd.Flags().StringVarP(&flagFormat, "format", "f", "completion",
        "Output format (completion, chat)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
    // Validate format flag
    format, err := generator.ParseFormat(flagFormat)
    if err != nil {
        return err // Returns clear error: "invalid format: xyz (must be 'completion' or 'chat')"
    }

    // Pass format to pipeline
    return commands.ExecuteGeneratePipeline(
        sourcePath, outputPath, flagProvider, flagModel,
        format, // NEW parameter
        flagPairsPer1K, flagValidPct, flagTestPct, flagTimeout, flagSkipMerge, flagYes, flagDryRun,
    )
}
```

### Converting Pairs to Chat Messages

```go
// In internal/writer/chat.go
func ConvertPairToChatMessage(pair generator.Pair, includeSystemMsg bool, systemContent string) ChatMessage {
    messages := []Message{}

    if includeSystemMsg && systemContent != "" {
        messages = append(messages, Message{
            Role:    "system",
            Content: systemContent,
        })
    }

    messages = append(messages,
        Message{Role: "user", Content: pair.Prompt},
        Message{Role: "assistant", Content: pair.Completion},
    )

    return ChatMessage{Messages: messages}
}
```

### Chat Template Example (train)

```
You are generating high-quality chat training data for an LLM.

Given the documentation below, generate exactly {{.PairCount}} single-turn conversations.

CRITICAL RULES:
- Each conversation is ONE user message + ONE assistant response
- User messages must be context-free: do NOT reference "the document", "the guide", "this text", or similar phrases
- Questions should stand alone as if asked by someone unfamiliar with the source
- Assistant responses must mirror the voice, tone, dialect, and conversational style of the source document
- If the document is casual, responses should be casual. If formal, responses should be formal.
- If the document uses specific jargon or patterns, responses should use them too

Question styles (mix these):
- Factual: "What is X?", "How does Y work?"
- Clarifying: "Why would someone use Z?", "What's the difference between A and B?"
- How-to: "How do I configure X?", "What steps are needed for Y?"
- Conceptual: "What are the tradeoffs of X?", "When should I avoid Y?"

Output format:
Return ONLY a JSON array of conversation objects in this exact form:
[
  {
    "user": "<context-free question>",
    "assistant": "<response in document's voice>"
  }
]

{{if .ExcludePairs}}
IMPORTANT:
Do NOT generate questions semantically similar to these previously generated pairs.
Avoid overlapping intent, not just wording.

Previously generated:
{{range .ExcludePairs}}
- User: "{{.Prompt}}"
  Assistant: "{{.Completion}}"
{{end}}
{{end}}

Document:
{{.DocumentContent}}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Prompt/completion pairs only | Messages format with roles | 2023-2024 | OpenAI moved to chat completions API; ecosystem followed |
| Generic system messages | Task-specific system message generation (SysGen) | 2024-2025 | Auto-generated system messages improve alignment by 15-20% on benchmarks |
| Multi-turn as default | Single-turn + explicit multi-turn support | 2025+ | Single-turn trains better; multi-turn requires specialized handling |
| Manual deduplication | MinHash LSH at scale | 2024+ | Handles trillion-token scale; probabilistic approach reduces compute |

**Deprecated/outdated:**
- **Completion-only format for chat models:** OpenAI deprecated `text-davinci-003` completion endpoints; chat completions are the standard API
- **Generic "You are a helpful assistant" system messages:** Research shows task-specific system messages improve performance significantly (SysGen paper, late 2024)
- **Assuming multi-turn is always better:** Recent research (2025) shows single-turn trains more reliably; multi-turn introduces context dependencies that can hurt generalization

## Open Questions

1. **System message auto-generation approach**
   - What we know: SysGen-style annotation (extracting key functionalities, filtering, verification) works well for existing datasets
   - What's unclear: Whether to auto-generate from document content or require user-provided string via flag
   - Recommendation: Start with optional user-provided system message via `--system-message "text"` flag. Auto-generation can be Phase 6+ enhancement based on user feedback.

2. **Merge format detection strategy**
   - What we know: Parsing JSONL to detect format is fragile; explicit flag is clearer
   - What's unclear: Whether to warn/error when mixed formats exist in output directory
   - Recommendation: Detect format mismatch and warn user, but allow merge to proceed. Add `--force` flag to skip warning. Never auto-detect format; always require explicit `--format` flag or use default.

3. **Deduplication across formats**
   - What we know: Current deduplication uses "prompt|||completion" key; chat uses nested messages
   - What's unclear: Should deduplication compare across formats (e.g., completion prompt vs chat user message)?
   - Recommendation: Keep formats separate for Phase 5. Deduplication within format only. Cross-format deduplication is a low-priority enhancement.

4. **Template instruction explicitness**
   - What we know: Context explicitly requires style matching: "Assistant responses must mirror the voice, tone, dialect, and conversational style of the source document"
   - What's unclear: Whether to include this verbatim in template or rely on LLM inference
   - Recommendation: Include explicit style-matching instruction in chat templates. Testing during implementation will validate if LLM follows it or needs refinement.

## Sources

### Primary (HIGH confidence)

- [OpenAI Chat Completions API Reference](https://platform.openai.com/docs/api-reference/chat) - Standard messages format structure
- [OpenAI Fine-tuning JSONL Format](https://cookbook.openai.com/examples/how_to_finetune_chat_models) - Official chat format for training data
- [Go Official Documentation - encoding/json](https://go.dev/blog/json) - JSON encoding patterns and interfaces
- [Go Official Documentation - text/template](https://pkg.go.dev/text/template) - Template syntax and execution
- [Cobra Official Documentation](https://cobra.dev/docs/how-to-guides/working-with-flags/) - CLI flag handling patterns
- Codebase analysis: internal/generator/generator.go, internal/writer/jsonl.go, internal/templates/templates.go - Existing patterns for SplitType enum, writer structure, template system

### Secondary (MEDIUM confidence)

- [Fine-Tuning LLMs for Multi-Turn Conversations](https://www.together.ai/blog/fine-tuning-llms-for-multi-turn-conversations-a-technical-deep-dive) - Single-turn vs multi-turn training considerations
- [System Message Generation (SysGen) Research](https://arxiv.org/html/2502.11330v2) - Auto-generating task-specific system messages from dataset context
- [Optimizing LLM Training Data in 2026](https://www.aqusag.com/blog/aqusag-technologies-blog-5/optimizing-llm-training-data-in-2026-fine-tuning-rlhf-red-teaming-and-beyond-136) - 2026 training data best practices (quality over quantity, expert annotations)
- [Go Enum Patterns Best Practices](https://www.bytesizego.com/blog/mastering-enums-in-go) - Type-safe string enums with validation
- [Go Strategy Pattern for Formatters](https://rednafi.com/go/strategy-pattern/) - Interface-based formatting with function types

### Tertiary (LOW confidence)

- Web search results on deduplication techniques - General approach guidance; needs validation against datakeg's specific use case
- Web search results on chat templates - Community patterns; OpenAI official docs are authoritative

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use, well-documented stdlib + cobra
- Architecture: HIGH - Patterns match existing codebase (SplitType enum, template system, writer package)
- Pitfalls: MEDIUM - Based on common LLM training issues and code analysis; needs validation during implementation
- Chat format structure: HIGH - OpenAI official documentation is authoritative
- System message approach: MEDIUM - SysGen research is recent (2024-2025) and not yet widely adopted; user-provided approach is simpler and lower risk

**Research date:** 2026-02-07
**Valid until:** 2026-03-07 (30 days - stable domain, Go stdlib and chat format standards evolve slowly)
