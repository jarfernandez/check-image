package commands

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSizeCommand(t *testing.T) {
	// Test that size command exists and has correct properties
	assert.NotNil(t, sizeCmd)
	assert.Equal(t, "size image", sizeCmd.Use)
	assert.Contains(t, sizeCmd.Short, "size")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, sizeCmd.Args)
	err := sizeCmd.Args(sizeCmd, []string{})
	assert.Error(t, err)

	err = sizeCmd.Args(sizeCmd, []string{"image"})
	assert.NoError(t, err)

	err = sizeCmd.Args(sizeCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestSizeCommandFlags(t *testing.T) {
	// Test that max-size flag exists
	flag := sizeCmd.Flags().Lookup("max-size")
	assert.NotNil(t, flag)
	assert.Equal(t, "s", flag.Shorthand)
	assert.Equal(t, "500", flag.DefValue)

	// Test that max-layers flag exists
	flag = sizeCmd.Flags().Lookup("max-layers")
	assert.NotNil(t, flag)
	assert.Equal(t, "y", flag.Shorthand)
	assert.Equal(t, "20", flag.DefValue)
}

func TestRunSize_WithinLimits(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 10 // 10 MB
	maxLayers = 5

	// Create test image with 3 layers, total ~3KB
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 3,
		layerSizes: []int64{1024, 1024, 1024}, // 1KB each
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when within size and layer limits")
}

func TestRunSize_ExceedsSizeLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 1 // 1 MB
	maxLayers = 10

	// Create test image with total size > 1MB
	// Using random data which doesn't compress, so sizes are predictable
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 2,
		layerSizes: []int64{600 * 1024, 600 * 1024}, // 600KB each = 1.2MB total
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when size exceeds limit")
}

func TestRunSize_ExceedsLayerLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 100
	maxLayers = 3

	// Create test image with 5 layers (exceeds limit of 3)
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{1024, 1024, 1024, 1024, 1024},
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when layer count exceeds limit")
}

func TestRunSize_ExceedsBothLimits(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 1 // 1 MB
	maxLayers = 2

	// Create test image exceeding both limits
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{500 * 1024, 500 * 1024, 500 * 1024, 500 * 1024, 500 * 1024},
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when both size and layer count exceed limits")
}

func TestRunSize_NoLayers(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 10
	maxLayers = 5

	// Create test image with no layers
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 0,
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed with no layers")
}

func TestRunSize_ExactlyAtSizeLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 1 // 1 MB = 1048576 bytes
	maxLayers = 5

	// Create test image with exactly 1MB
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024 * 1024}, // Exactly 1 MB
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when exactly at size limit")
}

func TestRunSize_ExactlyAtLayerLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 100
	maxLayers = 3

	// Create test image with exactly 3 layers
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 3,
		layerSizes: []int64{1024, 1024, 1024},
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should succeed when exactly at layer limit")
}

func TestRunSize_OneByteSizeOverLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 1 // 1 MB = 1048576 bytes
	maxLayers = 5

	// Create test image with 1 byte over 1MB
	// Using random data which doesn't compress much
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024*1024 + 1024}, // 1 MB + 1KB (to account for overhead)
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when even 1 byte over size limit")
}

func TestRunSize_OneLayerOverLimit(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 100
	maxLayers = 3

	// Create test image with 4 layers (1 over limit)
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 4,
		layerSizes: []int64{1024, 1024, 1024, 1024},
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should fail when 1 layer over limit")
}

func TestRunSize_PreservesPreviousFailure(t *testing.T) {
	// Set Result to ValidationFailed to simulate a previous failed check
	Result = ValidationFailed
	maxSize = 100
	maxLayers = 10

	// Create test image that would normally pass
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 2,
		layerSizes: []int64{1024, 1024},
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result, "Should preserve previous validation failure")
}

func TestRunSize_InvalidImageReference(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 100
	maxLayers = 10

	// Use invalid image reference
	err := runSize("oci:/nonexistent/path:latest")
	require.Error(t, err)
}

func TestRunSize_VeryLargeImage(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 1000 // 1GB
	maxLayers = 100

	// Create test image with many layers
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 50,
		layerSizes: []int64{1024 * 1024}, // 1MB per layer = 50MB total
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should handle large images with many layers")
}

func TestRunSize_VariableLayerSizes(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = 10 // 10 MB
	maxLayers = 10

	// Create test image with varying layer sizes
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{
			100 * 1024,  // 100 KB
			500 * 1024,  // 500 KB
			1024 * 1024, // 1 MB
			2048 * 1024, // 2 MB
			512 * 1024,  // 512 KB
		}, // Total: ~4.1 MB
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should handle variable layer sizes")
}

func TestRunSize_MaxSizeOverflow(t *testing.T) {
	// Reset global state
	Result = ValidationSkipped
	maxSize = math.MaxInt64/(1024*1024) + 1 // Value that would overflow when converted to bytes
	maxLayers = 10

	// Create test image
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024},
	})

	err := runSize(imageRef)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large", "Should return error for max-size overflow")
}

func TestRunSize_DefaultFlagValues(t *testing.T) {
	// Reset to default values
	Result = ValidationSkipped
	maxSize = 500  // default
	maxLayers = 20 // default

	// Create test image within defaults
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 10,
		layerSizes: []int64{1024 * 1024}, // 1MB per layer = 10MB total
	})

	err := runSize(imageRef)
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result, "Should work with default flag values")
}
