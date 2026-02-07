# Phase 6: Reasoning Format Support - Research

**Researched:** 2026-02-07
**Domain:** Chain-of-thought reasoning model training data generation
**Confidence:** MEDIUM

## Summary

Reasoning model training data has emerged as a major focus in the LLM ecosystem through 2025-2026, driven by models like DeepSeek-R1, OpenAI o1, and Qwen3-Thinking. The standard approach uses `<think>...</think>` tags to separate step-by-step reasoning traces from final answers. While the ecosystem has converged on this tag-based structure, there are variations in exact schema design (separate fields vs integrated text, metadata inclusion) and reasoning style (visible vs hidden, verbose vs concise).

Document-grounded reasoning training data should focus on multi-step questions that demand analytical thinking (comparisons, implications, why/how questions) rather than simple factual recall. Quality matters more than quantity—research shows models can be fine-tuned on reasoning with as few as 1,000 high-quality examples.

**Primary recommendation:** Implement a `<think>...</think>` tag-based format with separate fields (reasoning + answer) as the default, matching the dominant DeepSeek-R1 pattern observed across open-source implementations in 2026.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `<think>` tags | N/A | Reasoning trace delimiter | Universal across DeepSeek-R1, Qwen3-Thinking, Kimi K2 |
| JSONL | N/A | Output format | Standard for LLM training pipelines |
| Go `encoding/json` | stdlib | Serialization | Already in use, handles struct tags well |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/go-playground/validator/v10` | Latest | Struct validation | If implementing validation of reasoning outputs |
| `text/template` | stdlib | Template rendering | Already in use for prompts |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `<think>` tags | Step arrays (JSON) | Tags are ecosystem standard; arrays harder to parse from LLM output |
| Separate fields | Integrated text | Separate fields enable easier processing but complicate parsing |
| Tag-based validation | Manual checks | Validator library adds dependency but provides compile-time safety |

**Installation:**
```bash
# Only if adding validation library
go get github.com/go-playground/validator/v10
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── generator/
│   ├── generator.go          # Add FormatReasoning to enum
│   ├── reasoning_parser.go   # New: parse reasoning responses
│   └── reasoning_validator.go # New: validate reasoning quality
├── templates/
│   └── prompts/
│       ├── reasoning_train.tmpl
│       ├── reasoning_valid.tmpl
│       └── reasoning_test.tmpl
└── writer/
    ├── jsonl.go              # Add reasoning output methods
    └── reasoning_types.go    # New: ReasoningPair struct
```

### Pattern 1: Separate Fields with Metadata (RECOMMENDED)
**What:** JSONL with distinct `reasoning`, `answer`, and optional `metadata` fields
**When to use:** Default format - most flexible, matches ecosystem standards
**Example:**
```go
// Reasoning pair with separate fields
type ReasoningPair struct {
    Question  string              `json:"question"`
    Reasoning string              `json:"reasoning"`
    Answer    string              `json:"answer"`
    Metadata  *ReasoningMetadata  `json:"metadata,omitempty"`
}

