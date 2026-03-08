package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsCommand(t *testing.T) {
	assert.NotNil(t, secretsCmd)
	assert.Equal(t, "secrets image", secretsCmd.Use)
	assert.Contains(t, secretsCmd.Short, "sensitive")

	assert.NotNil(t, secretsCmd.Args)
	err := secretsCmd.Args(secretsCmd, []string{})
	assert.Error(t, err)

	err = secretsCmd.Args(secretsCmd, []string{"image"})
	assert.NoError(t, err)

	err = secretsCmd.Args(secretsCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestSecretsCommandFlags(t *testing.T) {
	flag := secretsCmd.Flags().Lookup("secrets-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "s", flag.Shorthand)

	flag = secretsCmd.Flags().Lookup("skip-env-vars")
	assert.NotNil(t, flag)

	flag = secretsCmd.Flags().Lookup("skip-files")
	assert.NotNil(t, flag)
}

func TestRunSecrets_NoSecrets(t *testing.T) {
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

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when no secrets detected")
}

func TestRunSecrets_SecretsInEnvVars(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123",
			"DATABASE_PASSWORD=mypassword",
			"AWS_SECRET_ACCESS_KEY=supersecret",
		},
	})

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when secrets detected in environment variables")
}

func TestRunSecrets_SecretsInFiles(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa":           "-----BEGIN RSA PRIVATE KEY-----",
				"/home/user/.aws/credentials": "[default]\naws_access_key_id=AKIA...",
			},
		},
	})

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when secrets detected in files")
}

func TestRunSecrets_SecretsBothEnvAndFiles(t *testing.T) {
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

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when secrets detected in both env vars and files")
}

func TestRunSecrets_SkipEnvVars(t *testing.T) {
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

	result, err := runSecrets(context.Background(), imageRef, "", true, false)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when skipping env vars and no file secrets")
}

func TestRunSecrets_SkipFiles(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
				"/root/.ssh/id_dsa": "-----BEGIN DSA PRIVATE KEY-----",
			},
		},
	})

	result, err := runSecrets(context.Background(), imageRef, "", false, true)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when skipping files and no env secrets")
}

func TestRunSecrets_SkipBoth(t *testing.T) {
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

	result, err := runSecrets(context.Background(), imageRef, "", true, true)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when skipping both env and file checks")
}

func TestRunSecrets_WithPolicyFile(t *testing.T) {
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

	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"API_KEY=secret123",
			"PUBLIC_KEY=pubkey",
		},
		layerFiles: []map[string]string{
			{
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
				"/var/log/app.log":  "password=secret",
			},
		},
	})

	result, err := runSecrets(context.Background(), imageRef, policyFile, false, false)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when all secrets are excluded by policy")
}

func TestRunSecrets_PolicyFileWithNonExcludedSecrets(t *testing.T) {
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

	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"PUBLIC_KEY=pubkey",
			"PASSWORD=secret123",
		},
		layerFiles: []map[string]string{
			{
				"/var/log/app.log":  "password=secret",
				"/root/.ssh/id_rsa": "-----BEGIN RSA PRIVATE KEY-----",
			},
		},
	})

	result, err := runSecrets(context.Background(), imageRef, policyFile, false, false)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when non-excluded secrets are detected")
}

func TestRunSecrets_InvalidPolicyFile(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		env: []string{"PATH=/usr/bin"},
	})

	_, err := runSecrets(context.Background(), imageRef, "/nonexistent/policy.json", false, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load secrets policy")
}

func TestRunSecrets_MultipleLayersWithSecrets(t *testing.T) {
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

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when secrets detected across multiple layers")
}

func TestRunSecrets_InvalidImageReference(t *testing.T) {
	_, err := runSecrets(context.Background(), "oci:/nonexistent/path:latest", "", false, false)
	require.Error(t, err)
}

func TestRunSecrets_EmptyImage(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{})

	result, err := runSecrets(context.Background(), imageRef, "", false, false)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed with empty image")
}

func TestRunSecrets_YAMLPolicyFile(t *testing.T) {
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

	imageRef := createTestImage(t, testImageOptions{
		env: []string{
			"PUBLIC_KEY=pubkey",
		},
		layerFiles: []map[string]string{
			{
				"/var/log/app.log": "password=secret",
			},
		},
	})

	result, err := runSecrets(context.Background(), imageRef, policyFile, false, false)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should work with YAML policy file")
}

func TestRunSecrets_PolicyFromStdin(t *testing.T) {
	t.Run("JSON policy from stdin — excludes env var", func(t *testing.T) {
		origStdin := os.Stdin
		defer func() { os.Stdin = origStdin }()

		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r

		go func() {
			_, _ = w.WriteString(`{"check-env-vars": true, "excluded-env-vars": ["SECRET_KEY"]}`)
			w.Close()
		}()

		imageRef := createTestImage(t, testImageOptions{
			env: []string{"SECRET_KEY=myvalue"},
		})

		result, err := runSecrets(context.Background(), imageRef, "-", false, false)
		require.NoError(t, err)
		// SECRET_KEY is excluded by the stdin policy, so no findings
		assert.True(t, result.Passed)
	})

	t.Run("YAML policy from stdin — excludes path", func(t *testing.T) {
		origStdin := os.Stdin
		defer func() { os.Stdin = origStdin }()

		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r

		go func() {
			_, _ = w.WriteString("check-env-vars: false\ncheck-files: false\n")
			w.Close()
		}()

		imageRef := createTestImage(t, testImageOptions{
			layerFiles: []map[string]string{
				{"/etc/id_rsa": "PRIVATE KEY content"},
			},
		})

		result, err := runSecrets(context.Background(), imageRef, "-", false, false)
		require.NoError(t, err)
		// Files check disabled by the stdin policy
		assert.True(t, result.Passed)
	})
}
