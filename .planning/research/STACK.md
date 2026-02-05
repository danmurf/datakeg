# Stack Research

**Domain:** Go CLI tool for LLM dataset generation
**Researched:** 2026-02-05
**Confidence:** MEDIUM (based on training data through Jan 2025; external verification unavailable)

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended | Confidence |
|------------|---------|---------|-----------------|------------|
| Go | 1.23+ | Language and runtime | Current stable release with embed support, excellent for CLI distribution | HIGH |
| Cobra | v1.8+ | CLI framework | Industry standard for Go CLIs (kubectl, gh, hugo). Rich flag handling, subcommands, automatic help generation | MEDIUM |
| Viper | v1.18+ | Configuration management | Pairs naturally with Cobra, handles flags/env/config files, used by most major Go CLIs | MEDIUM |

### LLM Integration

| Library | Version | Purpose | Why Recommended | Confidence |
|---------|---------|---------|-----------------|------------|
| ollama/ollama (Go SDK) | Latest | Ollama API client | Official Go client for Ollama, handles chat completions and model management | LOW* |
| net/http (stdlib) | stdlib | HTTP client | If official SDK unavailable, Ollama REST API is straightforward HTTP/JSON | HIGH |

*Cannot verify current Ollama Go SDK status without web access. Ollama provides REST API that works with standard Go http client.

### File Processing

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| embed (stdlib) | stdlib (Go 1.16+) | Embed prompt templates | Built into Go, zero dependencies for embedding files in binary | HIGH |
| filepath (stdlib) | stdlib | Path handling | Cross-platform file path operations | HIGH |
| encoding/json (stdlib) | stdlib | JSONL output | Standard library sufficient for line-by-line JSON writing | HIGH |
| bufio (stdlib) | stdlib | Efficient file I/O | Buffered reading/writing for large files | HIGH |

### Progress Reporting

| Library | Version | Purpose | When to Use | Confidence |
|---------|---------|---------|-------------|------------|
| schollz/progressbar/v3 | v3.14+ | Progress bars | Clean, customizable progress bars with ETA | MEDIUM |
| briandowns/spinner | v1.23+ | Spinners | For indeterminate operations (LLM calls) | MEDIUM |
| Standard output | stdlib | Simple logging | fmt.Fprintf(os.Stderr) sufficient for basic progress | HIGH |

### Development Tools

| Tool | Purpose | Notes | Confidence |
|------|---------|-------|------------|
| golangci-lint | Linting | Standard Go linter aggregator, catches common issues | HIGH |
| go test | Testing | Built-in testing framework, sufficient for CLI logic | HIGH |
| goreleaser | Binary distribution | Automates multi-platform builds, GitHub releases | MEDIUM |
| go build | Local builds | Standard toolchain, use with -ldflags for version info | HIGH |

## Installation

```bash
# Initialize Go module (already done)
go mod init github.com/danmurf/datakeg

# Core CLI framework
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest

# Ollama integration (verify official SDK first)
# Option 1: If official SDK exists
go get github.com/ollama/ollama-go@latest

# Option 2: Use standard library
# No additional dependencies needed - use net/http

# Progress UI (choose one)
go get github.com/schollz/progressbar/v3@latest
# OR
go get github.com/briandowns/spinner@latest

# Development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative | Confidence |
|-------------|-------------|-------------------------|------------|
| Cobra | urfave/cli | Simpler CLIs without subcommands. Cobra better for extensibility | MEDIUM |
| Cobra | flag (stdlib) | Tiny CLIs where dependencies matter. Cobra worth it for this project | HIGH |
| Viper | envconfig | Config via env vars only. Viper overkill if not using config files | MEDIUM |
| progressbar/v3 | bubbletea | If you want full TUI with interactivity. Overkill for progress reporting | LOW |
| Standard JSON | jsoniter | Only if profiling shows JSON is bottleneck (unlikely) | MEDIUM |

## What NOT to Use

| Avoid | Why | Use Instead | Confidence |
|-------|-----|-------------|------------|
| kingpin | Archived, no longer maintained | Cobra or urfave/cli | MEDIUM |
| go-flags | Less actively maintained than alternatives | Cobra for complex CLIs, flag stdlib for simple | MEDIUM |
| Third-party JSON libs initially | Premature optimization; stdlib works well | encoding/json until proven bottleneck | HIGH |
| CGo dependencies | Breaks static binary compilation, complicates cross-compilation | Pure Go libraries only | HIGH |

## Stack Patterns by Variant

**For Simple CLI (current recommendation):**
- Use Cobra for CLI structure
- Use Viper only if you add config file support (defer to v2)
- Use standard library for HTTP/JSON unless official Ollama SDK is well-maintained
- Use progressbar/v3 for progress reporting
- Embed templates with `//go:embed`

