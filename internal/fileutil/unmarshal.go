package fileutil

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalConfigFile unmarshals data from a configuration file (JSON or YAML)
// It first attempts to unmarshal as JSON, and if that fails, tries YAML.
// The filePath parameter is used to determine the format based on file extension.
func UnmarshalConfigFile(data []byte, v interface{}, filePath string) error {
	// Detect format by file extension
	if HasYAMLExtension(filePath) {
		if err := yaml.Unmarshal(data, v); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		return nil
	}

	// Default to JSON
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
