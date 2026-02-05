package registry

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/fileutil"
)

// Policy defines a registry allowlist or blocklist.
// Only one of TrustedRegistries or ExcludedRegistries should be specified:
// - If TrustedRegistries is set, only those registries are allowed (allowlist mode)
// - If ExcludedRegistries is set, all registries except those are allowed (blocklist mode)
type Policy struct {
	TrustedRegistries  []string `yaml:"trusted-registries,omitempty" json:"trusted-registries,omitempty"`
	ExcludedRegistries []string `yaml:"excluded-registries,omitempty" json:"excluded-registries,omitempty"`
}

// LoadRegistryPolicy loads a registry policy from a file, which can be in
// either YAML or JSON format, and returns the parsed Policy object.
// The policy must specify either trusted-registries or excluded-registries, but not both.
func LoadRegistryPolicy(path string) (*Policy, error) {
	// Read file securely
	data, err := fileutil.ReadSecureFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading registry policy file: %w", err)
	}

	var policy Policy

	// Unmarshal config file (JSON or YAML)
	if err := fileutil.UnmarshalConfigFile(data, &policy, path); err != nil {
		return nil, err
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
