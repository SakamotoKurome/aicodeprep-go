package interactive

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aicodeprep-go/internal/selector"
)

// InputHandler handles interactive user input
type InputHandler struct {
	scanner *bufio.Scanner
}

// New creates a new InputHandler
func New() *InputHandler {
	return &InputHandler{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// GetPrompt gets prompt input from user interactively
func (ih *InputHandler) GetPrompt() (string, error) {
	fmt.Print("请输入功能描述 (多行输入，空行结束):\n> ")

	var lines []string
	for {
		if !ih.scanner.Scan() {
			if err := ih.scanner.Err(); err != nil {
				return "", fmt.Errorf("failed to read input: %w", err)
			}
			break
		}

		line := ih.scanner.Text()
		if line == "" && len(lines) > 0 {
			break // Empty line ends input if we already have content
		}

		lines = append(lines, line)
		if line != "" {
			fmt.Print("> ")
		}
	}

	return strings.Join(lines, "\n"), nil
}

// GetFilePatterns gets file patterns from user interactively
func (ih *InputHandler) GetFilePatterns() ([]string, error) {
	fmt.Print("请输入文件模式 (如: *.go, src/**/*.js, 空行结束):\n> ")

	var patterns []string
	for {
		if !ih.scanner.Scan() {
			if err := ih.scanner.Err(); err != nil {
				return nil, fmt.Errorf("failed to read input: %w", err)
			}
			break
		}

		line := strings.TrimSpace(ih.scanner.Text())
		if line == "" {
			break // Empty line ends input
		}

		patterns = append(patterns, line)
		fmt.Print("> ")
	}

	// If no patterns provided, use current directory
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}

	return patterns, nil
}

// GetExcludePatterns gets exclude patterns from user interactively
func (ih *InputHandler) GetExcludePatterns() ([]string, error) {
	fmt.Print("请输入排除模式 (如: vendor/*, *_test.go, 空行结束):\n> ")

	var excludes []string
	for {
		if !ih.scanner.Scan() {
			if err := ih.scanner.Err(); err != nil {
				return nil, fmt.Errorf("failed to read input: %w", err)
			}
			break
		}

		line := strings.TrimSpace(ih.scanner.Text())
		if line == "" {
			break // Empty line ends input
		}

		excludes = append(excludes, line)
		fmt.Print("> ")
	}

	return excludes, nil
}

// SelectFromList allows user to select specific files from a list
func (ih *InputHandler) SelectFromList(files []selector.FileInfo) ([]selector.FileInfo, error) {
	if len(files) == 0 {
		return files, nil
	}

	fmt.Printf("\n找到 %d 个文件:\n", len(files))
	for i, file := range files {
		relPath := getDisplayPath(file.Path)
		fmt.Printf("%d. %s (%s)\n", i+1, relPath, formatBytes(file.Size))
	}

	fmt.Print("\n请选择要包含的文件 (输入编号，用空格分隔，Enter/a/all 以选择全部): ")

	if !ih.scanner.Scan() {
		if err := ih.scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}
		return nil, fmt.Errorf("no input received")
	}

	input := strings.TrimSpace(ih.scanner.Text())
	if input == "" || input == "a" || input == "all" {
		return files, nil // Return all files
	}

	// Parse selected indices
	parts := strings.Fields(input)
	var selected []selector.FileInfo

	for _, part := range parts {
		var index int
		if _, err := fmt.Sscanf(part, "%d", &index); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid index '%s', skipping\n", part)
			continue
		}

		if index < 1 || index > len(files) {
			fmt.Fprintf(os.Stderr, "Warning: Index %d out of range, skipping\n", index)
			continue
		}

		selected = append(selected, files[index-1])
	}

	return selected, nil
}

// GetOutputPath gets output path from user interactively
func (ih *InputHandler) GetOutputPath() (string, error) {
	fmt.Print("输出文件路径 (回车使用剪贴板): ")

	if !ih.scanner.Scan() {
		if err := ih.scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", nil
	}

	return strings.TrimSpace(ih.scanner.Text()), nil
}

// AskYesNo asks a yes/no question and returns the answer
func (ih *InputHandler) AskYesNo(question string, defaultYes bool) (bool, error) {
	var prompt string
	if defaultYes {
		prompt = fmt.Sprintf("%s (Y/n): ", question)
	} else {
		prompt = fmt.Sprintf("%s (y/N): ", question)
	}

	fmt.Print(prompt)

	if !ih.scanner.Scan() {
		if err := ih.scanner.Err(); err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}
		return defaultYes, nil
	}

	answer := strings.ToLower(strings.TrimSpace(ih.scanner.Text()))
	if answer == "" {
		return defaultYes, nil
	}

	return answer == "y" || answer == "yes", nil
}

// getDisplayPath returns a user-friendly display path
func getDisplayPath(path string) string {
	if wd, err := os.Getwd(); err == nil {
		if relPath, err := filepath.Rel(wd, path); err == nil {
			// Only use relative path if it's shorter and doesn't go up too many levels
			if len(relPath) < len(path) && !strings.HasPrefix(relPath, "../../") {
				return relPath
			}
		}
	}
	return path
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
