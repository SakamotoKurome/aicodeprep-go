package formatter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/schollz/progressbar/v3"

	"aicodeprep-go/internal/selector"
)

// PromptFormatter formats the prompt with file contents
type PromptFormatter struct {
	prompt  string
	files   []selector.FileInfo
	verbose bool
}

// New creates a new PromptFormatter
func New(prompt string, files []selector.FileInfo, verbose bool) *PromptFormatter {
	return &PromptFormatter{
		prompt:  prompt,
		files:   files,
		verbose: verbose,
	}
}

// Format generates the structured prompt text
func (pf *PromptFormatter) Format() (string, error) {
	var result strings.Builder

	// Add user prompt at the beginning
	result.WriteString("=== 用户需求 ===\n")
	if pf.prompt != "" {
		result.WriteString(pf.prompt)
	} else {
		result.WriteString("请分析以下代码文件。")
	}
	result.WriteString("\n\n")

	// Add file contents section
	result.WriteString("=== 文件内容开始 ===\n")

	totalSize := int64(0)
	processedFiles := 0

	// Create progress bar if verbose mode and multiple files
	var bar *progressbar.ProgressBar
	if pf.verbose && len(pf.files) > 1 {
		bar = progressbar.NewOptions(len(pf.files),
			progressbar.OptionSetDescription("Processing files..."),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerHead:    "█",
				SaucerPadding: "░",
				BarStart:      "│",
				BarEnd:        "│",
			}))
	}

	for i, file := range pf.files {
		if bar != nil {
			bar.Set(i)
		}

		content, err := pf.readFileContent(file.Path)
		if err != nil {
			if pf.verbose {
				fmt.Fprintf(os.Stderr, "\nWarning: Failed to read file %s: %v\n", file.Path, err)
			}
			continue
		}

		// Skip empty files
		if strings.TrimSpace(content) == "" {
			if pf.verbose {
				fmt.Fprintf(os.Stderr, "\nSkipping empty file: %s\n", file.Path)
			}
			continue
		}

		// Add file header with relative path for better readability
		displayPath := GetRelativePath(file.Path)
		result.WriteString(fmt.Sprintf("--- 文件: %s ---\n", displayPath))
		result.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n")

		totalSize += file.Size
		processedFiles++
	}

	if bar != nil {
		bar.Set(len(pf.files))
		bar.Close()
		fmt.Fprintf(os.Stderr, "\n")
	}

	result.WriteString("=== 文件内容结束 ===\n\n")

	// Add user prompt at the end again
	result.WriteString("=== 用户需求 ===\n")
	if pf.prompt != "" {
		result.WriteString(pf.prompt)
	} else {
		result.WriteString("请分析以上代码文件。")
	}
	result.WriteString("\n")

	if pf.verbose {
		fmt.Fprintf(os.Stderr, "Processed %d files, total size: %s\n",
			processedFiles, formatBytes(totalSize))
	}

	return result.String(), nil
}

// readFileContent reads and validates file content
func (pf *PromptFormatter) readFileContent(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	content, err := readFileAsString(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Validate UTF-8 encoding
	if !utf8.ValidString(content) {
		return "", fmt.Errorf("file contains invalid UTF-8 encoding")
	}

	return content, nil
}

// readFileAsString reads file content and handles different line endings
func readFileAsString(file *os.File) (string, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(file)

	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		result.WriteString(line)
		result.WriteString("\n")
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

// GetSummary returns a summary of what will be processed
func (pf *PromptFormatter) GetSummary() string {
	var result strings.Builder

	result.WriteString("Files to be processed:\n")

	totalSize := int64(0)
	for i, file := range pf.files {
		result.WriteString(fmt.Sprintf("%d. %s (%s)\n",
			i+1, file.Path, formatBytes(file.Size)))
		totalSize += file.Size
	}

	result.WriteString(fmt.Sprintf("\nTotal: %d files, %s\n",
		len(pf.files), formatBytes(totalSize)))

	if pf.prompt != "" {
		result.WriteString(fmt.Sprintf("\nPrompt: %s\n", pf.prompt))
	}

	return result.String()
}

// formatBytes formats byte count into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ValidateFiles checks if files exist and are readable
func ValidateFiles(files []selector.FileInfo) []selector.FileInfo {
	var validFiles []selector.FileInfo

	for _, file := range files {
		if info, err := os.Stat(file.Path); err == nil && info.Mode().IsRegular() {
			// Check if file is readable
			if f, err := os.Open(file.Path); err == nil {
				f.Close()
				validFiles = append(validFiles, file)
			}
		}
	}

	return validFiles
}

// GetRelativePath converts absolute path to relative path if possible
func GetRelativePath(path string) string {
	if wd, err := os.Getwd(); err == nil {
		if relPath, err := filepath.Rel(wd, path); err == nil {
			// Only use relative path if it's shorter and doesn't start with ../..
			if len(relPath) < len(path) && !strings.HasPrefix(relPath, "../..") {
				return relPath
			}
		}
	}
	return path
}