package commands

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootUserCommand(t *testing.T) {
	// Test that root-user command exists and has correct properties
	assert.NotNil(t, rootUserCmd)
	assert.Equal(t, "root-user image", rootUserCmd.Use)
	assert.Contains(t, rootUserCmd.Short, "non-root user")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, rootUserCmd.Args)
	err := rootUserCmd.Args(rootUserCmd, []string{})
	assert.Error(t, err)

	err = rootUserCmd.Args(rootUserCmd, []string{"image"})
	assert.NoError(t, err)

	err = rootUserCmd.Args(rootUserCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestRunRootUser(t *testing.T) {
	tests := []struct {
		name           string
		user           string
		expectedResult ValidationResult
		expectedOutput string
	}{
		{
			name:           "Non-root user (UID)",
			user:           "1000",
			expectedResult: ValidationSucceeded,
			expectedOutput: "Image is configured to run as a non-root user",
		},
		{
			name:           "Non-root user (username)",
			user:           "appuser",
			expectedResult: ValidationSucceeded,
			expectedOutput: "Image is configured to run as a non-root user",
		},
		{
			name:           "Non-root user (UID:GID)",
			user:           "1000:1000",
			expectedResult: ValidationSucceeded,
			expectedOutput: "Image is configured to run as a non-root user",
		},
		{
			name:           "Root user",
			user:           "root",
			expectedResult: ValidationFailed,
			expectedOutput: "Image is not configured to run as a non-root user",
		},
		{
			name:           "Empty user (defaults to root)",
			user:           "",
			expectedResult: ValidationFailed,
			expectedOutput: "Image is not configured to run as a non-root user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image with specific user
			imageRef := createTestImage(t, testImageOptions{
				user:    tt.user,
				created: time.Now(),
			})

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Reset Result
			Result = ValidationSkipped

			// Run command
			err := runRootUser(imageRef)
			require.NoError(t, err)

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			// Assert
			assert.Equal(t, tt.expectedResult, Result)
			assert.Contains(t, buf.String(), tt.expectedOutput)
		})
	}
}

func TestRunRootUser_InvalidImage(t *testing.T) {
	// Test with invalid image reference
	err := runRootUser("nonexistent:image")
	require.Error(t, err)
}
