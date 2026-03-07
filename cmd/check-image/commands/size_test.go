package commands

import (
	"math"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
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
	assert.Equal(t, "m", flag.Shorthand)
	assert.Equal(t, "500", flag.DefValue)

	// Test that max-layers flag exists
	flag = sizeCmd.Flags().Lookup("max-layers")
	assert.NotNil(t, flag)
	assert.Equal(t, "y", flag.Shorthand)
	assert.Equal(t, "20", flag.DefValue)
}

func TestRunSize_WithinLimits(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 3,
		layerSizes: []int64{1024, 1024, 1024},
	})

	result, err := runSize(imageRef, 10, 5)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when within size and layer limits")

	details, ok := result.Details.(output.SizeDetails)
	require.True(t, ok)
	assert.Equal(t, 3, details.LayerCount)
}

func TestRunSize_ExceedsSizeLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 2,
		layerSizes: []int64{600 * 1024, 600 * 1024},
	})

	result, err := runSize(imageRef, 1, 10)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when size exceeds limit")
}

func TestRunSize_ExceedsLayerLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{1024, 1024, 1024, 1024, 1024},
	})

	result, err := runSize(imageRef, 100, 3)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when layer count exceeds limit")
}

func TestRunSize_ExceedsBothLimits(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{500 * 1024, 500 * 1024, 500 * 1024, 500 * 1024, 500 * 1024},
	})

	result, err := runSize(imageRef, 1, 2)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when both size and layer count exceed limits")
}

func TestRunSize_NoLayers(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 0,
	})

	result, err := runSize(imageRef, 10, 5)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed with no layers")
}

func TestRunSize_ExactlyAtSizeLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024 * 1024},
	})

	result, err := runSize(imageRef, 1, 5)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when exactly at size limit")
}

func TestRunSize_ExactlyAtLayerLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 3,
		layerSizes: []int64{1024, 1024, 1024},
	})

	result, err := runSize(imageRef, 100, 3)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when exactly at layer limit")
}

func TestRunSize_OneByteSizeOverLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024*1024 + 1024},
	})

	result, err := runSize(imageRef, 1, 5)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when even 1 byte over size limit")
}

func TestRunSize_OneLayerOverLimit(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 4,
		layerSizes: []int64{1024, 1024, 1024, 1024},
	})

	result, err := runSize(imageRef, 100, 3)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when 1 layer over limit")
}

func TestRunSize_InvalidImageReference(t *testing.T) {
	_, err := runSize("oci:/nonexistent/path:latest", 100, 10)
	require.Error(t, err)
}

func TestRunSize_VeryLargeImage(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 50,
		layerSizes: []int64{1024 * 1024},
	})

	result, err := runSize(imageRef, 1000, 100)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should handle large images with many layers")
}

func TestRunSize_VariableLayerSizes(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 5,
		layerSizes: []int64{
			100 * 1024,
			500 * 1024,
			1024 * 1024,
			2048 * 1024,
			512 * 1024,
		},
	})

	result, err := runSize(imageRef, 10, 10)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should handle variable layer sizes")
}

func TestRunSize_MaxSizeOverflow(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 1,
		layerSizes: []int64{1024},
	})

	_, err := runSize(imageRef, math.MaxInt64/(1024*1024)+1, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large", "Should return error for max-size overflow")
}

func TestRunSize_DefaultFlagValues(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		layerCount: 10,
		layerSizes: []int64{1024 * 1024},
	})

	result, err := runSize(imageRef, 500, 20)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should work with default flag values")
}
