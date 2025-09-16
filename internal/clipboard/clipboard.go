package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// CopyToClipboard copies text to the system clipboard
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip")
		// Convert UTF-8 to GBK for Windows
		encoder := simplifiedchinese.GBK.NewEncoder()
		gbkText, _, err := transform.String(encoder, text)
		if err != nil {
			// If conversion fails, use original text
			gbkText = text
		}
		cmd.Stdin = strings.NewReader(gbkText)
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel as fallback
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// For non-Windows systems, set stdin here
	if runtime.GOOS != "windows" {
		cmd.Stdin = strings.NewReader(text)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// IsClipboardSupported checks if clipboard operations are supported on this system
func IsClipboardSupported() bool {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.LookPath("clip")
		return err == nil
	case "darwin":
		_, err := exec.LookPath("pbcopy")
		return err == nil
	case "linux":
		_, xclipErr := exec.LookPath("xclip")
		_, xselErr := exec.LookPath("xsel")
		return xclipErr == nil || xselErr == nil
	default:
		return false
	}
}

// WriteToOutput writes text to either clipboard or file based on the output parameter
func WriteToOutput(text, output string, verbose bool) error {
	if output == "" {
		// Try to copy to clipboard
		if IsClipboardSupported() {
			if err := CopyToClipboard(text); err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to copy to clipboard: %v\n", err)
					fmt.Fprintf(os.Stderr, "Writing to file 'prompt.txt' instead\n")
				}
				return writeToFile(text, "prompt.txt")
			}
			if verbose {
				fmt.Fprintf(os.Stderr, "Content copied to clipboard successfully\n")
			}
			return nil
		} else {
			if verbose {
				fmt.Fprintf(os.Stderr, "Clipboard not supported, writing to file 'prompt.txt'\n")
			}
			return writeToFile(text, "prompt.txt")
		}
	}

	// Write to specified file
	return writeToFile(text, output)
}

// writeToFile writes text to a specified file
func writeToFile(text, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(text); err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Content written to file: %s\n", filename)
	return nil
}