package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danmurf/datakeg/cmd/datakeg/commands"
)

// Build information set via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	goVersion = "unknown"
)

var RootCmd = &cobra.Command{
	Use:   "datakeg",
	Short: "Generate training datasets from documentation",
	Long: `datakeg transforms raw documentation into LLM training datasets.

Process markdown and text files from a source directory,
generate question-answer pairs using Ollama, and output
train/valid/test JSONL files.`,
	Version: version,
}

// Configuration flags for the generate command.
var (
	flagProvider              string
	flagModel                 string
	flagFormat                string
	flagSystemMessage         string
	flagReasoningFormat       string
	flagTrainPct              float64
	flagValidPct              float64
	flagTestPct               float64
	flagPairsPer1K            float64
	flagTimeout               int
	flagSkipMerge             bool
	flagYes                   bool
	flagDryRun                bool
	flagConvertTemplate       string
	flagConvertCustomTemplate string
	flagConvertSourceFormat   string
	flagConvertListTemplates  bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <source> <output>",
	Short: "Generate train/valid/test datasets",
	Long: `Process markdown and text files from source directory,
generate question-answer pairs using the configured LLM provider, and output
train/valid/test JSONL files to output directory.

Use --provider to select the LLM provider (ollama for local, openrouter for cloud).`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerate,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("datakeg %s\n", version)
		fmt.Printf("  commit:     %s\n", commit)
		fmt.Printf("  built:      %s\n", date)
		fmt.Printf("  go version: %s\n", goVersion)
	},
}

var mergeCmd = &cobra.Command{
	Use:   "merge <output>",
	Short: "Merge per-document JSONL files into master train/valid/test files",
	Long: `Merge per-document files (generated with --skip-merge) into master train.jsonl,
valid.jsonl, and test.jsonl files.`,
	Args: cobra.ExactArgs(1),
	RunE: runMerge,
}

var listProvidersCmd = &cobra.Command{
	Use:   "list-providers",
	Short: "List available LLM providers and their configuration status",
	Long: `List available LLM providers and their configuration status.
Shows whether each provider is configured and lists available models.`,
	RunE: runListProviders,
}

var convertCmd = &cobra.Command{
	Use:   "convert <input.jsonl> <output.jsonl>",
	Short: "Convert JSONL files to model-specific training formats",
	Long: `Convert generated JSONL files into model-specific training formats
using built-in or custom conversion templates.

Supports auto-detection of source format (completion, chat, reasoning) from the
JSONL structure. Use --template for built-in formats or --custom-template for
your own template files.

Examples:
  datakeg convert --template mistral-instruct train.jsonl train_mistral.jsonl
  datakeg convert --custom-template my-format.tmpl train.jsonl train_custom.jsonl
  datakeg convert --list-templates`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runConvert,
}

