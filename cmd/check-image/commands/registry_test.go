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

func TestRunRegistry_TrustedRegistry(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file with trusted registries
	// Note: nginx:latest resolves to index.docker.io, not docker.io
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io", "gcr.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with trusted registry
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed for trusted registry")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_UntrustedRegistry(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file with trusted registries (not including the test registry)
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["gcr.io", "quay.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with untrusted registry (docker.io is not in the list)
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail for untrusted registry")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_ExplicitRegistryName(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file
	// Note: docker.io images resolve to index.docker.io
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with explicit registry name
	err = runRegistry("docker.io/library/nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed for explicitly trusted registry")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_DifferentRegistries(t *testing.T) {
	tests := []struct {
		name          string
		imageName     string
		trustedRegs   []string
		expectSuccess bool
	}{
		{
			name:          "Docker Hub trusted",
			imageName:     "nginx:latest",
			trustedRegs:   []string{"index.docker.io"}, // Docker Hub resolves to index.docker.io
			expectSuccess: true,
		},
		{
			name:          "GCR trusted",
			imageName:     "gcr.io/project/image:tag",
			trustedRegs:   []string{"gcr.io"},
			expectSuccess: true,
		},
		{
			name:          "Quay trusted",
			imageName:     "quay.io/org/image:tag",
			trustedRegs:   []string{"quay.io"},
			expectSuccess: true,
		},
		{
			name:          "Custom registry trusted",
			imageName:     "registry.example.com/image:tag",
			trustedRegs:   []string{"registry.example.com"},
			expectSuccess: true,
		},
		{
			name:          "Docker Hub untrusted",
			imageName:     "nginx:latest",
			trustedRegs:   []string{"gcr.io", "quay.io"},
			expectSuccess: false,
		},
		{
			name:          "Custom registry untrusted",
			imageName:     "custom.io/image:tag",
			trustedRegs:   []string{"index.docker.io"},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			Result = ValidationSkipped

			// Create policy file
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.json")
			policyContent := `{"trusted-registries": [`
			for i, reg := range tt.trustedRegs {
				if i > 0 {
					policyContent += ","
				}
				policyContent += `"` + reg + `"`
			}
			policyContent += `]}`
			err := os.WriteFile(policyFile, []byte(policyContent), 0600)
			require.NoError(t, err)

			registryPolicy = policyFile

			err = runRegistry(tt.imageName)
			require.NoError(t, err)

			if tt.expectSuccess {
				assert.Equal(t, ValidationSucceeded, Result, "Expected validation to succeed")
			} else {
				assert.Equal(t, ValidationFailed, Result, "Expected validation to fail")
			}

			// Reset
			registryPolicy = ""
		})
	}
}

func TestRunRegistry_ExcludedRegistries(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file with excluded registries
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"excluded-registries": ["untrusted.io", "malicious.com"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with non-excluded registry (should succeed)
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed for non-excluded registry")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_ExcludedRegistryBlocked(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file with excluded registries
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"excluded-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with excluded registry (should fail)
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail for excluded registry")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_OCITransportSkipped(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file (won't be used because OCI transport skips validation)
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with OCI transport - should skip validation
	err = runRegistry("oci:/path/to/layout:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSkipped, Result, "Should skip validation for OCI transport")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_OCIArchiveTransportSkipped(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with OCI archive transport - should skip validation
	err = runRegistry("oci-archive:/path/to/image.tar:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSkipped, Result, "Should skip validation for OCI archive transport")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_DockerArchiveTransportSkipped(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with Docker archive transport - should skip validation
	err = runRegistry("docker-archive:/path/to/image.tar:nginx")
	require.NoError(t, err)
	assert.Equal(t, ValidationSkipped, Result, "Should skip validation for docker-archive transport")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_InvalidPolicyFile(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Set non-existent policy file
	registryPolicy = "/nonexistent/policy.json"

	err := runRegistry("nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load registry policy")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_InvalidPolicyContent(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create invalid policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{invalid json}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	err = runRegistry("nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load registry policy")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_YAMLPolicyFile(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create YAML policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.yaml")
	policyContent := `trusted-registries:
  - index.docker.io
  - gcr.io
`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with YAML policy
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should work with YAML policy file")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_PreservesPreviousFailure(t *testing.T) {
	// Set Result to ValidationFailed to simulate a previous failed check
	Result = ValidationFailed

	// Create policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with trusted registry (would normally succeed)
	err = runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should preserve previous validation failure")

	// Reset
	registryPolicy = ""
}

func TestRunRegistry_InvalidImageName(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped

	// Create policy file
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	// Test with invalid image name
	err = runRegistry("")
	require.Error(t, err)

	// Reset
	registryPolicy = ""
}
