package selector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileSelector handles file selection with glob patterns and exclusions
type FileSelector struct {
	patterns    []string
	excludes    []string
	maxFileSize int64
}

// FileInfo contains information about a selected file
type FileInfo struct {
	Path string
	Size int64
}

// New creates a new FileSelector
func New(patterns []string, excludes []string, maxFileSize int64) *FileSelector {
	return &FileSelector{
		patterns:    patterns,
		excludes:    excludes,
		maxFileSize: maxFileSize,
	}
}

// SelectFiles selects files based on patterns and exclusions
func (fs *FileSelector) SelectFiles() ([]FileInfo, error) {
	var files []FileInfo
	processedFiles := make(map[string]bool) // Prevent duplicates

	// If no patterns specified, use current directory
	patterns := fs.patterns
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}

	for _, pattern := range patterns {
		matches, err := fs.expandGlob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to expand pattern '%s': %w", pattern, err)
		}

		for _, match := range matches {
			if processedFiles[match] {
				continue
			}
			processedFiles[match] = true

			// Check if file should be excluded
			if fs.isExcluded(match) {
				continue
			}

			// Check if it's a regular file
			info, err := os.Stat(match)
			if err != nil {
				continue // Skip files that can't be accessed
			}

			if !info.Mode().IsRegular() {
				continue // Skip directories and special files
			}

			// Check file size
			if fs.maxFileSize > 0 && info.Size() > fs.maxFileSize {
				continue // Skip files that are too large
			}

			files = append(files, FileInfo{
				Path: match,
				Size: info.Size(),
			})
		}
	}

	return files, nil
}

// expandGlob expands a glob pattern, handling both simple globs and recursive patterns
func (fs *FileSelector) expandGlob(pattern string) ([]string, error) {
	// Handle recursive patterns like "src/**/*.go"
	if strings.Contains(pattern, "**") {
		return fs.expandRecursiveGlob(pattern)
	}

	// Handle simple glob patterns
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Convert to absolute paths
	var result []string
	for _, match := range matches {
		absPath, err := filepath.Abs(match)
		if err != nil {
			continue
		}
		result = append(result, absPath)
	}

	return result, nil
}

// expandRecursiveGlob handles recursive glob patterns with **
func (fs *FileSelector) expandRecursiveGlob(pattern string) ([]string, error) {
	var matches []string

	// Split pattern at first **
	parts := strings.SplitN(pattern, "**", 2)
	if len(parts) != 2 {
		return filepath.Glob(pattern)
	}

	prefix := parts[0]
	suffix := parts[1]

	// Remove trailing / from prefix and leading / from suffix
	prefix = strings.TrimSuffix(prefix, "/")
	suffix = strings.TrimPrefix(suffix, "/")

	// If prefix is empty, start from current directory
	if prefix == "" {
		prefix = "."
	}

	// Walk the directory tree
	err := filepath.WalkDir(prefix, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking even if there are errors
		}

		if d.IsDir() {
			return nil // We only want files
		}

		// Check if the file matches the suffix pattern
		if suffix == "" {
			// No suffix pattern, match everything
			absPath, err := filepath.Abs(path)
			if err == nil {
				matches = append(matches, absPath)
			}
		} else {
			// Extract the relative path from the prefix
			relPath, err := filepath.Rel(prefix, path)
			if err != nil {
				return nil
			}

			// Check if the relative path matches the suffix pattern
			matched, err := filepath.Match(suffix, relPath)
			if err != nil {
				return nil
			}

			if matched {
				absPath, err := filepath.Abs(path)
				if err == nil {
					matches = append(matches, absPath)
				}
			} else {
				// Try matching just the filename
				matched, err := filepath.Match(suffix, filepath.Base(path))
				if err == nil && matched {
					absPath, err := filepath.Abs(path)
					if err == nil {
						matches = append(matches, absPath)
					}
				}
			}
		}

		return nil
	})

	return matches, err
}

// isExcluded checks if a file path should be excluded based on exclude patterns
func (fs *FileSelector) isExcluded(path string) bool {
	for _, exclude := range fs.excludes {
		// Handle recursive exclusion patterns
		if strings.Contains(exclude, "**") {
			if fs.matchesRecursiveExclude(path, exclude) {
				return true
			}
		} else {
			// Simple pattern matching
			matched, err := filepath.Match(exclude, filepath.Base(path))
			if err == nil && matched {
				return true
			}

			// Also try matching the full path
			matched, err = filepath.Match(exclude, path)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

// matchesRecursiveExclude checks if a path matches a recursive exclude pattern
func (fs *FileSelector) matchesRecursiveExclude(path, pattern string) bool {
	// Split pattern at **
	parts := strings.SplitN(pattern, "**", 2)
	if len(parts) != 2 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	// Check if path contains the prefix
	if prefix != "" && !strings.Contains(path, prefix) {
		return false
	}

	// If there's a suffix, check if it matches
	if suffix != "" {
		matched, _ := filepath.Match(suffix, filepath.Base(path))
		return matched
	}

	// If no suffix, any path containing prefix matches
	return prefix == "" || strings.Contains(path, prefix)
}