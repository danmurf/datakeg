package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/processor"
	"github.com/danmurf/datakeg/internal/provider"
	"github.com/danmurf/datakeg/internal/usage"
	"github.com/danmurf/datakeg/internal/writer"
)

// ExecuteGeneratePipeline orchestrates the full generate pipeline:
// 1. Load documents from source directory
// 2. Create provider (Ollama or OpenRouter)
// 3. Generate prompt/completion pairs for each document (train → valid → test with exclusions)
// 4. Write per-document JSONL files during processing
// 5. Optionally merge into master files (skipable via skipMerge flag)
func ExecuteGeneratePipeline(
	sourceDir string,
	outputDir string,
	providerType string,
	model string,
	pairsPer1K float64,
	validPct float64,
	testPct float64,
	timeoutMinutes int,
	skipMerge bool,
	skipConfirm bool,
	dryRun bool,
) error {
	fmt.Printf("Starting pipeline...\n")

	// Step 1: Load documents from source directory
	fmt.Printf("Loading documents from %s...\n", sourceDir)
	documents, err := processor.LoadDocuments(sourceDir)
	if err != nil {
		return fmt.Errorf("could not read source directory: %s. Check that the path exists and you have read permissions", sourceDir)
	}
	fmt.Printf("Loaded %d documents\n", len(documents))

	if len(documents) == 0 {
		return fmt.Errorf("no markdown (.md) or text (.txt) files found in %s. Add some documentation files to the source directory and try again", sourceDir)
	}

	// Validate model for OpenRouter
	if providerType == string(provider.ProviderOpenRouter) && model == "" {
		return fmt.Errorf("openrouter requires an explicit model selection. Use --model to specify a model (e.g., --model meta-llama/llama-3.1-70b-instruct). Run 'datakeg list-providers' to see available models")
	}

	// Step 2: Create provider
	fmt.Printf("Creating %s provider...\n", providerType)
	p, err := provider.NewProvider(provider.ProviderType(providerType))
	if err != nil {
		return fmt.Errorf("could not create %s provider: %w", providerType, err)
	}

	// Step 3: Create generator with config
	// CLI uses 0.0-1.0 range but config uses 0-100 percentage
	genConfig := generator.Config{
		PairsPer1KChars: pairsPer1K,
		ValidPercent:    validPct * 100,
		TestPercent:     testPct * 100,
		Model:           model,
	}
	gen := generator.NewGenerator(p, genConfig)

	// Warning for paid providers
	if providerType == string(provider.ProviderOpenRouter) {
		fmt.Println("\nWARNING: Using OpenRouter may incur API costs.")
		fmt.Println("You are responsible for all charges incurred through your API usage.")
		fmt.Println()

		if dryRun {
			fmt.Println("Dry run - exiting without generation.")
			return nil
		}

		if !skipConfirm {
			fmt.Print("Continue? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(input) != "y" && strings.TrimSpace(input) != "Y" {
				return fmt.Errorf("operation cancelled by user")
			}
		}
	}

	// Step 4: Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory %s. Check that you have write permissions and the path is valid", outputDir)
	}

	// Step 5: Process documents and collect pairs sequentially with cross-split exclusion
	// Train → Valid (exclude train) → Test (exclude train+valid)
	// Per-document pairs reset between documents (no cross-document dedup)
	var trainPairs []writer.TrainingPair
	var validPairs []writer.TrainingPair
	var testPairs []writer.TrainingPair

	// Track usage for paid providers
	usageTracker := &usage.Tracker{}

	for i, doc := range documents {
		docPath, _ := filepath.Rel(sourceDir, doc.Path)
		fmt.Printf("[%d/%d] Processing: %s (%d chars)\n", i+1, len(documents), docPath, len(doc.Content))

		// Per-document timeout so each document gets the full timeout duration
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMinutes)*time.Minute)

		// Get pair counts from the generator (single source of truth
		totalPairs := gen.CalculatePairs(doc.Content)
		pairCounts := gen.CalculateSplitCounts(totalPairs)
		trainCount := pairCounts.Train
		validCount := pairCounts.Valid
		testCount := pairCounts.Test

		fmt.Printf("     → Target: %d train, %d valid, %d test pairs\n", trainCount, validCount, testCount)

		// Per-document pair tracking (resets each iteration for no cross-document dedup)
		var docTrainPairs []generator.Pair
		var docValidPairs []generator.Pair

		// Step 5a: Generate train pairs (no exclusions)
		if trainCount > 0 {
			fmt.Printf("     → Generating train pairs...\n")
			pairs, usage, err := gen.Generate(ctx, &doc, generator.SplitTrain, nil)
			if err != nil {
				cancel()
				return fmt.Errorf("failed to generate training pairs for %s. This may be a temporary issue. Try again, or try a different model", doc.Name)
			}
			docTrainPairs = pairs
			usageTracker.Add(usage)
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
					cancel()
					return fmt.Errorf("failed to write %s. Check disk space and write permissions", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}

		// Step 5b: Generate valid pairs (exclude train pairs)
		if validCount > 0 {
			fmt.Printf("     → Generating valid pairs (excluding %d train pairs)...\n", len(docTrainPairs))
			pairs, usage, err := gen.Generate(ctx, &doc, generator.SplitValid, docTrainPairs)
			if err != nil {
				cancel()
				return fmt.Errorf("failed to generate validation pairs for %s. This may be a temporary issue. Try again, or try a different model", doc.Name)
			}
			docValidPairs = pairs
			usageTracker.Add(usage)
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
					cancel()
					return fmt.Errorf("failed to write %s. Check disk space and write permissions", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}

		// Step 5c: Generate test pairs (exclude train + valid pairs)
		if testCount > 0 {
			// Combine train and valid exclusions
			allExclude := append(docTrainPairs, docValidPairs...)
			fmt.Printf("     → Generating test pairs (excluding %d train+valid pairs)...\n", len(allExclude))
			pairs, usage, err := gen.Generate(ctx, &doc, generator.SplitTest, allExclude)
			if err != nil {
				cancel()
				return fmt.Errorf("failed to generate test pairs for %s. This may be a temporary issue. Try again, or try a different model", doc.Name)
			}
			usageTracker.Add(usage)
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
					cancel()
					return fmt.Errorf("failed to write %s. Check disk space and write permissions", filepath.Base(perDocFile))
				}
				fmt.Printf("     → Written %d pairs to %s\n", len(pairs), filepath.Base(perDocFile))
			}
		}

		cancel()
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
			return fmt.Errorf("failed to write train.jsonl. Check disk space and write permissions")
		}
		fmt.Printf("  Written %d pairs to train.jsonl\n", len(trainPairs))
	}

	if len(validPairs) > 0 {
		outFile := filepath.Join(outputDir, "valid.jsonl")
		if err := writer.WriteJSONL(outFile, validPairs); err != nil {
			return fmt.Errorf("failed to write valid.jsonl. Check disk space and write permissions")
		}
		fmt.Printf("  Written %d pairs to valid.jsonl\n", len(validPairs))
	}

	if len(testPairs) > 0 {
		outFile := filepath.Join(outputDir, "test.jsonl")
		if err := writer.WriteJSONL(outFile, testPairs); err != nil {
			return fmt.Errorf("failed to write test.jsonl. Check disk space and write permissions")
		}
		fmt.Printf("  Written %d pairs to test.jsonl\n", len(testPairs))
	}

	fmt.Printf("Pipeline complete!\n")
	fmt.Printf("Total pairs: train=%d, valid=%d, test=%d\n", len(trainPairs), len(validPairs), len(testPairs))

	// Post-run summary for paid providers
	if providerType == string(provider.ProviderOpenRouter) {
		fmt.Printf("\nTokens used: %d prompt + %d completion = %d total\n",
			usageTracker.TotalPromptTokens, usageTracker.TotalCompletionTokens, usageTracker.TotalTokens)
	}

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
