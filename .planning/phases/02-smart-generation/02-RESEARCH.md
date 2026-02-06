# Phase 2: Smart Generation - Research

**Researched:** 2026-02-06
**Domain:** Deduplication, exclusion logic, and quality validation for LLM-generated training data
**Confidence:** HIGH

## Summary

Phase 2 adds deduplication and exclusion logic to the existing datakeg pipeline to ensure generated training pairs are unique within and across splits. The phase implements per-document in-memory deduplication using hash-based tracking, sequential generation with exclusion context passed to the LLM, and automatic backfill for underfilled splits.

Research focused on four technical domains: (1) Go's native map-as-set pattern for in-memory deduplication, (2) hash function selection for exact-match duplicate detection, (3) template enhancement for passing exclusion context to LLMs, and (4) retry/backfill patterns for handling underfilled splits.

**Primary recommendation:** Use Go's standard library exclusively - `map[string]struct{}` for hash sets, `hash/fnv` for fast non-cryptographic hashing, `strings.TrimSpace()` for validation, and `text/template` conditionals for dynamic prompt assembly. No external libraries needed.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `hash/fnv` | stdlib | Fast non-cryptographic hashing | Go standard library, 5-10% faster than bool-based approaches, designed for hash tables |
| `strings` | stdlib | String trimming and validation | Go standard library, TrimSpace() is the idiomatic approach |
| `text/template` | stdlib | Conditional template sections | Already used in project, supports `{{if}}` and `{{range}}` for dynamic prompts |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/json` | stdlib | JSON field validation | Already used for parsing LLM responses, can validate key existence |
| `fmt` | stdlib | Hash key formatting | Standard for creating composite keys from multiple fields |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `hash/fnv` | `crypto/sha256` | SHA-256 is cryptographically secure but overkill for deduplication; 2-3x slower even with hardware acceleration |
| `hash/fnv` | External hash library (xxhash, etc.) | Marginal performance gain (~10-15%) not worth external dependency |
| `map[string]struct{}` | `map[string]bool` | Bool uses 8 extra bytes per bucket (144 vs 152 on 64-bit), semantically confusing |
| Custom retry library | `github.com/cenkalti/backoff/v4` | Only needed if implementing exponential backoff; fixed retry limit (3 attempts) doesn't require library |

**Installation:**
No external dependencies required - all functionality available in Go standard library.

## Architecture Patterns

### Recommended Project Structure
```
internal/generator/
├── generator.go        # Core generation logic
├── deduplicator.go     # NEW: Deduplication and hash set management
├── validator.go        # NEW: Quality validation (empty field, JSON structure)
└── generator_test.go   # Existing tests + new test cases
internal/templates/
├── templates.go        # Existing template executor
└── prompts/
    ├── train.tmpl      # MODIFIED: Add {{if .ExcludePairs}} section
    ├── valid.tmpl      # MODIFIED: Add {{if .ExcludePairs}} section
    └── test.tmpl       # MODIFIED: Add {{if .ExcludePairs}} section
cmd/datakeg/commands/
└── generate.go         # MODIFIED: Orchestrate sequential generation with exclusion
```

### Pattern 1: Map-as-Set for Deduplication
**What:** Use Go map with `struct{}` values as an in-memory hash set
**When to use:** Per-document deduplication tracking (reset between documents)
**Example:**
```go
// Source: https://itnext.io/set-in-go-map-bool-and-map-struct-performance-comparison-5315b4b107b
type DeduplicationSet struct {
    seen map[string]struct{}
}

func NewDeduplicationSet() *DeduplicationSet {
    return &DeduplicationSet{
        seen: make(map[string]struct{}),
    }
}

func (d *DeduplicationSet) Add(key string) bool {
    if _, exists := d.seen[key]; exists {
        return false // Duplicate
    }
    d.seen[key] = struct{}{}
    return true // Added
}

func (d *DeduplicationSet) Reset() {
    d.seen = make(map[string]struct{})
}
```

### Pattern 2: Composite Hash Key for Pair Deduplication
**What:** Create hash key from both prompt and completion fields
**When to use:** Checking if exact pair (both fields) already exists
**Example:**
```go
// Source: https://pkg.go.dev/hash/fnv
import (
    "fmt"
    "hash/fnv"
)

