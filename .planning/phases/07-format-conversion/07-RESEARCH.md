# Phase 7: Format Conversion - Research

**Researched:** 2026-02-07
**Domain:** JSONL format conversion for model-specific training pipelines
**Confidence:** HIGH

## Summary

Format conversion transforms datakeg's generated JSONL files (completion, chat, reasoning) into model-specific training formats using customizable templates. The standard approach uses Go's text/template for transformation rules, embedded templates for built-in formats, and streaming line-by-line JSONL processing to handle large datasets efficiently. Common target formats include Mistral Instruct (`[INST]...[/INST]`), Llama 3 (`<|start_header_id|>...<|end_header_id|>`), and ChatML (`<|im_start|>...<|im_end|>`).

The key architectural pattern is a template-based converter that reads source JSONL line-by-line, unmarshals into the appropriate source struct (TrainingPair, ChatMessage, or ReasoningPair), executes a conversion template with that data, and writes the result to output JSONL. Templates define field mappings and special token insertion. Built-in templates ship embedded in the binary via `//go:embed`, while users can provide custom templates via file path for proprietary formats.

Implementation follows datakeg's existing patterns: templates in `internal/templates/conversions/`, writer functions in `internal/writer/`, and CLI command in `cmd/datakeg/commands/convert.go`. The converter must auto-detect source format from JSONL structure and validate template compatibility (e.g., reasoning templates require ReasoningPair source).

**Primary recommendation:** Implement streaming JSONL converter with text/template engine, embed common format templates, support custom template files, and provide clear error messages for template/format mismatches.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `text/template` | stdlib | Template execution | Already in use, Go standard for text transformation |
| `encoding/json` | stdlib | JSONL parsing/encoding | Already in use, handles line-by-line streaming with Decoder |
| `embed` | stdlib | Embed templates | Already in use for prompt templates, zero-dependency asset bundling |
| `bufio.Scanner` | stdlib | Line-by-line reading | Standard Go pattern for streaming text files |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `os` | stdlib | File I/O | Reading template files, creating output |
| `filepath` | stdlib | Path handling | Template file resolution, output naming |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `text/template` | String manipulation | Templates are declarative, reusable, safer than string concatenation |
| Line-by-line streaming | Load entire file | Streaming handles GB-sized JSONL; full-load causes OOM on large datasets |
| Embedded templates | External config files | Embedded templates ship in binary, zero-config for common formats |
| Format auto-detection | Require `--source-format` flag | Auto-detection is user-friendly; explicit flag is more robust |

