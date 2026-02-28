package labels

import (
	"fmt"
	"regexp"

	"github.com/jarfernandez/check-image/internal/fileutil"
)

// Policy defines required labels with validation rules.
type Policy struct {
	RequiredLabels []LabelRequirement `yaml:"required-labels" json:"required-labels"`
}

// LabelRequirement defines a single label requirement.
// Name is required. Value and Pattern are optional and mutually exclusive.
// - If neither Value nor Pattern is specified, only label existence is checked
// - If Value is specified, the label value must match exactly (case-sensitive)
// - If Pattern is specified, the label value must match the regex pattern
type LabelRequirement struct {
	Name    string `yaml:"name" json:"name"`
	Value   string `yaml:"value,omitempty" json:"value,omitempty"`
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
}

// LoadLabelsPolicy loads a labels policy from a file or stdin (if path is "-"),
// which can be in either YAML or JSON format, and returns the parsed Policy object.
// The policy must specify at least one required label, and each label must have a name.
// Labels cannot have both value and pattern specified (conflicting requirements).
func LoadLabelsPolicy(path string) (*Policy, error) {
	// Read file or stdin
	data, err := fileutil.ReadFileOrStdin(path)
	if err != nil {
		return nil, fmt.Errorf("error reading labels policy: %w", err)
	}

	var policy Policy

	// Unmarshal config data (JSON or YAML)
	if err := fileutil.UnmarshalConfigData(data, &policy, path); err != nil {
		return nil, err
	}

	// Validate policy
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	return &policy, nil
}

// Validate checks that the policy is well-formed
func (p *Policy) Validate() error {
	// Must have at least one required label
	if len(p.RequiredLabels) == 0 {
		return fmt.Errorf("policy must specify at least one required label")
	}

	// Check for duplicate label names
	seen := make(map[string]bool)

	for i, req := range p.RequiredLabels {
		// Each label must have a name
		if req.Name == "" {
			return fmt.Errorf("label requirement at index %d is missing a name", i)
		}

		// Check for duplicates
		if seen[req.Name] {
			return fmt.Errorf("duplicate label name %q in policy", req.Name)
		}
		seen[req.Name] = true

		// Cannot have both value and pattern
		if req.Value != "" && req.Pattern != "" {
			return fmt.Errorf("label %q cannot have both value and pattern requirements", req.Name)
		}

		// Validate pattern if specified
		if req.Pattern != "" {
			if _, err := regexp.Compile(req.Pattern); err != nil {
				return fmt.Errorf("invalid pattern for label %q: %w", req.Name, err)
			}
		}
	}

	return nil
}
