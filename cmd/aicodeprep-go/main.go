package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"aicodeprep-go/internal/clipboard"
	"aicodeprep-go/internal/config"
	"aicodeprep-go/internal/formatter"
	"aicodeprep-go/internal/interactive"
	"aicodeprep-go/internal/selector"
)

var (
	files       []string
	excludes    []string
	prompt      string
	interactive_mode bool
	output      string
	configPath  string
	dryRun      bool
	maxSize     int64
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "aicodeprep-go",
	Short: "Generate LLM prompts with multiple code files",
	Long: `aicodeprep-go is a command-line tool that helps developers quickly generate
LLM prompts containing multiple code file contents. It supports file selection
through patterns, exclusion rules, and interactive input.`,
	RunE: runCommand,
}

func init() {
	rootCmd.Flags().StringArrayVarP(&files, "files", "f", []string{}, "File patterns (can be used multiple times)")
	rootCmd.Flags().StringArrayVarP(&excludes, "exclude", "e", []string{}, "Exclude patterns (can be used multiple times)")
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt text")
	rootCmd.Flags().BoolVarP(&interactive_mode, "interactive", "i", false, "Interactive mode")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: clipboard)")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show files that would be processed")
	rootCmd.Flags().Int64Var(&maxSize, "max-size", 0, "Maximum file size in bytes (default: 1MB)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.DefaultConfig()
	if configPath != "" {
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		cfg = loadedCfg
	}

	// Set default max file size if not specified
	if maxSize == 0 {
		maxSize = cfg.MaxFileSize
	}

	// Merge command line options with config
	cfg.Merge(files, excludes, prompt, output, maxSize)

	// Handle interactive mode
	if interactive_mode {
		return runInteractiveMode(cfg)
	}

	// If no files specified and no config, ask for help
	if len(cfg.Files) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "No files specified, using current directory pattern\n")
		}
		cfg.Files = []string{"*"}
	}

	return runBatchMode(cfg)
}

func runInteractiveMode(cfg *config.Config) error {
	ih := interactive.New()

	// Get prompt if not provided
	if cfg.Prompt == "" {
		prompt, err := ih.GetPrompt()
		if err != nil {
			return fmt.Errorf("failed to get prompt: %w", err)
		}
		cfg.Prompt = prompt
	}

	// Get file patterns if not provided
	if len(cfg.Files) == 0 {
		patterns, err := ih.GetFilePatterns()
		if err != nil {
			return fmt.Errorf("failed to get file patterns: %w", err)
		}
		cfg.Files = patterns
	}

	// Get exclude patterns
	excludes, err := ih.GetExcludePatterns()
	if err != nil {
		return fmt.Errorf("failed to get exclude patterns: %w", err)
	}
	if len(excludes) > 0 {
		cfg.Exclude = append(cfg.Exclude, excludes...)
	}

	// Get output path if not specified
	if cfg.Output == "" {
		outputPath, err := ih.GetOutputPath()
		if err != nil {
			return fmt.Errorf("failed to get output path: %w", err)
		}
		cfg.Output = outputPath
	}

	// Select files
	fs := selector.New(cfg.Files, cfg.Exclude, cfg.MaxFileSize)
	selectedFiles, err := fs.SelectFiles()
	if err != nil {
		return fmt.Errorf("failed to select files: %w", err)
	}

	if len(selectedFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No files found matching the patterns\n")
		return nil
	}

	// Confirm file selection
	confirmed, err := ih.ConfirmFileSelection(selectedFiles)
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		fmt.Fprintf(os.Stderr, "Operation cancelled\n")
		return nil
	}

	// Allow user to select specific files
	finalFiles, err := ih.SelectFromList(selectedFiles)
	if err != nil {
		return fmt.Errorf("failed to select from list: %w", err)
	}

	return generateOutput(cfg, finalFiles)
}

func runBatchMode(cfg *config.Config) error {
	// Select files
	fs := selector.New(cfg.Files, cfg.Exclude, cfg.MaxFileSize)
	selectedFiles, err := fs.SelectFiles()
	if err != nil {
		return fmt.Errorf("failed to select files: %w", err)
	}

	if len(selectedFiles) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "No files found matching the patterns:\n")
			for _, pattern := range cfg.Files {
				fmt.Fprintf(os.Stderr, "  - %s\n", pattern)
			}
		}
		return fmt.Errorf("no files found matching the patterns")
	}

	// Validate files
	validFiles := formatter.ValidateFiles(selectedFiles)
	if len(validFiles) != len(selectedFiles) {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: %d files were skipped (not readable or not regular files)\n",
				len(selectedFiles)-len(validFiles))
		}
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid files found")
	}

	// Dry run mode
	if dryRun {
		pf := formatter.New("", validFiles, verbose)
		fmt.Print(pf.GetSummary())
		return nil
	}

	return generateOutput(cfg, validFiles)
}

func generateOutput(cfg *config.Config, files []selector.FileInfo) error {
	// Format the prompt
	pf := formatter.New(cfg.Prompt, files, verbose)

	if verbose {
		fmt.Fprintf(os.Stderr, "Formatting %d files...\n", len(files))
	}

	formattedPrompt, err := pf.Format()
	if err != nil {
		return fmt.Errorf("failed to format prompt: %w", err)
	}

	// Write output
	if err := clipboard.WriteToOutput(formattedPrompt, cfg.Output, verbose); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if verbose && cfg.Output == "" {
		fmt.Fprintf(os.Stderr, "Prompt generated successfully with %d files\n", len(files))
	}

	return nil
}

// getConfigPath returns the default config path if not specified
func getConfigPath() string {
	if configPath != "" {
		return configPath
	}

	// Try to find config file in common locations
	candidates := []string{
		".aicodeprep.yaml",
		".aicodeprep.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "aicodeprep", "config.yaml"),
		filepath.Join(os.Getenv("HOME"), ".aicodeprep.yaml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}