package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelsCommand(t *testing.T) {
	// Test that the command is properly registered
	assert.NotNil(t, labelsCmd)
	assert.Equal(t, "labels image", labelsCmd.Use)
	assert.True(t, labelsCmd.Flags().Lookup("labels-policy").NoOptDefVal == "")
}

func TestLabelsCommand_MissingPolicyFlag(t *testing.T) {
	// Test that --labels-policy flag is required
	err := labelsCmd.MarkFlagRequired("labels-policy")
	require.NoError(t, err)

	// Verify the flag is marked as required
	flag := labelsCmd.Flags().Lookup("labels-policy")
	require.NotNil(t, flag)
}

func TestRunLabels_AllValid(t *testing.T) {
	// Create test image with OCI layout
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"maintainer":                       "John Doe <john@example.com>",
		"org.opencontainers.image.version": "v1.2.3",
		"org.opencontainers.image.vendor":  "Acme Inc",
	})

	// Create policy file
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"},
			{"name": "org.opencontainers.image.version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"},
			{"name": "org.opencontainers.image.vendor", "value": "Acme Inc"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	// Set global variable
	labelsPolicy = policyFile

	// Run the check
	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result
	assert.Equal(t, "labels", result.Check)
	assert.True(t, result.Passed)
	assert.Equal(t, "All required labels are present and valid", result.Message)

	// Verify details
	details := result.Details.(output.LabelsDetails)
	assert.Len(t, details.RequiredLabels, 3)
	assert.Len(t, details.ActualLabels, 3)
	assert.Empty(t, details.MissingLabels)
	assert.Empty(t, details.InvalidLabels)
}

func TestRunLabels_MissingLabels(t *testing.T) {
	// Create test image with only some labels
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"maintainer": "John Doe",
	})

	// Create policy requiring multiple labels
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"},
			{"name": "version"},
			{"name": "team"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Passed)
	assert.Equal(t, "Image does not meet label requirements", result.Message)

	details := result.Details.(output.LabelsDetails)
	assert.Len(t, details.MissingLabels, 2)
	assert.Contains(t, details.MissingLabels, "version")
	assert.Contains(t, details.MissingLabels, "team")
}

func TestRunLabels_InvalidValue(t *testing.T) {
	// Create test image with wrong value
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"vendor": "Other Company",
	})

	// Create policy expecting specific value
	policyContent := `{
		"required-labels": [
			{"name": "vendor", "value": "Acme Inc"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Passed)

	details := result.Details.(output.LabelsDetails)
	assert.Len(t, details.InvalidLabels, 1)
	assert.Equal(t, "vendor", details.InvalidLabels[0].Name)
	assert.Equal(t, "Other Company", details.InvalidLabels[0].ActualValue)
	assert.Equal(t, "Acme Inc", details.InvalidLabels[0].ExpectedValue)
}

func TestRunLabels_InvalidPattern(t *testing.T) {
	// Create test image with version that doesn't match pattern
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"version": "1.2", // Missing patch version
	})

	// Create policy with semver pattern
	policyContent := `{
		"required-labels": [
			{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Passed)

	details := result.Details.(output.LabelsDetails)
	assert.Len(t, details.InvalidLabels, 1)
	assert.Equal(t, "version", details.InvalidLabels[0].Name)
	assert.Equal(t, "1.2", details.InvalidLabels[0].ActualValue)
	assert.Equal(t, "^v?\\d+\\.\\d+\\.\\d+$", details.InvalidLabels[0].ExpectedPattern)
}

func TestRunLabels_NoLabelsInImage(t *testing.T) {
	// Create test image with no labels
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, nil)

	// Create policy requiring a label
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Passed)

	details := result.Details.(output.LabelsDetails)
	assert.Empty(t, details.ActualLabels)
	assert.Len(t, details.MissingLabels, 1)
	assert.Contains(t, details.MissingLabels, "maintainer")
}