**Installation:**
No additional dependencies required beyond existing datakeg stack.

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── converter/
│   ├── converter.go           # New: Core conversion logic
│   ├── format_detector.go     # New: Auto-detect source format
│   └── template_validator.go  # New: Validate template compatibility
├── templates/
│   ├── conversions/           # New: Conversion templates
│   │   ├── mistral-instruct.tmpl
│   │   ├── llama3-instruct.tmpl
│   │   ├── chatml.tmpl
│   │   └── deepseek-r1.tmpl
│   └── templates.go           # Extend: Add conversion template loader
├── writer/
│   └── jsonl.go               # Existing: WriteJSONLAppend reused
cmd/datakeg/commands/
└── convert.go                 # New: CLI command for conversion
```

### Pattern 1: Streaming JSONL Converter (CORE)
**What:** Read JSONL line-by-line, unmarshal to struct, execute template, write output line
**When to use:** All format conversions - handles datasets of any size
**Example:**
```go
// Core conversion loop
func ConvertJSONL(inputPath, outputPath string, tmpl *template.Template, sourceFormat FormatType) error {
    inFile, err := os.Open(inputPath)
    if err != nil {
        return fmt.Errorf("open input: %w", err)
    }
    defer inFile.Close()

    outFile, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("create output: %w", err)
    }
    defer outFile.Close()

    decoder := json.NewDecoder(inFile)
    encoder := json.NewEncoder(outFile)

    lineNum := 0
    for decoder.More() {
        lineNum++

        // Unmarshal based on source format
        var data interface{}
        switch sourceFormat {
        case FormatCompletion:
            var pair writer.TrainingPair
            if err := decoder.Decode(&pair); err != nil {
                return fmt.Errorf("line %d: decode: %w", lineNum, err)
            }
            data = pair
        case FormatChat:
            var msg writer.ChatMessage
            if err := decoder.Decode(&msg); err != nil {
                return fmt.Errorf("line %d: decode: %w", lineNum, err)
            }
            data = msg
        case FormatReasoning:
            var rp writer.ReasoningPair
            if err := decoder.Decode(&rp); err != nil {
                return fmt.Errorf("line %d: decode: %w", lineNum, err)
            }
            data = rp
        }

        // Execute template
        var buf strings.Builder
        if err := tmpl.Execute(&buf, data); err != nil {
            return fmt.Errorf("line %d: template: %w", lineNum, err)
        }

        // Parse template output and write as JSON
        var result interface{}
        if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
            return fmt.Errorf("line %d: template output not valid JSON: %w", lineNum, err)
        }

        if err := encoder.Encode(result); err != nil {
            return fmt.Errorf("line %d: encode output: %w", lineNum, err)
        }
    }

    return nil
}
```

### Pattern 2: Format Auto-Detection
**What:** Detect source format by parsing first JSONL line and checking field structure
**When to use:** User doesn't specify `--source-format` flag
**Example:**
```go
// Detect format from JSONL structure
func DetectFormat(jsonLine []byte) (FormatType, error) {
    var raw map[string]interface{}
    if err := json.Unmarshal(jsonLine, &raw); err != nil {
        return "", fmt.Errorf("invalid JSON: %w", err)
    }

    // Check for chat format: {"messages": [...]}
    if messages, ok := raw["messages"].([]interface{}); ok && len(messages) > 0 {
        return FormatChat, nil
    }

    // Check for reasoning format: {"question": "...", "reasoning": "...", "answer": "..."}
    if _, hasQuestion := raw["question"]; hasQuestion {
        if _, hasReasoning := raw["reasoning"]; hasReasoning {
            if _, hasAnswer := raw["answer"]; hasAnswer {
                return FormatReasoning, nil
            }
        }
    }

    // Check for completion format: {"prompt": "...", "completion": "..."}
    if _, hasPrompt := raw["prompt"]; hasPrompt {
        if _, hasCompletion := raw["completion"]; hasCompletion {
            return FormatCompletion, nil
        }
    }

    return "", fmt.Errorf("unknown format: fields %v", keys(raw))
}
```

### Pattern 3: Embedded Template Loader
**What:** Extend existing template system to load conversion templates from embedded filesystem
**When to use:** Loading built-in format templates
**Example:**
```go
// In internal/templates/templates.go
//go:embed conversions/*.tmpl
var conversionFS embed.FS

// LoadConversionTemplate loads a built-in conversion template by name
func LoadConversionTemplate(name string) (*template.Template, error) {
    tmpl, err := template.ParseFS(conversionFS, "conversions/"+name+".tmpl")
    if err != nil {
        return nil, fmt.Errorf("parse embedded template %s: %w", name, err)
    }
    return tmpl, nil
}

