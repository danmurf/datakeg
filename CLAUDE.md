# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

datakeg transforms raw documentation into LLM training datasets. It processes markdown/text files, generates question-answer pairs using Ollama models, and outputs train/valid/test JSONL files.

## Build and Test Commands

```bash
# Build the binary (builds to ./datakeg)
make build

# Run all tests (uses Go's built-in test cache for speed)
make test

# Run all tests without cache (useful after external changes)
make test-nocache

# Run tests with coverage report (generates coverage.html)
make coverage

# Run linting (automatically installs golangci-lint if needed)
make lint

# Run a single test
go test -v -run TestName ./path/to/package

# Run a single test in a specific package
go test -v -run TestLoadDocuments ./internal/processor

# Clean build artifacts and test cache
make clean

# Install to $GOPATH/bin
make install
```

### Test Caching

Go automatically caches test results for faster subsequent runs. Tests are re-run when:
- Source files change
- Test files change
- Dependencies change

To force tests to run without cache, use `make test-nocache` or `go test -count=1 ./...`.

## Architecture

The codebase follows a pipeline architecture with clear separation of concerns:

### Pipeline Flow
1. **processor** package: Recursively loads `.md` and `.txt` files from source directory
2. **generator** package: Calls Ollama LLM to create question-answer pairs for each document
3. **writer** package: Outputs JSONL files (train.jsonl, valid.jsonl, test.jsonl)

### Key Components

**cmd/datakeg/commands/generate.go**: Orchestrates the entire pipeline
- Calls processor.LoadDocuments()
- Creates ollama.Client
- Creates generator.Generator with Config
- Iterates documents, calling gen.Generate() for each split type (train/valid/test)
- Collects all pairs and writes JSONL files via writer.WriteJSONL()

**internal/generator/generator.go**: Core generation logic
- `Config` specifies PairsPer1KChars, ValidPercent, TestPercent, Model
- `calculatePairs()`: Determines total pairs based on document character count
- `calculateSplitCounts()`: Distributes pairs across train/valid/test splits with special handling for small documents
- `Generate()`: Executes template, calls Ollama, parses JSON response
- `parseResponse()`: Extracts JSON array from LLM response, handles malformed output by padding with empty pairs

**internal/templates/templates.go**: Template system
- Uses `//go:embed` to embed `.tmpl` files at compile time
- Templates are in `internal/templates/prompts/` (train.tmpl, valid.tmpl, test.tmpl)
- Templates receive `PromptData{DocumentContent, PairCount, DocumentName}`
- LLM is instructed to return JSON array: `[{"prompt": "...", "completion": "..."}, ...]`

**internal/ollama/client.go**: Ollama API wrapper
- Wraps official `github.com/ollama/ollama/api` client
- `NewClient()`: Creates client from OLLAMA_HOST environment variable
- `Generate()`: Streams response from model, respects context cancellation

**internal/processor/files.go**: Document loading
- `LoadDocuments()`: Walks directory tree, filters .md/.txt files
- Returns `[]Document` with Name (filename without extension), Path, Content

**internal/writer/jsonl.go**: JSONL output
- `WriteJSONL()`: Creates new file, encodes each TrainingPair as JSON line
- `WriteJSONLAppend()`: Appends to existing file

### Version Information

The build system uses ldflags to inject version info into the binary:
- Version, commit, date, and Go version are set via Makefile
- Appends "-dirty" suffix if working directory has uncommitted changes
- Access via `datakeg version` command

## Important Implementation Details

### Split Calculation Edge Cases
For documents with very few pairs (1-2 pairs), the generator has special logic:
- 1 pair: All to train, none to valid/test
- 2 pairs: 1 to train, 1 to valid, 0 to test
- 3+ pairs: Use percentage-based distribution

### Template System
Templates are embedded at compile time via `//go:embed`. Changes to `.tmpl` files require recompilation.

### LLM Response Parsing
The generator expects Ollama to return a JSON array. `parseResponse()` extracts JSON between `[` and `]`, handles double-encoded JSON, and pads with empty pairs if the LLM returns fewer than expected.

### Testing Philosophy
The codebase uses table-driven tests. See `internal/generator/generator_test.go` and `internal/processor/files_test.go` for examples.
