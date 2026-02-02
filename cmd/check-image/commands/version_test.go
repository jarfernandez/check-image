package commands

import (
	"bytes"
	"io"
	"os"
	"testing"

	ver "check-image/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		versionValue   string
		expectedOutput string
	}{
		{
			name:           "Version with injected value",
			versionValue:   "v0.1.0",
			expectedOutput: "v0.1.0\n",
		},
		{
			name:           "Version with default dev value",
			versionValue:   "dev",
			expectedOutput: "dev\n",
		},
		{
			name:           "Version with whitespace",
			versionValue:   "  v1.2.3  ",
			expectedOutput: "v1.2.3\n",
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Preserve original version
			originalVersion := ver.Version
			defer func() { ver.Version = originalVersion }()

			// Set the version value
			ver.Version = tt.versionValue

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the version command
			err := runVersion()

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			// Assert
			require.NoError(t, err, "runVersion should not return an error")
			assert.Equal(t, tt.expectedOutput, buf.String(), "Output should match expected version")
		})
	}
}

func TestVersionCommand_Properties(t *testing.T) {
	assert.NotNil(t, versionCmd, "versionCmd should not be nil")
	assert.Equal(t, "version", versionCmd.Use, "Command should have correct Use value")
	assert.Equal(t, "Show the check-image version", versionCmd.Short, "Command should have correct Short description")
	assert.NotEmpty(t, versionCmd.Long, "Command should have Long description")
	assert.NotEmpty(t, versionCmd.Example, "Command should have Example")
}

func TestVersionCommand_NoArgs(t *testing.T) {
	// Test that the command rejects arguments
	err := versionCmd.Args(versionCmd, []string{"extra"})
	assert.Error(t, err, "Command should reject extra arguments")
}

func TestVersionCommand_RunE(t *testing.T) {
	// Preserve original version
	originalVersion := ver.Version
	defer func() { ver.Version = originalVersion }()

	// Set a test version
	ver.Version = "v1.0.0"

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the RunE function
	err := versionCmd.RunE(versionCmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	// Assert
	require.NoError(t, err, "RunE should not return an error")
	assert.Equal(t, "v1.0.0\n", buf.String(), "Output should contain the version")
}

func TestRunVersion_EmptyVersion(t *testing.T) {
	// Preserve original version
	originalVersion := ver.Version
	defer func() { ver.Version = originalVersion }()

	// Test behavior with empty version string
	ver.Version = ""

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute runVersion
	err := runVersion()

	// Restore stdout
	_ = w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	// Assert
	require.NoError(t, err, "runVersion should not return an error with empty version")
	assert.Equal(t, "dev\n", buf.String(), "Output should be a newline for empty version")
}
