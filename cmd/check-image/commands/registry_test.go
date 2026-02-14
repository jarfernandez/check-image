package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryCommand(t *testing.T) {
	assert.NotNil(t, registryCmd)
	assert.Equal(t, "registry image", registryCmd.Use)
	assert.Contains(t, registryCmd.Short, "registry")

	assert.NotNil(t, registryCmd.Args)
	err := registryCmd.Args(registryCmd, []string{})
	assert.Error(t, err)

	err = registryCmd.Args(registryCmd, []string{"image"})
	assert.NoError(t, err)

	err = registryCmd.Args(registryCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestRegistryCommandFlags(t *testing.T) {
	flag := registryCmd.Flags().Lookup("registry-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "r", flag.Shorthand)
}

func TestRegistryCommand_MissingPolicy(t *testing.T) {
	registryPolicy = ""

	err := registryCmd.RunE(registryCmd, []string{"nginx:latest"})
	require.Error(t, err)
	assert.NotNil(t, err)
}

func TestRunRegistry_TrustedRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io", "gcr.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed for trusted registry")

	registryPolicy = ""
}

func TestRunRegistry_UntrustedRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["gcr.io", "quay.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail for untrusted registry")

	registryPolicy = ""
}

func TestRunRegistry_ExplicitRegistryName(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("docker.io/library/nginx:latest")
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed for explicitly trusted registry")

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
			trustedRegs:   []string{"index.docker.io"},
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

			result, err := runRegistry(tt.imageName)
			require.NoError(t, err)

			if tt.expectSuccess {
				assert.True(t, result.Passed, "Expected validation to succeed")
			} else {
				assert.False(t, result.Passed, "Expected validation to fail")
			}

			registryPolicy = ""
		})
	}
}

func TestRunRegistry_ExcludedRegistries(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"excluded-registries": ["untrusted.io", "malicious.com"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed for non-excluded registry")

	registryPolicy = ""
}

func TestRunRegistry_ExcludedRegistryBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"excluded-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail for excluded registry")

	registryPolicy = ""
}

func TestRunRegistry_OCITransportSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("oci:/path/to/layout:latest")
	require.NoError(t, err)

	details, ok := result.Details.(output.RegistryDetails)
	require.True(t, ok)
	assert.True(t, details.Skipped, "Should skip validation for OCI transport")

	registryPolicy = ""
}

func TestRunRegistry_OCIArchiveTransportSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("oci-archive:/path/to/image.tar:latest")
	require.NoError(t, err)

	details, ok := result.Details.(output.RegistryDetails)
	require.True(t, ok)
	assert.True(t, details.Skipped, "Should skip validation for OCI archive transport")

	registryPolicy = ""
}

func TestRunRegistry_DockerArchiveTransportSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("docker-archive:/path/to/image.tar:nginx")
	require.NoError(t, err)

	details, ok := result.Details.(output.RegistryDetails)
	require.True(t, ok)
	assert.True(t, details.Skipped, "Should skip validation for docker-archive transport")

	registryPolicy = ""
}

func TestRunRegistry_InvalidPolicyFile(t *testing.T) {
	registryPolicy = "/nonexistent/policy.json"

	_, err := runRegistry("nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load registry policy")

	registryPolicy = ""
}

func TestRunRegistry_InvalidPolicyContent(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{invalid json}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	_, err = runRegistry("nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load registry policy")

	registryPolicy = ""
}

func TestRunRegistry_YAMLPolicyFile(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.yaml")
	policyContent := `trusted-registries:
  - index.docker.io
  - gcr.io
`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	result, err := runRegistry("nginx:latest")
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should work with YAML policy file")

	registryPolicy = ""
}

func TestRunRegistry_InvalidImageName(t *testing.T) {
	tmpDir := t.TempDir()
	policyFile := filepath.Join(tmpDir, "policy.json")
	policyContent := `{"trusted-registries": ["index.docker.io"]}`
	err := os.WriteFile(policyFile, []byte(policyContent), 0600)
	require.NoError(t, err)

	registryPolicy = policyFile

	_, err = runRegistry("")
	require.Error(t, err)

	registryPolicy = ""
}
