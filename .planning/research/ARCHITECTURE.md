# Architecture Research

**Domain:** Go CLI tools for LLM-powered document processing
**Researched:** 2026-02-05
**Confidence:** MEDIUM (based on Go standard patterns and CLI best practices; verification sources unavailable)

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Layer                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  Flags   │  │  Args    │  │ Progress │  │  Output  │        │
│  │  Parser  │  │  Parser  │  │ Reporter │  │ Formatter│        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
│       │             │              │             │               │
├───────┴─────────────┴──────────────┴─────────────┴───────────────┤
│                     Orchestration Layer                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Pipeline Orchestrator                  │   │
│  │  (coordinates phases: read → process → dedupe → merge)    │   │
│  └──┬────────────────────┬────────────────────┬─────────────┘   │
│     │                    │                    │                  │
├─────┴────────────────────┴────────────────────┴──────────────────┤
│                      Processing Layer                            │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐          │
│  │  Document   │  │     LLM      │  │  Deduplicator │          │
│  │   Reader    │  │   Processor  │  │   + Filler    │          │
│  └──────┬──────┘  └──────┬───────┘  └───────┬───────┘          │
│         │                │                   │                   │
├─────────┴────────────────┴───────────────────┴───────────────────┤
│                      Integration Layer                           │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐          │
│  │ File System │  │    Ollama    │  │     JSONL     │          │
│  │   Reader    │  │    Client    │  │    Writer     │          │
│  └─────────────┘  └──────────────┘  └───────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **CLI Parser** | Parse flags, validate inputs, display help | cobra or stdlib flag package |
| **Progress Reporter** | Track and display progress of long-running operations | Custom goroutine writing to stderr |
| **Pipeline Orchestrator** | Coordinate execution phases, manage concurrency | Main workflow with errgroup for parallelism |
| **Document Reader** | Discover and read source files from directory | filepath.Walk or fs.WalkDir + ioutil.ReadFile |
| **LLM Processor** | Call Ollama API, manage retries, handle rate limiting | HTTP client with retry logic, worker pool pattern |
| **Deduplicator** | Hash-based exact match detection across sets | Map[string]bool with content hashing |
| **Filler** | Regenerate pairs when dedup reduces count below target | Re-invoke LLM processor with tracking |
| **JSONL Writer** | Write individual and merged output files | Buffered writer with JSON encoder |
| **Template Manager** | Embed and interpolate prompt templates | embed.FS + text/template |

## Recommended Project Structure

```
datakeg/
├── cmd/
│   └── datakeg/
│       └── main.go           # Entry point, CLI setup
├── internal/
│   ├── cli/
│   │   ├── flags.go          # Flag definitions and parsing
│   │   ├── progress.go       # Progress reporting
│   │   └── output.go         # Output formatting
│   ├── pipeline/
│   │   ├── orchestrator.go   # Main pipeline coordination
│   │   ├── phases.go         # Individual phase implementations
│   │   └── config.go         # Pipeline configuration
│   ├── document/
│   │   ├── reader.go         # File discovery and reading
│   │   ├── types.go          # Document data structures
│   │   └── filters.go        # Extension filtering (.md, .txt)
│   ├── llm/
│   │   ├── client.go         # Ollama HTTP client
│   │   ├── processor.go      # LLM processing logic
│   │   ├── retry.go          # Retry and backoff logic
│   │   └── pool.go           # Worker pool for concurrent requests
│   ├── dataset/
│   │   ├── types.go          # Dataset structures (TrainingPair, etc)
│   │   ├── deduplicator.go   # Deduplication logic
│   │   ├── filler.go         # Fill logic to reach target counts
│   │   └── splitter.go       # Train/valid/test splitting
│   ├── output/
│   │   ├── jsonl.go          # JSONL writing
│   │   └── merger.go         # Per-doc → master file merging
│   └── templates/
│       ├── embed.go          # Embedded templates
│       └── render.go         # Template rendering
├── templates/
│   ├── prompt_train.txt      # Embedded training prompt
│   ├── prompt_valid.txt      # Embedded validation prompt
│   └── prompt_test.txt       # Embedded test prompt
├── go.mod
├── go.sum
└── README.md
```

### Structure Rationale

