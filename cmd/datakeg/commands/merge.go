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
		return fmt.Errorf("directory '%s' does not exist\nRun 'datakeg generate --skip-merge <source> <output>' first to create per-document files.", outputDir)
	}

	fmt.Printf("Merging per-document files in %s...\n", outputDir)

	// Define split types to look for
	splits := []string{"train", "valid", "test"}

	// Track totals across all splits
	totalPairsBySplit := make(map[string]int)
	totalFilesFound := 0

	fmt.Printf("Scanning for per-document files...\n")

	// First pass: count all per-document files
	for _, split := range splits {
		pattern := "*_" + split + ".jsonl"
		matches, _ := filepath.Glob(filepath.Join(outputDir, pattern))
		for _, filePath := range matches {
			fileName := filepath.Base(filePath)
			if !strings.HasPrefix(fileName, split+".") {
				totalFilesFound++
			}
		}
	}

	if totalFilesFound == 0 {
		return fmt.Errorf("no per-document files (*_train.jsonl, *_valid.jsonl, *_test.jsonl) found in %s\nRun 'datakeg generate --skip-merge' first to create them.", outputDir)
	}

	fmt.Printf("  Found %d train, %d valid, %d test files\n",
		countFiles(outputDir, "train"),
		countFiles(outputDir, "valid"),
		countFiles(outputDir, "test"))

	// Process each split type
	for _, split := range splits {
		pairs, filesProcessed := mergeSplitFiles(outputDir, split)
		totalPairsBySplit[split] = len(pairs)

		if len(pairs) > 0 {
			// Write master file for this split
			masterFile := filepath.Join(outputDir, split+".jsonl")
			if err := writer.WriteJSONL(masterFile, pairs); err != nil {
				return fmt.Errorf("could not write %s.jsonl\nCheck disk space and write permissions.", split)
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

	// Count files to report
	for _, filePath := range matches {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, split+".") {
			continue // This is a master file like train.jsonl, skip it
		}
		filesProcessed++
	}

	fmt.Printf("  Found %d %s files\n", filesProcessed, split)
	fmt.Printf("  Merging...\n")

	for _, filePath := range matches {
		// Skip if this looks like a master file (no underscore prefix)
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, split+".") {
			continue // This is a master file like train.jsonl, skip it
		}

		// Read pairs from this file
		pairs, err := readJSONLFile(filePath)
		if err != nil {
			fmt.Printf("  Warning: could not read %s: %v\nThe file may be corrupted or permissions are incorrect.\n", fileName, err)
			continue
		}

		if len(pairs) > 0 {
			allPairs = append(allPairs, pairs...)
		}
	}

	return allPairs, filesProcessed
}

// readJSONLFile reads a JSONL file and returns all TrainingPair objects.
func readJSONLFile(filePath string) ([]writer.TrainingPair, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read %s\nThe file may be corrupted or permissions are incorrect.", filepath.Base(filePath))
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

// countFiles counts per-document JSONL files for a given split type.
func countFiles(outputDir string, split string) int {
	pattern := "*_" + split + ".jsonl"
	matches, _ := filepath.Glob(filepath.Join(outputDir, pattern))
	count := 0
	for _, filePath := range matches {
		fileName := filepath.Base(filePath)
		if !strings.HasPrefix(fileName, split+".") {
			count++
		}
	}
	return count
}
