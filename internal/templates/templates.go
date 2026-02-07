// Package templates provides embedded prompt templates for dataset generation.
// Templates are embedded at build time using the //go:embed directive.
package templates

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptFS embed.FS

//go:embed conversions/*.tmpl
var conversionFS embed.FS

// ExcludePair represents a previously generated pair to exclude from new generation.
// This avoids duplicate pairs across dataset splits.
type ExcludePair struct {
	Prompt     string
	Completion string
}

// PromptData contains the data passed to template execution.
type PromptData struct {
	DocumentContent string
	PairCount       int
	DocumentName    string // Reserved for future use
	ExcludePairs    []ExcludePair
}

// ExecuteTemplate loads and executes the specified template with the given data.
// The name parameter should be one of: "train.tmpl", "valid.tmpl", or "test.tmpl".
//
// It returns the executed template as a string, or an error if the template
// cannot be loaded or executed.
func ExecuteTemplate(name string, data PromptData) (string, error) {
	// Parse the template from the embedded filesystem
	tmpl, err := template.ParseFS(promptFS, "prompts/"+name)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}

	// Execute the template with the provided data
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// jsonEscape escapes special JSON characters in a string.
// This is used in conversion templates to ensure template output is valid JSON.
func jsonEscape(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	// json.Marshal returns a quoted string, so we need to remove the quotes
	return string(b[1 : len(b)-1])
}

// LoadConversionTemplate loads a built-in conversion template by name.
// The name should be one of: "mistral-instruct", "llama3-instruct", "chatml", or "deepseek-r1".
// The template is loaded from the embedded filesystem.
func LoadConversionTemplate(name string) (*template.Template, error) {
	content, err := conversionFS.ReadFile("conversions/" + name + ".tmpl")
	if err != nil {
		return nil, fmt.Errorf("read embedded template %s: %w", name, err)
	}

	tmpl, err := template.New(name).
		Funcs(template.FuncMap{"jsonEscape": jsonEscape}).
		Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	return tmpl, nil
}

// LoadCustomConversionTemplate loads a user-provided conversion template from a file path.
// The template file should contain Go template syntax.
func LoadCustomConversionTemplate(filePath string) (*template.Template, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", filePath)
	}

	// Parse the template from the file
	tmpl, err := template.New(filepath.Base(filePath)).
		Funcs(template.FuncMap{"jsonEscape": jsonEscape}).
		ParseFiles(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse template file %s: %w", filePath, err)
	}
	return tmpl, nil
}

// ListConversionTemplates returns a slice of available built-in conversion template names.
func ListConversionTemplates() ([]string, error) {
	entries, err := conversionFS.ReadDir("conversions")
	if err != nil {
		return nil, fmt.Errorf("read conversion templates directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			// Remove .tmpl extension
			name := strings.TrimSuffix(entry.Name(), ".tmpl")
			names = append(names, name)
		}
	}
	return names, nil
}
