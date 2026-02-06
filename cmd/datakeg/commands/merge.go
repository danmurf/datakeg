package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danmurf/datakeg/internal/writer"
)

// ExecuteMergePipeline merges per-document JSONL files into master train/valid/test files.
// It scans the output directory for files matching {name}_{split}.jsonl pattern,
// collects all pairs, and writes consolidated master files.
func ExecuteMergePipeline(outputDir string) error {
	// Verify output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", outputDir)
	}

	fmt.Printf("Merging per-document files in %s...\n", outputDir)

	// Define split types to look for
	splits := []string{"train", "valid", "test"}

	// Track totals across all splits
	totalPairsBySplit := make(map[string]int)

	// Process each split type
	for _, split := range splits {
		pairs, filesProcessed := mergeSplitFiles(outputDir, split)
		totalPairsBySplit[split] = len(pairs)

		if len(pairs) > 0 {
			// Write master file for this split
			masterFile := filepath.Join(outputDir, split+".jsonl")
			if err := writer.WriteJSONL(masterFile, pairs); err != nil {
				return fmt.Errorf("write %s.jsonl: %w", split, err)
			}
			fmt.Printf("  Merged %d pairs from %d files into %s.jsonl\n", len(pairs), filesProcessed, split)
		} else {
			fmt.Printf("  No pairs found for %s split (no master file created)\n", split)
		}
	}

	// Summary
	fmt.Printf("Merge complete!\n")
	fmt.Printf("  Train: %d total pairs\n", totalPairsBySplit["train"])
	fmt.Printf("  Valid: %d total pairs\n", totalPairsBySplit["valid"])
	fmt.Printf("  Test:  %d total pairs\n", totalPairsBySplit["test"])

	return nil
}

// mergeSplitFiles finds all per-document files for a given split and merges them.
func mergeSplitFiles(outputDir string, split string) ([]writer.TrainingPair, int) {
	pattern := "*_" + split + ".jsonl"

	// Find all matching files
	matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
	if err != nil {
		fmt.Printf("  Error searching for %s files: %v\n", split, err)
		return nil, 0
	}

	var allPairs []writer.TrainingPair
	filesProcessed := 0

	for _, filePath := range matches {
		// Skip if this looks like a master file (no underscore prefix)
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, split+".") {
			continue // This is a master file like train.jsonl, skip it
		}

		// Read pairs from this file
		pairs, err := readJSONLFile(filePath)
		if err != nil {
			fmt.Printf("  Warning: failed to read %s: %v\n", fileName, err)
			continue
		}

		if len(pairs) > 0 {
			allPairs = append(allPairs, pairs...)
			filesProcessed++
		}
	}

	return allPairs, filesProcessed
}

// readJSONLFile reads a JSONL file and returns all TrainingPair objects.
func readJSONLFile(filePath string) ([]writer.TrainingPair, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Parse each line as JSON
	lines := strings.Split(string(data), "\n")
	var pairs []writer.TrainingPair

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		var pair writer.TrainingPair
		if err := json.Unmarshal([]byte(line), &pair); err != nil {
			return nil, fmt.Errorf("parse line %d: %w", i+1, err)
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}