- **cmd/datakeg/:** Standard Go convention for CLI entry point, keeps main.go minimal
- **internal/:** Prevents external import, signals this is not a library
- **Separation by concern:** Each package has single responsibility (document reading vs LLM processing vs output writing)
- **cli/ package:** All CLI-specific code isolated, makes testing core logic easier
- **pipeline/ package:** Orchestration logic separate from implementation details
- **templates/ as embedded:** Using Go 1.16+ embed makes binary self-contained

## Architectural Patterns

### Pattern 1: Pipeline Orchestrator

**What:** Central coordinator that executes phases sequentially, managing state between them
**When to use:** Multi-phase workflows where each phase depends on previous output
**Trade-offs:** Simple to reason about but phases can't overlap; suitable for datakeg's linear workflow

**Example:**
```go
type Orchestrator struct {
    config   *Config
    progress *ProgressReporter
    docReader *document.Reader
    llmProc   *llm.Processor
    deduper   *dataset.Deduplicator
    writer    *output.Writer
}

func (o *Orchestrator) Run(ctx context.Context) error {
    // Phase 1: Read documents
    docs, err := o.docReader.ReadAll(ctx, o.config.SourceDir)
    if err != nil {
        return fmt.Errorf("read phase: %w", err)
    }
    o.progress.PhaseComplete("read", len(docs))

    // Phase 2: Process through LLM
    pairs, err := o.llmProc.ProcessBatch(ctx, docs)
    if err != nil {
        return fmt.Errorf("process phase: %w", err)
    }
    o.progress.PhaseComplete("process", len(pairs))

    // Phase 3: Deduplicate
    deduped := o.deduper.Deduplicate(pairs)
    o.progress.PhaseComplete("dedupe", len(deduped))

    // Phase 4: Fill if needed
    if len(deduped) < o.config.TargetCount {
        filled, err := o.llmProc.Fill(ctx, deduped, o.config.TargetCount)
        if err != nil {
            return fmt.Errorf("fill phase: %w", err)
        }
        deduped = filled
    }

    // Phase 5: Write outputs
    return o.writer.WriteAll(ctx, deduped, o.config.OutputDir)
}
```

### Pattern 2: Worker Pool for LLM Calls

**What:** Fixed-size pool of goroutines processing documents concurrently
**When to use:** I/O-bound operations (API calls) that benefit from parallelism
**Trade-offs:** Complexity vs performance; needed for datakeg since LLM calls are slow

**Example:**
```go
type Processor struct {
    client      *Client
    workerCount int
}

func (p *Processor) ProcessBatch(ctx context.Context, docs []Document) ([]TrainingPair, error) {
    jobs := make(chan Document, len(docs))
    results := make(chan ProcessResult, len(docs))

    // Start workers
    g, ctx := errgroup.WithContext(ctx)
    for i := 0; i < p.workerCount; i++ {
        g.Go(func() error {
            return p.worker(ctx, jobs, results)
        })
    }

    // Send jobs
    go func() {
        for _, doc := range docs {
            jobs <- doc
        }
        close(jobs)
    }()

    // Collect results
    var pairs []TrainingPair
    go func() {
        for r := range results {
            if r.Error == nil {
                pairs = append(pairs, r.Pairs...)
            }
        }
    }()

    return pairs, g.Wait()
}
```

### Pattern 3: Context-Based Cancellation

**What:** Use context.Context throughout to support graceful shutdown
**When to use:** Always in CLI tools with long-running operations
**Trade-offs:** Small complexity overhead but critical for user experience

**Example:**
```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle Ctrl-C
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigChan
        fmt.Fprintln(os.Stderr, "\nShutting down gracefully...")
        cancel()
    }()

    orch := NewOrchestrator(config)
    if err := orch.Run(ctx); err != nil {
        if errors.Is(err, context.Canceled) {
            os.Exit(130) // Standard shell exit code for SIGINT
        }
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Pattern 4: Progressive Enhancement Output

**What:** Write per-document files immediately, merge at the end
**When to use:** When partial results are valuable (crash recovery)
**Trade-offs:** More I/O but provides resumability

**Example:**
```go
// Write per-document JSONL immediately after processing
func (w *Writer) WriteDocumentPairs(docName string, pairs []TrainingPair) error {
    f, err := os.Create(filepath.Join(w.outputDir, "per-doc", docName+".jsonl"))
    if err != nil {
        return err
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    for _, pair := range pairs {
        if err := enc.Encode(pair); err != nil {
            return err
        }
    }
    return nil
}

// Merge at the end
func (w *Writer) MergeAll() error {
    // Read all per-doc files, split by type, merge into master files
}
```

### Pattern 5: Retry with Exponential Backoff

**What:** Automatic retry for transient LLM API failures
**When to use:** Network calls, especially to external services
**Trade-offs:** Increases latency on failures but improves reliability

**Example:**
```go
func (c *Client) CallWithRetry(ctx context.Context, prompt string, maxRetries int) (string, error) {
    var lastErr error
    backoff := time.Second

    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return "", ctx.Err()
            case <-time.After(backoff):
                backoff *= 2 // Exponential
            }
        }

        resp, err := c.call(ctx, prompt)
        if err == nil {
            return resp, nil
        }

        // Only retry on retryable errors
        if !isRetryable(err) {
            return "", err
        }
        lastErr = err
    }

    return "", fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}
