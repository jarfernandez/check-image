package labels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLabels_AllValid(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
			{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			{Name: "vendor", Value: "Acme Inc"},
		},
	}

	imageLabels := map[string]string{
		"maintainer": "John Doe <john@example.com>",
		"version":    "v1.2.3",
		"vendor":     "Acme Inc",
		"extra":      "ignored",
	}

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.Empty(t, result.MissingLabels)
	assert.Empty(t, result.InvalidLabels)
}

func TestValidateLabels_MissingLabels(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
			{Name: "version"},
			{Name: "team"},
		},
	}

	imageLabels := map[string]string{
		"maintainer": "John Doe",
	}

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.MissingLabels, 2)
	assert.Contains(t, result.MissingLabels, "version")
	assert.Contains(t, result.MissingLabels, "team")
	assert.Empty(t, result.InvalidLabels)
}

func TestValidateLabels_ExactValueMatch(t *testing.T) {
	tests := []struct {
		name          string
		requirement   LabelRequirement
		imageLabels   map[string]string
		expectedPass  bool
		expectedValid bool
	}{
		{
			name:        "Exact match succeeds",
			requirement: LabelRequirement{Name: "vendor", Value: "Acme Inc"},
			imageLabels: map[string]string{
				"vendor": "Acme Inc",
			},
			expectedPass:  true,
			expectedValid: true,
		},
		{
			name:        "Value mismatch fails",
			requirement: LabelRequirement{Name: "vendor", Value: "Acme Inc"},
			imageLabels: map[string]string{
				"vendor": "Other Corp",
			},
			expectedPass:  false,
			expectedValid: false,
		},
		{
			name:        "Case sensitive check",
			requirement: LabelRequirement{Name: "env", Value: "production"},
			imageLabels: map[string]string{
				"env": "Production",
			},
			expectedPass:  false,
			expectedValid: false,
		},
		{
			name:        "Empty value matches empty requirement",
			requirement: LabelRequirement{Name: "optional", Value: ""},
			imageLabels: map[string]string{
				"optional": "",
			},
			expectedPass:  true,
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &Policy{
				RequiredLabels: []LabelRequirement{tt.requirement},
			}

			result, err := ValidateLabels(tt.imageLabels, policy)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, result.Passed)

			if tt.expectedValid {
				assert.Empty(t, result.InvalidLabels)
			} else {
				assert.Len(t, result.InvalidLabels, 1)
				assert.Equal(t, tt.requirement.Name, result.InvalidLabels[0].Name)
				assert.Equal(t, tt.imageLabels[tt.requirement.Name], result.InvalidLabels[0].ActualValue)
				assert.Equal(t, tt.requirement.Value, result.InvalidLabels[0].ExpectedValue)
			}
		})
	}
}

func TestValidateLabels_PatternMatch(t *testing.T) {
	tests := []struct {
		name          string
		requirement   LabelRequirement
		imageLabels   map[string]string
		expectedPass  bool
		expectedValid bool
	}{
		{
			name:        "Semver pattern matches",
			requirement: LabelRequirement{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			imageLabels: map[string]string{
				"version": "v1.2.3",
			},
			expectedPass:  true,
			expectedValid: true,
		},
		{
			name:        "Semver pattern matches without v prefix",
			requirement: LabelRequirement{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			imageLabels: map[string]string{
				"version": "1.2.3",
			},
			expectedPass:  true,
			expectedValid: true,
		},
		{
			name:        "Semver pattern fails on incomplete version",
			requirement: LabelRequirement{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			imageLabels: map[string]string{
				"version": "1.2",
			},
			expectedPass:  false,
			expectedValid: false,
		},
		{
			name:        "URL pattern matches",
			requirement: LabelRequirement{Name: "source", Pattern: "^https://github\\.com/.+/.+$"},
			imageLabels: map[string]string{
				"source": "https://github.com/user/repo",
			},
			expectedPass:  true,
			expectedValid: true,
		},
		{
			name:        "URL pattern fails on invalid URL",
			requirement: LabelRequirement{Name: "source", Pattern: "^https://github\\.com/.+/.+$"},
			imageLabels: map[string]string{
				"source": "https://gitlab.com/user/repo",
			},
			expectedPass:  false,
			expectedValid: false,
		},
		{
			name:        "Empty pattern matches anything",
			requirement: LabelRequirement{Name: "any", Pattern: ".*"},
			imageLabels: map[string]string{
				"any": "anything goes here",
			},
			expectedPass:  true,
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &Policy{
				RequiredLabels: []LabelRequirement{tt.requirement},
			}

			result, err := ValidateLabels(tt.imageLabels, policy)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, result.Passed)

			if tt.expectedValid {
				assert.Empty(t, result.InvalidLabels)
			} else {
				assert.Len(t, result.InvalidLabels, 1)
				assert.Equal(t, tt.requirement.Name, result.InvalidLabels[0].Name)
				assert.Equal(t, tt.imageLabels[tt.requirement.Name], result.InvalidLabels[0].ActualValue)
				assert.Equal(t, tt.requirement.Pattern, result.InvalidLabels[0].ExpectedPattern)
			}
		})
	}
}

func TestValidateLabels_ExistenceOnly(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
			{Name: "description"},
		},
	}

	imageLabels := map[string]string{
		"maintainer":  "Anyone",
		"description": "Any value is fine here",
	}

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Empty(t, result.MissingLabels)
	assert.Empty(t, result.InvalidLabels)
}

