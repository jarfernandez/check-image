package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformCommand(t *testing.T) {
	assert.NotNil(t, platformCmd)
	assert.Equal(t, "platform image", platformCmd.Use)
	assert.Contains(t, platformCmd.Short, "platform")

	// Requires exactly 1 argument
	assert.NotNil(t, platformCmd.Args)
	err := platformCmd.Args(platformCmd, []string{})
	assert.Error(t, err)

	err = platformCmd.Args(platformCmd, []string{"image"})
	assert.NoError(t, err)

	err = platformCmd.Args(platformCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestPlatformCommandFlags(t *testing.T) {
	flag := platformCmd.Flags().Lookup("allowed-platforms")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
}

func TestParseAllowedPlatforms_Required(t *testing.T) {
	origAllowedPlatforms := allowedPlatforms
	defer func() { allowedPlatforms = origAllowedPlatforms }()

	allowedPlatforms = ""
	_, err := parseAllowedPlatforms()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--allowed-platforms is required")
}

func TestParseAllowedPlatforms_CommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single platform",
			input:    "linux/amd64",
			expected: []string{"linux/amd64"},
		},
		{
			name:     "multiple platforms",
			input:    "linux/amd64,linux/arm64",
			expected: []string{"linux/amd64", "linux/arm64"},
		},
		{
			name:     "with variant",
			input:    "linux/amd64,linux/arm/v7",
			expected: []string{"linux/amd64", "linux/arm/v7"},
		},
		{
			name:     "with whitespace",
			input:    " linux/amd64 , linux/arm64 ",
			expected: []string{"linux/amd64", "linux/arm64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origAllowedPlatforms := allowedPlatforms
			defer func() { allowedPlatforms = origAllowedPlatforms }()

			allowedPlatforms = tt.input
			result, err := parseAllowedPlatforms()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAllowedPlatforms_FromFile(t *testing.T) {
	t.Run("JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "platforms.json")
		content := `{"allowed-platforms": ["linux/amd64", "linux/arm64"]}`
		err := os.WriteFile(filePath, []byte(content), 0600)
		require.NoError(t, err)

		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "@" + filePath
		result, err := parseAllowedPlatforms()
		require.NoError(t, err)
		assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, result)
	})

	t.Run("YAML file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "platforms.yaml")
		content := "allowed-platforms:\n  - linux/amd64\n  - linux/arm64\n"
		err := os.WriteFile(filePath, []byte(content), 0600)
		require.NoError(t, err)

		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "@" + filePath
		result, err := parseAllowedPlatforms()
		require.NoError(t, err)
		assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, result)
	})

	t.Run("file not found", func(t *testing.T) {
		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "@/nonexistent/platforms.json"
		_, err := parseAllowedPlatforms()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

func TestParseAllowedPlatforms_FromStdin(t *testing.T) {
	t.Run("JSON from stdin", func(t *testing.T) {
		content := `{"allowed-platforms": ["linux/amd64", "linux/arm64"]}`
		r, w, err := os.Pipe()
		require.NoError(t, err)
		_, err = w.WriteString(content)
		require.NoError(t, err)
		require.NoError(t, w.Close())

		origStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = origStdin }()

		origAllowedPlatforms := allowedPlatforms
		defer func() { allowedPlatforms = origAllowedPlatforms }()

		allowedPlatforms = "@-"
		result, err := parseAllowedPlatforms()
		require.NoError(t, err)
		assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, result)
	})
}

func TestRunPlatform_Pass(t *testing.T) {
	origAllowedPlatformsList := allowedPlatformsList
	defer func() { allowedPlatformsList = origAllowedPlatformsList }()

	allowedPlatformsList = []string{"linux/amd64", "linux/arm64"}

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "amd64",
	})

	result, err := runPlatform(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed)
	assert.Equal(t, "platform", result.Check)
	assert.Contains(t, result.Message, "linux/amd64")
	assert.Contains(t, result.Message, "allowed")

	details, ok := result.Details.(output.PlatformDetails)
	require.True(t, ok)
	assert.Equal(t, "linux/amd64", details.Platform)
	assert.Equal(t, []string{"linux/amd64", "linux/arm64"}, details.AllowedPlatforms)
}

func TestRunPlatform_Fail(t *testing.T) {
	origAllowedPlatformsList := allowedPlatformsList
	defer func() { allowedPlatformsList = origAllowedPlatformsList }()

	allowedPlatformsList = []string{"linux/amd64"}

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "arm64",
	})

	result, err := runPlatform(imageRef)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "platform", result.Check)
	assert.Contains(t, result.Message, "linux/arm64")
	assert.Contains(t, result.Message, "not in the allowed list")

	details, ok := result.Details.(output.PlatformDetails)
	require.True(t, ok)
	assert.Equal(t, "linux/arm64", details.Platform)
	assert.Equal(t, []string{"linux/amd64"}, details.AllowedPlatforms)
}

func TestRunPlatform_WithVariant(t *testing.T) {
	origAllowedPlatformsList := allowedPlatformsList
	defer func() { allowedPlatformsList = origAllowedPlatformsList }()

	t.Run("variant matches", func(t *testing.T) {
		allowedPlatformsList = []string{"linux/arm/v7"}

		imageRef := createTestImage(t, testImageOptions{
			os:           "linux",
			architecture: "arm",
			variant:      "v7",
		})

		result, err := runPlatform(imageRef)
		require.NoError(t, err)
		assert.True(t, result.Passed)

		details := result.Details.(output.PlatformDetails)
		assert.Equal(t, "linux/arm/v7", details.Platform)
	})

	t.Run("variant not in list", func(t *testing.T) {
		allowedPlatformsList = []string{"linux/arm64"}

		imageRef := createTestImage(t, testImageOptions{
			os:           "linux",
			architecture: "arm",
			variant:      "v7",
		})

		result, err := runPlatform(imageRef)
		require.NoError(t, err)
		assert.False(t, result.Passed)

		details := result.Details.(output.PlatformDetails)
		assert.Equal(t, "linux/arm/v7", details.Platform)
	})
}

func TestRunPlatform_InvalidImage(t *testing.T) {
	origAllowedPlatformsList := allowedPlatformsList
	defer func() { allowedPlatformsList = origAllowedPlatformsList }()

	allowedPlatformsList = []string{"linux/amd64"}

	_, err := runPlatform("oci:/nonexistent/path:latest")
	require.Error(t, err)
}

func TestRunPlatform_JSONOutput(t *testing.T) {
	origAllowedPlatformsList := allowedPlatformsList
	defer func() { allowedPlatformsList = origAllowedPlatformsList }()

	allowedPlatformsList = []string{"linux/amd64"}

	imageRef := createTestImage(t, testImageOptions{
		os:           "linux",
		architecture: "amd64",
	})

	result, err := runPlatform(imageRef)
	require.NoError(t, err)

	// Verify JSON serialisation uses kebab-case keys
	data, err := json.Marshal(result.Details)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.True(t, strings.Contains(jsonStr, `"platform"`))
	assert.True(t, strings.Contains(jsonStr, `"allowed-platforms"`))
}
