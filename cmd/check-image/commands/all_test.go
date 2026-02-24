package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
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

	flag = allCmd.Flags().Lookup("include")
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

	flag = allCmd.Flags().Lookup("allowed-platforms")
	assert.NotNil(t, flag)
}

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

func TestDetermineChecks(t *testing.T) {
	t.Run("no config no skip runs all 10 checks", func(t *testing.T) {
		checks := determineChecks(nil, nil, nil)
		assert.Len(t, checks, 10)

		names := make([]string, len(checks))
		for i, c := range checks {
			names[i] = c.name
		}
		assert.Equal(t, []string{"age", "size", "ports", "registry", "root-user", "secrets", "healthcheck", "labels", "entrypoint", "platform"}, names)
	})

	t.Run("skip excludes checks", func(t *testing.T) {
		skipMap := map[string]bool{"registry": true, "secrets": true}
		checks := determineChecks(nil, skipMap, nil)
		assert.Len(t, checks, 8)

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
		checks := determineChecks(cfg, nil, nil)
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
		checks := determineChecks(cfg, skipMap, nil)
		assert.Len(t, checks, 2)
		assert.Equal(t, "age", checks[0].name)
		assert.Equal(t, "root-user", checks[1].name)
	})

	t.Run("skip all gives 0 checks", func(t *testing.T) {
		skipMap := map[string]bool{
			"age": true, "size": true, "ports": true,
			"registry": true, "root-user": true, "secrets": true,
			"healthcheck": true, "labels": true, "entrypoint": true,
			"platform": true,
		}
		checks := determineChecks(nil, skipMap, nil)
		assert.Len(t, checks, 0)
	})

	t.Run("all runners have non-nil render function", func(t *testing.T) {
		// Every checkRunner must carry its own renderer so the all command
		// path never falls back to a stringly-typed dispatch.
		checks := determineChecks(nil, nil, nil)
		for _, c := range checks {
			assert.NotNilf(t, c.render, "checkRunner %q has nil render function", c.name)
		}
	})

	t.Run("runners from config also have non-nil render function", func(t *testing.T) {
		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:         &ageCheckConfig{},
				Size:        &sizeCheckConfig{},
				Ports:       &portsCheckConfig{},
				Registry:    &registryCheckConfig{},
				RootUser:    &rootUserCheckConfig{},
				Secrets:     &secretsCheckConfig{},
				Healthcheck: &healthcheckCheckConfig{},
				Labels:      &labelsCheckConfig{},
				Entrypoint:  &entrypointCheckConfig{},
				Platform:    &platformCheckConfig{},
			},
		}
		checks := determineChecks(cfg, nil, nil)
		for _, c := range checks {
			assert.NotNilf(t, c.render, "checkRunner %q has nil render function", c.name)
		}
	})

	t.Run("runners from includeMap have non-nil render function", func(t *testing.T) {
		includeMap := map[string]bool{"age": true, "size": true, "healthcheck": true}
		checks := determineChecks(nil, nil, includeMap)
		for _, c := range checks {
			assert.NotNilf(t, c.render, "checkRunner %q has nil render function", c.name)
		}
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

func TestFormatAllowedPlatforms(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "slice of platforms",
			input:    []any{"linux/amd64", "linux/arm64"},
			expected: "linux/amd64,linux/arm64",
		},
		{
			name:     "string value",
			input:    "linux/amd64,linux/arm64",
			expected: "linux/amd64,linux/arm64",
		},
		{
			name:     "empty slice",
			input:    []any{},
			expected: "",
		},
		{
			name:     "single platform with variant",
			input:    []any{"linux/arm/v7"},
			expected: "linux/arm/v7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAllowedPlatforms(tt.input)
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

// resetAllGlobals resets all package-level variables used by commands to their defaults.
func resetAllGlobals() {
	Result = ValidationSkipped
	OutputFmt = output.FormatText
	colorMode = "auto"
	maxAge = 90
	maxSize = 500
	maxLayers = 20
	allowedPorts = ""
	allowedPortsList = nil
	registryPolicy = ""
	secretsPolicy = ""
	skipEnvVars = false
	skipFiles = false
	allowShellForm = false
	configFile = ""
	skipChecks = ""
	includeChecks = ""
	failFast = false
	allowedPlatforms = ""
	allowedPlatformsList = nil
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
	skipChecks = "registry,labels,platform" // skip checks that require policy files

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
		healthcheck: &v1.HealthConfig{
			Test: []string{"CMD", "/health.sh"},
		},
		entrypoint: []string{"/docker-entrypoint.sh"},
	})

	output := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, output, "── age")
	assert.Contains(t, output, "── size")
	assert.Contains(t, output, "── root-user")
	assert.Contains(t, output, "── secrets")
	assert.NotContains(t, output, "── registry")
}

