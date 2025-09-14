package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Files       []string `yaml:"files"`
	Exclude     []string `yaml:"exclude"`
	Prompt      string   `yaml:"prompt"`
	MaxFileSize int64    `yaml:"max_file_size"`
	Output      string   `yaml:"output"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Files:       []string{},
		Exclude:     []string{"vendor/**", "node_modules/**", ".git/**"},
		Prompt:      "",
		MaxFileSize: 1048576, // 1MB
		Output:      "",       // Empty means clipboard
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return default config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Merge merges command line options into the configuration
func (c *Config) Merge(files []string, exclude []string, prompt string, output string, maxFileSize int64) {
	if len(files) > 0 {
		c.Files = append(c.Files, files...)
	}
	if len(exclude) > 0 {
		c.Exclude = append(c.Exclude, exclude...)
	}
	if prompt != "" {
		c.Prompt = prompt
	}
	if output != "" {
		c.Output = output
	}
	if maxFileSize > 0 {
		c.MaxFileSize = maxFileSize
	}
}