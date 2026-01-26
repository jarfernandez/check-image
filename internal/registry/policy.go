package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Policy defines a registry allowlist or blocklist.
// Only one of TrustedRegistries or ExcludedRegistries should be specified:
// - If TrustedRegistries is set, only those registries are allowed (allowlist mode)
// - If ExcludedRegistries is set, all registries except those are allowed (blocklist mode)
type Policy struct {
	TrustedRegistries  []string `yaml:"trusted-registries,omitempty" json:"trusted-registries,omitempty"`
	ExcludedRegistries []string `yaml:"excluded-registries,omitempty" json:"excluded-registries,omitempty"`
}

// readSecureFile reads a file securely using os.OpenRoot to prevent directory traversal
func readSecureFile(path string) ([]byte, error) {
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

// LoadRegistryPolicy loads a registry policy from a file, which can be in
// either YAML or JSON format, and returns the parsed Policy object.
// The policy must specify either trusted-registries or excluded-registries, but not both.
func LoadRegistryPolicy(path string) (*Policy, error) {
	// Read file securely
	data, err := readSecureFile(path)
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

	// Validate that only one mode is specified
	hasTrusted := len(policy.TrustedRegistries) > 0
	hasExcluded := len(policy.ExcludedRegistries) > 0

	if hasTrusted && hasExcluded {
		return nil, fmt.Errorf("policy must specify either trusted-registries or excluded-registries, not both")
	}

	if !hasTrusted && !hasExcluded {
		return nil, fmt.Errorf("policy must specify either trusted-registries or excluded-registries")
	}

	return &policy, nil
}

func hasYAMLExtension(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// IsRegistryAllowed checks if the given registry is allowed based on the policy.
// If trusted-registries is set (allowlist mode), only registries in that list are allowed.
// If excluded-registries is set (blocklist mode), all registries except those in the list are allowed.
func (p *Policy) IsRegistryAllowed(registry string) bool {
	// Allowlist mode: only trusted registries are allowed
	if len(p.TrustedRegistries) > 0 {
		for _, trusted := range p.TrustedRegistries {
			if registry == trusted {
				return true
			}
		}
		return false
	}

	// Blocklist mode: all registries except excluded ones are allowed
	if len(p.ExcludedRegistries) > 0 {
		for _, excluded := range p.ExcludedRegistries {
			if registry == excluded {
				return false
			}
		}
		return true
	}

	// This should not happen if LoadRegistryPolicy validation works correctly
	return false
}
