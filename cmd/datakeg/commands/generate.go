package commands

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/ollama"
	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/writer"
)

// ExecuteGeneratePipeline orchestrates the full generate pipeline:
// 1. Load documents from source directory
// 2. Create Ollama client
// 3. Generate prompt/completion pairs for each document (train → valid → test with exclusions)
// 4. Write per-document JSONL files during processing
// 5. Optionally merge into master files (skipable via skipMerge flag)
func ExecuteGeneratePipeline(
	sourceDir string,
	outputDir string,
	model string,
	pairsPer1K float64,
	validPct float64,
	testPct float64,
	timeoutMinutes int,
	skipMerge bool,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMinutes)*time.Minute)
	defer cancel()

	fmt.Printf("Starting pipeline...\n")

	// Step 1: Load documents from source directory
	fmt.Printf("Loading documents from %s...\n", sourceDir)
	documents, err := processor.LoadDocuments(sourceDir)
	if err != nil {
		return fmt.Errorf("could not read source directory: %s\nCheck that the path exists and you have read permissions.", sourceDir)
	}
	fmt.Printf("Loaded %d documents\n", len(documents))

	if len(documents) == 0 {
		return fmt.Errorf("no markdown (.md) or text (.txt) files found in %s\nAdd some documentation files to the source directory and try again.", sourceDir)
	}

	// Step 2: Create Ollama client
	fmt.Printf("Connecting to Ollama...\n")
	ollamaClient, err := ollama.NewClient()
	if err != nil {
		return fmt.Errorf("could not connect to Ollama\nMake sure Ollama is running (ollama serve) and the model is available (ollama pull %s).", model)
	}

	// Step 3: Create generator with config
	// CLI uses 0.0-1.0 range but config uses 0-100 percentage
	genConfig := generator.Config{
		PairsPer1KChars: pairsPer1K,
		ValidPercent:    validPct * 100,
		TestPercent:     testPct * 100,
		Model:           model,
	}
	gen := generator.NewGenerator(ollamaClient, genConfig)

	// Step 4: Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory %s\nCheck that you have write permissions and the path is valid.", outputDir)
	}

	// Step 5: Process documents and collect pairs sequentially with cross-split exclusion
	// Train → Valid (exclude train) → Test (exclude train+valid)
	// Per-document pairs reset between documents (no cross-document dedup)
	var trainPairs []writer.TrainingPair
	var validPairs []writer.TrainingPair
	var testPairs []writer.TrainingPair

	for i, doc := range documents {
		docPath, _ := filepath.Rel(sourceDir, doc.Path)
		fmt.Printf("[%d/%d] Processing: %s (%d chars)\n", i+1, len(documents), docPath, len(doc.Content))

		// Get pair counts from generator config
		genCfg := gen.GetConfig()
		totalPairs := int(math.Ceil(float64(len(doc.Content)) / 1000 * pairsPer1K))
		// Use the same split calculation logic as the generator
		validCount := int(math.Ceil(float64(totalPairs) * genCfg.ValidPercent / 100))
		testCount := int(math.Ceil(float64(totalPairs) * genCfg.TestPercent / 100))
		trainCount := totalPairs - validCount - testCount

		// Handle edge cases for small documents (same logic as generator)
		if totalPairs == 1 {
			trainCount = 1
			validCount = 0
			testCount = 0
		} else if totalPairs == 2 {
			trainCount = 1
			validCount = 1
			testCount = 0
		}

		fmt.Printf("     → Target: %d train, %d valid, %d test pairs\n", trainCount, validCount, testCount)

		// Per-document pair tracking (resets each iteration for no cross-document dedup)
		var docTrainPairs []generator.Pair
		var docValidPairs []generator.Pair

		// Step 5a: Generate train pairs (no exclusions)
		if trainCount > 0 {
			fmt.Printf("     → Generating train pairs...\n")
			pairs, err := gen.Generate(ctx, &doc, generator.SplitTrain, nil)
			if err != nil {
				return fmt.Errorf("failed to generate training pairs for %s\nThis may be a temporary Ollama issue. Try again, or try a different model.", doc.Name)
			}
			docTrainPairs = pairs
			fmt.Printf("     → Train: %d pairs generated\n", len(pairs))

			// Convert and append to global train slice
			for _, p := range pairs {
				trainPairs = append(trainPairs, writer.TrainingPair{
					Prompt:     p.Prompt,
					Completion: p.Completion,
				})
			}

			// Write per-document train file
			if len(pairs) > 0 {
				perDocFile := sanitizeDocName(doc, outputDir, "train")
				if err := writer.WriteJSONL(perDocFile, convertPairs(pairs)); err != nil {
					return fmt.Errorf("failed to write %s\nCheck disk space and write permissions.", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}

		// Step 5b: Generate valid pairs (exclude train pairs)
		if validCount > 0 {
			fmt.Printf("     → Generating valid pairs (excluding %d train pairs)...\n", len(docTrainPairs))
			pairs, err := gen.Generate(ctx, &doc, generator.SplitValid, docTrainPairs)
			if err != nil {
				return fmt.Errorf("failed to generate validation pairs for %s\nThis may be a temporary Ollama issue. Try again, or try a different model.", doc.Name)
			}
			docValidPairs = pairs
			fmt.Printf("     → Valid: %d pairs generated\n", len(pairs))

			// Convert and append to global valid slice
			for _, p := range pairs {
				validPairs = append(validPairs, writer.TrainingPair{
					Prompt:     p.Prompt,
					Completion: p.Completion,
				})
			}

			// Write per-document valid file
			if len(pairs) > 0 {
				perDocFile := sanitizeDocName(doc, outputDir, "valid")
				if err := writer.WriteJSONL(perDocFile, convertPairs(pairs)); err != nil {
					return fmt.Errorf("failed to write %s\nCheck disk space and write permissions.", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}

		// Step 5c: Generate test pairs (exclude train + valid pairs)
		if testCount > 0 {
			// Combine train and valid exclusions
			allExclude := append(docTrainPairs, docValidPairs...)
			fmt.Printf("     → Generating test pairs (excluding %d train+valid pairs)...\n", len(allExclude))
			pairs, err := gen.Generate(ctx, &doc, generator.SplitTest, allExclude)
			if err != nil {
				return fmt.Errorf("failed to generate test pairs for %s\nThis may be a temporary Ollama issue. Try again, or try a different model.", doc.Name)
			}
			fmt.Printf("     → Test: %d pairs generated\n", len(pairs))

			// Convert and append to global test slice
			for _, p := range pairs {
				testPairs = append(testPairs, writer.TrainingPair{
					Prompt:     p.Prompt,
					Completion: p.Completion,
				})
			}

			// Write per-document test file
			if len(pairs) > 0 {
				perDocFile := sanitizeDocName(doc, outputDir, "test")
				if err := writer.WriteJSONL(perDocFile, convertPairs(pairs)); err != nil {
					return fmt.Errorf("failed to write %s\nCheck disk space and write permissions.", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}
	}

	// Step 6: Write master output files (skipable via --skip-merge)
	fmt.Printf("Writing output files to %s...\n", outputDir)

	// Skip writing master files if skipMerge is true
	if skipMerge {
		fmt.Printf("Skipping master file merge (--skip-merge specified)\n")
		fmt.Printf("Per-document files written: %d\n", countPerDocFiles(outputDir))
		return nil
	}

	if len(trainPairs) > 0 {
		outFile := filepath.Join(outputDir, "train.jsonl")
		if err := writer.WriteJSONL(outFile, trainPairs); err != nil {
			return fmt.Errorf("failed to write train.jsonl\nCheck disk space and write permissions.")
		}
		fmt.Printf("  Written %d pairs to train.jsonl\n", len(trainPairs))
	}

	if len(validPairs) > 0 {
		outFile := filepath.Join(outputDir, "valid.jsonl")
		if err := writer.WriteJSONL(outFile, validPairs); err != nil {
			return fmt.Errorf("failed to write valid.jsonl\nCheck disk space and write permissions.")
		}
		fmt.Printf("  Written %d pairs to valid.jsonl\n", len(validPairs))
	}

	if len(testPairs) > 0 {
		outFile := filepath.Join(outputDir, "test.jsonl")
		if err := writer.WriteJSONL(outFile, testPairs); err != nil {
			return fmt.Errorf("failed to write test.jsonl\nCheck disk space and write permissions.")
		}
		fmt.Printf("  Written %d pairs to test.jsonl\n", len(testPairs))
	}

	fmt.Printf("Pipeline complete!\n")
	fmt.Printf("Total pairs: train=%d, valid=%d, test=%d\n", len(trainPairs), len(validPairs), len(testPairs))

	return nil
}

// sanitizeDocName creates a sanitized filename from a document for per-document output.
// Removes file extension and replaces spaces with underscores.
func sanitizeDocName(doc processor.Document, outputDir string, split string) string {
	docName := filepath.Base(doc.Path)
	ext := filepath.Ext(docName)
	if ext != "" {
		docName = docName[:len(docName)-len(ext)]
	}
	docName = strings.ReplaceAll(docName, " ", "_")
	return filepath.Join(outputDir, docName+"_"+split+".jsonl")
}

// convertPairs converts generator.Pair slice to writer.TrainingPair slice.
func convertPairs(pairs []generator.Pair) []writer.TrainingPair {
	result := make([]writer.TrainingPair, len(pairs))
	for i, p := range pairs {
		result[i] = writer.TrainingPair{
			Prompt:     p.Prompt,
			Completion: p.Completion,
		}
	}
	return result
}

// countPerDocFiles counts the number of per-document JSONL files in the output directory.
func countPerDocFiles(outputDir string) int {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") && !strings.HasPrefix(e.Name(), "train.jsonl") && !strings.HasPrefix(e.Name(), "valid.jsonl") && !strings.HasPrefix(e.Name(), "test.jsonl") {
			count++
		}
	}
	return count
}
