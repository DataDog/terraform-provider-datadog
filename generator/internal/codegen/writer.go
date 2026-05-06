package codegen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// FormatGoSource formats Go source code using the standard gofmt formatting.
func FormatGoSource(src []byte) ([]byte, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return nil, fmt.Errorf("formatting Go source: %w", err)
	}
	return formatted, nil
}

// WriteGoFile formats the Go source code and writes it to the given path.
// Creates parent directories as needed.
func WriteGoFile(path string, content []byte) error {
	formatted, err := FormatGoSource(content)
	if err != nil {
		return fmt.Errorf("formatting %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// WriteIfNotExists writes the content to the given path only if the file
// does not already exist. Used for hooks files that should not be overwritten.
func WriteIfNotExists(path string, content []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil // File exists, skip
	}

	return WriteGoFile(path, content)
}