func TestValidateLabels_MultipleFailures(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
			{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			{Name: "vendor", Value: "Acme Inc"},
			{Name: "team"},
		},
	}

	imageLabels := map[string]string{
		"version": "1.2",      // Invalid pattern
		"vendor":  "Other Co", // Wrong value
		// maintainer and team are missing
	}

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	assert.False(t, result.Passed)

	// Should have 2 missing labels
	assert.Len(t, result.MissingLabels, 2)
	assert.Contains(t, result.MissingLabels, "maintainer")
	assert.Contains(t, result.MissingLabels, "team")

	// Should have 2 invalid labels
	assert.Len(t, result.InvalidLabels, 2)

	// Find version and vendor in invalid labels
	foundVersion := false
	foundVendor := false
	for _, inv := range result.InvalidLabels {
		if inv.Name == "version" {
			foundVersion = true
			assert.Equal(t, "1.2", inv.ActualValue)
			assert.Equal(t, "^v?\\d+\\.\\d+\\.\\d+$", inv.ExpectedPattern)
		}
		if inv.Name == "vendor" {
			foundVendor = true
			assert.Equal(t, "Other Co", inv.ActualValue)
			assert.Equal(t, "Acme Inc", inv.ExpectedValue)
		}
	}
	assert.True(t, foundVersion, "version should be in invalid labels")
	assert.True(t, foundVendor, "vendor should be in invalid labels")
}

func TestValidateLabels_EmptyImageLabels(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
			{Name: "version"},
		},
	}

	// Empty labels map
	imageLabels := make(map[string]string)

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Len(t, result.MissingLabels, 2)
	assert.Contains(t, result.MissingLabels, "maintainer")
	assert.Contains(t, result.MissingLabels, "version")
	assert.Empty(t, result.InvalidLabels)
}

func TestValidateLabels_NilImageLabels(t *testing.T) {
	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "maintainer"},
		},
	}

	// Nil labels (as might come from config.Config.Labels)
	var imageLabels map[string]string

	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Len(t, result.MissingLabels, 1)
	assert.Contains(t, result.MissingLabels, "maintainer")
}

func TestValidateLabels_InvalidPatternAtRuntime(t *testing.T) {
	// This should not happen if policy validation works correctly,
	// but we test it for completeness

	policy := &Policy{
		RequiredLabels: []LabelRequirement{
			{Name: "version", Pattern: "[invalid("},
		},
	}

	imageLabels := map[string]string{
		"version": "1.0",
	}

	// Bypass policy validation to test runtime error handling
	// In practice, this would be caught by policy.Validate()
	result, err := ValidateLabels(imageLabels, policy)

	// Should return an error because the pattern can't be compiled
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compile pattern")
	assert.Nil(t, result)
}

func TestValidateLabels_EmptyPolicy(t *testing.T) {
	// This should not happen if policy validation works correctly,
	// but we test the validator's behavior

	policy := &Policy{
		RequiredLabels: []LabelRequirement{},
	}

	imageLabels := map[string]string{
		"any": "value",
	}

	// With no required labels, validation should pass
	result, err := ValidateLabels(imageLabels, policy)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Empty(t, result.MissingLabels)
	assert.Empty(t, result.InvalidLabels)
}
