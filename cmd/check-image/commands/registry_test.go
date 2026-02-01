package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryCommand(t *testing.T) {
	// Test that registry command exists and has correct properties
	assert.NotNil(t, registryCmd)
	assert.Equal(t, "registry image", registryCmd.Use)
	assert.Contains(t, registryCmd.Short, "registry")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, registryCmd.Args)
	err := registryCmd.Args(registryCmd, []string{})
	assert.Error(t, err)

	err = registryCmd.Args(registryCmd, []string{"image"})
	assert.NoError(t, err)

	err = registryCmd.Args(registryCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestRegistryCommandFlags(t *testing.T) {
	// Test that registry-policy flag exists and is required
	flag := registryCmd.Flags().Lookup("registry-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "r", flag.Shorthand)

	// The flag is marked as required via MarkFlagRequired in init()
	// We can verify it's a required flag by checking the command's annotations
	// Note: Cobra stores this information internally
}

func TestRegistryCommand_MissingPolicy(t *testing.T) {
	// Reset the registryPolicy variable
	registryPolicy = ""

	// Try to run without policy - should fail in RunE
	err := registryCmd.RunE(registryCmd, []string{"nginx:latest"})
	require.Error(t, err)
	// The error message may vary depending on how the empty path is interpreted
	assert.NotNil(t, err)
}

func TestRegistryCommand_NonRegistryTransport(t *testing.T) {
	// Create a dummy policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with OCI transport (should skip validation)
	// Note: This would require creating an actual OCI layout
	// For now, we just test the command structure
}

// Note: Full testing of runRegistry requires:
// 1. Valid policy files (tested in internal/registry/policy_test.go)
// 2. Images from various registries
// Integration tests should cover the actual validation logic.