```

## Data Flow

### Overall Pipeline Flow

```
[Source Directory]
    ↓ (walk + filter .md/.txt)
[Document Reader] → []Document
    ↓ (fan-out to worker pool)
[LLM Processor] → []TrainingPair (per doc, 3 types each)
    ↓ (write immediately)
[Per-Doc JSONL Writer] → output/per-doc/*.jsonl
    ↓ (collect all)
[Splitter] → train[], valid[], test[] sets
    ↓ (deduplicate within and across sets)
[Deduplicator] → deduplicated sets
    ↓ (if count < target)
[Filler] → regenerate missing pairs
    ↓ (merge sets)
[Merger] → output/{train,valid,test}.jsonl
    ↓
[Final Output]
```

### Concurrent Processing Flow

```
Main Goroutine:
    ↓
[Document Reader] → Channel<Document>
    ↓ ↓ ↓ (fan-out)
[Worker 1] [Worker 2] [Worker 3] ... [Worker N]
    ↓ ↓ ↓ (fan-in)
Channel<Result> → [Result Collector]
    ↓
[Continue Pipeline]
```

### Key Data Flows

1. **Document → LLM Pairs:** Each document generates exactly 3 training pairs (train, valid, test types). Worker pool ensures N documents processed in parallel.

2. **Deduplication:** Hash-based comparison within and across sets. Order matters: dedupe train first, then valid (against train), then test (against train+valid).

3. **Fill Mechanism:** When deduplication reduces count below target, re-invoke LLM with tracking to avoid regenerating existing pairs.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-100 docs | Single-threaded is fine, worker pool of 3-5 sufficient |
| 100-1k docs | Worker pool of 10-20, consider streaming JSONL output |
| 1k-10k docs | Increase workers to match Ollama capacity, add checkpoint/resume |
| 10k+ docs | Consider splitting into batches, external queue (not needed for datakeg MVP) |

### Scaling Priorities

1. **First bottleneck:** LLM API calls (I/O bound). Solution: Worker pool with configurable concurrency, size based on Ollama's capacity (CPU/GPU limits).

2. **Second bottleneck:** Memory for deduplication. If deduping 10k+ pairs, map of hashes could grow large. Solution: Disk-backed deduplication or streaming comparison (not needed for MVP).

3. **Third bottleneck:** Disk I/O for JSONL writes. Solution: Buffered writes, batch merging (already in design).

## Anti-Patterns

### Anti-Pattern 1: Global State for Configuration

**What people do:** Use package-level variables for configuration
**Why it's wrong:** Makes testing impossible, prevents running multiple pipelines
**Do this instead:** Pass config explicitly through constructors, use dependency injection

```go
// BAD
var (
    sourceDir string
    outputDir string
)

func ProcessDocuments() { ... }

// GOOD
type Config struct {
    SourceDir string
    OutputDir string
}

func NewOrchestrator(cfg *Config) *Orchestrator { ... }
```

### Anti-Pattern 2: Ignoring Context Cancellation

**What people do:** Long-running operations that don't check ctx.Done()
**Why it's wrong:** User can't interrupt, leads to zombie processes
**Do this instead:** Check context in loops, pass to all blocking operations

```go
// BAD
for _, doc := range docs {
    process(doc) // No way to cancel
}

// GOOD
for _, doc := range docs {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        if err := process(ctx, doc); err != nil {
            return err
        }
    }
}
```

### Anti-Pattern 3: Unbounded Concurrency

**What people do:** Spawn goroutine per document without limit
**Why it's wrong:** Can overwhelm system/API, cause OOM or rate limiting
**Do this instead:** Use fixed-size worker pool with errgroup

```go
// BAD
for _, doc := range docs {
    go func(d Document) {
        process(d) // No coordination
    }(doc)
}

// GOOD
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(workerCount) // Limit concurrency
for _, doc := range docs {
    doc := doc // Capture loop var
    g.Go(func() error {
        return process(ctx, doc)
    })
}
return g.Wait()
```

### Anti-Pattern 4: Silent Failures in Concurrent Code

**What people do:** Log errors but don't propagate them
**Why it's wrong:** Pipeline completes "successfully" with missing data
**Do this instead:** Use errgroup to collect first error and cancel remaining work

```go
// BAD
for _, doc := range docs {
    go func(d Document) {
        if err := process(d); err != nil {
            log.Println(err) // Lost!
        }
    }(doc)
}

// GOOD
g, ctx := errgroup.WithContext(ctx)
for _, doc := range docs {
    doc := doc
    g.Go(func() error {
        return process(ctx, doc) // Propagates up
    })
}
if err := g.Wait(); err != nil {
    return fmt.Errorf("processing failed: %w", err)
}
```

### Anti-Pattern 5: Tightly Coupled Components

**What people do:** LLM processor directly writes to files
**Why it's wrong:** Can't test LLM logic without file system, can't change output format
**Do this instead:** Return data structures, separate I/O concerns

```go
// BAD
func (p *Processor) Process(doc Document) error {
    pairs := p.generatePairs(doc)
    return writeToFile(pairs) // Tight coupling
}

// GOOD
func (p *Processor) Process(ctx context.Context, doc Document) ([]TrainingPair, error) {
    return p.generatePairs(ctx, doc), nil
}

// Caller decides what to do with pairs
pairs, err := processor.Process(ctx, doc)
if err != nil { ... }
writer.Write(pairs)
```

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Ollama | HTTP client with JSON API | Local service, typically http://localhost:11434 |
| File System | stdlib os/filepath | Use filepath.Walk for directory traversal |
| Templates | embed.FS + text/template | Embed at compile time for single binary |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| CLI ↔ Orchestrator | Direct function calls | CLI parses flags, creates config, calls Run() |
| Orchestrator ↔ Processors | Direct calls + channels | Orchestrator owns pipeline, processors return data |
| Processor ↔ LLM Client | Request/response + retry | Client handles HTTP, processor handles business logic |
| Reader ↔ File System | Synchronous reads | Small files, no streaming needed |
| Writer ↔ File System | Buffered writes | Use bufio.Writer for performance |

## Build Order Implications

Based on component dependencies, suggested build order:

1. **Foundation (no dependencies):**
   - Config structures (`pipeline/config.go`)
   - Data types (`document/types.go`, `dataset/types.go`)
   - Template embedding (`templates/embed.go`)

2. **Integration layer (external dependencies only):**
   - Document reader (`document/reader.go`)
   - JSONL writer (`output/jsonl.go`)
   - Ollama client (`llm/client.go`)

3. **Processing layer (depends on integration):**
   - Template renderer (`templates/render.go`)
   - LLM processor (`llm/processor.go`)
   - Deduplicator (`dataset/deduplicator.go`)

4. **Orchestration (depends on processing):**
   - Pipeline orchestrator (`pipeline/orchestrator.go`)
   - Progress reporter (`cli/progress.go`)

5. **CLI (depends on orchestration):**
   - Flag parsing (`cli/flags.go`)
   - Main entry point (`cmd/datakeg/main.go`)

This order allows testing each layer independently before composing.

## Sources

**Confidence note:** This research is based on established Go patterns and CLI best practices from my training data (January 2025). Web search verification was unavailable. Patterns are MEDIUM confidence - standard in Go ecosystem but specific library choices should be verified during implementation.

Key patterns drawn from:
- Go standard library patterns (stdlib flag, context, errgroup)
- Common CLI tool structures (cobra pattern, even if not using cobra)
- Document processing pipeline patterns (reader → processor → writer)
- Worker pool patterns (golang.org/x/sync/errgroup)

---
*Architecture research for: datakeg - Go CLI for LLM fine-tuning dataset generation*
*Researched: 2026-02-05*
