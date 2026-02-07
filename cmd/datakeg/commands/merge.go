package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExecuteMergePipeline merges per-document JSONL files into master train/valid/test files.
// It scans the output directory for files matching {name}_{split}.jsonl pattern,
// collects all lines, and writes consolidated master files.
// This merge is format-agnostic - it concatenates lines as-is without parsing,
// allowing it to work with both completion format and chat format files.
func ExecuteMergePipeline(outputDir string) error {
	// Verify output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist. Run 'datakeg generate --skip-merge <source> <output>' first to create per-document files", outputDir)
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
		return fmt.Errorf("no per-document files (*_train.jsonl, *_valid.jsonl, *_test.jsonl) found in %s. Run 'datakeg generate --skip-merge' first to create them", outputDir)
	}

	fmt.Printf("  Found %d train, %d valid, %d test files\n",
		countFiles(outputDir, "train"),
		countFiles(outputDir, "valid"),
		countFiles(outputDir, "test"))

	// Process each split type using raw line concatenation (format-agnostic)
	for _, split := range splits {
		lines, filesProcessed := mergeSplitFilesRaw(outputDir, split)
		totalPairsBySplit[split] = lines

		if lines > 0 {
			// Write master file for this split
			masterFile := filepath.Join(outputDir, split+".jsonl")
			if err := writeLinesToFile(masterFile, outputDir, split); err != nil {
				return fmt.Errorf("could not write %s.jsonl. Check disk space and write permissions", split)
			}
			fmt.Printf("  Merged %d pairs from %d files into %s.jsonl\n", lines, filesProcessed, split)
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

// mergeSplitFilesRaw finds all per-document files for a given split and merges them using raw line concatenation.
func mergeSplitFilesRaw(outputDir string, split string) (int, int) {
	pattern := "*_" + split + ".jsonl"

	matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
	if err != nil {
		fmt.Printf("  Error searching for %s files: %v\n", split, err)
		return 0, 0
	}

	filesProcessed := 0
	lineCount := 0

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

		// Read file and count lines
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("  Warning: could not read %s: %v\nThe file may be corrupted or permissions are incorrect.\n", fileName, err)
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				lineCount++
			}
		}
	}

	return lineCount, filesProcessed
}

// writeLinesToFile writes all lines from per-document files to a master file.
func writeLinesToFile(masterFile string, outputDir string, split string) error {
	pattern := "*_" + split + ".jsonl"
	matches, _ := filepath.Glob(filepath.Join(outputDir, pattern))

	file, err := os.Create(masterFile)
	if err != nil {
		return fmt.Errorf("create file %s: %w", masterFile, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close file %s: %w", masterFile, closeErr)
		}
	}()

	for _, filePath := range matches {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, split+".") {
			continue // Skip master files
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", fileName, err)
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if _, err := file.WriteString(line + "\n"); err != nil {
					return fmt.Errorf("write line to %s: %w", masterFile, err)
				}
			}
		}
	}

	return nil
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