func init() {
	// Persistent flags available to all subcommands
	RootCmd.PersistentFlags().StringVarP(&flagModel, "model", "m", "gpt-oss:20b", "Model to use (provider-specific)")
	RootCmd.PersistentFlags().StringVarP(&flagProvider, "provider", "", "ollama", "LLM provider to use (ollama, openrouter)")

	// Local flags for generate command only
	generateCmd.Flags().Float64VarP(&flagTrainPct, "train-pct", "", 0.6, "Training set percentage (0.0-1.0)")
	generateCmd.Flags().Float64VarP(&flagValidPct, "valid-pct", "", 0.2, "Validation set percentage (0.0-1.0)")
	generateCmd.Flags().Float64VarP(&flagTestPct, "test-pct", "", 0.2, "Test set percentage (0.0-1.0)")
	generateCmd.Flags().Float64VarP(&flagPairsPer1K, "pairs-per-1k", "", 10.0, "Target pairs per 1000 characters")
	generateCmd.Flags().IntVarP(&flagTimeout, "timeout", "t", 60, "Per-document timeout in minutes")
	generateCmd.Flags().BoolVarP(&flagSkipMerge, "skip-merge", "", false, "Skip merging per-document files into master files")
	generateCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts (auto-confirm cost estimates)")
	generateCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "", false, "Print cost estimate and exit without generating")
	generateCmd.Flags().StringVarP(&flagFormat, "format", "f", "completion", "Output format (completion, chat, reasoning)")
	generateCmd.Flags().StringVarP(&flagSystemMessage, "system-message", "", "", "System message to include in chat format output")
	generateCmd.Flags().StringVarP(&flagReasoningFormat, "reasoning-format", "", "separate", "Reasoning output format: 'separate' (question/reasoning/answer fields) or 'integrated' (prompt/completion with inline tags)")

	// Convert command flags
	convertCmd.Flags().StringVarP(&flagConvertTemplate, "template", "t", "", "Built-in conversion template name (e.g., mistral-instruct)")
	convertCmd.Flags().StringVarP(&flagConvertCustomTemplate, "custom-template", "", "", "Path to custom conversion template file")
	convertCmd.Flags().StringVarP(&flagConvertSourceFormat, "source-format", "", "", "Source JSONL format (completion, chat, reasoning). Auto-detected if omitted")
	convertCmd.Flags().BoolVarP(&flagConvertListTemplates, "list-templates", "l", false, "List available built-in conversion templates")

	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(mergeCmd)
	RootCmd.AddCommand(listProvidersCmd)
	RootCmd.AddCommand(convertCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]
	outputPath := args[1]

	fmt.Printf("Generate command invoked with:\n")
	fmt.Printf("  Source:   %s\n", sourcePath)
	fmt.Printf("  Output:   %s\n", outputPath)
	fmt.Printf("  Provider: %s\n", flagProvider)
	fmt.Printf("  Model:    %s\n", flagModel)
	fmt.Printf("  Format:   %s\n", flagFormat)
	if flagFormat == "chat" && flagSystemMessage != "" {
		fmt.Printf("  System Message: %s\n", flagSystemMessage)
	}
	if flagFormat == "reasoning" {
		fmt.Printf("  Reasoning Format: %s\n", flagReasoningFormat)
	}
	fmt.Printf("  Splits:   train=%.0f%%, valid=%.0f%%, test=%.0f%%\n", flagTrainPct*100, flagValidPct*100, flagTestPct*100)
	fmt.Printf("  Pairs per 1K chars: %.1f\n", flagPairsPer1K)

	return commands.ExecuteGeneratePipeline(sourcePath, outputPath, flagProvider, flagModel, flagFormat, flagSystemMessage, flagReasoningFormat, flagPairsPer1K, flagValidPct, flagTestPct, flagTimeout, flagSkipMerge, flagYes, flagDryRun)
}

func runMerge(cmd *cobra.Command, args []string) error {
	outputPath := args[0]

	fmt.Printf("Merge command invoked with:\n")
	fmt.Printf("  Output:  %s\n", outputPath)

	return commands.ExecuteMergePipeline(outputPath)
}

func runListProviders(cmd *cobra.Command, args []string) error {
	return commands.ExecuteListProviders()
}

func runConvert(cmd *cobra.Command, args []string) error {
	// Handle --list-templates
	if flagConvertListTemplates {
		return commands.ListBuiltinTemplates()
	}

	// Normal conversion requires exactly 2 args
	if len(args) != 2 {
		return fmt.Errorf("convert requires exactly 2 arguments: <input.jsonl> <output.jsonl>\nRun 'datakeg convert --help' for usage")
	}

	inputPath := args[0]
	outputPath := args[1]

	// Require either --template or --custom-template
	if flagConvertTemplate == "" && flagConvertCustomTemplate == "" {
		return fmt.Errorf("specify a template with --template <name> or --custom-template <file>\nRun 'datakeg convert --list-templates' to see available templates")
	}

	// Cannot use both
	if flagConvertTemplate != "" && flagConvertCustomTemplate != "" {
		return fmt.Errorf("specify either --template or --custom-template, not both")
	}

	fmt.Printf("Convert command invoked with:\n")
	fmt.Printf("  Input:  %s\n", inputPath)
	fmt.Printf("  Output: %s\n", outputPath)
	if flagConvertTemplate != "" {
		fmt.Printf("  Template: %s (built-in)\n", flagConvertTemplate)
	} else {
		fmt.Printf("  Template: %s (custom)\n", flagConvertCustomTemplate)
	}
	if flagConvertSourceFormat != "" {
		fmt.Printf("  Source format: %s\n", flagConvertSourceFormat)
	} else {
		fmt.Printf("  Source format: (auto-detect)\n")
	}

	return commands.ExecuteConvertPipeline(inputPath, outputPath, flagConvertTemplate, flagConvertCustomTemplate, flagConvertSourceFormat)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
