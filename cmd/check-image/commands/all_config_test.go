package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCheckNameList(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string]bool
		expectErr bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single valid check",
			input:    "age",
			expected: map[string]bool{"age": true},
		},
		{
			name:     "multiple valid checks",
			input:    "age,size,ports",
			expected: map[string]bool{"age": true, "size": true, "ports": true},
		},
		{
			name:     "all checks",
			input:    "age,size,ports,registry,root-user,secrets",
			expected: map[string]bool{"age": true, "size": true, "ports": true, "registry": true, "root-user": true, "secrets": true},
		},
		{
			name:     "with whitespace",
			input:    " age , size , ports ",
			expected: map[string]bool{"age": true, "size": true, "ports": true},
		},
		{
			name:     "duplicates",
			input:    "age,age,size",
			expected: map[string]bool{"age": true, "size": true},
		},
		{
			name:      "invalid check name",
			input:     "age,invalid",
			expectErr: true,
		},
		{
			name:      "completely invalid",
			input:     "foobar",
			expectErr: true,
		},
		{
			name:     "trailing comma",
			input:    "age,size,",
			expected: map[string]bool{"age": true, "size": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCheckNameList(tt.input)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unknown check name")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadAllConfig(t *testing.T) {
	t.Run("valid YAML config", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, "config.yaml")
		content := `checks:
  age:
    max-age: 30
  size:
    max-size: 200
    max-layers: 10
  root-user: {}
`
		err := os.WriteFile(cfgFile, []byte(content), 0600)
		require.NoError(t, err)

		cfg, err := loadAllConfig(cfgFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.NotNil(t, cfg.Checks.Age)
		require.NotNil(t, cfg.Checks.Age.MaxAge)
		assert.Equal(t, uint(30), *cfg.Checks.Age.MaxAge)

		require.NotNil(t, cfg.Checks.Size)
		require.NotNil(t, cfg.Checks.Size.MaxSize)
		assert.Equal(t, uint(200), *cfg.Checks.Size.MaxSize)
		require.NotNil(t, cfg.Checks.Size.MaxLayers)
		assert.Equal(t, uint(10), *cfg.Checks.Size.MaxLayers)

		require.NotNil(t, cfg.Checks.RootUser)

		assert.Nil(t, cfg.Checks.Ports)
		assert.Nil(t, cfg.Checks.Registry)
		assert.Nil(t, cfg.Checks.Secrets)
	})

	t.Run("valid JSON config", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, "config.json")
		content := `{
			"checks": {
				"age": {"max-age": 60},
				"ports": {"allowed-ports": [80, 443]},
				"secrets": {"secrets-policy": "policy.json", "skip-env-vars": true}
			}
		}`
		err := os.WriteFile(cfgFile, []byte(content), 0600)
		require.NoError(t, err)

		cfg, err := loadAllConfig(cfgFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.NotNil(t, cfg.Checks.Age)
		require.NotNil(t, cfg.Checks.Ports)
		require.NotNil(t, cfg.Checks.Secrets)
		assert.Nil(t, cfg.Checks.Size)
		assert.Nil(t, cfg.Checks.Registry)
		assert.Nil(t, cfg.Checks.RootUser)
	})

	t.Run("root-user with empty object", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, "config.yaml")
		content := `checks:
  root-user: {}
`
		err := os.WriteFile(cfgFile, []byte(content), 0600)
		require.NoError(t, err)

		cfg, err := loadAllConfig(cfgFile)
		require.NoError(t, err)
		require.NotNil(t, cfg.Checks.RootUser)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := loadAllConfig("/nonexistent/config.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("invalid content", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgFile := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(cfgFile, []byte("not valid json"), 0600)
		require.NoError(t, err)

		_, err = loadAllConfig(cfgFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})
}

