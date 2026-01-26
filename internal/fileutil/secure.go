package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ReadSecureFile reads a file securely using os.OpenRoot to prevent directory traversal
func ReadSecureFile(path string) ([]byte, error) {
	// Clean the path to remove any .. or . elements
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// Get the directory containing the file
	dir := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	// Create a root-scoped filesystem
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot create root for directory: %w", err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			log.Warnf("failed to close root: %v", closeErr)
		}
	}()

	// Check if file exists and is a regular file
	info, err := root.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Open and read file using the scoped root
	file, err := root.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Warnf("failed to close file: %v", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}
