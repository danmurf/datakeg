package processor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDocuments(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string // Returns temp dir path
		wantCount int
		wantErr   bool
		validate  func(t *testing.T, docs []Document)
	}{
		{
			name: "load single markdown file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := "# Test Document\n\nThis is a test."
				err := os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				if docs[0].Name != "test" {
					t.Errorf("got Name = %q, want %q", docs[0].Name, "test")
				}
				if !filepath.IsAbs(docs[0].Path) {
					t.Error("Path should be absolute")
				}
				if docs[0].Content != "# Test Document\n\nThis is a test." {
					t.Errorf("got Content = %q, want %q", docs[0].Content, "# Test Document\n\nThis is a test.")
				}
			},
		},
		{
			name: "load single text file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				content := "Plain text content"
				err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				if docs[0].Name != "notes" {
					t.Errorf("got Name = %q, want %q", docs[0].Name, "notes")
				}
				if docs[0].Content != "Plain text content" {
					t.Errorf("got Content = %q, want %q", docs[0].Content, "Plain text content")
				}
			},
		},
		{
			name: "load multiple files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := map[string]string{
					"doc1.md":  "Content 1",
					"doc2.md":  "Content 2",
					"doc3.txt": "Content 3",
				}
				for name, content := range files {
					err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				return tmpDir
			},
			wantCount: 3,
			wantErr:   false,
			validate:  nil,
		},
		{
			name: "skip non-md-txt files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := map[string]string{
					"doc.md":    "Markdown",
					"notes.txt": "Text",
					"code.go":   "Go code",
					"data.json": "{}",
					"README":    "No extension",
				}
				for name, content := range files {
					err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				return tmpDir
			},
			wantCount: 2,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				// Should only have doc.md and notes.txt
				for _, doc := range docs {
					if doc.Name != "doc" && doc.Name != "notes" {
						t.Errorf("unexpected document: %q", doc.Name)
					}
				}
			},
		},
		{
			name: "recursive directory traversal",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create nested structure
				subDir := filepath.Join(tmpDir, "subdir")
				deepDir := filepath.Join(subDir, "deep")
				if err := os.MkdirAll(deepDir, 0755); err != nil {
					t.Fatal(err)
				}
				files := map[string]string{
					filepath.Join(tmpDir, "root.md"):     "Root",
					filepath.Join(subDir, "sub.txt"):     "Subdir",
					filepath.Join(deepDir, "deep.md"):    "Deep",
				}
				for path, content := range files {
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						t.Fatal(err)
					}
				}
				return tmpDir
			},
			wantCount: 3,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				names := make(map[string]bool)
				for _, doc := range docs {
					names[doc.Name] = true
				}
				expected := []string{"root", "sub", "deep"}
				for _, name := range expected {
					if !names[name] {
						t.Errorf("missing document: %q", name)
					}
				}
			},
		},
		{
			name: "empty directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantCount: 0,
			wantErr:   false,
			validate:  nil,
		},
		{
			name: "case insensitive extensions",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := map[string]string{
					"upper.MD":  "Upper MD",
					"upper.TXT": "Upper TXT",
					"mixed.Md":  "Mixed Md",
					"mixed.Txt": "Mixed Txt",
				}
				for name, content := range files {
					err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				return tmpDir
			},
			wantCount: 4,
			wantErr:   false,
			validate:  nil,
		},
		{
			name: "empty files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				err := os.WriteFile(filepath.Join(tmpDir, "empty.md"), []byte(""), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantCount: 1,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				if docs[0].Content != "" {
					t.Errorf("got Content = %q, want empty string", docs[0].Content)
				}
			},
		},
		{
			name: "nonexistent directory",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/path/that/does/not/exist"
			},
			wantCount: 0,
			wantErr:   true,
			validate:  nil,
		},
		{
			name: "files with dots in name",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				files := map[string]string{
					"my.file.name.md":  "Dots in name",
					"version.2.0.txt": "Version file",
				}
				for name, content := range files {
					err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
					if err != nil {
						t.Fatal(err)
					}
				}
				return tmpDir
			},
			wantCount: 2,
			wantErr:   false,
			validate: func(t *testing.T, docs []Document) {
				names := make(map[string]bool)
				for _, doc := range docs {
					names[doc.Name] = true
				}
				// Should strip only the final extension
				expected := []string{"my.file.name", "version.2.0"}
				for _, name := range expected {
					if !names[name] {
						t.Errorf("missing document with name: %q", name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := tt.setupFunc(t)
			docs, err := LoadDocuments(rootDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadDocuments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(docs) != tt.wantCount {
					t.Errorf("LoadDocuments() got %d documents, want %d", len(docs), tt.wantCount)
				}

				if tt.validate != nil {
					tt.validate(t, docs)
				}
			}
		})
	}
}
