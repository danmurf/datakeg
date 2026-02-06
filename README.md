# datakeg

A command-line tool that transforms raw documentation into LLM training datasets.

## Overview

datakeg processes markdown and text files from a source directory, generates question-answer pairs using Ollama, and outputs structured JSONL files ready for training, validation, and testing.

## Features

- Process markdown (`.md`) and text (`.txt`) files automatically
- Generate question-answer pairs using local Ollama models
- Automatic train/validation/test split generation
- Configurable pair density per document
- JSONL output format compatible with popular ML frameworks

## Prerequisites

- Go 1.25.4 or later
- [Ollama](https://ollama.ai/) running locally
- An Ollama model downloaded (default: `gpt-oss:20b`)

## Installation

```bash
go build -o datakeg ./cmd/datakeg
```

## Usage

### Basic Command

```bash
datakeg generate <source-directory> <output-directory>
```

### Example

```bash
# Generate training data from documentation in ./docs
datakeg generate ./docs ./output
```

### Options

```bash
datakeg generate [flags] <source> <output>

Flags:
  -m, --model string          Ollama model to use (default "gpt-oss:20b")
      --train-pct float       Training set percentage 0.0-1.0 (default 0.6)
      --valid-pct float       Validation set percentage 0.0-1.0 (default 0.2)
      --test-pct float        Test set percentage 0.0-1.0 (default 0.2)
      --pairs-per-1k float    Target pairs per 1000 characters (default 1.0)
  -t, --timeout int           Operation timeout in minutes (default 30)
```

### Example with Options

```bash
datakeg generate \
  --model llama2 \
  --pairs-per-1k 2.0 \
  --timeout 60 \
  ./docs ./output
```

## Output Format

The tool generates three JSONL files in the output directory:

- `train.jsonl` - Training dataset
- `valid.jsonl` - Validation dataset
- `test.jsonl` - Test dataset

Each line in the JSONL files contains a JSON object with:

```json
{
  "prompt": "What is...",
  "completion": "The answer is..."
}
```

## Version Information

```bash
datakeg version
```

## License

See LICENSE file for details.
