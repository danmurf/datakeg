package commands

import (
	"context"
	"fmt"
	"math"
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
		docPath, _ := filepath.Rel(sourceDir, doc.Path)
		fmt.Printf("[%d/%d] Processing: %s (%d chars)\n", i+1, len(documents), docPath, len(doc.Content))

		// Calculate and announce pair counts for each split
		totalPairs := int(math.Ceil(float64(len(doc.Content)) / 1000 * pairsPer1K))

		// Use the generator's config to calculate splits
		genConfig := gen.GetConfig()
		validCount := int(math.Ceil(float64(totalPairs) * genConfig.ValidPercent / 100))
		testCount := int(math.Ceil(float64(totalPairs) * genConfig.TestPercent / 100))
		trainCount := totalPairs - validCount - testCount

		// Handle edge cases for small documents
		if totalPairs == 1 {
			trainCount = 1
			validCount = 0
			testCount = 0
		} else if totalPairs == 2 {
			trainCount = 1
			validCount = 1
			testCount = 0
		}

		fmt.Printf("     → Generating: %d train, %d valid, %d test pairs\n", trainCount, validCount, testCount)

		// Generate for each split type (only if count > 0)
		splits := []struct {
			name      string
			splitType generator.SplitType
			pairs     *[]writer.TrainingPair
			count     int
		}{
			{"train", generator.SplitTrain, &trainPairs, trainCount},
			{"valid", generator.SplitValid, &validPairs, validCount},
			{"test", generator.SplitTest, &testPairs, testCount},
		}

		for _, split := range splits {
			if split.count <= 0 {
				fmt.Printf("     → Skipping %s split (0 pairs required)\n", split.name)
				continue
			}

			fmt.Printf("     → Calling LLM for %s split...\n", split.name)
			pairs, err := gen.Generate(ctx, &doc, split.splitType)
			if err != nil {
				return fmt.Errorf("generate %s pairs for %s: %w", split.splitType, doc.Name, err)
			}
			fmt.Printf("     → %s split: %d pairs generated\n", split.name, len(pairs))

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
