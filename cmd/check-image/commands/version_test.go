package commands

import (
	"runtime"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	ver "github.com/jarfernandez/check-image/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveBuildState saves and restores all version package globals and the shortVersion flag.
func saveBuildState(t *testing.T) {
	t.Helper()
	origVersion := ver.Version
	origCommit := ver.Commit
	origBuildDate := ver.BuildDate
	origShort := shortVersion
	origFmt := OutputFmt
	t.Cleanup(func() {
		ver.Version = origVersion
		ver.Commit = origCommit
		ver.BuildDate = origBuildDate
		shortVersion = origShort
		OutputFmt = origFmt
	})
}

func TestVersionCommand_Short_Text(t *testing.T) {
	tests := []struct {
		name           string
		versionValue   string
		expectedOutput string
	}{
		{
			name:           "injected version",
			versionValue:   "v0.1.0",
			expectedOutput: "v0.1.0\n",
		},
		{
			name:           "default dev value",
			versionValue:   "dev",
			expectedOutput: "dev\n",
		},
		{
			name:           "version with whitespace is trimmed",
			versionValue:   "  v1.2.3  ",
			expectedOutput: "v1.2.3\n",
		},
		{
			name:           "empty version falls back to dev",
			versionValue:   "",
			expectedOutput: "dev\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveBuildState(t)
			ver.Version = tt.versionValue
			OutputFmt = output.FormatText
			shortVersion = true

			var err error
			got := captureStdout(t, func() { err = runVersion() })
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}

func TestVersionCommand_Full_Text(t *testing.T) {
	saveBuildState(t)
	ver.Version = "v1.2.3"
	ver.Commit = "abc1234"
	ver.BuildDate = "2026-02-18T12:34:56Z"
	OutputFmt = output.FormatText
	shortVersion = false

	var err error
	got := captureStdout(t, func() { err = runVersion() })
	require.NoError(t, err)

	assert.Contains(t, got, "check-image version v1.2.3")
	assert.Contains(t, got, "commit:     abc1234")
	assert.Contains(t, got, "built at:   2026-02-18T12:34:56Z")
	assert.Contains(t, got, "go version: "+runtime.Version())
	assert.Contains(t, got, "platform:   "+runtime.GOOS+"/"+runtime.GOARCH)
}

func TestVersionCommand_Full_JSON(t *testing.T) {
	saveBuildState(t)
	ver.Version = "v1.2.3"
	ver.Commit = "abc1234"
	ver.BuildDate = "2026-02-18T12:34:56Z"
	OutputFmt = output.FormatJSON
	shortVersion = false

	var err error
	got := captureStdout(t, func() { err = runVersion() })
	require.NoError(t, err)

	assert.Contains(t, got, `"version": "v1.2.3"`)
	assert.Contains(t, got, `"commit": "abc1234"`)
	assert.Contains(t, got, `"built_at": "2026-02-18T12:34:56Z"`)
	assert.Contains(t, got, `"go_version": "`+runtime.Version()+`"`)
	assert.Contains(t, got, `"platform": "`+runtime.GOOS+"/"+runtime.GOARCH+`"`)
}

func TestVersionCommand_Short_JSON(t *testing.T) {
	saveBuildState(t)
	ver.Version = "v1.2.3"
	OutputFmt = output.FormatJSON
	shortVersion = true

	var err error
	got := captureStdout(t, func() { err = runVersion() })
	require.NoError(t, err)

	assert.Contains(t, got, `"version": "v1.2.3"`)
	// Full build fields must NOT appear in --short JSON output
	assert.NotContains(t, got, `"commit"`)
	assert.NotContains(t, got, `"built_at"`)
	assert.NotContains(t, got, `"go_version"`)
	assert.NotContains(t, got, `"platform"`)
}

func TestVersionCommand_Properties(t *testing.T) {
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Show the check-image version", versionCmd.Short)
	assert.NotEmpty(t, versionCmd.Long)
	assert.NotEmpty(t, versionCmd.Example)
	assert.NotNil(t, versionCmd.Flags().Lookup("short"), "--short flag should be registered")
}

func TestVersionCommand_NoArgs(t *testing.T) {
	err := versionCmd.Args(versionCmd, []string{"extra"})
	assert.Error(t, err, "command should reject extra arguments")
}

func TestVersionCommand_RunE(t *testing.T) {
	saveBuildState(t)
	ver.Version = "v1.0.0"
	OutputFmt = output.FormatText
	shortVersion = true

	var err error
	got := captureStdout(t, func() {
		err = versionCmd.RunE(versionCmd, []string{})
	})
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0\n", got)
}
