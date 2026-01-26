package fileutil

import "strings"

// HasYAMLExtension checks if a file path has a YAML extension (.yaml or .yml)
func HasYAMLExtension(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}