**For Advanced CLI (future):**
- Add Viper for config file support (.datakeg.yaml)
- Add bubbletea for interactive TUI
- Add concurrent processing with worker pools
- Add structured logging (zerolog or slog in Go 1.21+)

**For Minimal Dependencies:**
- Skip Cobra, use flag (stdlib)
- Skip progress bars, use fmt.Fprintf
- HTTP client + JSON encoding (stdlib only)
- Still embed templates (stdlib)

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Go 1.23+ | All stdlib packages | Embed requires 1.16+, slog available in 1.21+ |
| Cobra v1.8+ | Viper v1.18+ | Cobra and Viper maintained by same team, well-integrated |
| progressbar/v3 | Go 1.13+ | No special version requirements |

## Architecture Implications

**CLI Framework Choice → Code Structure:**
- Cobra encourages cmd/ directory with one file per subcommand
- Root command in cmd/root.go, main.go is minimal

**Embedding → Build Process:**
- Templates in templates/ directory
- `//go:embed templates/*.txt` in code
- No runtime template loading needed

**Ollama Integration → Error Handling:**
- LLM calls can fail (timeout, model not available, out of memory)
- Need retry logic with exponential backoff
- Need graceful degradation if LLM unavailable

**JSONL Output → Streaming:**
- Don't load entire dataset in memory
- Write JSONL line-by-line as pairs are generated
- Use bufio.Writer for efficient I/O

## Key Technical Decisions

### Decision 1: Cobra vs flag (stdlib)
**Recommendation:** Cobra
**Why:** Project requirements show multiple flags (source, output, model, percentages, pairs_per_1k). Cobra's flag handling, help generation, and validation make it worth the dependency.
**Confidence:** HIGH

### Decision 2: Official Ollama SDK vs HTTP client
**Recommendation:** Try official SDK first, fall back to http.Client
**Why:** Official SDKs provide better types and error handling. Ollama's REST API is simple enough that http.Client + encoding/json works fine as fallback.
**Confidence:** MEDIUM (cannot verify SDK quality without access)

### Decision 3: Progress library vs manual
**Recommendation:** Use progressbar/v3
**Why:** Generates multiple pairs per doc × multiple docs = long-running operation. Progress bars improve UX significantly. Library is mature, small dependency.
**Confidence:** MEDIUM

### Decision 4: Dependency minimization
**Recommendation:** Moderate dependency approach
**Why:** Go's excellent standard library means we can avoid many dependencies, but quality dependencies (Cobra, progressbar) improve DX and UX significantly. Avoid dependencies for things stdlib does well (JSON, HTTP, file I/O).
**Confidence:** HIGH

## Research Limitations

Due to unavailability of web search and fetch tools during research:

1. **Cannot verify current versions:** Version numbers are based on training data through Jan 2025. Verify latest versions before installation.

2. **Cannot verify Ollama Go SDK:** Training data suggests Ollama may have official Go client, but cannot confirm current status, API, or quality. **HIGH PRIORITY:** Verify if github.com/ollama/ollama-go or similar exists and is maintained.

3. **Cannot verify Cobra/Viper latest releases:** Recommended versions are from training data. Both are mature, stable projects, but check for recent updates.

4. **Cannot verify current Go best practices:** Stack follows patterns from training data. Go 1.23 may have introduced new stdlib features that obsolete some recommendations.

## Verification Needed

Before implementing:

- [ ] Check if Ollama has official Go SDK at github.com/ollama/ollama or github.com/ollama/ollama-go
- [ ] Verify Ollama REST API documentation at ollama.ai or GitHub
- [ ] Check latest Cobra version (github.com/spf13/cobra/releases)
- [ ] Check latest progressbar version (github.com/schollz/progressbar/releases)
- [ ] Verify Go 1.23+ embed syntax hasn't changed
- [ ] Check if Go 1.23+ has new stdlib features relevant to CLI development

## Sources

- Training data through January 2025 (Go ecosystem, Cobra, Viper, common CLI patterns)
- Existing project context from /Users/dan/code/datakeg/.planning/PROJECT.md
- Go version 1.25.4 from /Users/dan/code/datakeg/go.mod

**Note:** All library versions and recommendations require verification with official sources before implementation. This research provides directional guidance based on historical patterns but cannot guarantee current accuracy without web access.

---
*Stack research for: Go CLI tool for LLM dataset generation*
*Researched: 2026-02-05*
*Confidence: MEDIUM - requires verification of current versions and library availability*
