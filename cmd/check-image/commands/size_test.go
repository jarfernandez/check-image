package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

// Note: Testing runSize requires actual image with layers which is complex.
// The function relies on GetImage() which would need network/daemon access.
// For unit tests, we focus on testing the command structure and flags.
// Integration tests should cover the actual validation logic.