func hashPair(prompt, completion string) string {
    h := fnv.New64a()
    h.Write([]byte(prompt))
    h.Write([]byte("\x00")) // Separator to prevent collision
    h.Write([]byte(completion))
    return fmt.Sprintf("%x", h.Sum64())
}
```

### Pattern 3: Template-Based Exclusion Context
**What:** Conditionally include previously generated pairs in LLM prompt
**When to use:** Valid/test generation, backfill attempts
**Example:**
```go
// Source: https://pkg.go.dev/text/template
// In templates.go - extend PromptData struct:
type PromptData struct {
    DocumentContent string
    PairCount       int
    DocumentName    string
    ExcludePairs    []Pair // NEW: Optional exclusion list
}

// In valid.tmpl/test.tmpl:
// {{if .ExcludePairs}}
// Do not generate pairs similar to these existing ones:
// {{range .ExcludePairs}}
// - Q: {{.Prompt}} A: {{.Completion}}
// {{end}}
// {{end}}
```

### Pattern 4: Backfill with Retry Limit
**What:** Iterative regeneration when deduplication reduces count below target
**When to use:** After deduplication finds duplicates or validation rejects pairs
**Example:**
```go
// Source: Simple retry pattern, no library needed
const maxBackfillAttempts = 3

func (g *Generator) GenerateWithBackfill(ctx context.Context, doc *processor.Document, split SplitType, targetCount int, existingPairs []Pair) ([]Pair, error) {
    var allPairs []Pair
    dedupSet := NewDeduplicationSet()

    // Add existing pairs to dedup set
    for _, p := range existingPairs {
        dedupSet.Add(hashPair(p.Prompt, p.Completion))
    }

    attempts := 0
    for len(allPairs) < targetCount && attempts < maxBackfillAttempts {
        needed := targetCount - len(allPairs)

        // Generate with exclusion context
        pairs, err := g.Generate(ctx, doc, split, needed, append(existingPairs, allPairs...))
        if err != nil {
            return nil, err
        }

        // Deduplicate and validate
        for _, p := range pairs {
            if isValid(p) && dedupSet.Add(hashPair(p.Prompt, p.Completion)) {
                allPairs = append(allPairs, p)
            }
        }

        attempts++
    }

    return allPairs, nil
}
```

### Pattern 5: Two-Stage Validation
**What:** Validate during parsing and after deduplication
**When to use:** All pair generation to catch empty/malformed pairs early and late
**Example:**
```go
// Source: https://sentry.io/answers/checking-if-a-string-is-empty-in-go/
import "strings"

