package commands

import (
	"fmt"
	"text/template"

	"github.com/danmurf/datakeg/internal/converter"
	"github.com/danmurf/datakeg/internal/generator"
	"github.com/danmurf/datakeg/internal/templates"
)

// ExecuteConvertPipeline orchestrates the JSONL format conversion pipeline.
// It loads a template, detects or validates the source format, validates the template
// compatibility, and then performs the conversion.
func ExecuteConvertPipeline(inputPath, outputPath, templateName, customTemplatePath, sourceFormat string) error {
	fmt.Printf("Converting %s...\n", inputPath)

	// Load template (built-in or custom)
	var tmpl *template.Template
	var err error
	if customTemplatePath != "" {
		tmpl, err = templates.LoadCustomConversionTemplate(customTemplatePath)
		if err != nil {
			return fmt.Errorf("load custom template: %w", err)
		}
		fmt.Printf("Using custom template: %s\n", customTemplatePath)
	} else {
		tmpl, err = templates.LoadConversionTemplate(templateName)
		if err != nil {
			return fmt.Errorf("load template '%s': %w. Run 'datakeg convert --list-templates' to see available templates", templateName, err)
		}
		fmt.Printf("Using built-in template: %s\n", templateName)
	}

	// Detect or parse source format
	var format generator.FormatType
	if sourceFormat != "" {
		format, err = generator.ParseFormat(sourceFormat)
		if err != nil {
			return err
		}
		fmt.Printf("Source format: %s (specified)\n", format)
	} else {
		format, err = converter.DetectFormatFromFile(inputPath)
		if err != nil {
			return fmt.Errorf("could not detect source format: %w. Specify --source-format explicitly", err)
		}
		fmt.Printf("Detected source format: %s\n", format)
	}

	// Validate template compatibility
	if err := converter.ValidateTemplate(tmpl, format); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Convert
	linesConverted, err := converter.ConvertJSONL(inputPath, outputPath, tmpl, format)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	fmt.Printf("Conversion complete: %s (%d lines converted)\n", outputPath, linesConverted)
	return nil
}

// ListBuiltinTemplates prints the available built-in conversion templates with descriptions.
func ListBuiltinTemplates() error {
	templateNames, err := templates.ListConversionTemplates()
	if err != nil {
		return fmt.Errorf("list templates: %w", err)
	}

	type templateInfo struct {
		format string
		desc   string
	}
	info := map[string]templateInfo{
		"mistral-instruct": {"completion", "Mistral Instruct format with [INST] tags"},
		"llama3-instruct":  {"chat", "Llama 3 Instruct with header tags"},
		"chatml":           {"chat", "ChatML format with <|im_start|> tags"},
		"deepseek-r1":      {"reasoning", "DeepSeek-R1 integrated reasoning format"},
	}

	fmt.Println("Built-in conversion templates:")
	fmt.Println()

	for _, name := range templateNames {
		if ti, ok := info[name]; ok {
			fmt.Printf("  %-20s (%-10s) - %s\n", name, ti.format, ti.desc)
		} else {
			fmt.Printf("  %-20s                   - (no description)\n", name)
		}
	}

	fmt.Println()
	fmt.Println("Usage: datakeg convert --template <name> <input.jsonl> <output.jsonl>")
	fmt.Println("Custom: datakeg convert --custom-template <file.tmpl> <input.jsonl> <output.jsonl>")

	return nil
}
