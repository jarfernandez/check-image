package registry

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// Policy contains the list of trusted and excluded registries
type Policy struct {
	TrustedRegistries  []string `yaml:"trusted-registries" json:"trusted-registries"`
	ExcludedRegistries []string `yaml:"excluded-registries,omitempty" json:"excluded-registries,omitempty"`
}

// LoadRegistryPolicy loads a registry policy from a file, which can be in
// either YAML or JSON format, and returns the parsed RegistryPolicy object.
func LoadRegistryPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading registry policy file: %w", err)
	}

	var policy Policy

	// Detect format by file extension
	switch {
	case hasYAMLExtension(path):
		if err := yaml.Unmarshal(data, &policy); err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
	default: // JSON as fallback
		if err := json.Unmarshal(data, &policy); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}

	return &policy, nil
}

func hasYAMLExtension(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// IsRegistryAllowed checks if the given registry is allowed based on the policy
func (p *Policy) IsRegistryAllowed(registry string) bool {
	// If the registry is in the excluded list, deny always
	for _, excluded := range p.ExcludedRegistries {
		if registry == excluded {
			return false
		}
	}

	// If "*" is in the list of trusted registries, all registries are allowed
	for _, trusted := range p.TrustedRegistries {
		if trusted == "*" {
			return true
		}
	}

	// If the registry is in the trusted list, allow
	for _, trusted := range p.TrustedRegistries {
		if registry == trusted {
			return true
		}
	}

	// It's not in the list of trusted registries and there's no wildcard
	return false
}