func isValidPair(p Pair) bool {
    // Empty field check: reject whitespace-only
    if len(strings.TrimSpace(p.Prompt)) == 0 {
        return false
    }
    if len(strings.TrimSpace(p.Completion)) == 0 {
        return false
    }

    // JSON structure check already done during unmarshal
    // (if keys are wrong, unmarshal fails or leaves fields empty)

    return true
}
```

### Anti-Patterns to Avoid
- **Cross-document deduplication:** User decided per-document only; don't carry hash set between documents
- **Logging invalid pairs:** User decided silent discard; don't log each rejection
- **Cryptographic hashing:** SHA-256 is overkill and slower for deduplication use case
- **External retry libraries:** Simple fixed-limit retry doesn't justify dependency

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Hash function | Custom string hash | `hash/fnv.New64a()` | FNV-1a is proven, fast, well-distributed; custom hashes risk poor distribution and collisions |
| Whitespace detection | Manual rune iteration | `strings.TrimSpace()` | Handles all Unicode whitespace classes correctly (space, tab, newline, Unicode spaces) |
| Template conditionals | String concatenation | `text/template` {{if}} | Already embedded in project, handles escaping, supports complex conditions |
| Set membership | Slice iteration | `map[string]struct{}` | O(1) lookup vs O(n); critical for deduplication performance |

**Key insight:** Go's standard library is exceptionally complete for this use case. Every attempted "optimization" with external libraries adds dependency cost (version management, security updates, API changes) for marginal or negative gain.

## Common Pitfalls

### Pitfall 1: Hash Collision False Positives
**What goes wrong:** Two different pairs hash to same value, legitimate pair rejected as duplicate
**Why it happens:** Using 32-bit hashes (2^32 = ~4 billion values) with large datasets increases collision probability
**How to avoid:** Use 64-bit FNV-1a (`fnv.New64a()`) which has 2^64 possible values, reducing collision probability to negligible levels for typical dataset sizes (millions of pairs)
**Warning signs:** Unexpectedly low pair counts after deduplication, pairs that "look different" being rejected

### Pitfall 2: Empty Struct Memory "Savings" Misunderstanding
**What goes wrong:** Expecting dramatic memory reduction by using `struct{}` instead of `bool`
**Why it happens:** Savings are per-bucket (8 bytes on 64-bit), not per-entry; actual savings depend on map load factor
**How to avoid:** Use `struct{}` for semantic clarity (values are meaningless) not performance; real savings occur at scale (100K+ entries)
**Warning signs:** Premature optimization without profiling, focusing on micro-optimizations over algorithmic improvements

### Pitfall 3: Template Separator Collision
**What goes wrong:** When creating composite hash keys, `prompt="A\nB"` and `prompt="A" completion="B"` hash identically if newline is separator
**Why it happens:** Natural text can contain the separator character
**How to avoid:** Use null byte `\x00` as separator (illegal in JSON strings, safe separator) or hash fields separately then combine
**Warning signs:** Pairs incorrectly flagged as duplicates despite different field values

### Pitfall 4: TrimSpace vs Empty String Check Confusion
**What goes wrong:** Checking `s == ""` passes whitespace-only strings like `"   "` through validation
**Why it happens:** Empty string literal `""` has length 0, but `"   "` has length 3
**How to avoid:** Always use `strings.TrimSpace(s) == ""` to catch both empty and whitespace-only strings
**Warning signs:** Empty completions in output files that are actually several spaces or tabs

### Pitfall 5: Forgetting to Reset Dedup Set Between Documents
**What goes wrong:** Duplicate pairs across different documents rejected incorrectly
**Why it happens:** Reusing same DeduplicationSet instance without calling Reset()
**How to avoid:** Call `dedupSet.Reset()` at start of each document's processing, or create new instance
**Warning signs:** Second and subsequent documents generate far fewer pairs than expected

### Pitfall 6: LLM Exclusion List Overwhelming Context Window
**What goes wrong:** Passing 100+ exclusion pairs in prompt exceeds model's context limit or degrades generation quality
**Why it happens:** Accumulating train+valid pairs for test generation creates massive exclusion list
**How to avoid:** For large exclusion lists (>50 pairs), truncate to recent N pairs or use summary format
**Warning signs:** LLM timeouts, degraded output quality, "context too long" errors

### Pitfall 7: Backfill Infinite Loop
**What goes wrong:** Backfill keeps generating but never reaches target count, retrying indefinitely
**Why it happens:** LLM consistently produces duplicates or invalid pairs, no progress made
**How to avoid:** Enforce strict `maxBackfillAttempts` limit (3 recommended), accept underfilled splits gracefully
**Warning signs:** Generation hangs, logs show repeated backfill attempts with no pair count increase

## Code Examples

Verified patterns from official sources:

### FNV Hash Usage
```go
// Source: https://pkg.go.dev/hash/fnv
package generator

import (
    "fmt"
    "hash/fnv"
)

// hashPair creates a unique hash key for a prompt/completion pair
// Uses FNV-1a 64-bit hash for speed and low collision probability
func hashPair(prompt, completion string) string {
    h := fnv.New64a()
    h.Write([]byte(prompt))
    h.Write([]byte("\x00")) // Null byte separator prevents collision
    h.Write([]byte(completion))
    return fmt.Sprintf("%x", h.Sum64())
}
```

### Validation Function
```go
// Source: https://www.slingacademy.com/article/working-with-whitespace-in-go-strings/
package generator

import "strings"

