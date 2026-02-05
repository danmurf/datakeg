package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/ollama"
	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/writer"
)

// ExecuteGeneratePipeline orchestrates the full generate pipeline:
// 1. Load documents from source directory
// 2. Create Ollama client
// 3. Generate prompt/completion pairs for each document
// 4. Write train/valid/test JSONL files to output directory
func ExecuteGeneratePipeline(
	sourceDir string,
	outputDir string,
	model string,
	pairsPer1K float64,
	timeoutMinutes int,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMinutes)*time.Minute)
	defer cancel()

	fmt.Printf("Starting pipeline...\n")

	// Step 1: Load documents from source directory
	fmt.Printf("Loading documents from %s...\n", sourceDir)
	documents, err := processor.LoadDocuments(sourceDir)
	if err != nil {
		return fmt.Errorf("load documents: %w", err)
	}
	fmt.Printf("Loaded %d documents\n", len(documents))

	if len(documents) == 0 {
		return fmt.Errorf("no .md or .txt files found in %s", sourceDir)
	}

	// Step 2: Create Ollama client
	fmt.Printf("Connecting to Ollama...\n")
	ollamaClient, err := ollama.NewClient()
	if err != nil {
		return fmt.Errorf("create ollama client: %w", err)
	}

	// Step 3: Create generator with config
	genConfig := generator.Config{
		PairsPer1KChars: pairsPer1K,
		ValidPercent:    10.0,
		TestPercent:     10.0,
		Model:           model,
	}
	gen := generator.NewGenerator(ollamaClient, genConfig)

	// Step 4: Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Step 5: Process documents and collect pairs
	var trainPairs []writer.TrainingPair
	var validPairs []writer.TrainingPair
	var testPairs []writer.TrainingPair

	for i, doc := range documents {
		fmt.Printf("Processing document %d/%d: %s\n", i+1, len(documents), doc.Name)

		// Generate for each split type
		splits := []struct {
			splitType generator.SplitType
			pairs     *[]writer.TrainingPair
		}{
			{generator.SplitTrain, &trainPairs},
			{generator.SplitValid, &validPairs},
			{generator.SplitTest, &testPairs},
		}

		for _, split := range splits {
			pairs, err := gen.Generate(ctx, &doc, split.splitType)
			if err != nil {
				return fmt.Errorf("generate %s pairs for %s: %w", split.splitType, doc.Name, err)
			}

			// Convert generator.Pair to writer.TrainingPair
			for _, p := range pairs {
				*split.pairs = append(*split.pairs, writer.TrainingPair{
					Prompt:     p.Prompt,
					Completion: p.Completion,
				})
			}
		}
	}

	// Step 6: Write output files
	fmt.Printf("Writing output files to %s...\n", outputDir)

	if len(trainPairs) > 0 {
		outFile := filepath.Join(outputDir, "train.jsonl")
		if err := writer.WriteJSONL(outFile, trainPairs); err != nil {
			return fmt.Errorf("write train.jsonl: %w", err)
		}
		fmt.Printf("  Written %d pairs to train.jsonl\n", len(trainPairs))
	}

	if len(validPairs) > 0 {
		outFile := filepath.Join(outputDir, "valid.jsonl")
		if err := writer.WriteJSONL(outFile, validPairs); err != nil {
			return fmt.Errorf("write valid.jsonl: %w", err)
		}
		fmt.Printf("  Written %d pairs to valid.jsonl\n", len(validPairs))
	}

	if len(testPairs) > 0 {
		outFile := filepath.Join(outputDir, "test.jsonl")
		if err := writer.WriteJSONL(outFile, testPairs); err != nil {
			return fmt.Errorf("write test.jsonl: %w", err)
		}
		fmt.Printf("  Written %d pairs to test.jsonl\n", len(testPairs))
	}

	fmt.Printf("Pipeline complete!\n")
	fmt.Printf("Total pairs: train=%d, valid=%d, test=%d\n", len(trainPairs), len(validPairs), len(testPairs))

	return nil
}
