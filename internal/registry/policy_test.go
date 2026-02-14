package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRegistryPolicy_JSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid allowlist policy",
			content: `{
				"trusted-registries": ["docker.io", "gcr.io"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.TrustedRegistries, 2)
				assert.Contains(t, p.TrustedRegistries, "docker.io")
				assert.Contains(t, p.TrustedRegistries, "gcr.io")
				assert.Empty(t, p.ExcludedRegistries)
			},
		},
		{
			name: "Valid blocklist policy",
			content: `{
				"excluded-registries": ["bad-registry.com"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.ExcludedRegistries, 1)
				assert.Contains(t, p.ExcludedRegistries, "bad-registry.com")
				assert.Empty(t, p.TrustedRegistries)
			},
		},
		{
			name: "Both allowlist and blocklist specified",
			content: `{
				"trusted-registries": ["docker.io"],
				"excluded-registries": ["bad.com"]
			}`,
			wantErr:     true,
			errContains: "not both",
		},
		{
			name:        "Neither allowlist nor blocklist specified",
			content:     `{}`,
			wantErr:     true,
			errContains: "must specify either",
		},
		{
			name:        "Invalid JSON",
			content:     `{invalid json}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.json")
			err := os.WriteFile(policyFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			policy, err := LoadRegistryPolicy(policyFile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestLoadRegistryPolicy_YAML(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid YAML allowlist",
			content: `trusted-registries:
  - docker.io
  - gcr.io
  - quay.io`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.TrustedRegistries, 3)
				assert.Contains(t, p.TrustedRegistries, "docker.io")
				assert.Contains(t, p.TrustedRegistries, "gcr.io")
				assert.Contains(t, p.TrustedRegistries, "quay.io")
			},
		},
		{
			name: "Valid YAML blocklist",
			content: `excluded-registries:
  - untrusted.com
  - malicious.io`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.ExcludedRegistries, 2)
				assert.Contains(t, p.ExcludedRegistries, "untrusted.com")
			},
		},
		{
			name: "Invalid YAML",
			content: `invalid:
  yaml:
  - [unclosed`,
			wantErr:     true,
			errContains: "invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.yaml")
			err := os.WriteFile(policyFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			policy, err := LoadRegistryPolicy(policyFile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestLoadRegistryPolicy_FileErrors(t *testing.T) {
	t.Run("Nonexistent file", func(t *testing.T) {
		_, err := LoadRegistryPolicy("/nonexistent/path/policy.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error reading registry policy")
	})

	t.Run("Directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := LoadRegistryPolicy(tmpDir)
		require.Error(t, err)
	})
}

func TestIsRegistryAllowed_AllowlistMode(t *testing.T) {
	policy := &Policy{
		TrustedRegistries: []string{"docker.io", "gcr.io", "quay.io"},
	}

	tests := []struct {
		name     string
		registry string
		want     bool
	}{
		{
			name:     "Trusted registry docker.io",
			registry: "docker.io",
			want:     true,
		},
		{
			name:     "Trusted registry gcr.io",
			registry: "gcr.io",
			want:     true,
		},
		{
			name:     "Untrusted registry",
			registry: "untrusted.com",
			want:     false,
		},
		{
			name:     "Empty registry",
			registry: "",
			want:     false,
		},
		{
			name:     "Case sensitive mismatch",
			registry: "Docker.io",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.IsRegistryAllowed(tt.registry)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRegistryAllowed_BlocklistMode(t *testing.T) {
	policy := &Policy{
		ExcludedRegistries: []string{"bad-registry.com", "malicious.io"},
	}

	tests := []struct {
		name     string
		registry string
		want     bool
	}{
		{
			name:     "Allowed registry docker.io",
			registry: "docker.io",
			want:     true,
		},
		{
			name:     "Allowed registry gcr.io",
			registry: "gcr.io",
			want:     true,
		},
		{
			name:     "Excluded registry",
			registry: "bad-registry.com",
			want:     false,
		},
		{
			name:     "Another excluded registry",
			registry: "malicious.io",
			want:     false,
		},
		{
			name:     "Empty registry allowed in blocklist mode",
			registry: "",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.IsRegistryAllowed(tt.registry)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsRegistryAllowed_EmptyPolicy(t *testing.T) {
	// This should not happen if LoadRegistryPolicy validation works
	// but test it for completeness
	policy := &Policy{}

	got := policy.IsRegistryAllowed("docker.io")
	assert.False(t, got, "Empty policy should deny all registries")
}

func TestLoadRegistryPolicy_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name:    "JSON from stdin",
			input:   `{"trusted-registries": ["docker.io", "ghcr.io"]}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.TrustedRegistries, 2)
				assert.Contains(t, p.TrustedRegistries, "docker.io")
				assert.Contains(t, p.TrustedRegistries, "ghcr.io")
			},
		},
		{
			name: "YAML from stdin",
			input: `trusted-registries:
  - docker.io
  - gcr.io`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.TrustedRegistries, 2)
				assert.Contains(t, p.TrustedRegistries, "docker.io")
			},
		},
		{
			name:        "Invalid JSON from stdin",
			input:       `{invalid}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "Empty stdin",
			input:       "",
			wantErr:     true,
			errContains: "stdin is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// Create pipe to mock stdin
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdin = r

			// Write test data
			go func() {
				_, _ = w.Write([]byte(tt.input))
				w.Close()
			}()

			policy, err := LoadRegistryPolicy("-")
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestLoadRegistryPolicyFromObject(t *testing.T) {
	tests := []struct {
		name        string
		obj         interface{}
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid allowlist object",
			obj: map[string]any{
				"trusted-registries": []any{"docker.io", "ghcr.io"},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.TrustedRegistries, 2)
				assert.Contains(t, p.TrustedRegistries, "docker.io")
				assert.Contains(t, p.TrustedRegistries, "ghcr.io")
			},
		},
		{
			name: "Valid blocklist object",
			obj: map[string]any{
				"excluded-registries": []any{"bad-registry.com"},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.ExcludedRegistries, 1)
				assert.Contains(t, p.ExcludedRegistries, "bad-registry.com")
			},
		},
		{
			name: "Both allowlist and blocklist",
			obj: map[string]any{
				"trusted-registries":  []any{"docker.io"},
				"excluded-registries": []any{"bad.com"},
			},
			wantErr:     true,
			errContains: "not both",
		},
		{
			name:        "Empty object",
			obj:         map[string]any{},
			wantErr:     true,
			errContains: "must specify either",
		},
		{
			name:        "Invalid object type",
			obj:         "not an object",
			wantErr:     true,
			errContains: "invalid policy object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := LoadRegistryPolicyFromObject(tt.obj)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}