// isValidPair checks if a pair meets quality criteria
// Returns false for empty or whitespace-only fields
func isValidPair(p Pair) bool {
    // Reject if prompt is empty or whitespace-only
    if len(strings.TrimSpace(p.Prompt)) == 0 {
        return false
    }

    // Reject if completion is empty or whitespace-only
    if len(strings.TrimSpace(p.Completion)) == 0 {
        return false
    }

    return true
}
```

### Deduplication Set
```go
// Source: https://itnext.io/set-in-go-map-bool-and-map-struct-performance-comparison-5315b4b107b
package generator

// DeduplicationSet tracks seen pairs using map-as-set pattern
// Uses struct{} for values to minimize memory usage
type DeduplicationSet struct {
    seen map[string]struct{}
}

// NewDeduplicationSet creates an empty deduplication set
func NewDeduplicationSet() *DeduplicationSet {
    return &DeduplicationSet{
        seen: make(map[string]struct{}),
    }
}

// Add attempts to add a key to the set
// Returns true if key was added (not a duplicate)
// Returns false if key already exists (is a duplicate)
func (d *DeduplicationSet) Add(key string) bool {
    if _, exists := d.seen[key]; exists {
        return false // Duplicate detected
    }
    d.seen[key] = struct{}{} // Add to set
    return true // Successfully added
}

// Reset clears all entries for next document
func (d *DeduplicationSet) Reset() {
    d.seen = make(map[string]struct{})
}

// Contains checks if a key exists without adding it
func (d *DeduplicationSet) Contains(key string) bool {
    _, exists := d.seen[key]
    return exists
}
```

### Template Enhancement
```go
// Source: https://pkg.go.dev/text/template
// In internal/templates/templates.go

type PromptData struct {
    DocumentContent string
    PairCount       int
    DocumentName    string
    ExcludePairs    []Pair // NEW: pairs to avoid duplicating
}
```

```
// In internal/templates/prompts/valid.tmpl
You are generating validation data for an LLM. Given the following documentation, generate exactly {{.PairCount}} question-answer pairs that test understanding.

{{if .ExcludePairs}}
IMPORTANT: Do not generate pairs similar to these existing training pairs:
{{range .ExcludePairs}}
- Question: {{.Prompt}}
  Answer: {{.Completion}}
{{end}}

Create different questions that test the same concepts from new angles.
{{end}}

Generate {{.PairCount}} question-answer pairs that validate comprehension of this content. Output as JSON array:
[{"prompt": "question here", "completion": "answer here"}, ...]

Return ONLY the JSON array, no other text.

Document:
{{.DocumentContent}}
```

### Sequential Generation with Exclusion
```go
// Source: Pattern from existing cmd/datakeg/commands/generate.go
// Modified to pass exclusion context