func TestApplyConfigValues(t *testing.T) {
	t.Run("config values applied when no CLI flags", func(t *testing.T) {
		// Save and reset globals
		origMaxAge := maxAge
		origMaxSize := maxSize
		origMaxLayers := maxLayers
		origAllowedPorts := allowedPorts
		origRegistryPolicy := registryPolicy
		origSecretsPolicy := secretsPolicy
		origSkipEnvVars := skipEnvVars
		origSkipFiles := skipFiles
		defer func() {
			maxAge = origMaxAge
			maxSize = origMaxSize
			maxLayers = origMaxLayers
			allowedPorts = origAllowedPorts
			registryPolicy = origRegistryPolicy
			secretsPolicy = origSecretsPolicy
			skipEnvVars = origSkipEnvVars
			skipFiles = origSkipFiles
		}()

		// Set defaults
		maxAge = 90
		maxSize = 500
		maxLayers = 20
		allowedPorts = ""
		registryPolicy = ""
		secretsPolicy = ""
		skipEnvVars = false
		skipFiles = false

		age := uint(30)
		ms := uint(200)
		ml := uint(10)
		sev := true
		sf := true

		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:      &ageCheckConfig{MaxAge: &age},
				Size:     &sizeCheckConfig{MaxSize: &ms, MaxLayers: &ml},
				Ports:    &portsCheckConfig{AllowedPorts: []any{float64(80), float64(443)}},
				Registry: &registryCheckConfig{RegistryPolicy: "registry-policy.yaml"},
				Secrets:  &secretsCheckConfig{SecretsPolicy: "secrets-policy.yaml", SkipEnvVars: &sev, SkipFiles: &sf},
			},
		}

		// Create a command with no flags changed
		cmd := &cobra.Command{}
		cmd.Flags().UintVar(&maxAge, "max-age", 90, "")
		cmd.Flags().UintVar(&maxSize, "max-size", 500, "")
		cmd.Flags().UintVar(&maxLayers, "max-layers", 20, "")
		cmd.Flags().StringVar(&allowedPorts, "allowed-ports", "", "")
		cmd.Flags().StringVar(&registryPolicy, "registry-policy", "", "")
		cmd.Flags().StringVar(&secretsPolicy, "secrets-policy", "", "")
		cmd.Flags().BoolVar(&skipEnvVars, "skip-env-vars", false, "")
		cmd.Flags().BoolVar(&skipFiles, "skip-files", false, "")

		applyConfigValues(cmd, cfg)

		assert.Equal(t, uint(30), maxAge)
		assert.Equal(t, uint(200), maxSize)
		assert.Equal(t, uint(10), maxLayers)
		assert.Equal(t, "80,443", allowedPorts)
		assert.Equal(t, "registry-policy.yaml", registryPolicy)
		assert.Equal(t, "secrets-policy.yaml", secretsPolicy)
		assert.True(t, skipEnvVars)
		assert.True(t, skipFiles)
	})

	t.Run("CLI flags override config values", func(t *testing.T) {
		origMaxAge := maxAge
		origMaxSize := maxSize
		defer func() {
			maxAge = origMaxAge
			maxSize = origMaxSize
		}()

		age := uint(30)
		ms := uint(200)

		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:  &ageCheckConfig{MaxAge: &age},
				Size: &sizeCheckConfig{MaxSize: &ms},
			},
		}

		// Create command and mark max-age as changed (simulating CLI --max-age 60)
		cmd := &cobra.Command{}
		cmd.Flags().UintVar(&maxAge, "max-age", 90, "")
		cmd.Flags().UintVar(&maxSize, "max-size", 500, "")
		require.NoError(t, cmd.Flags().Set("max-age", "60"))

		applyConfigValues(cmd, cfg)

		// max-age should keep CLI value (60), NOT config value (30)
		assert.Equal(t, uint(60), maxAge)
		// max-size should use config value since CLI didn't change it
		assert.Equal(t, uint(200), maxSize)
	})

	t.Run("nil config fields do not override defaults", func(t *testing.T) {
		origMaxAge := maxAge
		defer func() { maxAge = origMaxAge }()

		maxAge = 90

		cfg := &allConfig{
			Checks: allChecksConfig{
				Age: &ageCheckConfig{MaxAge: nil},
			},
		}

		cmd := &cobra.Command{}
		cmd.Flags().UintVar(&maxAge, "max-age", 90, "")

		applyConfigValues(cmd, cfg)
		assert.Equal(t, uint(90), maxAge)
	})
}

