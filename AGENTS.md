# AGENTS.md

This file provides guidance for agentic coding agents working with the datakeg codebase.

## Mandatory Verification After Changes

**After any code changes, you MUST run both linter and tests:**

```bash
# Run linter first
make lint

# Then run tests
make test
```

Do not commit or submit code until both commands pass without errors.

### Build Commands
```bash
# Build the binary (builds to ./datakeg)
make build

# Build with specific version
make build VERSION=1.0.0

# Clean build artifacts
make clean
```

### Test Commands
```bash
# Run all tests (uses Go's built-in test cache for speed)
make test

# Run all tests without cache (useful after external changes)
make test-nocache

# Run tests with coverage report (generates coverage.html)
make coverage

# Run a single test
go test -v -run TestName ./path/to/package

# Run a single test in a specific package
go test -v -run TestLoadDocuments ./internal/processor

# Run tests with verbose output
go test -v ./...
```

### Linting Commands
```bash
# Run golangci-lint (automatically installs if needed)
make lint

# Run lint with specific configurations
golangci-lint run --timeout=5m ./...
```

## Code Style Guidelines

### Import Organization
- Use standard Go import grouping: standard library, third-party, local packages
- Group imports with blank lines between groups
- Use relative imports for local packages
- Alphabetize within groups (except standard library which has specific order)

```go
import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    
    "github.com/danmurf/datakeg/internal/generator"
    "github.com/danmurf/datakeg/internal/ollama"
    "github.com/danmurf/datakeg/internal/processor"
)
```

### Formatting
- Use `goimports` to automatically format imports and code
- Run `goimports -w .` before committing
- Line length should not exceed 100 characters
- Use tabs for indentation (4 spaces per tab)

### Naming Conventions
- **Package names**: lowercase, descriptive, avoid underscores
- **Public functions/methods**: PascalCase (exported)
- **Private functions/methods**: camelCase (unexported)
- **Struct names**: PascalCase, descriptive
- **Interface names**: PascalCase, typically end with "-er"
- **Constants**: PascalCase
- **Variables**: camelCase for private, PascalCase for exported
- **Error variables**: use "err" prefix

```go
// Good naming
type Generator struct {
    client *ollama.Client
    config Config
}

func (g *Generator) Generate(ctx context.Context, doc Document) ([]Pair, error) {
    return nil, nil
}

const (
    SplitTrain SplitType = "train"
    SplitValid SplitType = "valid"
    SplitTest  SplitType = "test"
)
```

### Error Handling
- Always wrap errors with context using `fmt.Errorf` and `%w`
- Handle errors at appropriate level, don't just pass them up
- Use `errors.Is()` and `errors.As()` for error checking
- Return descriptive error messages

```go
// Good error handling
documents, err := processor.LoadDocuments(sourceDir)
if err != nil {
    return fmt.Errorf("load documents: %w", err)
}

if len(documents) == 0 {
    return fmt.Errorf("no .md or .txt files found in %s", sourceDir)
}
```

### Type Safety
- Use strong typing over string constants where possible
- Define custom types for domain-specific concepts
- Use interfaces for abstraction

```go
// Good type usage
type SplitType string

const (
    SplitTrain SplitType = "train"
    SplitValid SplitType = "valid"
    SplitTest  SplitType = "test"
)

type Config struct {
    PairsPer1KChars float64
    ValidPercent    float64
    TestPercent     float64
    Model           string
}
```

### Testing Guidelines
- Use table-driven tests for complex logic
- Test both happy path and error cases
- Use testify/assert for assertions when available
- Mock external dependencies in tests

```go
func TestCalculateSplitCounts(t *testing.T) {
    tests := []struct {
        name        string
        totalPairs  int
        validPercent float64
        testPercent  float64
        wantTrain   int
        wantValid   int
        wantTest    int
    }{
        {
            name:        "small document",
            totalPairs:  1,
            validPercent: 10.0,
            testPercent:  10.0,
            wantTrain:   1,
            wantValid:   0,
            wantTest:    0,
        },
        {
            name:        "medium document",
            totalPairs:  10,
            validPercent: 10.0,
            testPercent:  10.0,
            wantTrain:   8,
            wantValid:   1,
            wantTest:    1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            train, valid, test := calculateSplitCounts(tt.totalPairs, tt.validPercent, tt.testPercent)
            assert.Equal(t, tt.wantTrain, train)
            assert.Equal(t, tt.wantValid, valid)
            assert.Equal(t, tt.wantTest, test)
        })
    }
}
```

### Code Organization
- Follow the package structure: cmd/, internal/, templates/
- Keep business logic separate from infrastructure concerns
- Use dependency injection for external services
- Maintain single responsibility principle

### Template System
- Templates are embedded at compile time via `//go:embed`
- Changes to `.tmpl` files require recompilation
- Templates receive structured data via PromptData struct
- LLMs should return JSON arrays of pairs

### Version Information
- Version is injected via ldflags during build
- Access via `datakeg version` command
- Version includes git commit and build date
- Appends "-dirty" suffix if working directory has uncommitted changes

### Testing Philosophy
- The codebase uses table-driven tests extensively
- Tests are cached by Go for faster subsequent runs
- Tests are re-run when source files, test files, or dependencies change
- Use `make test-nocache` to force tests to run without cache

### Important Implementation Details
- For documents with very few pairs (1-2), use special split logic
- Ollama responses are parsed to extract JSON arrays
- Templates are embedded at compile time and require recompilation
- Error handling should provide clear, actionable messages
- Use context cancellation for timeout handling