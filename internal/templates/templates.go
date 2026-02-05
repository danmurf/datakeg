// Package templates provides embedded prompt templates for dataset generation.
// Templates are embedded at build time using the //go:embed directive.
package templates

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompts/*.tmpl
var promptFS embed.FS

// PromptData contains the data passed to template execution.
type PromptData struct {
	DocumentContent string
	PairCount       int
	DocumentName    string // Reserved for future use
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
