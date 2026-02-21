package commands

import (
	"testing"
	"time"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntrypointCommand(t *testing.T) {
	assert.NotNil(t, entrypointCmd)
	assert.Equal(t, "entrypoint image", entrypointCmd.Use)
	assert.Contains(t, entrypointCmd.Short, "entrypoint")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, entrypointCmd.Args)

	err := entrypointCmd.Args(entrypointCmd, []string{})
	assert.Error(t, err)

	err = entrypointCmd.Args(entrypointCmd, []string{"image"})
	assert.NoError(t, err)

	err = entrypointCmd.Args(entrypointCmd, []string{"image1", "image2"})
	assert.Error(t, err)

	// Test that --allow-shell-form flag exists
	flag := entrypointCmd.Flags().Lookup("allow-shell-form")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestIsShellFormCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		expected bool
	}{
		{
			name:     "shell form with /bin/sh",
			cmd:      []string{"/bin/sh", "-c", "nginx -g 'daemon off;'"},
			expected: true,
		},
		{
			name:     "shell form with /bin/bash",
			cmd:      []string{"/bin/bash", "-c", "nginx -g 'daemon off;'"},
			expected: true,
		},
		{
			name:     "exec form",
			cmd:      []string{"nginx", "-g", "daemon off;"},
			expected: false,
		},
		{
			name:     "exec form with single element",
			cmd:      []string{"/docker-entrypoint.sh"},
			expected: false,
		},
		{
			name:     "/bin/sh without -c (not shell form)",
			cmd:      []string{"/bin/sh"},
			expected: false,
		},
		{
			name:     "empty slice",
			cmd:      []string{},
			expected: false,
		},
		{
			name:     "nil-equivalent empty",
			cmd:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isShellFormCommand(tt.cmd))
		})
	}
}

func TestRunEntrypoint(t *testing.T) {
	tests := []struct {
		name                 string
		entrypoint           []string
		cmd                  []string
		allowShellFormFlag   bool
		expectedPass         bool
		expectedMsg          string
		expectedHas          bool
		expectedExecForm     bool
		expectedShellAllowed bool
	}{
		{
			name:               "exec form entrypoint only",
			entrypoint:         []string{"nginx", "-g", "daemon off;"},
			cmd:                nil,
			allowShellFormFlag: false,
			expectedPass:       true,
			expectedMsg:        "Image has a valid exec-form entrypoint",
			expectedHas:        true,
			expectedExecForm:   true,
		},
		{
			name:               "exec form cmd only (no entrypoint)",
			entrypoint:         nil,
			cmd:                []string{"nginx", "-g", "daemon off;"},
			allowShellFormFlag: false,
			expectedPass:       true,
			expectedMsg:        "Image has a valid exec-form entrypoint",
			expectedHas:        true,
			expectedExecForm:   true,
		},
		{
			name:               "exec form both entrypoint and cmd",
			entrypoint:         []string{"/docker-entrypoint.sh"},
			cmd:                []string{"nginx", "-g", "daemon off;"},
			allowShellFormFlag: false,
			expectedPass:       true,
			expectedMsg:        "Image has a valid exec-form entrypoint",
			expectedHas:        true,
			expectedExecForm:   true,
		},
		{
			name:               "shell form in entrypoint, no allow-shell-form",
			entrypoint:         []string{"/bin/sh", "-c", "nginx -g 'daemon off;'"},
			cmd:                nil,
			allowShellFormFlag: false,
			expectedPass:       false,
			expectedMsg:        "Image uses shell form for entrypoint or cmd",
			expectedHas:        true,
			expectedExecForm:   false,
		},
		{
			name:                 "shell form in entrypoint, with allow-shell-form",
			entrypoint:           []string{"/bin/sh", "-c", "nginx -g 'daemon off;'"},
			cmd:                  nil,
			allowShellFormFlag:   true,
			expectedPass:         true,
			expectedMsg:          "Image uses shell form but it is allowed",
			expectedHas:          true,
			expectedExecForm:     false,
			expectedShellAllowed: true,
		},
		{
			name:               "shell form in cmd, no allow-shell-form",
			entrypoint:         nil,
			cmd:                []string{"/bin/sh", "-c", "nginx -g 'daemon off;'"},
			allowShellFormFlag: false,
			expectedPass:       false,
			expectedMsg:        "Image uses shell form for entrypoint or cmd",
			expectedHas:        true,
			expectedExecForm:   false,
		},
		{
			name:                 "shell form with /bin/bash, with allow-shell-form",
			entrypoint:           []string{"/bin/bash", "-c", "start.sh"},
			cmd:                  nil,
			allowShellFormFlag:   true,
			expectedPass:         true,
			expectedMsg:          "Image uses shell form but it is allowed",
			expectedHas:          true,
			expectedExecForm:     false,
			expectedShellAllowed: true,
		},
		{
			name:               "no entrypoint and no cmd",
			entrypoint:         nil,
			cmd:                nil,
			allowShellFormFlag: false,
			expectedPass:       false,
			expectedMsg:        "Image has no entrypoint or cmd defined",
			expectedHas:        false,
			expectedExecForm:   false,
		},
		{
			name:               "no entrypoint and no cmd with allow-shell-form (still fails)",
			entrypoint:         nil,
			cmd:                nil,
			allowShellFormFlag: true,
			expectedPass:       false,
			expectedMsg:        "Image has no entrypoint or cmd defined",
			expectedHas:        false,
			expectedExecForm:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set package-level flag for this test case
			allowShellForm = tt.allowShellFormFlag
			t.Cleanup(func() { allowShellForm = false })

			imageRef := createTestImage(t, testImageOptions{
				user:       "1000",
				created:    time.Now(),
				entrypoint: tt.entrypoint,
				cmd:        tt.cmd,
			})

			result, err := runEntrypoint(imageRef)
			require.NoError(t, err)

			assert.Equal(t, "entrypoint", result.Check)
			assert.Equal(t, imageRef, result.Image)
			assert.Equal(t, tt.expectedPass, result.Passed)
			assert.Equal(t, tt.expectedMsg, result.Message)

			details, ok := result.Details.(output.EntrypointDetails)
			require.True(t, ok)
			assert.Equal(t, tt.expectedHas, details.HasEntrypoint)
			assert.Equal(t, tt.expectedExecForm, details.ExecForm)
			assert.Equal(t, tt.expectedShellAllowed, details.ShellFormAllowed)

			if tt.expectedHas {
				// Entrypoint and Cmd fields should reflect what was set
				assert.Equal(t, tt.entrypoint, details.Entrypoint)
				assert.Equal(t, tt.cmd, details.Cmd)
			}
		})
	}
}

func TestRunEntrypoint_InvalidImage(t *testing.T) {
	_, err := runEntrypoint("nonexistent:image")
	require.Error(t, err)
}
