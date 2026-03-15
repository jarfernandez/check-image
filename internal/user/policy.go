package user

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/fileutil"
)

// Policy defines user validation rules for container images.
// All fields use pointers/slices so "not set" is distinguishable from zero values.
type Policy struct {
	MinUID         *uint    `json:"min-uid,omitempty"         yaml:"min-uid,omitempty"`
	MaxUID         *uint    `json:"max-uid,omitempty"         yaml:"max-uid,omitempty"`
	BlockedUsers   []string `json:"blocked-users,omitempty"   yaml:"blocked-users,omitempty"`
	RequireNumeric *bool    `json:"require-numeric,omitempty" yaml:"require-numeric,omitempty"`
}

// LoadUserPolicy loads a user policy from a file or stdin (if path is "-"),
// which can be in either YAML or JSON format, and returns the parsed Policy object.
func LoadUserPolicy(path string) (*Policy, error) {
	data, err := fileutil.ReadFileOrStdin(path)
	if err != nil {
		return nil, fmt.Errorf("error reading user policy: %w", err)
	}

	var policy Policy
	if err := fileutil.UnmarshalConfigData(data, &policy, path); err != nil {
		return nil, err
	}

	if err := policy.Validate(); err != nil {
		return nil, err
	}

	return &policy, nil
}

// Validate checks that the policy is internally consistent.
func (p *Policy) Validate() error {
	if p.MinUID != nil && p.MaxUID != nil && *p.MinUID > *p.MaxUID {
		return fmt.Errorf("min-uid (%d) must not exceed max-uid (%d)", *p.MinUID, *p.MaxUID)
	}
	return nil
}