func TestRunLabels_MultipleFailures(t *testing.T) {
	// Create test image with some labels missing and some invalid
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"version": "1.2",      // Invalid pattern
		"vendor":  "Wrong Co", // Wrong value
		// maintainer and team are missing
	})

	// Create policy with multiple requirements
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"},
			{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"},
			{"name": "vendor", "value": "Acme Inc"},
			{"name": "team"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Passed)

	details := result.Details.(output.LabelsDetails)
	// Should have 2 missing labels
	assert.Len(t, details.MissingLabels, 2)
	assert.Contains(t, details.MissingLabels, "maintainer")
	assert.Contains(t, details.MissingLabels, "team")

	// Should have 2 invalid labels
	assert.Len(t, details.InvalidLabels, 2)
}

func TestRunLabels_InvalidPolicy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid policy (empty required labels)
	policyContent := `{
		"required-labels": []
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	// Should fail to load policy
	result, err := runLabels("nginx:latest")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unable to load labels policy")
}

func TestRunLabels_NonexistentImage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid policy
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	// Try to check nonexistent image
	result, err := runLabels("oci:/nonexistent/path")
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestLabelsCommand_RunE(t *testing.T) {
	// Save original Result and OutputFmt
	origResult := Result
	origOutputFmt := OutputFmt
	defer func() {
		Result = origResult
		OutputFmt = origOutputFmt
	}()

	t.Run("successful validation", func(t *testing.T) {
		// Reset Result
		Result = ValidationSkipped
		OutputFmt = output.FormatText

		// Create test image with valid labels
		tmpDir := t.TempDir()
		imageDir := filepath.Join(tmpDir, "test-image")
		createTestOCIImage(t, imageDir, map[string]string{
			"maintainer": "John Doe",
		})

		// Create policy file
		policyContent := `{"required-labels": [{"name": "maintainer"}]}`
		policyFile := filepath.Join(tmpDir, "policy.json")
		err := os.WriteFile(policyFile, []byte(policyContent), 0600)
		require.NoError(t, err)

		labelsPolicy = policyFile

		// Capture output
		output := captureStdout(t, func() {
			err := labelsCmd.RunE(labelsCmd, []string{"oci:" + imageDir + ":latest"})
			require.NoError(t, err)
		})

		// Verify Result was updated to ValidationSucceeded
		assert.Equal(t, ValidationSucceeded, Result)
		assert.Contains(t, output, "All required labels are present and valid")
	})

	t.Run("failed validation", func(t *testing.T) {
		// Reset Result
		Result = ValidationSkipped
		OutputFmt = output.FormatText

		// Create test image with missing labels
		tmpDir := t.TempDir()
		imageDir := filepath.Join(tmpDir, "test-image")
		createTestOCIImage(t, imageDir, nil)

		// Create policy file
		policyContent := `{"required-labels": [{"name": "maintainer"}]}`
		policyFile := filepath.Join(tmpDir, "policy.json")
		err := os.WriteFile(policyFile, []byte(policyContent), 0600)
		require.NoError(t, err)

		labelsPolicy = policyFile

		// Capture output
		output := captureStdout(t, func() {
			err := labelsCmd.RunE(labelsCmd, []string{"oci:" + imageDir + ":latest"})
			require.NoError(t, err)
		})

		// Verify Result was updated to ValidationFailed
		assert.Equal(t, ValidationFailed, Result)
		assert.Contains(t, output, "Image does not meet label requirements")
	})
}

func TestLabelsCommand_StdinInput(t *testing.T) {
	tests := []struct {
		name          string
		stdinContent  string
		expectedError bool
		errorContains string
	}{
		{
			name: "JSON policy from stdin",
			stdinContent: `{
				"required-labels": [
					{"name": "maintainer"}
				]
			}`,
			expectedError: false,
		},
		{
			name: "YAML policy from stdin",
			stdinContent: `required-labels:
  - name: maintainer`,
			expectedError: false,
		},
		{
			name:          "Malformed JSON from stdin",
			stdinContent:  `{invalid json`,
			expectedError: true,
			errorContains: "unable to load labels policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test image
			imageDir := filepath.Join(tmpDir, "test-image")
			createTestOCIImage(t, imageDir, map[string]string{
				"maintainer": "John Doe",
			})

			// Create stdin file and redirect
			stdinFile := filepath.Join(tmpDir, "stdin")
			err := os.WriteFile(stdinFile, []byte(tt.stdinContent), 0600)
			require.NoError(t, err)

			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			f, err := os.Open(stdinFile)
			require.NoError(t, err)
			defer f.Close()
			os.Stdin = f

			// Set policy to stdin
			labelsPolicy = "-"

			// Run check
			result, err := runLabels("oci:" + imageDir + ":latest")

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Passed)
		})
	}
}

func TestLabelsCommand_JSONOutput(t *testing.T) {
	// Save original OutputFmt
	origOutputFmt := OutputFmt
	defer func() { OutputFmt = origOutputFmt }()

	// Create test image
	tmpDir := t.TempDir()
	imageDir := filepath.Join(tmpDir, "test-image")
	createTestOCIImage(t, imageDir, map[string]string{
		"maintainer": "John Doe",
		"version":    "1.2.3",
	})

	// Create policy file
	policyContent := `{
		"required-labels": [
			{"name": "maintainer"},
			{"name": "version", "pattern": "^\\d+\\.\\d+\\.\\d+$"}
		]
	}`
	policyFile := filepath.Join(tmpDir, "policy.json")
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	labelsPolicy = policyFile

	// Run check
	result, err := runLabels("oci:" + imageDir + ":latest")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Set JSON output mode
	OutputFmt = output.FormatJSON

	// Capture JSON output
	jsonOutput := captureStdout(t, func() {
		err := renderResult(result)
		require.NoError(t, err)
	})

	// Verify JSON structure
	assert.Contains(t, jsonOutput, `"check": "labels"`)
	assert.Contains(t, jsonOutput, `"passed": true`)
	assert.Contains(t, jsonOutput, `"required-labels"`)
	assert.Contains(t, jsonOutput, `"actual-labels"`)
	assert.Contains(t, jsonOutput, `"maintainer"`)
	assert.Contains(t, jsonOutput, `"version"`)
}

func TestRunLabels_InvalidPolicyFormat(t *testing.T) {
	tests := []struct {
		name          string
		policyContent string
		errorContains string
	}{
		{
			name:          "Malformed JSON",
			policyContent: `{invalid json}`,
			errorContains: "unable to load labels policy",
		},
		{
			name: "Malformed YAML",
			policyContent: `required-labels:
  - name: test
    invalid:
    - [unclosed`,
			errorContains: "unable to load labels policy",
		},
		{
			name: "Invalid regex pattern",
			policyContent: `{
				"required-labels": [
					{"name": "version", "pattern": "[invalid(regex"}
				]
			}`,
			errorContains: "unable to load labels policy",
		},
		{
			name: "Both value and pattern specified",
			policyContent: `{
				"required-labels": [
					{"name": "version", "value": "1.0", "pattern": "^v?\\d+"}
				]
			}`,
			errorContains: "unable to load labels policy",
		},
		{
			name: "Label without name",
			policyContent: `{
				"required-labels": [
					{"value": "test"}
				]
			}`,
			errorContains: "unable to load labels policy",
		},
		{
			name: "Duplicate label names",
			policyContent: `{
				"required-labels": [
					{"name": "version"},
					{"name": "version", "pattern": "^v?\\d+"}
				]
			}`,
			errorContains: "unable to load labels policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create policy file
			policyFile := filepath.Join(tmpDir, "policy.json")
			err := os.WriteFile(policyFile, []byte(tt.policyContent), 0600)
			require.NoError(t, err)

			labelsPolicy = policyFile

			// Try to run check - should fail to load policy
			result, err := runLabels("nginx:latest")
			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}