type ReasoningMetadata struct {
    Difficulty int    `json:"difficulty,omitempty"`
    Steps      int    `json:"steps,omitempty"`
}
```
**JSONL Output:**
```jsonl
{"question":"Why does X lead to Y in this system?","reasoning":"<think>First, I need to understand...</think>","answer":"X leads to Y because..."}
```

### Pattern 2: Integrated Text Format
**What:** Single text field with inline `<think>` tags
**When to use:** Alternative format for systems expecting continuous text
**Example:**
```go
type ReasoningTextPair struct {
    Prompt     string `json:"prompt"`
    Completion string `json:"completion"`
}
```
**JSONL Output:**
```jsonl
{"prompt":"Why does X lead to Y?","completion":"<think>First, I need to understand...</think>\n\nX leads to Y because..."}
```

### Pattern 3: Format-Specific Template Selection
**What:** Extend `getTemplateName()` pattern to include reasoning templates
**When to use:** Core implementation pattern
**Example:**
```go
// Source: datakeg existing generator.go pattern
func (g *Generator) getTemplateName(format FormatType, split SplitType) string {
    switch format {
    case FormatReasoning:
        switch split {
        case SplitTrain:
            return "reasoning_train.tmpl"
        case SplitValid:
            return "reasoning_valid.tmpl"
        case SplitTest:
            return "reasoning_test.tmpl"
        }
    // ... existing completion, chat cases
    }
}
```

### Pattern 4: Reasoning Response Parsing
**What:** Parse LLM responses containing `<think>` tags into separate fields
**When to use:** Core implementation for reasoning format
**Example:**
```go
func (g *Generator) parseReasoningResponse(response string) []ReasoningPair {
    // Extract JSON array similar to existing parseResponse()
    // Then split each pair's content into reasoning + answer
    var pairs []ReasoningPair

    // Find JSON array boundaries
    start := strings.Index(response, "[")
    end := strings.LastIndex(response, "]")

    if start != -1 && end != -1 {
        jsonStr := response[start : end+1]
        // Parse into intermediate struct, extract <think> tags
        // Similar to existing parseChatResponse() pattern
    }

    return pairs
}
```

### Anti-Patterns to Avoid
- **Storing reasoning in prompt field:** Keep reasoning separate from question - don't conflate input and reasoning process
- **Generating shallow questions:** Avoid questions answerable with simple lookup - demand multi-step thinking
- **Fixed verbosity:** Match reasoning depth to question complexity, don't force verbose reasoning on trivial questions
- **Unvalidated responses:** LLMs may skip `<think>` tags or provide incomplete reasoning - validate and handle gracefully

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Struct field validation | Custom if/else chains | `go-playground/validator` tags | Handle edge cases, compose validations, compile-time safety |
| Tag extraction from text | Regex parsing | `strings.Index()` + substring | Regex overkill for simple delimiters, harder to debug |
| JSON schema evolution | Breaking changes | `omitempty` + optional fields | Backward compatibility, graceful degradation |
| Reasoning quality filtering | Hard thresholds | Template instructions + LLM discretion | LLM better judges what needs reasoning than rule-based filters |

**Key insight:** The LLM generating the data is better positioned to determine whether a question warrants reasoning than any rule-based filter we could write. Trust the template instructions to guide quality.

## Common Pitfalls

### Pitfall 1: Training on Wrong Problem Types
**What goes wrong:** LLMs trained on math/code reasoning fail on open-ended questions
**Why it happens:** Models rewarded for correct outcomes on concrete problems (coding, math) don't learn good reasoning *processes* for subjective domains
**How to avoid:** Document-grounded approach naturally constrains to factual reasoning based on source material - no extrapolation beyond documentation
**Warning signs:** Questions asking "what do you think?" instead of "why does the document state X?"

### Pitfall 2: Pattern Matching Instead of Reasoning
**What goes wrong:** Models produce plausible-sounding reasoning chains with logical flaws
**Why it happens:** Training data focuses on correct answers, not valid reasoning steps - models learn to pattern-match surface features
**How to avoid:** Template should emphasize step-by-step logical connections, not just conclusion. Consider including self-correction examples
**Warning signs:** Reasoning traces that jump to conclusions, skip critical steps, or contain contradictions

### Pitfall 3: Sycophantic Reasoning
**What goes wrong:** Model generates reasoning that confirms user assumptions rather than challenging incorrect beliefs
**Why it happens:** LLMs trained to provide pleasing responses avoid challenging users
**How to avoid:** Document-grounded approach prevents this - reasoning must reflect source material, not user preference
**Warning signs:** Reasoning that always agrees with question framing rather than analyzing objectively

### Pitfall 4: Error Propagation in Multi-Step Reasoning
**What goes wrong:** Early incorrect step invalidates entire reasoning chain
**Why it happens:** CoT lacks mechanisms to re-evaluate earlier steps once mistake occurs
**How to avoid:** Consider including examples with self-correction ("Wait, that's not quite right..."). Accept that some generated pairs will be invalid - deduplication/validation catches these
**Warning signs:** Reasoning chains with early factual errors that aren't corrected

### Pitfall 5: Verbosity Without Value
**What goes wrong:** Reasoning traces become unnecessarily verbose, adding computation without improving accuracy
**Why it happens:** Models trained to "think more" apply verbose reasoning even when simple answer suffices
**How to avoid:** Template instruction to match reasoning depth to question complexity. Skip reasoning generation for trivial factual questions
**Warning signs:** Every question gets multi-paragraph reasoning trace regardless of complexity

### Pitfall 6: Incomplete Data Sets
**What goes wrong:** Models miss essential patterns and correlations
**Why it happens:** Manual curation is time-consuming, prone to errors
**How to avoid:** Synthetic generation from documentation provides diverse coverage. Template includes exclusion mechanism to avoid duplicates
**Warning signs:** All questions follow same pattern, no diversity in reasoning styles

## Code Examples

Verified patterns from official sources:

### Reasoning Format Detection
```go
// Detect if response contains reasoning tags
func containsReasoningTags(response string) bool {
    return strings.Contains(response, "<think>") && strings.Contains(response, "</think>")
}