func TestRunAll_OneCheckFails_OthersContinue(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck

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
	assert.Contains(t, output, "── age")
	assert.Contains(t, output, "── size")
	assert.Contains(t, output, "── root-user")
	assert.Contains(t, output, "── secrets")
}

func TestRunAll_SkipFailingCheck(t *testing.T) {
	resetAllGlobals()
	skipChecks = "root-user,registry,healthcheck,labels,entrypoint,platform" // skip root-user (would fail) and checks that require policy files or missing healthcheck/entrypoint

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
	assert.NotContains(t, output, "── root-user")
	assert.NotContains(t, output, "── registry")
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
	assert.Contains(t, output, "── age")
	assert.Contains(t, output, "── root-user")
	assert.NotContains(t, output, "── size")
	assert.NotContains(t, output, "── ports")
	assert.NotContains(t, output, "── registry")
	assert.NotContains(t, output, "── secrets")
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
	assert.Contains(t, output, "── age")
	assert.Contains(t, output, "── root-user")
	assert.NotContains(t, output, "── size")
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
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck
	maxAge = 1                                          // Very strict: 1 day

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
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck

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
	skipChecks = "age,size,ports,registry,root-user,secrets,healthcheck,labels,entrypoint,platform"

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
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck
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
	assert.Contains(t, output, "── root-user")
	// secrets comes after root-user, should NOT have run
	assert.NotContains(t, output, "── secrets")
}

func TestRunAll_FailFast_StopsOnExecutionError(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck
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

	assert.Equal(t, ExecutionError, Result)
	// ports should have run (and errored)
	assert.Contains(t, output, "── ports")
	// root-user and secrets come after ports, should NOT have run
	assert.NotContains(t, output, "── root-user")
	assert.NotContains(t, output, "── secrets")
}

