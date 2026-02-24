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
