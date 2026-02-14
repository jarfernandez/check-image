package fileutil

import (
	"bytes"
	"strings"
)

// HasYAMLExtension checks if a file path has a YAML extension (.yaml or .yml)
func HasYAMLExtension(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// IsYAML returns true if content appears to be YAML, false if JSON
func IsYAML(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false // default to JSON
	}
	// JSON always starts with { or [
	firstChar := trimmed[0]
	if firstChar == '{' || firstChar == '[' {
		return false // JSON
	}
	return true // YAML
}