// Extract reasoning and answer from response with tags
func extractReasoningAndAnswer(text string) (reasoning, answer string) {
    thinkStart := strings.Index(text, "<think>")
    thinkEnd := strings.Index(text, "</think>")

    if thinkStart != -1 && thinkEnd != -1 && thinkEnd > thinkStart {
        reasoning = text[thinkStart : thinkEnd+len("</think>")]
        answer = strings.TrimSpace(text[thinkEnd+len("</think>"):])
        return
    }

    // No tags found - treat entire text as answer
    return "", text
}
```

### Template Pattern for Reasoning Generation
```go
// Example template instruction structure
// Based on research of effective CoT prompts
const reasoningPromptPattern = `
You are generating chain-of-thought reasoning training data from documentation.

Generate exactly {{.PairCount}} question-answer pairs with step-by-step reasoning.

CRITICAL RULES:
- Questions must demand multi-step reasoning (Why? Compare? What if? How does X lead to Y?)
- Do NOT generate simple factual questions answerable by lookup
- Reasoning must be document-grounded - no extrapolation beyond source material
- Wrap reasoning in <think>...</think> tags
- Match reasoning depth to question complexity
- Skip generating pairs for trivial content that doesn't warrant reasoning

Output format:
[
  {
    "question": "<multi-step question requiring analysis>",
    "reasoning": "<think>Step 1: ... Step 2: ... Therefore...</think>",
    "answer": "<final answer derived from reasoning>"
  }
]

Document:
{{.DocumentContent}}
`
```

### Validation Pattern
```go
// Validate reasoning pair quality
func validateReasoningPair(p ReasoningPair) bool {
    // Basic non-empty checks
    if strings.TrimSpace(p.Question) == "" ||
       strings.TrimSpace(p.Answer) == "" {
        return false
    }

    // Reasoning field should contain think tags if populated
    if p.Reasoning != "" {
        if !strings.Contains(p.Reasoning, "<think>") ||
           !strings.Contains(p.Reasoning, "</think>") {
            return false
        }
    }

    return true
}
```

### Format Variant Support
```go
// Support multiple reasoning format variants
type ReasoningFormat string

const (
    ReasoningFormatSeparate  ReasoningFormat = "separate"   // Default: separate fields
    ReasoningFormatIntegrated ReasoningFormat = "integrated" // Inline tags in text
)

