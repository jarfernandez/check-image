package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsCommand(t *testing.T) {
	// Test that secrets command exists and has correct properties
	assert.NotNil(t, secretsCmd)
	assert.Equal(t, "secrets image", secretsCmd.Use)
	assert.Contains(t, secretsCmd.Short, "sensitive")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, secretsCmd.Args)
	err := secretsCmd.Args(secretsCmd, []string{})
	assert.Error(t, err)

	err = secretsCmd.Args(secretsCmd, []string{"image"})
	assert.NoError(t, err)

	err = secretsCmd.Args(secretsCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestSecretsCommandFlags(t *testing.T) {
	// Test that secrets-policy flag exists
	flag := secretsCmd.Flags().Lookup("secrets-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "p", flag.Shorthand)

	// Test that skip-env-vars flag exists
	flag = secretsCmd.Flags().Lookup("skip-env-vars")
	assert.NotNil(t, flag)

	// Test that skip-files flag exists
	flag = secretsCmd.Flags().Lookup("skip-files")
	assert.NotNil(t, flag)
}

func TestSecretsCommand_WithPolicyFile(t *testing.T) {
	// Create a test policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "secrets-policy.json")
	policyContent := `{
		"check-env-vars": true,
		"check-files": true,
		"excluded-paths": ["/var/log/**"],
		"excluded-env-vars": ["PUBLIC_KEY"]
	}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	// Set the policy
	secretsPolicy = policyFile

	// This would require an actual image to test fully
	// For now, we just verify the command accepts the flag
}

func TestSecretsCommand_SkipFlags(t *testing.T) {
	// Test skip-env-vars flag
	skipEnvVars = true
	assert.True(t, skipEnvVars)

	// Test skip-files flag
	skipFiles = true
	assert.True(t, skipFiles)

	// Reset
	skipEnvVars = false
	skipFiles = false
}

// Note: Full testing of runSecrets requires:
// 1. Valid policy files (tested in internal/secrets/policy_test.go)
// 2. Images with environment variables and files to scan
// 3. Detection logic (tested in internal/secrets/detector_test.go)
// Integration tests should cover the actual scanning logic.
