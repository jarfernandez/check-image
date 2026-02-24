package fileutil

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalConfigData unmarshals data using content detection for stdin
func UnmarshalConfigData(data []byte, v any, filePath string) error {
	var isYAML bool
	if filePath == "-" {
		isYAML = IsYAML(data)
	} else {
		isYAML = HasYAMLExtension(filePath)
	}

	if isYAML {
		if err := yaml.Unmarshal(data, v); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		return nil
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