func TestFormatAllowedList(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Numeric slices (ports use case)
		{
			name:     "slice of numbers",
			input:    []any{float64(80), float64(443)},
			expected: "80,443",
		},
		{
			name:     "empty slice",
			input:    []any{},
			expected: "",
		},
		{
			name:     "single port",
			input:    []any{float64(8080)},
			expected: "8080",
		},
		// String slices (platforms use case)
		{
			name:     "slice of platforms",
			input:    []any{"linux/amd64", "linux/arm64"},
			expected: "linux/amd64,linux/arm64",
		},
		{
			name:     "single platform with variant",
			input:    []any{"linux/arm/v7"},
			expected: "linux/arm/v7",
		},
		// Passthrough string
		{
			name:     "string passthrough ports",
			input:    "80,443",
			expected: "80,443",
		},
		{
			name:     "string passthrough platforms",
			input:    "linux/amd64,linux/arm64",
			expected: "linux/amd64,linux/arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAllowedList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyPlatformConfig(t *testing.T) {
	t.Run("config value applied when flag not changed", func(t *testing.T) {
		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = ""

		cmd := &cobra.Command{}
		cmd.Flags().String("allowed-platforms", "", "")

		cfg := &platformCheckConfig{
			AllowedPlatforms: "linux/amd64,linux/arm64",
		}

		applyPlatformConfig(cmd, cfg)
		assert.Equal(t, "linux/amd64,linux/arm64", allowedPlatforms)
	})

	t.Run("inline array applied when flag not changed", func(t *testing.T) {
		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = ""

		cmd := &cobra.Command{}
		cmd.Flags().String("allowed-platforms", "", "")

		cfg := &platformCheckConfig{
			AllowedPlatforms: []any{"linux/amd64", "linux/arm64"},
		}

		applyPlatformConfig(cmd, cfg)
		assert.Equal(t, "linux/amd64,linux/arm64", allowedPlatforms)
	})

	t.Run("config value skipped when flag changed", func(t *testing.T) {
		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "linux/amd64"

		cmd := &cobra.Command{}
		cmd.Flags().String("allowed-platforms", "", "")
		cmd.Flags().Set("allowed-platforms", "linux/amd64")

		cfg := &platformCheckConfig{
			AllowedPlatforms: "linux/arm64",
		}

		applyPlatformConfig(cmd, cfg)
		assert.Equal(t, "linux/amd64", allowedPlatforms)
	})

	t.Run("nil config does nothing", func(t *testing.T) {
		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "linux/amd64"

		cmd := &cobra.Command{}
		cmd.Flags().String("allowed-platforms", "", "")

		applyPlatformConfig(cmd, nil)
		assert.Equal(t, "linux/amd64", allowedPlatforms)
	})
}

func TestInlinePolicyToTempFile_RegistryPolicy(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result string)
	}{
		{
			name:    "String file path",
			input:   "config/registry-policy.json",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "config/registry-policy.json", result)
			},
		},
		{
			name: "Inline object",
			input: map[string]any{
				"trusted-registries": []any{"docker.io", "ghcr.io"},
			},
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// Should create a temp file
				assert.Contains(t, result, "registry-policy-")
				assert.Contains(t, result, ".json")
				// Verify file exists and contains the policy
				data, err := os.ReadFile(result)
				require.NoError(t, err)
				assert.Contains(t, string(data), "trusted-registries")
			},
		},
		{
			name:        "Invalid type",
			input:       123,
			wantErr:     true,
			errContains: "must be either a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, cleanup, err := inlinePolicyToTempFile("registry-policy", tt.input)
			t.Cleanup(cleanup)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestInlinePolicyToTempFile_SecretsPolicy(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result string)
	}{
		{
			name:    "String file path",
			input:   "config/secrets-policy.json",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "config/secrets-policy.json", result)
			},
		},
		{
			name: "Inline object",
			input: map[string]any{
				"check-env-vars": true,
				"check-files":    false,
			},
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// Should create a temp file
				assert.Contains(t, result, "secrets-policy-")
				assert.Contains(t, result, ".json")
				// Verify file exists and contains the policy
				data, err := os.ReadFile(result)
				require.NoError(t, err)
				assert.Contains(t, string(data), "check-env-vars")
			},
		},
		{
			name:        "Invalid type",
			input:       []string{"not", "valid"},
			wantErr:     true,
			errContains: "must be either a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, cleanup, err := inlinePolicyToTempFile("secrets-policy", tt.input)
			t.Cleanup(cleanup)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestLoadAllConfig_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *allConfig)
	}{
		{
			name: "JSON from stdin",
			input: `{
				"checks": {
					"age": {"max-age": 30},
					"size": {"max-size": 200}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *allConfig) {
				require.NotNil(t, cfg.Checks.Age)
				assert.Equal(t, uint(30), *cfg.Checks.Age.MaxAge)
				require.NotNil(t, cfg.Checks.Size)
				assert.Equal(t, uint(200), *cfg.Checks.Size.MaxSize)
			},
		},
		{
			name: "YAML from stdin",
			input: `checks:
  age:
    max-age: 45
  registry:
    registry-policy: policy.json`,
			wantErr: false,
			validate: func(t *testing.T, cfg *allConfig) {
				require.NotNil(t, cfg.Checks.Age)
				assert.Equal(t, uint(45), *cfg.Checks.Age.MaxAge)
				require.NotNil(t, cfg.Checks.Registry)
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

			cfg, err := loadAllConfig("-")

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadAllConfig_InlinePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "inline-config.json")

	// Config with inline registry policy
	content := `{
		"checks": {
			"registry": {
				"registry-policy": {
					"trusted-registries": ["docker.io", "ghcr.io"]
				}
			},
			"secrets": {
				"secrets-policy": {
					"check-env-vars": true,
					"excluded-paths": ["/usr/share/**"]
				}
			}
		}
	}`

	err := os.WriteFile(configFile, []byte(content), 0600)
	require.NoError(t, err)

	cfg, err := loadAllConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify registry policy is loaded as object
	require.NotNil(t, cfg.Checks.Registry)
	require.NotNil(t, cfg.Checks.Registry.RegistryPolicy)
	policyObj, ok := cfg.Checks.Registry.RegistryPolicy.(map[string]any)
	require.True(t, ok, "registry-policy should be a map")
	assert.Contains(t, policyObj, "trusted-registries")

	// Verify secrets policy is loaded as object
	require.NotNil(t, cfg.Checks.Secrets)
	require.NotNil(t, cfg.Checks.Secrets.SecretsPolicy)
	secretsObj, ok := cfg.Checks.Secrets.SecretsPolicy.(map[string]any)
	require.True(t, ok, "secrets-policy should be a map")
	assert.Contains(t, secretsObj, "check-env-vars")
}

func TestInlinePolicyToTempFile_LabelsPolicy(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result string)
	}{
		{
			name:    "String file path",
			input:   "config/labels-policy.json",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "config/labels-policy.json", result)
			},
		},
		{
			name: "Inline object",
			input: map[string]any{
				"required-labels": []any{
					map[string]any{"name": "maintainer"},
					map[string]any{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result string) {
				// Should create a temp file
				assert.Contains(t, result, "labels-policy-")
				assert.Contains(t, result, ".json")
				// Verify file exists and contains the policy
				data, err := os.ReadFile(result)
				require.NoError(t, err)
				assert.Contains(t, string(data), "required-labels")
				assert.Contains(t, string(data), "maintainer")
			},
		},
		{
			name:        "Invalid type",
			input:       123,
			wantErr:     true,
			errContains: "must be either a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, cleanup, err := inlinePolicyToTempFile("labels-policy", tt.input)
			t.Cleanup(cleanup)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestApplyLabelsConfig(t *testing.T) {
	t.Run("config value applied when flag not changed", func(t *testing.T) {
		// Save and reset globals
		origLabelsPolicy := labelsPolicy
		defer func() { labelsPolicy = origLabelsPolicy }()

		// Set default
		labelsPolicy = ""

		// Create a mock command with flags not marked as changed
		cmd := &cobra.Command{}
		cmd.Flags().String("labels-policy", "", "")

		cfg := &labelsCheckConfig{
			LabelsPolicy: "config/labels-policy.yaml",
		}

		applyLabelsConfig(cmd, cfg)
		assert.Equal(t, "config/labels-policy.yaml", labelsPolicy)
	})

	t.Run("config value skipped when flag changed", func(t *testing.T) {
		// Save and reset globals
		origLabelsPolicy := labelsPolicy
		defer func() { labelsPolicy = origLabelsPolicy }()

		// Set CLI flag value
		labelsPolicy = "cli/policy.json"

		// Create a mock command with flag marked as changed
		cmd := &cobra.Command{}
		cmd.Flags().String("labels-policy", "", "")
		cmd.Flags().Set("labels-policy", "cli/policy.json")

		cfg := &labelsCheckConfig{
			LabelsPolicy: "config/labels-policy.yaml",
		}

		applyLabelsConfig(cmd, cfg)
		// Should keep CLI value, not apply config
		assert.Equal(t, "cli/policy.json", labelsPolicy)
	})

	t.Run("nil config does nothing", func(t *testing.T) {
		origLabelsPolicy := labelsPolicy
		defer func() { labelsPolicy = origLabelsPolicy }()

		labelsPolicy = "original"

		cmd := &cobra.Command{}
		cmd.Flags().String("labels-policy", "", "")

		applyLabelsConfig(cmd, nil)
		assert.Equal(t, "original", labelsPolicy)
	})
}

func TestLoadAllConfig_InlineLabelsPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	content := `{
		"checks": {
			"labels": {
				"labels-policy": {
					"required-labels": [
						{"name": "maintainer"},
						{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"}
					]
				}
			}
		}
	}`

	err := os.WriteFile(configFile, []byte(content), 0600)
	require.NoError(t, err)

	cfg, err := loadAllConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify labels policy is loaded as object
	require.NotNil(t, cfg.Checks.Labels)
	require.NotNil(t, cfg.Checks.Labels.LabelsPolicy)
	policyObj, ok := cfg.Checks.Labels.LabelsPolicy.(map[string]any)
	require.True(t, ok, "labels-policy should be a map")
	assert.Contains(t, policyObj, "required-labels")

	// Verify required-labels is a list
	requiredLabels, ok := policyObj["required-labels"].([]any)
	require.True(t, ok, "required-labels should be a list")
	require.Len(t, requiredLabels, 2)
}

func TestLoadAllConfig_Healthcheck(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := `checks:
  age:
    max-age: 30
  root-user: {}
  healthcheck: {}
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	cfg, err := loadAllConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg.Checks.Age)
	require.NotNil(t, cfg.Checks.RootUser)
	require.NotNil(t, cfg.Checks.Healthcheck)
	assert.Nil(t, cfg.Checks.Size)
	assert.Nil(t, cfg.Checks.Ports)
	assert.Nil(t, cfg.Checks.Registry)
	assert.Nil(t, cfg.Checks.Secrets)
	assert.Nil(t, cfg.Checks.Labels)
}

// TestApplyRegistryConfig_InvalidPolicyType tests that applyRegistryConfig handles
// an invalid policy type gracefully by logging an error and leaving registryPolicy unchanged.
func TestApplyRegistryConfig_InvalidPolicyType(t *testing.T) {
	origRegistryPolicy := registryPolicy
	defer func() { registryPolicy = origRegistryPolicy }()

	registryPolicy = "original-policy.json"

	cmd := &cobra.Command{}
	cmd.Flags().String("registry-policy", "", "")

	// Pass an integer as RegistryPolicy to trigger the default case in formatRegistryPolicy
	cfg := &registryCheckConfig{RegistryPolicy: 42}

	applyRegistryConfig(cmd, cfg)

	// registryPolicy should NOT have changed (function returns early on error)
	assert.Equal(t, "original-policy.json", registryPolicy)
}

// TestApplyLabelsConfig_InvalidPolicyType tests that applyLabelsConfig handles
// an invalid policy type gracefully by logging an error and leaving labelsPolicy unchanged.
func TestApplyLabelsConfig_InvalidPolicyType(t *testing.T) {
	origLabelsPolicy := labelsPolicy
	defer func() { labelsPolicy = origLabelsPolicy }()

	labelsPolicy = "original-labels-policy.json"

	cmd := &cobra.Command{}
	cmd.Flags().String("labels-policy", "", "")

	// Pass an integer as LabelsPolicy to trigger the default case in formatLabelsPolicy
	cfg := &labelsCheckConfig{LabelsPolicy: 42}

	applyLabelsConfig(cmd, cfg)

	// labelsPolicy should NOT have changed (function returns early on error)
	assert.Equal(t, "original-labels-policy.json", labelsPolicy)
}

func TestParseAllowedListFromFile(t *testing.T) {
	type dest struct {
		Items []string `json:"items" yaml:"items"`
	}

	tests := []struct {
		name        string
		fileName    string
		content     string
		wantItems   []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid JSON file",
			fileName:  "list.json",
			content:   `{"items": ["a", "b", "c"]}`,
			wantItems: []string{"a", "b", "c"},
		},
		{
			name:      "valid YAML file",
			fileName:  "list.yaml",
			content:   "items:\n  - a\n  - b\n",
			wantItems: []string{"a", "b"},
		},
		{
			name:        "file not found",
			fileName:    "",
			content:     "",
			wantErr:     true,
			errContains: "failed to read file",
		},
		{
			name:        "invalid JSON",
			fileName:    "bad.json",
			content:     `{invalid}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.fileName == "" {
				path = "/nonexistent/path/list.json"
			} else {
				tmpDir := t.TempDir()
				path = filepath.Join(tmpDir, tt.fileName)
				require.NoError(t, os.WriteFile(path, []byte(tt.content), 0600))
			}

			var d dest
			err := parseAllowedListFromFile(path, &d)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantItems, d.Items)
		})
	}
}

func TestApplyInlinePolicy(t *testing.T) {
	t.Run("nil policyVal is a no-op", func(t *testing.T) {
		target := "original"
		cmd := &cobra.Command{}
		cmd.Flags().String("my-policy", "", "")

		cleanup := applyInlinePolicy(cmd, "my-policy", nil, &target)
		t.Cleanup(cleanup)

		assert.Equal(t, "original", target)
	})

	t.Run("flag already changed is a no-op", func(t *testing.T) {
		target := "original"
		cmd := &cobra.Command{}
		cmd.Flags().String("my-policy", "", "")
		require.NoError(t, cmd.Flags().Set("my-policy", "cli-value.json"))

		cleanup := applyInlinePolicy(cmd, "my-policy", "config-value.json", &target)
		t.Cleanup(cleanup)

		assert.Equal(t, "original", target)
	})

	t.Run("string path sets target directly", func(t *testing.T) {
		target := ""
		cmd := &cobra.Command{}
		cmd.Flags().String("my-policy", "", "")

		cleanup := applyInlinePolicy(cmd, "my-policy", "config/policy.json", &target)
		t.Cleanup(cleanup)

		assert.Equal(t, "config/policy.json", target)
	})

	t.Run("inline map creates temp file and sets target", func(t *testing.T) {
		target := ""
		cmd := &cobra.Command{}
		cmd.Flags().String("my-policy", "", "")

		policy := map[string]any{"trusted-registries": []any{"docker.io"}}
		cleanup := applyInlinePolicy(cmd, "my-policy", policy, &target)

		require.NotEmpty(t, target)
		assert.Contains(t, target, "my-policy-")
		assert.Contains(t, target, ".json")
		data, err := os.ReadFile(target)
		require.NoError(t, err)
		assert.Contains(t, string(data), "trusted-registries")

		cleanup()
		_, err = os.Stat(target)
		assert.True(t, os.IsNotExist(err), "temp file should be removed by cleanup")
	})

	t.Run("invalid type logs error and leaves target unchanged", func(t *testing.T) {
		target := "original"
		cmd := &cobra.Command{}
		cmd.Flags().String("my-policy", "", "")

		cleanup := applyInlinePolicy(cmd, "my-policy", 42, &target)
		t.Cleanup(cleanup)

		assert.Equal(t, "original", target)
	})
}

func TestApplyRegistryConfig(t *testing.T) {
	t.Run("config value applied when flag not changed", func(t *testing.T) {
		origRegistryPolicy := registryPolicy
		defer func() { registryPolicy = origRegistryPolicy }()

		registryPolicy = ""

		cmd := &cobra.Command{}
		cmd.Flags().String("registry-policy", "", "")

		cfg := &registryCheckConfig{RegistryPolicy: "config/registry-policy.yaml"}
		cleanup := applyRegistryConfig(cmd, cfg)
		t.Cleanup(cleanup)

		assert.Equal(t, "config/registry-policy.yaml", registryPolicy)
	})

	t.Run("config value skipped when flag changed", func(t *testing.T) {
		origRegistryPolicy := registryPolicy
		defer func() { registryPolicy = origRegistryPolicy }()

		registryPolicy = "cli/policy.json"

		cmd := &cobra.Command{}
		cmd.Flags().String("registry-policy", "", "")
		require.NoError(t, cmd.Flags().Set("registry-policy", "cli/policy.json"))

		cfg := &registryCheckConfig{RegistryPolicy: "config/registry-policy.yaml"}
		cleanup := applyRegistryConfig(cmd, cfg)
		t.Cleanup(cleanup)

		assert.Equal(t, "cli/policy.json", registryPolicy)
	})

	t.Run("nil config does nothing", func(t *testing.T) {
		origRegistryPolicy := registryPolicy
		defer func() { registryPolicy = origRegistryPolicy }()

		registryPolicy = "original"

		cmd := &cobra.Command{}
		cmd.Flags().String("registry-policy", "", "")

		cleanup := applyRegistryConfig(cmd, nil)
		t.Cleanup(cleanup)

		assert.Equal(t, "original", registryPolicy)
	})
}
