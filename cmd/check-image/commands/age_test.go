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

func TestAgeCommand(t *testing.T) {
	// Test that age command exists and has correct properties
	assert.NotNil(t, ageCmd)
	assert.Equal(t, "age image", ageCmd.Use)
	assert.Contains(t, ageCmd.Short, "age")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, ageCmd.Args)
	err := ageCmd.Args(ageCmd, []string{})
	assert.Error(t, err)

	err = ageCmd.Args(ageCmd, []string{"image"})
	assert.NoError(t, err)

	err = ageCmd.Args(ageCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestAgeCommandFlags(t *testing.T) {
	// Test that max-age flag exists
	flag := ageCmd.Flags().Lookup("max-age")
	assert.NotNil(t, flag)
	assert.Equal(t, "a", flag.Shorthand)
	assert.Equal(t, "90", flag.DefValue)
}

func TestRunAge(t *testing.T) {
	tests := []struct {
		name           string
		imageAge       time.Duration
		maxAge         uint
		expectedResult ValidationResult
		expectedPass   bool
	}{
		{
			name:           "Recent image within limit",
			imageAge:       10 * 24 * time.Hour, // 10 days
			maxAge:         90,
			expectedResult: ValidationSucceeded,
			expectedPass:   true,
		},
		{
			name:           "Image just under limit",
			imageAge:       89 * 24 * time.Hour, // 89 days
			maxAge:         90,
			expectedResult: ValidationSucceeded,
			expectedPass:   true,
		},
		{
			name:           "Old image beyond limit",
			imageAge:       100 * 24 * time.Hour, // 100 days
			maxAge:         90,
			expectedResult: ValidationFailed,
			expectedPass:   false,
		},
		{
			name:           "Very old image",
			imageAge:       365 * 24 * time.Hour, // 1 year
			maxAge:         90,
			expectedResult: ValidationFailed,
			expectedPass:   false,
		},
		{
			name:           "Brand new image",
			imageAge:       1 * time.Hour,
			maxAge:         90,
			expectedResult: ValidationSucceeded,
			expectedPass:   true,
		},
		{
			name:           "Image with strict limit",
			imageAge:       5 * 24 * time.Hour, // 5 days
			maxAge:         1,                  // 1 day limit
			expectedResult: ValidationFailed,
			expectedPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test image with specific creation date
			createdAt := time.Now().Add(-tt.imageAge)
			imageRef := createTestImage(t, testImageOptions{
				user:    "1000",
				created: createdAt,
			})

			// Set max age
			maxAge = tt.maxAge

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Reset Result
			Result = ValidationSkipped

			// Run command
			err := runAge(imageRef)
			require.NoError(t, err)

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Assert
			assert.Equal(t, tt.expectedResult, Result)

			if tt.expectedPass {
				assert.Contains(t, output, "less than")
			} else {
				assert.Contains(t, output, "older than")
			}
		})
	}
}

func TestRunAge_InvalidImage(t *testing.T) {
	// Test with invalid image reference
	err := runAge("nonexistent:image")
	require.Error(t, err)
}
