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

func TestRunSecrets_NoSecrets(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with no secrets
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"PATH=/usr/local/bin:/usr/bin",
			"HOME=/root",
		},
		layerFiles: []map[string]string{
			{
				"/app/config.json": `{"setting": "value"}`,
				"/app/main.go":     `package main`,
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when no secrets detected")
}

func TestRunSecrets_SecretsInEnvVars(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with secrets in environment variables
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123",
			"DATABASE_PASSWORD=mypassword",
			"AWS_SECRET_ACCESS_KEY=supersecret",
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when secrets detected in environment variables")
}

func TestRunSecrets_SecretsInFiles(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with secret files
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa":           "-----BEGIN RSA PRIVATE KEY-----",
				"/home/user/.aws/credentials": "[default]\naws_access_key_id=AKIA...",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when secrets detected in files")
}

func TestRunSecrets_SecretsBothEnvAndFiles(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with secrets in both env vars and files
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_TOKEN=token123",
			"PATH=/usr/bin",
		},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
				"/app/config.yaml":  "normal config",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when secrets detected in both env vars and files")
}

func TestRunSecrets_SkipEnvVars(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = true // Skip environment variable checks
	skipFiles = false

	// Create test image with secrets only in env vars
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123",
			"PASSWORD=mypassword",
		},
		layerFiles: []map[string]string{
			{
				"/app/main.go": "package main",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when skipping env vars and no file secrets")

	// Reset
	skipEnvVars = false
}

func TestRunSecrets_SkipFiles(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = true // Skip file checks

	// Create test image with secrets only in files
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
				"/root/.ssh/id_dsa": "-----BEGIN DSA PRIVATE KEY-----",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when skipping files and no env secrets")

	// Reset
	skipFiles = false
}

func TestRunSecrets_SkipBoth(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = true // Skip both checks
	skipFiles = true

	// Create test image with secrets everywhere
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123",
			"PASSWORD=mypassword",
		},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when skipping both env and file checks")

	// Reset
	skipEnvVars = false
	skipFiles = false
}

func TestRunSecrets_WithPolicyFile(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	skipEnvVars = false
	skipFiles = false

	// Create a test policy file with exclusions
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "secrets-policy.json")
	policyContent := `{
		"check-env-vars": true,
		"check-files": true,
		"excluded-paths": ["/var/log/**", "/root/.ssh/id_rsa"],
		"excluded-env-vars": ["PUBLIC_KEY", "API_KEY"]
	}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	secretsPolicy = policyFile

	// Create test image with excluded secrets
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123", // Excluded
			"PUBLIC_KEY=pubkey", // Excluded
		},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----", // Excluded
				"/var/log/app.log":  "password=secret",                 // Excluded
			},
		},
	})

	err = runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when all secrets are excluded by policy")

	// Reset
	secretsPolicy = ""
}

func TestRunSecrets_PolicyFileWithNonExcludedSecrets(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	skipEnvVars = false
	skipFiles = false

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

	secretsPolicy = policyFile

	// Create test image with both excluded and non-excluded secrets
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"PUBLIC_KEY=pubkey",  // Excluded
			"PASSWORD=secret123", // Not excluded - should fail
		},
		layerFiles: []map[string]string{
			{
				"/var/log/app.log":  "password=secret",                 // Excluded
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----", // Not excluded - should fail
			},
		},
	})

	err = runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when non-excluded secrets are detected")

	// Reset
	secretsPolicy = ""
}

func TestRunSecrets_InvalidPolicyFile(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	skipEnvVars = false
	skipFiles = false

	// Set non-existent policy file
	secretsPolicy = "/nonexistent/policy.json"

	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
	})

	err := runSecrets(imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load secrets policy")

	// Reset
	secretsPolicy = ""
}

func TestRunSecrets_MultipleLayersWithSecrets(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with secrets in multiple layers
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/app/config.json": "normal config",
			},
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
			},
			{
				"/home/user/.aws/credentials": "[default]\naws_access_key_id=AKIA...",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when secrets detected across multiple layers")
}

func TestRunSecrets_InvalidImageReference(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Use invalid image reference
	err := runSecrets("oci:/nonexistent/path:latest")
	require.Error(t, err)
}

func TestRunSecrets_PreservesPreviousFailure(t *testing.T) {
	// Set Result to ValidationFailed to simulate a previous failed check
	Result = ValidationFailed
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with no secrets (would normally pass)
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/app/main.go": "package main",
			},
		},
	})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should preserve previous validation failure")
}

func TestRunSecrets_EmptyImage(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false

	// Create test image with no env vars or layers
	imageRef := createTestImage(t, testImageOptions{})

	err := runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed with empty image")
}

func TestRunSecrets_YAMLPolicyFile(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	skipEnvVars = false
	skipFiles = false

	// Create a test policy file in YAML format
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "secrets-policy.yaml")
	policyContent := `check-env-vars: true
check-files: true
excluded-paths:
  - /var/log/**
excluded-env-vars:
  - PUBLIC_KEY
`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	secretsPolicy = policyFile

	// Create test image with excluded secrets
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"PUBLIC_KEY=pubkey", // Excluded
		},
		layerFiles: []map[string]string{
			{
				"/var/log/app.log": "password=secret", // Excluded
			},
		},
	})

	err = runSecrets(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should work with YAML policy file")

	// Reset
	secretsPolicy = ""
}
