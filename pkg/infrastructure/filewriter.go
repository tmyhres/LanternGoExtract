package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteBytesToDisk writes the given bytes to a file at the specified path.
// It creates the directory structure if it doesn't exist.
// Returns an error if the write fails.
func WriteBytesToDisk(data []byte, filePath, fileName string) error {
	if data == nil || filePath == "" {
		return nil
	}

	// Create directory structure if it doesn't exist
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filePath, err)
	}

	fullPath := filepath.Join(filePath, fileName)

	// Write the file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirectoryExists checks if a directory exists at the given path.
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDirectoryExists creates the directory if it doesn't exist.
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}
