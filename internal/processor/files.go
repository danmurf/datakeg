package processor

import (
	"os"
	"path/filepath"
	"strings"
)

// Document represents a loaded document for processing.
type Document struct {
	Name    string // Filename without extension
	Path    string // Full path to the file
	Content string // File contents
}

// LoadDocuments recursively finds .md and .txt files in the given directory
// and returns them as a slice of Document structs.
func LoadDocuments(rootDir string) ([]Document, error) {
	var documents []Document

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		// Read file contents
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Get filename without extension for the Name field
		baseName := filepath.Base(path)
		name := strings.TrimSuffix(baseName, ext)

		documents = append(documents, Document{
			Name:    name,
			Path:    path,
			Content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return documents, nil
}