// Convert between formats
func convertToFormat(pairs []ReasoningPair, format ReasoningFormat) interface{} {
    switch format {
    case ReasoningFormatIntegrated:
        // Convert to single text field with inline tags
        var textPairs []ReasoningTextPair
        for _, p := range pairs {
            completion := p.Reasoning + "\n\n" + p.Answer
            textPairs = append(textPairs, ReasoningTextPair{
                Prompt:     p.Question,
                Completion: completion,
            })
        }
        return textPairs
    default:
        // Return as-is (separate fields)
        return pairs
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hidden reasoning tokens (o1-style) | Visible `<think>` tags (DeepSeek-R1) | Late 2025 | Transparency vs safety tradeoff - visible reasoning aids interpretability |
| Padding with empty pairs | LLM discretion on quantity | 2025-2026 | Quality over quantity - fewer but better reasoning pairs |
| Uniform verbosity | Complexity-matched reasoning | 2026 | Efficiency - avoid verbose reasoning on simple questions |
| Single reasoning style | Hybrid thinking modes | 2026 (Qwen3) | Flexibility - toggle between deep reasoning and rapid response |
| Reward correct answers only | Train on self-correction traces | 2025-2026 | Better reasoning processes, not just outcomes |

**Deprecated/outdated:**
- **Separate reasoning and chat formats:** Reasoning is now recognized as its own format type, not a chat variant
- **JSON Schema enforcement in thinking mode:** Qwen3 explicitly states thinking mode doesn't support structured output - must parse afterward
- **Forced system prompts for reasoning:** DeepSeek-R1 guidance recommends avoiding system prompts, containing all instructions in user prompt

## Open Questions

Things that couldn't be fully resolved:

1. **Self-correction in training data**
   - What we know: Research shows training on self-correction traces improves reasoning, but models trained this way learn to avoid major edits
   - What's unclear: Whether to include explicit backtracking in generated traces ("Wait, that's wrong...") or keep reasoning linear
   - Recommendation: Start without self-correction, add as enhancement if needed. Template can generate correction examples but not required initially

2. **Optimal reasoning/answer separation**
   - What we know: Ecosystem has examples of both separate fields and integrated text with tags
   - What's unclear: Which approach is preferred by fine-tuning pipelines in practice
   - Recommendation: Default to separate fields (more flexible), provide `--reasoning-format` flag to toggle. Can add integrated format as variant

3. **Metadata necessity**
   - What we know: Some datasets include difficulty, topic tags, step count metadata
   - What's unclear: Whether this metadata improves fine-tuning outcomes or is just informational
   - Recommendation: Start minimal (question, reasoning, answer only), add metadata fields as optional enhancement later

4. **Validation strictness**
   - What we know: Can validate for `<think>` tag presence, non-empty fields
   - What's unclear: Whether to enforce stricter validation (minimum reasoning length, step count) or trust LLM output
   - Recommendation: Lenient validation initially - just check tags and non-empty. Strict rules risk filtering valid edge cases

## Sources

### Primary (HIGH confidence)
- [DeepSeek-R1 GitHub](https://github.com/deepseek-ai/DeepSeek-R1) - Official format specification, `<think>` tag usage
- [Fine-tune DeepSeek-R1 with synthetic reasoning data](https://huggingface.co/blog/sdiazlor/fine-tune-deepseek-with-a-synthetic-reasoning-data) - JSONL schema example, temperature settings
- [General Inquiry Thinking Chain-of-Thought Dataset](https://huggingface.co/datasets/moremilk/General_Inquiry_Thinking-Chain-Of-Thought) - Schema with reasoning field structure

### Secondary (MEDIUM confidence)
- [How to teach chain of thought reasoning to your LLM](https://invisibletech.ai/blog/how-to-teach-chain-of-thought-reasoning-to-your-llm) - CoT data structure best practices
- [Fine-Tuning DeepSeek R1 (Reasoning Model) | DataCamp](https://www.datacamp.com/tutorial/fine-tuning-deepseek-r1-reasoning-model) - Training approaches with `<think>` tags
- [Top 10 Open-source Reasoning Models in 2026](https://www.clarifai.com/blog/top-10-open-source-reasoning-models-in-2026) - Ecosystem overview, format variants
- [Reasoning Models | OpenAI API](https://platform.openai.com/docs/guides/reasoning) - Reasoning tokens architecture
- [Qwen3 Structured Output with Thinking Mode](https://www.dataleadsfuture.com/build-autogen-agents-with-qwen3-structured-output-thinking-mode/) - Format limitations, tag behavior

### Secondary (MEDIUM confidence) - Training Data Quality
- [Chain-of-Thought Training Data | EmergentMind](https://www.emergentmind.com/topics/chain-of-thought-training-data) - Data generation methods, quality criteria
- [ThoughtSource: A central hub for large language model reasoning data](https://www.nature.com/articles/s41597-023-02433-3) - Chain-of-thought datasets, standardized formats
- [Build Custom Reasoning Models with Advanced, Open Post-Training Datasets | NVIDIA](https://developer.nvidia.com/blog/build-custom-reasoning-models-with-advanced-open-post-training-datasets/) - Dataset structure recommendations
- [Learning from Reasoning Failures via Synthetic Data Generation](https://arxiv.org/abs/2504.14523) - Synthetic data generation from LMM failures

### Secondary (MEDIUM confidence) - Common Pitfalls
- [AI's Reasoning Failures Can Impact Critical Fields | IEEE Spectrum](https://spectrum.ieee.org/ai-reasoning-failures) - Pattern-matching vs reasoning, training biases
- [Pitfalls of large language models in medical ethics reasoning | npj Digital Medicine](https://www.nature.com/articles/s41746-025-01792-y) - Sycophancy, cognitive biases, mistake rates
- [Verifying Chain-of-Thought Reasoning via Its Computational Graph](https://arxiv.org/html/2510.09312v1) - Error correction, silent failures
- [Training language models to self-correct](https://proceedings.iclr.cc/paper_files/paper/2025/file/871ac99fdc5282d0301934d23945ebaa-Paper-Conference.pdf) - Self-correction trace training

### Secondary (MEDIUM confidence) - Question Generation
- [Multi-Hop Reasoning Question Generation and Its Application | ResearchGate](https://www.researchgate.net/publication/350883543_Multi-hop_Reasoning_Question_Generation_and_Its_Application) - Multi-step question design
- [Reasoning in Trees: Improving Retrieval-Augmented Generation](https://arxiv.org/pdf/2601.11255) - Multi-hop QA patterns, error propagation
- [SimpleQA | OpenAI](https://openai.com/index/introducing-simpleqa/) - Reasoning models choosing not to attempt simple questions

### Tertiary (LOW confidence)
- [Structured Reasoning for Large Language Models](https://arxiv.org/abs/2601.07180) - Recent 2026 research on reasoning frameworks
- [JustRL: Scaling a 1.5B LLM with a Simple RL Recipe](https://iclr-blogposts.github.io/2026/blog/2026/justrl/) - Simplicity in training approaches

### Go Implementation References
- [Go JSON encoding documentation](https://pkg.go.dev/encoding/json) - Struct tag usage, marshaling
- [go-playground/validator GitHub](https://github.com/go-playground/validator) - Struct validation patterns
- [Making Go Struct Tags Work with JSON | Medium](https://medium.com/@AlexanderObregon/making-go-struct-tags-work-with-json-and-databases-7c698095b73a) - Multiple format tags

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - `<think>` tags are universal across DeepSeek-R1, Qwen3, verified in official docs
- Architecture: MEDIUM - Separate fields pattern verified in multiple datasets, but implementation details need testing
- Pitfalls: HIGH - Research-backed (IEEE, Nature publications) on reasoning failure modes
- Question generation: MEDIUM - Multi-step reasoning patterns verified but optimal mix unclear

**Research date:** 2026-02-07
**Valid until:** Approximately 30 days (reasoning model ecosystem stable but rapidly evolving)
