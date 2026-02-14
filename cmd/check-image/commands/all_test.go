package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllCommand(t *testing.T) {
	assert.NotNil(t, allCmd)
	assert.Equal(t, "all image", allCmd.Use)
	assert.Contains(t, allCmd.Short, "all")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, allCmd.Args)
	err := allCmd.Args(allCmd, []string{})
	assert.Error(t, err)

	err = allCmd.Args(allCmd, []string{"image"})
	assert.NoError(t, err)

	err = allCmd.Args(allCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestAllCommandFlags(t *testing.T) {
	flag := allCmd.Flags().Lookup("config")
	assert.NotNil(t, flag)
	assert.Equal(t, "c", flag.Shorthand)

	flag = allCmd.Flags().Lookup("skip")
	assert.NotNil(t, flag)

	flag = allCmd.Flags().Lookup("max-age")
	assert.NotNil(t, flag)
	assert.Equal(t, "a", flag.Shorthand)
	assert.Equal(t, "90", flag.DefValue)

	flag = allCmd.Flags().Lookup("max-size")
	assert.NotNil(t, flag)
	assert.Equal(t, "m", flag.Shorthand)
	assert.Equal(t, "500", flag.DefValue)

	flag = allCmd.Flags().Lookup("max-layers")
	assert.NotNil(t, flag)
	assert.Equal(t, "y", flag.Shorthand)
	assert.Equal(t, "20", flag.DefValue)

	flag = allCmd.Flags().Lookup("allowed-ports")
	assert.NotNil(t, flag)
	assert.Equal(t, "p", flag.Shorthand)

	flag = allCmd.Flags().Lookup("registry-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "r", flag.Shorthand)

	flag = allCmd.Flags().Lookup("secrets-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "s", flag.Shorthand)

	flag = allCmd.Flags().Lookup("skip-env-vars")
	assert.NotNil(t, flag)

	flag = allCmd.Flags().Lookup("skip-files")
	assert.NotNil(t, flag)
}

func TestParseSkipList(t *testing.T) {
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
			result, err := parseSkipList(tt.input)
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

func TestDetermineChecks(t *testing.T) {
	t.Run("no config no skip runs all 6 checks", func(t *testing.T) {
		checks := determineChecks(nil, nil)
		assert.Len(t, checks, 6)

		names := make([]string, len(checks))
		for i, c := range checks {
			names[i] = c.name
		}
		assert.Equal(t, []string{"age", "size", "ports", "registry", "root-user", "secrets"}, names)
	})

	t.Run("skip excludes checks", func(t *testing.T) {
		skipMap := map[string]bool{"registry": true, "secrets": true}
		checks := determineChecks(nil, skipMap)
		assert.Len(t, checks, 4)

		for _, c := range checks {
			assert.NotEqual(t, "registry", c.name)
			assert.NotEqual(t, "secrets", c.name)
		}
	})

	t.Run("config selects subset", func(t *testing.T) {
		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:      &ageCheckConfig{},
				RootUser: &rootUserCheckConfig{},
			},
		}
		checks := determineChecks(cfg, nil)
		assert.Len(t, checks, 2)
		assert.Equal(t, "age", checks[0].name)
		assert.Equal(t, "root-user", checks[1].name)
	})

	t.Run("config plus skip", func(t *testing.T) {
		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:      &ageCheckConfig{},
				Size:     &sizeCheckConfig{},
				RootUser: &rootUserCheckConfig{},
			},
		}
		skipMap := map[string]bool{"size": true}
		checks := determineChecks(cfg, skipMap)
		assert.Len(t, checks, 2)
		assert.Equal(t, "age", checks[0].name)
		assert.Equal(t, "root-user", checks[1].name)
	})

	t.Run("skip all gives 0 checks", func(t *testing.T) {
		skipMap := map[string]bool{
			"age": true, "size": true, "ports": true,
			"registry": true, "root-user": true, "secrets": true,
		}
		checks := determineChecks(nil, skipMap)
		assert.Len(t, checks, 0)
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
		rp := "registry-policy.yaml"
		sp := "secrets-policy.yaml"
		sev := true
		sf := true

		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:      &ageCheckConfig{MaxAge: &age},
				Size:     &sizeCheckConfig{MaxSize: &ms, MaxLayers: &ml},
				Ports:    &portsCheckConfig{AllowedPorts: []any{float64(80), float64(443)}},
				Registry: &registryCheckConfig{RegistryPolicy: &rp},
				Secrets:  &secretsCheckConfig{SecretsPolicy: &sp, SkipEnvVars: &sev, SkipFiles: &sf},
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

func TestFormatAllowedPorts(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "slice of numbers",
			input:    []any{float64(80), float64(443)},
			expected: "80,443",
		},
		{
			name:     "string value",
			input:    "80,443",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAllowedPorts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// resetAllGlobals resets all package-level variables used by commands to their defaults.
func resetAllGlobals() {
	Result = ValidationSkipped
	OutputFmt = output.FormatText
	maxAge = 90
	maxSize = 500
	maxLayers = 20
	allowedPorts = ""
	allowedPortsList = nil
	registryPolicy = ""
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false
	configFile = ""
	skipChecks = ""
	failFast = false
}

// captureStdout captures stdout output during fn execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestRunAll_AllChecksPass(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, output, "=== age ===")
	assert.Contains(t, output, "=== size ===")
	assert.Contains(t, output, "=== root-user ===")
	assert.Contains(t, output, "=== secrets ===")
	assert.NotContains(t, output, "=== registry ===")
}

func TestRunAll_OneCheckFails_OthersContinue(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)

	// Image runs as root -> root-user check fails
	imageRef := createTestImage(t, testImageOptions{
		user:       "root",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	// Verify other checks still ran
	assert.Contains(t, output, "=== age ===")
	assert.Contains(t, output, "=== size ===")
	assert.Contains(t, output, "=== root-user ===")
	assert.Contains(t, output, "=== secrets ===")
}

func TestRunAll_SkipFailingCheck(t *testing.T) {
	resetAllGlobals()
	skipChecks = "root-user,registry" // skip root-user (would fail) and registry (requires policy)

	// Image runs as root but we skip root-user check
	imageRef := createTestImage(t, testImageOptions{
		user:       "root",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.NotContains(t, output, "=== root-user ===")
	assert.NotContains(t, output, "=== registry ===")
}

func TestRunAll_ConfigSelectsSubset(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := `checks:
  age:
    max-age: 90
  root-user: {}
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, output, "=== age ===")
	assert.Contains(t, output, "=== root-user ===")
	assert.NotContains(t, output, "=== size ===")
	assert.NotContains(t, output, "=== ports ===")
	assert.NotContains(t, output, "=== registry ===")
	assert.NotContains(t, output, "=== secrets ===")
	assert.Contains(t, output, "Running 2 checks")
}

func TestRunAll_ConfigPlusSkip(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := `checks:
  age:
    max-age: 90
  size:
    max-size: 500
  root-user: {}
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile
	skipChecks = "size"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, output, "=== age ===")
	assert.Contains(t, output, "=== root-user ===")
	assert.NotContains(t, output, "=== size ===")
	assert.Contains(t, output, "Running 2 checks")
}

func TestRunAll_CLIOverridesConfig(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := `checks:
  age:
    max-age: 1
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile

	// Image is 10 days old; config says max-age: 1 (would fail)
	// But CLI overrides with max-age 90 (should pass)
	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now().Add(-10 * 24 * time.Hour),
	})

	// Mark max-age as changed to simulate CLI override
	require.NoError(t, allCmd.Flags().Set("max-age", "90"))
	defer func() {
		// Reset the flag
		require.NoError(t, allCmd.Flags().Set("max-age", "90"))
	}()

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, output, "less than 90 days")
}

func TestRunAll_WithoutConfig_FlagsWork(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)
	maxAge = 1              // Very strict: 1 day

	// Image is 10 days old -> should fail age check
	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	assert.Contains(t, output, "older than 1 days")
}

func TestRunAll_RegistryRequiresPolicy(t *testing.T) {
	resetAllGlobals()

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	// Without --registry-policy and without skipping registry, runAll should error
	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--registry-policy is required")
}

func TestRunAll_PortsWithoutAllowedPorts(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)

	// Image exposes ports but no allowed-ports provided -> ports check fails
	imageRef := createTestImage(t, testImageOptions{
		user:         "1000",
		created:      time.Now().Add(-10 * 24 * time.Hour),
		layerCount:   2,
		exposedPorts: map[string]struct{}{"8080/tcp": {}},
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	assert.Contains(t, output, "No allowed ports were provided")
}

func TestRunAll_SkipAll(t *testing.T) {
	resetAllGlobals()
	skipChecks = "age,size,ports,registry,root-user,secrets"

	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSkipped, Result)
	assert.Contains(t, output, "No checks to run")
}

func TestRunAll_InvalidConfigFile(t *testing.T) {
	resetAllGlobals()
	configFile = "/nonexistent/config.yaml"

	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestRunAll_InvalidSkipList(t *testing.T) {
	resetAllGlobals()
	skipChecks = "age,invalid"

	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check name")
}

func TestAllCommandFailFastFlag(t *testing.T) {
	flag := allCmd.Flags().Lookup("fail-fast")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestRunAll_FailFast_StopsOnValidationFailure(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)
	failFast = true

	// Image runs as root -> root-user check fails
	// age and size run before root-user, secrets runs after
	imageRef := createTestImage(t, testImageOptions{
		user:       "root",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	// root-user should have run and failed
	assert.Contains(t, output, "=== root-user ===")
	// secrets comes after root-user, should NOT have run
	assert.NotContains(t, output, "=== secrets ===")
}

func TestRunAll_FailFast_StopsOnExecutionError(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)
	failFast = true

	// Provide invalid allowed-ports to cause an execution error in ports check
	allowedPorts = "invalid-port"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	// ports should have run (and errored)
	assert.Contains(t, output, "=== ports ===")
	// root-user and secrets come after ports, should NOT have run
	assert.NotContains(t, output, "=== root-user ===")
	assert.NotContains(t, output, "=== secrets ===")
}

func TestRunAll_FailFastDisabled_RunsAllChecks(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry" // skip registry (requires policy file)
	failFast = false

	// Image runs as root -> root-user check fails
	imageRef := createTestImage(t, testImageOptions{
		user:       "root",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	// All checks should have run despite root-user failure
	assert.Contains(t, output, "=== age ===")
	assert.Contains(t, output, "=== size ===")
	assert.Contains(t, output, "=== root-user ===")
	assert.Contains(t, output, "=== secrets ===")
}
