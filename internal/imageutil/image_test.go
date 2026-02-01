package imageutil

import (
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetImageRegistry(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Docker Hub official image",
			imageName: "nginx:latest",
			want:      "index.docker.io",
			wantErr:   false,
		},
		{
			name:      "Docker Hub with explicit registry",
			imageName: "docker.io/nginx:latest",
			want:      "index.docker.io",
			wantErr:   false,
		},
		{
			name:      "Google Container Registry",
			imageName: "gcr.io/project/image:tag",
			want:      "gcr.io",
			wantErr:   false,
		},
		{
			name:      "Quay.io registry",
			imageName: "quay.io/organization/repo:v1.0",
			want:      "quay.io",
			wantErr:   false,
		},
		{
			name:      "Custom registry with port",
			imageName: "registry.example.com:5000/image:tag",
			want:      "registry.example.com:5000",
			wantErr:   false,
		},
		{
			name:      "Image with digest",
			imageName: "nginx@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			want:      "index.docker.io",
			wantErr:   false,
		},
		{
			name:      "OCI transport not applicable",
			imageName: "oci:/path/to/layout:tag",
			wantErr:   true,
		},
		{
			name:      "OCI archive transport not applicable",
			imageName: "oci-archive:/path/to/image.tar:tag",
			wantErr:   true,
		},
		{
			name:      "Docker archive transport not applicable",
			imageName: "docker-archive:/path/to/image.tar:tag",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImageRegistry(tt.imageName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetImageConfig(t *testing.T) {
	// Create a test image with specific configuration
	baseImg := empty.Image

	// Create config with known values
	cfg, err := baseImg.ConfigFile()
	require.NoError(t, err)

	cfg.Architecture = "amd64"
	cfg.OS = "linux"
	cfg.Config.User = "1000:1000"
	cfg.Config.Env = []string{"PATH=/usr/bin", "HOME=/home/user"}
	cfg.Created = v1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}

	img, err := mutate.ConfigFile(baseImg, cfg)
	require.NoError(t, err)

	// Test GetImageConfig
	loadedConfig, err := GetImageConfig(img)
	require.NoError(t, err)
	require.NotNil(t, loadedConfig)

	assert.Equal(t, "amd64", loadedConfig.Architecture)
	assert.Equal(t, "linux", loadedConfig.OS)
	assert.Equal(t, "1000:1000", loadedConfig.Config.User)
	assert.Contains(t, loadedConfig.Config.Env, "PATH=/usr/bin")
	assert.Equal(t, 2024, loadedConfig.Created.Year())
}

func TestGetImageAndConfig_OCI(t *testing.T) {
	// Test the full flow with OCI layout
	tmpDir := t.TempDir()
	layoutPath := tmpDir + "/oci-layout"

	_, digest := createOCILayoutWithTag(t, layoutPath, "v1.0")

	// Use OCI transport
	imageName := "oci:" + layoutPath + ":v1.0"

	img, cfg, err := GetImageAndConfig(imageName)
	require.NoError(t, err)
	require.NotNil(t, img)
	require.NotNil(t, cfg)

	// Verify we got the right image
	loadedDigest, err := img.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)

	// Config should exist (random images may not have all fields set)
	assert.NotNil(t, cfg)
}

func TestGetImage_OCITransport(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := tmpDir + "/oci-layout"

	_, digest := createOCILayoutWithTag(t, layoutPath, "latest")

	tests := []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "OCI with tag",
			imageName: "oci:" + layoutPath + ":latest",
			wantErr:   false,
		},
		{
			name:      "OCI with digest",
			imageName: "oci:" + layoutPath + "@" + digest.String(),
			wantErr:   false,
		},
		{
			name:      "OCI without tag or digest",
			imageName: "oci:" + layoutPath,
			wantErr:   true,
		},
		{
			name:      "OCI with nonexistent tag",
			imageName: "oci:" + layoutPath + ":nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := GetImage(tt.imageName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, img)

			// Verify we can get config
			cfg, err := img.ConfigFile()
			require.NoError(t, err)
			assert.NotNil(t, cfg)
		})
	}
}

func TestGetImage_InvalidTransport(t *testing.T) {
	// Test with a reference that contains invalid characters
	// References cannot contain spaces or other special characters
	_, err := GetImage("invalid reference with spaces")
	require.Error(t, err)
	// The error should come from the name parsing, not from our ParseReference
}

func TestGetImage_UnsupportedTransport(t *testing.T) {
	// This test would require implementing a new transport type
	// For now, we test that unsupported transports are properly rejected
	// by testing the error path in GetImage

	// All currently defined transports are supported, so we can't easily
	// test this without modifying the code. Skip for now.
	t.Skip("All defined transports are currently supported")
}

func TestGetImageConfig_ValidImage(t *testing.T) {
	// Create random image
	img, err := random.Image(512, 2)
	require.NoError(t, err)

	cfg, err := GetImageConfig(img)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Random images have basic config structure
	assert.NotNil(t, cfg.RootFS)
	// Architecture and OS may or may not be set by random.Image
}

func TestGetImageConfig_EmptyImage(t *testing.T) {
	img := empty.Image

	cfg, err := GetImageConfig(img)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Empty image should still have a valid config
	assert.NotNil(t, cfg.RootFS)
}

func TestGetImageAndConfig_InvalidReference(t *testing.T) {
	// Test with a reference that contains invalid characters
	_, _, err := GetImageAndConfig("invalid reference with spaces")
	require.Error(t, err)
	// The error should come from the name parsing, not from our ParseReference
}

func TestGetImageAndConfig_OCIWithoutTag(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := tmpDir + "/oci-layout"
	createOCILayout(t, layoutPath)

	// OCI transport requires tag or digest
	_, _, err := GetImageAndConfig("oci:" + layoutPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires tag or digest")
}

// TestGetImage_OCIArchiveTransport tests the oci-archive transport
// Note: This requires actual implementation of extractOCIArchive which may not be complete
func TestGetImage_OCIArchiveTransport(t *testing.T) {
	// Skip test if OCI archive functionality is not yet implemented
	t.Skip("OCI archive extraction not yet implemented")
}

// TestGetImage_DockerArchiveTransport tests the docker-archive transport
func TestGetImage_DockerArchiveTransport(t *testing.T) {
	// Skip test - requires actual Docker archive file
	t.Skip("Docker archive test requires actual archive file")
}

// TestGetLocalImage and TestGetRemoteImage would require mocking
// the Docker daemon and registry clients. These are integration tests
// that should be run separately with actual Docker daemon or test registries.

func TestGetLocalImage_RequiresDaemon(t *testing.T) {
	// This test requires Docker daemon - skip in unit tests
	// Users should run integration tests separately
	t.Skip("Requires Docker daemon - run as integration test")
}

func TestGetRemoteImage_RequiresRegistry(t *testing.T) {
	// This test requires network access - skip in unit tests
	// Users should run integration tests separately
	t.Skip("Requires network access - run as integration test")
}

func TestGetImage_DaemonRegistryFallback(t *testing.T) {
	// Testing the fallback behavior would require:
	// 1. Mocking the daemon to return an error
	// 2. Mocking the registry to return success
	// This is better suited for integration tests
	t.Skip("Daemon/registry fallback requires mocking - run as integration test")
}