// LoadCustomConversionTemplate loads a user-provided template from file
func LoadCustomConversionTemplate(filePath string) (*template.Template, error) {
    tmpl, err := template.ParseFiles(filePath)
    if err != nil {
        return nil, fmt.Errorf("parse template file %s: %w", filePath, err)
    }
    return tmpl, nil
}
```

### Pattern 4: Template Validation
**What:** Verify template is compatible with source format before processing
**When to use:** After loading template, before conversion starts
**Example:**
```go
// Validate template produces valid JSON output
func ValidateTemplate(tmpl *template.Template, sourceFormat FormatType) error {
    // Create sample data based on format
    var testData interface{}
    switch sourceFormat {
    case FormatCompletion:
        testData = writer.TrainingPair{Prompt: "test", Completion: "test"}
    case FormatChat:
        testData = writer.ChatMessage{
            Messages: []writer.Message{
                {Role: "user", Content: "test"},
                {Role: "assistant", Content: "test"},
            },
        }
    case FormatReasoning:
        testData = writer.ReasoningPair{
            Question: "test",
            Reasoning: "<think>test</think>",
            Answer: "test",
        }
    }

    // Execute template with test data
    var buf strings.Builder
    if err := tmpl.Execute(&buf, testData); err != nil {
        return fmt.Errorf("template execution failed: %w", err)
    }

    // Verify output is valid JSON
    var result interface{}
    if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
        return fmt.Errorf("template output is not valid JSON: %w\nOutput: %s", err, buf.String())
    }

    return nil
}
```

### Anti-Patterns to Avoid
- **Loading entire JSONL into memory:** Use streaming decoder, not ioutil.ReadFile + json.Unmarshal of array
- **Hardcoding format mappings:** Use templates for all conversions, even simple ones, to maintain consistency
- **Ignoring template errors:** Template execution errors indicate malformed templates; fail fast with clear messages
- **Assuming single-line JSON:** JSONL must have one JSON object per line; don't try to parse pretty-printed multi-line JSON

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Text transformation rules | String concatenation/sprintf | text/template | Declarative, reusable, injection-safe, user-customizable |
| Streaming large files | bufio.Reader + manual parsing | json.Decoder.More() loop | Built-in JSONL streaming, handles edge cases (trailing newlines, empty lines) |
| Template bundling | config files + install script | //go:embed | Zero-config deployment, templates versioned with binary |
| Format detection heuristics | Custom logic for each format | JSON field presence checks | Simple, reliable, extensible to new formats |

**Key insight:** Go's stdlib provides everything needed for JSONL conversion. Don't add dependencies for problems that encoding/json, text/template, and embed already solve elegantly.

## Common Pitfalls

### Pitfall 1: Memory Exhaustion on Large Files
**What goes wrong:** Loading multi-GB JSONL files into memory causes OOM crashes
**Why it happens:** Using json.Unmarshal on entire file instead of streaming with json.Decoder
**How to avoid:** Always use json.NewDecoder(file) with decoder.More() loop for line-by-line processing
**Warning signs:** Memory usage grows proportional to input file size, crashes on files >500MB

### Pitfall 2: Template Output Not Valid JSON
**What goes wrong:** Template executes successfully but produces malformed JSON, causing encoding errors
**Why it happens:** Template syntax errors (missing quotes, unclosed braces) or unescaped field content
**How to avoid:**
- Use template validation before processing
- Escape JSON values with `{{.Field | js}}` or marshal to JSON explicitly
- Test templates with edge cases (quotes in content, newlines, Unicode)
**Warning signs:** "invalid character" errors during encoding, garbage in output JSONL

### Pitfall 3: Format Mismatch Between Source and Template
**What goes wrong:** Template expects ReasoningPair fields but source is ChatMessage format
**Why it happens:** User specifies wrong template for their data or auto-detection fails
**How to avoid:**
- Validate template compatibility with source format before conversion
- Provide clear error: "Template 'deepseek-r1' requires reasoning format, but source is chat format"
- List compatible formats in template metadata or filename convention
**Warning signs:** "field not found" template execution errors, nil pointer panics

### Pitfall 4: Character Encoding Issues
**What goes wrong:** Non-ASCII characters become garbled ("café" → "cafÃ©") or conversion fails
**Why it happens:** Not handling UTF-8 encoding consistently throughout pipeline
**How to avoid:**
- Go's encoding/json handles UTF-8 by default
- Ensure input/output files are opened as UTF-8 (default in Go)
- Don't use io.Reader wrappers that change encoding
**Warning signs:** Corrupted Unicode in output, mojibake characters, "invalid UTF-8" errors

### Pitfall 5: Incomplete Error Context
**What goes wrong:** Conversion fails with generic "template error" but user can't identify which line or field caused the issue
**Why it happens:** Not tracking line numbers or including data context in error messages
**How to avoid:**
- Include line number in all decode/template/encode errors
- Log problematic JSON line content (truncated if long)
- Provide actionable guidance: "Check template field references match source format"
**Warning signs:** Users report "conversion failed" but can't debug without reading source code

### Pitfall 6: Special Token Escaping
**What goes wrong:** Model-specific special tokens get double-escaped or mangled (e.g., `<|im_start|>` becomes `&lt;|im_start|&gt;`)
**Why it happens:** Using html/template instead of text/template, or escaping tokens as JSON strings when they should be literal
**How to avoid:**
- Use text/template, not html/template (which auto-escapes HTML entities)
- Keep special tokens outside JSON string values in template output
- Test templates with actual model tokenizers to verify token preservation
**Warning signs:** Model training fails to recognize special tokens, tokenization errors during fine-tuning

## Code Examples

### Mistral Instruct Conversion Template
```go
// File: internal/templates/conversions/mistral-instruct.tmpl
// Converts completion format to Mistral Instruct format
// Input: writer.TrainingPair{Prompt: "...", Completion: "..."}
// Output: {"text": "<s>[INST] ... [/INST] ... </s>"}
{
  "text": "<s>[INST] {{.Prompt}} [/INST] {{.Completion}}</s>"
}
```

### Llama 3 Instruct Conversion Template
```go
// File: internal/templates/conversions/llama3-instruct.tmpl
// Converts chat format to Llama 3 format
// Input: writer.ChatMessage{Messages: [...]}
// Output: {"text": "<|begin_of_text|><|start_header_id|>..."}
{
  "text": "<|begin_of_text|>{{range .Messages}}<|start_header_id|>{{.Role}}<|end_header_id|>\n\n{{.Content}}<|eot_id|>{{end}}"
}
```

### ChatML Conversion Template
```go
// File: internal/templates/conversions/chatml.tmpl
// Converts chat format to ChatML format
// Input: writer.ChatMessage{Messages: [...]}
// Output: {"text": "<|im_start|>user\n...<|im_end|>..."}
{
  "text": "{{range .Messages}}<|im_start|>{{.Role}}\n{{.Content}}<|im_end|>\n{{end}}"
}
```

### DeepSeek-R1 Reasoning Conversion Template
```go
// File: internal/templates/conversions/deepseek-r1.tmpl
// Converts reasoning format (separate fields) to DeepSeek-R1 format (integrated)
// Input: writer.ReasoningPair{Question: "...", Reasoning: "<think>...</think>", Answer: "..."}
// Output: {"prompt": "...", "completion": "<think>...</think>\n\n..."}
{
  "prompt": "{{.Question}}",
  "completion": "{{.Reasoning}}\n\n{{.Answer}}"
}
```

### CLI Command Pattern
```go
// In cmd/datakeg/commands/convert.go
func ExecuteConvertPipeline(inputPath, outputPath, templateName, customTemplatePath string, sourceFormat string) error {
    fmt.Printf("Converting %s...\n", inputPath)

    // Load template (built-in or custom)
    var tmpl *template.Template
    var err error
    if customTemplatePath != "" {
        tmpl, err = templates.LoadCustomConversionTemplate(customTemplatePath)
        if err != nil {
            return fmt.Errorf("load custom template: %w", err)
        }
        fmt.Printf("Using custom template: %s\n", customTemplatePath)
    } else {
        tmpl, err = templates.LoadConversionTemplate(templateName)
        if err != nil {
            return fmt.Errorf("load template '%s': %w. Run 'datakeg convert --list-templates' to see available templates", templateName, err)
        }
        fmt.Printf("Using built-in template: %s\n", templateName)
    }

    // Auto-detect source format if not specified
    var format FormatType
    if sourceFormat != "" {
        format, err = generator.ParseFormat(sourceFormat)
        if err != nil {
            return err
        }
    } else {
        format, err = converter.DetectFormatFromFile(inputPath)
        if err != nil {
            return fmt.Errorf("could not detect format: %w. Specify --source-format explicitly", err)
        }
        fmt.Printf("Detected source format: %s\n", format)
    }

    // Validate template compatibility
    if err := converter.ValidateTemplate(tmpl, format); err != nil {
        return fmt.Errorf("template validation failed: %w", err)
    }

    // Convert
    if err := converter.ConvertJSONL(inputPath, outputPath, tmpl, format); err != nil {
        return fmt.Errorf("conversion failed: %w", err)
    }

    fmt.Printf("Conversion complete: %s\n", outputPath)
    return nil
}
```

### List Built-in Templates
```go
// Helper to list available built-in templates
func ListBuiltinTemplates() error {
    templates := []struct {
        Name   string
        Format string
        Desc   string
    }{
        {"mistral-instruct", "completion", "Mistral Instruct format with [INST] tags"},
        {"llama3-instruct", "chat", "Llama 3 Instruct with header tags"},
        {"chatml", "chat", "ChatML format with <|im_start|> tags"},
        {"deepseek-r1", "reasoning", "DeepSeek-R1 integrated reasoning format"},
    }

    fmt.Println("Built-in conversion templates:")
    fmt.Println()
    for _, t := range templates {
        fmt.Printf("  %-20s (%-10s) - %s\n", t.Name, t.Format, t.Desc)
    }
    fmt.Println()
    fmt.Println("Usage: datakeg convert --template <name> <input.jsonl> <output.jsonl>")
    fmt.Println("Custom: datakeg convert --custom-template <file.tmpl> <input.jsonl> <output.jsonl>")

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Format-specific scripts | Template-based converters | 2024-2025 | Single tool for all formats vs scattered conversion scripts |
| Config files for formats | Embedded templates | 2025+ (embed directive) | Zero-config deployment, templates versioned with code |
| Load entire file | Streaming JSONL | Always best practice | Handle datasets >RAM, production-ready for TB-scale data |
| Manual field mapping | Declarative templates | Always preferred | Templates are documentation, easier to audit/customize |

**Deprecated/outdated:**
- **Python conversion scripts:** Go offers better performance and single-binary deployment for CLI tools
- **Pretty-printed JSON:** JSONL (one object per line) is standard for LLM training pipelines, not arrays
- **Format-specific tools:** Universal converters with templates replace per-format utilities

## Open Questions

Things that couldn't be fully resolved:

1. **Template versioning/compatibility**
   - What we know: Model formats evolve (e.g., Llama 3.1 vs 3.2 template changes)
   - What's unclear: Whether to version templates or always use "latest" format
   - Recommendation: Start with single template per format, add versioning (mistral-instruct-v1, mistral-instruct-v2) if ecosystem fragmentation requires it

2. **Multi-file batch conversion**
   - What we know: Users may want to convert train.jsonl, valid.jsonl, test.jsonl in one command
   - What's unclear: Whether to support glob patterns, directory input, or require separate invocations
   - Recommendation: Start with single-file conversion (simple, clear). Add batch mode as enhancement if users request it

3. **Template output format flexibility**
   - What we know: Templates currently output JSON objects (one per line)
   - What's unclear: Whether some formats need plain text output (no JSON wrapper) or multi-line entries
   - Recommendation: Require templates to produce JSON output (maintains JSONL standard). If plain text needed, users can post-process with jq or similar tools

4. **Error recovery strategies**
   - What we know: Single malformed line can stop conversion
   - What's unclear: Whether to skip invalid lines with warning or fail fast
   - Recommendation: Fail fast by default (data integrity). Add `--skip-errors` flag for lenient mode that logs warnings and continues

## Sources

### Primary (HIGH confidence)
- [Go text/template documentation](https://pkg.go.dev/text/template) - Template syntax, execution, functions
- [Go encoding/json documentation](https://pkg.go.dev/encoding/json) - Decoder.More() streaming pattern
- [Go embed documentation](https://pkg.go.dev/embed) - Template embedding patterns
- [Mistral Instruct format specification](https://huggingface.co/mistralai/Mistral-7B-Instruct-v0.1) - Official [INST] tag format
- [Llama 3 Model Cards and Prompt Formats](https://www.llama.com/docs/model-cards-and-prompt-formats/meta-llama-3/) - Official Llama 3 format spec
- Codebase analysis: internal/templates/templates.go (existing embed pattern), internal/writer/jsonl.go (JSONL handling), cmd/datakeg/commands/merge.go (line-by-line file processing)

### Secondary (MEDIUM confidence)
- [ChatML format specification](https://www.shshell.com/blog/fine-tuning-module-6-lesson-1-formats) - ChatML structure with im_start/im_end tags
- [OpenAI ChatML usage](https://medium.com/@gireeshm/open-source-instruction-tuning-for-openais-chatml-dfbef62057e9) - ChatML as industry standard
- [DeepSeek-R1 GitHub](https://github.com/deepseek-ai/DeepSeek-R1) - Reasoning format with <think> tags
- [JSONL streaming processing in Go](https://itnext.io/parsing-18-billion-lines-json-with-go-738be6ee5ed2) - Best practices for handling large JSONL files
- [Go Template examples](https://github.com/phcollignon/Go-Template) - Template-driven code generation patterns
- [LLM Dataset Formats](https://huggingface.co/blog/tegridydev/llm-dataset-formats-101-hugging-face) - Common training data format pitfalls

### Tertiary (LOW confidence)
- WebSearch results on format conversion approaches - General guidance, needs validation with official specs
- WebSearch results on character encoding issues - Common pitfalls applicable to any text processing

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All stdlib, well-documented, already in use in datakeg
- Architecture: HIGH - Patterns match existing codebase (templates, writer, CLI commands)
- Pitfalls: HIGH - Based on Go best practices, JSONL processing patterns, verified in production use
- Format specifications: HIGH - Official documentation from Mistral, Meta, OpenAI
- Template examples: MEDIUM - Based on format specs but need validation with actual model training

**Research date:** 2026-02-07
**Valid until:** 2026-04-07 (60 days - format specs stable, Go stdlib stable, low churn domain)
