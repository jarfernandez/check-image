package commands

import (
	"testing"
	"time"

	"github.com/jarfernandez/check-image/internal/output"
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
		name         string
		user         string
		expectedPass bool
		expectedMsg  string
	}{
		{
			name:         "Non-root user (UID)",
			user:         "1000",
			expectedPass: true,
			expectedMsg:  "Image is configured to run as a non-root user",
		},
		{
			name:         "Non-root user (username)",
			user:         "appuser",
			expectedPass: true,
			expectedMsg:  "Image is configured to run as a non-root user",
		},
		{
			name:         "Non-root user (UID:GID)",
			user:         "1000:1000",
			expectedPass: true,
			expectedMsg:  "Image is configured to run as a non-root user",
		},
		{
			name:         "Root user",
			user:         "root",
			expectedPass: false,
			expectedMsg:  "Image is not configured to run as a non-root user",
		},
		{
			name:         "Empty user (defaults to root)",
			user:         "",
			expectedPass: false,
			expectedMsg:  "Image is not configured to run as a non-root user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image with specific user
			imageRef := createTestImage(t, testImageOptions{
				user:    tt.user,
				created: time.Now(),
			})

			// Run command
			result, err := runRootUser(imageRef)
			require.NoError(t, err)

			// Assert on struct
			assert.Equal(t, "root-user", result.Check)
			assert.Equal(t, imageRef, result.Image)
			assert.Equal(t, tt.expectedPass, result.Passed)
			assert.Equal(t, tt.expectedMsg, result.Message)

			details, ok := result.Details.(output.RootUserDetails)
			require.True(t, ok)
			assert.Equal(t, tt.user, details.User)
		})
	}
}

func TestRunRootUser_InvalidImage(t *testing.T) {
	// Test with invalid image reference
	_, err := runRootUser("nonexistent:image")
	require.Error(t, err)
}
