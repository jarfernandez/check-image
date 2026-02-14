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