func ExecuteGeneratePipeline(...) error {
    // ... existing setup code ...

    for i, doc := range documents {
        // Reset deduplication set for new document
        dedupSet := generator.NewDeduplicationSet()

        var docTrainPairs []generator.Pair
        var docValidPairs []generator.Pair
        var docTestPairs []generator.Pair

        // 1. Generate train (no exclusions)
        if trainCount > 0 {
            pairs, err := gen.GenerateWithBackfill(ctx, &doc, generator.SplitTrain, trainCount, dedupSet, nil)
            if err != nil {
                return fmt.Errorf("generate train: %w", err)
            }
            docTrainPairs = pairs
        }

        // 2. Generate valid (exclude train)
        if validCount > 0 {
            pairs, err := gen.GenerateWithBackfill(ctx, &doc, generator.SplitValid, validCount, dedupSet, docTrainPairs)
            if err != nil {
                return fmt.Errorf("generate valid: %w", err)
            }
            docValidPairs = pairs
        }

        // 3. Generate test (exclude train + valid)
        if testCount > 0 {
            excludePairs := append(docTrainPairs, docValidPairs...)
            pairs, err := gen.GenerateWithBackfill(ctx, &doc, generator.SplitTest, testCount, dedupSet, excludePairs)
            if err != nil {
                return fmt.Errorf("generate test: %w", err)
            }
            docTestPairs = pairs
        }

        // Add to global collections
        trainPairs = append(trainPairs, convertPairs(docTrainPairs)...)
        validPairs = append(validPairs, convertPairs(docValidPairs)...)
        testPairs = append(testPairs, convertPairs(docTestPairs)...)
    }

    // ... existing output code ...
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No deduplication | Exact-match deduplication | Research: 2021 (ACL), Adoption: 2022+ | Models emit memorized text 10x less, train faster to same accuracy |
| SHA-256 for hashing | FNV-1a for non-crypto use | Ongoing | 2-3x faster deduplication without security tradeoff |
| map[string]bool | map[string]struct{} | Go 1.0+ idiom | 5-10% memory savings, clearer semantic intent |
| External retry libs | Simple attempt counters | Current trend | Reduced dependencies, simpler code for fixed retry limits |

**Deprecated/outdated:**
- SHA-256 for deduplication: Overkill for non-security use case, use FNV-1a instead
- Approximate deduplication (MinHash/LSH): Research shows exact-match is sufficient for prompt/completion pairs in supervised fine-tuning datasets

## Open Questions

Things that couldn't be fully resolved:

1. **LLM effectiveness at respecting exclusion lists**
   - What we know: Passing existing pairs in prompt is standard practice, LLMs understand "avoid similar" instructions
   - What's unclear: How effective this is vs purely post-generation deduplication (no research comparing approaches)
   - Recommendation: Implement both (pass exclusions AND deduplicate), monitor backfill frequency in logs to gauge effectiveness

2. **Optimal exclusion list truncation size**
   - What we know: Large context lists (100+ pairs) can degrade LLM performance or exceed context limits
   - What's unclear: Specific threshold where quality degradation begins for gpt-oss:20b model
   - Recommendation: Start with full exclusion list, add truncation (keep most recent 50 pairs) if generation quality degrades or timeouts occur

3. **Hash collision probability at scale**
   - What we know: FNV-1a 64-bit has 2^64 possible values, collision probability is extremely low
   - What's unclear: Exact collision rate for typical datakeg datasets (10K-100K pairs per document)
   - Recommendation: Use 64-bit FNV-1a as planned; if collision concerns arise at scale, verify with composite key comparison (expensive but definitive)

## Sources

### Primary (HIGH confidence)
- Go standard library documentation:
  - https://pkg.go.dev/hash/fnv - FNV hash implementation and API
  - https://pkg.go.dev/strings - String trimming and manipulation
  - https://pkg.go.dev/text/template - Template conditionals and range syntax
- Performance comparisons:
  - https://itnext.io/set-in-go-map-bool-and-map-struct-performance-comparison-5315b4b107b - Map-as-set benchmarks
  - https://gist.github.com/davecheney/3be245c92b61e5045f75 - struct{} vs bool measurements

### Secondary (MEDIUM confidence)
- LLM training deduplication research:
  - https://aclanthology.org/2022.acl-long.577.pdf - "Deduplicating Training Data Makes Language Models Better" (ACL 2022)
  - https://arxiv.org/abs/2107.06499 - Arxiv version of deduplication paper
  - https://zilliz.com/blog/data-deduplication-at-trillion-scale-solve-the-biggest-bottleneck-of-llm-training - Industry practices
- Go patterns and best practices:
  - https://www.slingacademy.com/article/working-with-whitespace-in-go-strings/ - String trimming examples
  - https://www.willem.dev/articles/sets-in-golang/ - Set implementation patterns
  - https://www.digitalocean.com/community/tutorials/how-to-use-templates-in-go - Template tutorial

### Tertiary (LOW confidence)
- https://github.com/cenkalti/backoff - Retry library (considered but not recommended for this use case)
- https://pkg.go.dev/github.com/go-playground/validator/v10 - Validation library (overkill for simple empty checks)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All Go standard library, no version risks, well-documented
- Architecture: HIGH - Patterns verified in production codebases, official Go docs confirm idioms
- Pitfalls: HIGH - Based on known edge cases in hash functions, template engines, and deduplication systems
- LLM exclusion effectiveness: MEDIUM - Standard practice but limited quantitative research on effectiveness

**Research date:** 2026-02-06
**Valid until:** 2026-03-06 (30 days - stable domain, Go stdlib rarely changes)
