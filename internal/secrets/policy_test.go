package secrets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSecretsPolicy_DefaultPolicy(t *testing.T) {
	// Empty path should return default policy
	policy, err := LoadSecretsPolicy("")
	require.NoError(t, err)
	require.NotNil(t, policy)

	assert.True(t, policy.CheckEnvVars, "Default policy should check env vars")
	assert.True(t, policy.CheckFiles, "Default policy should check files")
	assert.Equal(t, DefaultExcludedEnvVars, policy.ExcludedEnvVars)
	assert.Empty(t, policy.ExcludedPaths)
	assert.Empty(t, policy.CustomEnvPatterns)
}

func TestLoadSecretsPolicy_JSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Full custom policy",
			content: `{
				"check-env-vars": false,
				"check-files": true,
				"excluded-paths": ["/var/log/**", "*.log"],
				"excluded-env-vars": ["MY_PUBLIC_KEY"],
				"custom-env-patterns": ["custom_secret"],
				"custom-file-patterns": ["*.privatekey"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.False(t, p.CheckEnvVars)
				assert.True(t, p.CheckFiles)
				assert.Contains(t, p.ExcludedPaths, "/var/log/**")
				assert.Contains(t, p.ExcludedEnvVars, "MY_PUBLIC_KEY")
				assert.Contains(t, p.CustomEnvPatterns, "custom_secret")
				assert.Contains(t, p.CustomFilePatterns, "*.privatekey")
			},
		},
		{
			name: "Minimal policy with defaults",
			content: `{
				"check-env-vars": true,
				"check-files": true
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.True(t, p.CheckEnvVars)
				assert.True(t, p.CheckFiles)
				// Should get default excluded env vars
				assert.Equal(t, DefaultExcludedEnvVars, p.ExcludedEnvVars)
			},
		},
		{
			name: "Disable both checks",
			content: `{
				"check-env-vars": false,
				"check-files": false
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.False(t, p.CheckEnvVars)
				assert.False(t, p.CheckFiles)
			},
		},
		{
			name:        "Invalid JSON",
			content:     `{invalid}`,
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

			policy, err := LoadSecretsPolicy(policyFile)
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

func TestLoadSecretsPolicy_YAML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(t *testing.T, p *Policy)
	}{
		{
			name: "YAML with exclusions",
			content: `check-env-vars: true
check-files: true
excluded-paths:
  - /tmp/**
  - /var/cache/**
excluded-env-vars:
  - PUBLIC_KEY
  - DISPLAY
custom-env-patterns:
  - apikey
custom-file-patterns:
  - "*.pem"`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.True(t, p.CheckEnvVars)
				assert.True(t, p.CheckFiles)
				assert.Len(t, p.ExcludedPaths, 2)
				assert.Contains(t, p.ExcludedPaths, "/tmp/**")
				assert.Contains(t, p.ExcludedEnvVars, "PUBLIC_KEY")
				assert.Contains(t, p.CustomEnvPatterns, "apikey")
				assert.Contains(t, p.CustomFilePatterns, "*.pem")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.yaml")
			err := os.WriteFile(policyFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			policy, err := LoadSecretsPolicy(policyFile)
			if tt.wantErr {
				require.Error(t, err)
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

func TestLoadSecretsPolicy_FileErrors(t *testing.T) {
	t.Run("Nonexistent file", func(t *testing.T) {
		_, err := LoadSecretsPolicy("/nonexistent/policy.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error reading secrets policy")
	})
}

func TestGetEnvPatterns(t *testing.T) {
	policy := &Policy{
		CustomEnvPatterns: []string{"custom1", "custom2"},
	}

	patterns := policy.GetEnvPatterns()

	// Should contain all default patterns
	for _, defaultPattern := range DefaultEnvPatterns {
		assert.Contains(t, patterns, defaultPattern)
	}

	// Should contain custom patterns
	assert.Contains(t, patterns, "custom1")
	assert.Contains(t, patterns, "custom2")

	// Total length should be defaults + custom
	expectedLen := len(DefaultEnvPatterns) + len(policy.CustomEnvPatterns)
	assert.Len(t, patterns, expectedLen)
}

func TestGetFilePatterns(t *testing.T) {
	policy := &Policy{
		CustomFilePatterns: []string{"*.custom", "custom.key"},
	}

	patterns := policy.GetFilePatterns()

	// Should contain all default patterns
	for defaultPattern := range DefaultFilePatterns {
		assert.Contains(t, patterns, defaultPattern)
	}

	// Should contain custom patterns
	assert.Contains(t, patterns, "*.custom")
	assert.Contains(t, patterns, "custom.key")

	// Minimum length should be defaults + custom
	minLen := len(DefaultFilePatterns) + len(policy.CustomFilePatterns)
	assert.GreaterOrEqual(t, len(patterns), minLen)
}

func TestDefaultFilePatterns(t *testing.T) {
	// Ensure default patterns are defined correctly
	assert.NotEmpty(t, DefaultFilePatterns)

	// Check some important patterns exist
	importantPatterns := []string{
		"id_rsa",
		"id_ed25519",
		".aws/credentials",
		".kube/config",
		"*.key",
	}

	for _, pattern := range importantPatterns {
		description, exists := DefaultFilePatterns[pattern]
		assert.True(t, exists, "Pattern %s should exist in DefaultFilePatterns", pattern)
		assert.NotEmpty(t, description, "Pattern %s should have a description", pattern)
	}
}

func TestDefaultEnvPatterns(t *testing.T) {
	// Ensure default env patterns include common keywords
	assert.NotEmpty(t, DefaultEnvPatterns)

	expectedKeywords := []string{
		"password",
		"secret",
		"token",
		"key",
		"api",
	}

	for _, keyword := range expectedKeywords {
		assert.Contains(t, DefaultEnvPatterns, keyword)
	}
}

func TestDefaultExcludedEnvVars(t *testing.T) {
	// Public keys should be in the default exclusion list
	assert.Contains(t, DefaultExcludedEnvVars, "PUBLIC_KEY")
	assert.Contains(t, DefaultExcludedEnvVars, "SSH_PUBLIC_KEY")
}

func TestLoadSecretsPolicy_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "JSON from stdin",
			input: `{
				"check-env-vars": true,
				"check-files": false,
				"excluded-paths": ["/usr/share/**"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.True(t, p.CheckEnvVars)
				assert.False(t, p.CheckFiles)
				assert.Len(t, p.ExcludedPaths, 1)
				assert.Contains(t, p.ExcludedPaths, "/usr/share/**")
			},
		},
		{
			name: "YAML from stdin",
			input: `check-env-vars: false
check-files: true
excluded-env-vars:
  - PUBLIC_KEY
  - TEST_VAR`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.False(t, p.CheckEnvVars)
				assert.True(t, p.CheckFiles)
				assert.Len(t, p.ExcludedEnvVars, 2)
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

			policy, err := LoadSecretsPolicy("-")
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

func TestLoadSecretsPolicyFromObject(t *testing.T) {
	tests := []struct {
		name     string
		obj      any
		wantErr  bool
		validate func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid policy object",
			obj: map[string]any{
				"check-env-vars": true,
				"check-files":    false,
				"excluded-paths": []any{"/usr/share/**", "/tmp/**"},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.True(t, p.CheckEnvVars)
				assert.False(t, p.CheckFiles)
				assert.Len(t, p.ExcludedPaths, 2)
			},
		},
		{
			name: "Policy with custom patterns",
			obj: map[string]any{
				"check-env-vars":       true,
				"custom-env-patterns":  []any{"my_secret", "my_key"},
				"custom-file-patterns": []any{"*.custom"},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.True(t, p.CheckEnvVars)
				assert.Len(t, p.CustomEnvPatterns, 2)
				assert.Len(t, p.CustomFilePatterns, 1)
			},
		},
		{
			name:    "Empty object uses defaults",
			obj:     map[string]any{},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				// Should use default excluded env vars
				assert.Equal(t, DefaultExcludedEnvVars, p.ExcludedEnvVars)
			},
		},
		{
			name:    "Invalid object type",
			obj:     "not an object",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := LoadSecretsPolicyFromObject(tt.obj)
			if tt.wantErr {
				require.Error(t, err)
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
