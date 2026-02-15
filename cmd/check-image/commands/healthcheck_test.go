package commands

import (
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthcheckCommand(t *testing.T) {
	assert.NotNil(t, healthcheckCmd)
	assert.Equal(t, "healthcheck image", healthcheckCmd.Use)
	assert.Contains(t, healthcheckCmd.Short, "healthcheck")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, healthcheckCmd.Args)

	err := healthcheckCmd.Args(healthcheckCmd, []string{})
	assert.Error(t, err)

	err = healthcheckCmd.Args(healthcheckCmd, []string{"image"})
	assert.NoError(t, err)

	err = healthcheckCmd.Args(healthcheckCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestRunHealthcheck(t *testing.T) {
	tests := []struct {
		name         string
		healthcheck  *v1.HealthConfig
		expectedPass bool
		expectedMsg  string
	}{
		{
			name: "Image with CMD-SHELL healthcheck",
			healthcheck: &v1.HealthConfig{
				Test:     []string{"CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"},
				Interval: 30 * time.Second,
				Timeout:  3 * time.Second,
				Retries:  3,
			},
			expectedPass: true,
			expectedMsg:  "Image has a healthcheck defined",
		},
		{
			name: "Image with CMD healthcheck",
			healthcheck: &v1.HealthConfig{
				Test: []string{"CMD", "/health.sh"},
			},
			expectedPass: true,
			expectedMsg:  "Image has a healthcheck defined",
		},
		{
			name:         "Image without healthcheck (nil)",
			healthcheck:  nil,
			expectedPass: false,
			expectedMsg:  "Image does not have a healthcheck defined",
		},
		{
			name: "Image with empty test slice",
			healthcheck: &v1.HealthConfig{
				Test: []string{},
			},
			expectedPass: false,
			expectedMsg:  "Image does not have a healthcheck defined",
		},
		{
			name: "Image with NONE healthcheck (explicitly disabled)",
			healthcheck: &v1.HealthConfig{
				Test: []string{"NONE"},
			},
			expectedPass: false,
			expectedMsg:  "Image does not have a healthcheck defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageRef := createTestImage(t, testImageOptions{
				user:        "1000",
				created:     time.Now(),
				healthcheck: tt.healthcheck,
			})

			result, err := runHealthcheck(imageRef)
			require.NoError(t, err)

			assert.Equal(t, "healthcheck", result.Check)
			assert.Equal(t, imageRef, result.Image)
			assert.Equal(t, tt.expectedPass, result.Passed)
			assert.Equal(t, tt.expectedMsg, result.Message)

			details, ok := result.Details.(output.HealthcheckDetails)
			require.True(t, ok)
			assert.Equal(t, tt.expectedPass, details.HasHealthcheck)
		})
	}
}

func TestRunHealthcheck_InvalidImage(t *testing.T) {
	_, err := runHealthcheck("nonexistent:image")
	require.Error(t, err)
}