func TestRunAll_FailFastDisabled_RunsAllChecks(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,healthcheck,labels,platform" // skip checks that require policy files or missing healthcheck
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
	assert.Contains(t, output, "── age")
	assert.Contains(t, output, "── size")
	assert.Contains(t, output, "── root-user")
	assert.Contains(t, output, "── secrets")
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

func TestRunAll_LabelsRequiresPolicy(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,platform" // skip checks that require policy files

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	// Without --labels-policy and without skipping labels, runAll should error
	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--labels-policy is required")
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

func TestRunAll_HealthcheckPassesWithHealthcheck(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,labels,platform" // skip checks that require policy files

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
		healthcheck: &v1.HealthConfig{
			Test: []string{"CMD", "/health.sh"},
		},
		entrypoint: []string{"/docker-entrypoint.sh"},
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── healthcheck")
	assert.Contains(t, out, "Image has a healthcheck defined")
}

func TestRunAll_HealthcheckFailsWithoutHealthcheck(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,labels,platform" // skip checks that require policy files

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	assert.Contains(t, out, "── healthcheck")
	assert.Contains(t, out, "Image does not have a healthcheck defined")
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

func TestDetermineChecks_HealthcheckInConfig(t *testing.T) {
	cfg := &allConfig{
		Checks: allChecksConfig{
			Age:         &ageCheckConfig{},
			RootUser:    &rootUserCheckConfig{},
			Healthcheck: &healthcheckCheckConfig{},
		},
	}
	checks := determineChecks(cfg, nil, nil)
	assert.Len(t, checks, 3)
	assert.Equal(t, "age", checks[0].name)
	assert.Equal(t, "root-user", checks[1].name)
	assert.Equal(t, "healthcheck", checks[2].name)
}

func TestDetermineChecks_IncludeMap(t *testing.T) {
	t.Run("includeMap selects subset", func(t *testing.T) {
		includeMap := map[string]bool{"age": true, "size": true}
		checks := determineChecks(nil, nil, includeMap)
		assert.Len(t, checks, 2)
		assert.Equal(t, "age", checks[0].name)
		assert.Equal(t, "size", checks[1].name)
	})

	t.Run("includeMap overrides config selection", func(t *testing.T) {
		cfg := &allConfig{
			Checks: allChecksConfig{
				Age:      &ageCheckConfig{},
				RootUser: &rootUserCheckConfig{},
			},
		}
		// Config enables age and root-user, but --include asks only for size
		includeMap := map[string]bool{"size": true}
		checks := determineChecks(cfg, nil, includeMap)
		assert.Len(t, checks, 1)
		assert.Equal(t, "size", checks[0].name)
	})

	t.Run("includeMap single check", func(t *testing.T) {
		includeMap := map[string]bool{"healthcheck": true}
		checks := determineChecks(nil, nil, includeMap)
		assert.Len(t, checks, 1)
		assert.Equal(t, "healthcheck", checks[0].name)
	})

	t.Run("includeMap all checks", func(t *testing.T) {
		includeMap := map[string]bool{
			"age": true, "size": true, "ports": true, "registry": true,
			"root-user": true, "secrets": true, "healthcheck": true, "labels": true, "entrypoint": true,
			"platform": true,
		}
		checks := determineChecks(nil, nil, includeMap)
		assert.Len(t, checks, 10)
	})
}

func TestNonRunningCheckNames(t *testing.T) {
	t.Run("with skip map", func(t *testing.T) {
		skipMap := map[string]bool{"age": true, "size": true}
		names := nonRunningCheckNames(skipMap, nil)
		assert.Equal(t, []string{"age", "size"}, names)
	})

	t.Run("with include map", func(t *testing.T) {
		includeMap := map[string]bool{"age": true, "size": true}
		names := nonRunningCheckNames(nil, includeMap)
		assert.Len(t, names, 8)
		assert.NotContains(t, names, "age")
		assert.NotContains(t, names, "size")
		assert.Contains(t, names, "ports")
		assert.Contains(t, names, "registry")
		assert.Contains(t, names, "root-user")
		assert.Contains(t, names, "secrets")
		assert.Contains(t, names, "healthcheck")
		assert.Contains(t, names, "labels")
		assert.Contains(t, names, "entrypoint")
		assert.Contains(t, names, "platform")
	})

	t.Run("with neither", func(t *testing.T) {
		names := nonRunningCheckNames(nil, nil)
		assert.Nil(t, names)
	})

	t.Run("include all checks returns nil", func(t *testing.T) {
		includeMap := map[string]bool{
			"age": true, "size": true, "ports": true, "registry": true,
			"root-user": true, "secrets": true, "healthcheck": true, "labels": true, "entrypoint": true,
			"platform": true,
		}
		names := nonRunningCheckNames(nil, includeMap)
		assert.Nil(t, names)
	})

	t.Run("empty skip map returns nil", func(t *testing.T) {
		names := nonRunningCheckNames(map[string]bool{}, nil)
		assert.Nil(t, names)
	})
}

func TestRunAll_IncludeAndSkipMutuallyExclusive(t *testing.T) {
	resetAllGlobals()
	skipChecks = "age"
	includeChecks = "size"

	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestRunAll_IncludeSelectsSubset(t *testing.T) {
	resetAllGlobals()
	includeChecks = "age,root-user"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── age")
	assert.Contains(t, out, "── root-user")
	assert.NotContains(t, out, "── size")
	assert.NotContains(t, out, "── ports")
	assert.NotContains(t, out, "── registry")
	assert.NotContains(t, out, "── secrets")
	assert.NotContains(t, out, "── healthcheck")
	assert.NotContains(t, out, "── labels")
	assert.Contains(t, out, "Running 2 checks")
}

func TestRunAll_IncludeOverridesConfig(t *testing.T) {
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
	// Config enables age and root-user, but --include overrides to only size
	includeChecks = "size"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── size")
	assert.NotContains(t, out, "── age")
	assert.NotContains(t, out, "── root-user")
	assert.Contains(t, out, "Running 1 checks")
}

func TestRunAll_IncludeWithConfig_ValuesApply(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	// Use max-layers instead of max-age because allCmd's max-age flag may be
	// permanently marked as Changed by TestRunAll_CLIOverridesConfig.
	content := `checks:
  size:
    max-layers: 1
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile
	// --include selects size, config sets max-layers: 1
	includeChecks = "size"

	// Image has 3 layers; config says max-layers: 1 -> should fail
	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 3,
	})

	captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
}

func TestRunAll_InvalidIncludeList(t *testing.T) {
	resetAllGlobals()
	includeChecks = "age,invalid"

	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check name")
}

func TestRunAll_EntrypointPassesWithExecForm(t *testing.T) {
	resetAllGlobals()
	includeChecks = "entrypoint" // run only entrypoint check

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now(),
		entrypoint: []string{"/docker-entrypoint.sh"},
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── entrypoint")
	assert.Contains(t, out, "exec-form entrypoint")
}

func TestRunAll_EntrypointFailsWithShellForm(t *testing.T) {
	resetAllGlobals()
	includeChecks = "entrypoint"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now(),
		entrypoint: []string{"/bin/sh", "-c", "nginx -g 'daemon off;'"},
	})

	captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
}

func TestRunAll_EntrypointSkipped(t *testing.T) {
	resetAllGlobals()
	skipChecks = "registry,healthcheck,labels,entrypoint,platform"

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now().Add(-10 * 24 * time.Hour),
		layerCount: 2,
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.NotContains(t, out, "── entrypoint")
}

func TestRunAll_EntrypointWithAllowShellFormViaConfig(t *testing.T) {
	resetAllGlobals()

	allow := true
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := `checks:
  entrypoint:
    allow-shell-form: true
`
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile
	_ = allow // used via config

	imageRef := createTestImage(t, testImageOptions{
		user:       "1000",
		created:    time.Now(),
		entrypoint: []string{"/bin/sh", "-c", "nginx"},
	})

	captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	// With allow-shell-form: true from config, shell form should pass
	assert.Equal(t, ValidationSucceeded, Result)
}

func TestRunAll_PlatformPasses(t *testing.T) {
	resetAllGlobals()
	includeChecks = "platform"
	allowedPlatforms = "linux/amd64,linux/arm64"

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "amd64",
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── platform")
	assert.Contains(t, out, "linux/amd64")
}

func TestRunAll_PlatformFails(t *testing.T) {
	resetAllGlobals()
	includeChecks = "platform"
	allowedPlatforms = "linux/amd64"

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "arm64",
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationFailed, Result)
	assert.Contains(t, out, "── platform")
	assert.Contains(t, out, "not in the allowed list")
}

func TestRunAll_PlatformFailsWhenNoPlatformsProvided(t *testing.T) {
	resetAllGlobals()
	includeChecks = "platform"
	// allowedPlatforms is "" after reset

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "amd64",
	})

	err := runAll(allCmd, imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--allowed-platforms is required")
}

func TestRunAll_PlatformWithConfig(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := "checks:\n  platform:\n    allowed-platforms: \"linux/amd64,linux/arm64\"\n"
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "amd64",
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "── platform")
}

func TestDetermineChecks_PlatformInConfig(t *testing.T) {
	cfg := &allConfig{
		Checks: allChecksConfig{
			Platform: &platformCheckConfig{
				AllowedPlatforms: "linux/amd64",
			},
		},
	}

	checks := determineChecks(cfg, nil, nil)
	names := make([]string, 0, len(checks))
	for _, c := range checks {
		names = append(names, c.name)
	}

	assert.Contains(t, names, "platform")
	assert.NotContains(t, names, "age")
	assert.NotContains(t, names, "registry")
}

func TestRunAll_PlatformWithInlineConfig(t *testing.T) {
	resetAllGlobals()

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	content := "checks:\n  platform:\n    allowed-platforms:\n      - linux/amd64\n      - linux/arm64\n"
	err := os.WriteFile(cfgFile, []byte(content), 0600)
	require.NoError(t, err)

	configFile = cfgFile

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "arm64",
	})

	// arm64 is in the allowed list -> should pass
	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "linux/arm64")
}

// TestRenderAllJSON_AllPassing tests renderAllJSON when all checks pass.
func TestRenderAllJSON_AllPassing(t *testing.T) {
	resetAllGlobals()
	Result = ValidationSucceeded

	results := []output.CheckResult{
		{Check: "age", Image: "nginx:latest", Passed: true, Message: "Image is recent"},
		{Check: "size", Image: "nginx:latest", Passed: true, Message: "Image size ok"},
	}

	captured := captureStdout(t, func() {
		err := renderAllJSON("nginx:latest", results, nil, nil)
		require.NoError(t, err)
	})

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(captured), &data))
	assert.Equal(t, "nginx:latest", data["image"])
	assert.Equal(t, true, data["passed"])
	summary := data["summary"].(map[string]any)
	assert.Equal(t, float64(2), summary["total"])
	assert.Equal(t, float64(2), summary["passed"])
	assert.Equal(t, float64(0), summary["failed"])
	assert.Equal(t, float64(0), summary["errored"])
	assert.Nil(t, summary["skipped"]) // no skipped checks
}

// TestRenderAllJSON_WithFailures tests renderAllJSON when some checks fail or error.
func TestRenderAllJSON_WithFailures(t *testing.T) {
	resetAllGlobals()
	Result = ValidationFailed

	results := []output.CheckResult{
		{Check: "age", Image: "nginx:latest", Passed: true, Message: "Image is recent"},
		{Check: "root-user", Image: "nginx:latest", Passed: false, Message: "Image runs as root"},
		{Check: "size", Image: "nginx:latest", Passed: false, Message: "check failed with error: some error", Error: "some error"},
	}

	captured := captureStdout(t, func() {
		err := renderAllJSON("nginx:latest", results, nil, nil)
		require.NoError(t, err)
	})

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(captured), &data))
	assert.Equal(t, false, data["passed"])
	summary := data["summary"].(map[string]any)
	assert.Equal(t, float64(3), summary["total"])
	assert.Equal(t, float64(1), summary["passed"])
	assert.Equal(t, float64(1), summary["failed"])
	assert.Equal(t, float64(1), summary["errored"])
}

// TestRenderAllJSON_WithSkipMap tests renderAllJSON with a skip map.
func TestRenderAllJSON_WithSkipMap(t *testing.T) {
	resetAllGlobals()
	Result = ValidationSucceeded

	results := []output.CheckResult{
		{Check: "age", Image: "nginx:latest", Passed: true, Message: "Image is recent"},
	}
	skipMap := map[string]bool{"registry": true, "secrets": true}

	captured := captureStdout(t, func() {
		err := renderAllJSON("nginx:latest", results, skipMap, nil)
		require.NoError(t, err)
	})

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(captured), &data))
	assert.Equal(t, true, data["passed"])
	summary := data["summary"].(map[string]any)
	skipped := summary["skipped"].([]any)
	assert.Contains(t, skipped, "registry")
	assert.Contains(t, skipped, "secrets")
}

// TestRenderAllJSON_WithIncludeMap tests renderAllJSON with an include map.
func TestRenderAllJSON_WithIncludeMap(t *testing.T) {
	resetAllGlobals()
	Result = ValidationSucceeded

	results := []output.CheckResult{
		{Check: "age", Image: "nginx:latest", Passed: true, Message: "Image is recent"},
	}
	includeMap := map[string]bool{"age": true}

	captured := captureStdout(t, func() {
		err := renderAllJSON("nginx:latest", results, nil, includeMap)
		require.NoError(t, err)
	})

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(captured), &data))
	assert.Equal(t, true, data["passed"])
	summary := data["summary"].(map[string]any)
	// All checks except "age" should appear in skipped
	skipped := summary["skipped"].([]any)
	assert.NotContains(t, skipped, "age")
	assert.Contains(t, skipped, "size")
	assert.Contains(t, skipped, "registry")
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

// TestRunAll_EntrypointWithCmdField tests that the entrypoint check renders
// both Entrypoint and Cmd fields when an image defines both.
func TestRunAll_EntrypointWithCmdField(t *testing.T) {
	resetAllGlobals()
	includeChecks = "entrypoint"

	imageRef := createTestImage(t, testImageOptions{
		entrypoint: []string{"/docker-entrypoint.sh"},
		cmd:        []string{"nginx", "-g", "daemon off;"},
	})

	out := captureStdout(t, func() {
		err := runAll(allCmd, imageRef)
		require.NoError(t, err)
	})

	assert.Equal(t, ValidationSucceeded, Result)
	assert.Contains(t, out, "Entrypoint:")
	assert.Contains(t, out, "Cmd:")
}
