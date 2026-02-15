package imageutil

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
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
			imageName: "quay.io/organization/repo:1.0",
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

	_, digest := createOCILayoutWithTag(t, layoutPath, "1.0")

	// Use OCI transport
	imageName := "oci:" + layoutPath + ":1.0"

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

// createTarballFromOCILayout creates a tarball archive from an OCI layout directory
func createTarballFromOCILayout(t *testing.T, layoutPath string, tarPath string) {
	t.Helper()

	tarFile, err := os.Create(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	// Walk the layout directory and add all files to the tarball
	err = filepath.Walk(layoutPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from layoutPath
		relPath, err := filepath.Rel(layoutPath, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself (.)
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file (not a directory), write its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
	require.NoError(t, err)
}

// TestGetImage_OCIArchiveTransport tests the oci-archive transport
func TestGetImage_OCIArchiveTransport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an OCI layout
	layoutPath := filepath.Join(tmpDir, "layout")
	img, digest := createOCILayoutWithTag(t, layoutPath, "test-tag")
	require.NotNil(t, img)

	// Create a tarball from the OCI layout
	tarPath := filepath.Join(tmpDir, "image.tar")
	createTarballFromOCILayout(t, layoutPath, tarPath)

	// Test with tag reference
	loadedImg, err := GetImage("oci-archive:" + tarPath + ":test-tag")
	require.NoError(t, err)
	require.NotNil(t, loadedImg)

	// Verify it's the same image by comparing digest
	loadedDigest, err := loadedImg.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)

	// Test with digest reference
	loadedImg2, err := GetImage("oci-archive:" + tarPath + "@" + digest.String())
	require.NoError(t, err)
	require.NotNil(t, loadedImg2)

	// Test error: missing reference
	_, err = GetImage("oci-archive:" + tarPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires tag or digest")
}

// TestGetImage_DockerArchiveTransport tests the docker-archive transport
func TestGetImage_DockerArchiveTransport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a random test image
	img, err := random.Image(512, 1)
	require.NoError(t, err)

	// Create a Docker archive (tarball) - tarball.WriteToFile needs a tag
	tarPath := filepath.Join(tmpDir, "docker-image.tar")
	tag, err := name.NewTag("test/image:latest")
	require.NoError(t, err)

	err = tarball.WriteToFile(tarPath, tag, img)
	require.NoError(t, err)

	// Load the image from docker-archive without specifying tag
	// (will load the first/only image in the archive)
	loadedImg, err := GetImage("docker-archive:" + tarPath)
	require.NoError(t, err)
	require.NotNil(t, loadedImg)

	// Verify it's the same image by comparing digest
	origDigest, err := img.Digest()
	require.NoError(t, err)
	loadedDigest, err := loadedImg.Digest()
	require.NoError(t, err)
	assert.Equal(t, origDigest, loadedDigest)
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

// Tests for GetLocalImage error paths

func TestGetLocalImage_InvalidImageName(t *testing.T) {
	_, err := GetLocalImage("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing the reference")
}

func TestGetLocalImage_InvalidReference(t *testing.T) {
	_, err := GetLocalImage("INVALID@IMAGE@NAME")
	require.Error(t, err)
}

// Tests for GetRemoteImage error paths

func TestGetRemoteImage_InvalidImageName(t *testing.T) {
	_, err := GetRemoteImage("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing the reference")
}

func TestGetRemoteImage_InvalidReference(t *testing.T) {
	_, err := GetRemoteImage("INVALID@IMAGE@NAME")
	require.Error(t, err)
}

// Tests for GetDockerArchiveImage

func TestGetDockerArchiveImage_NonExistentFile(t *testing.T) {
	_, err := GetDockerArchiveImage("/nonexistent/file.tar", "nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error loading docker archive")
}

func TestGetDockerArchiveImage_InvalidTag(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create an empty tar file
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	_, err = GetDockerArchiveImage(tarPath, "INVALID TAG WITH SPACES")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing tag")
}

func TestGetDockerArchiveImage_EmptyTag(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create an empty tar file
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Empty tag should not error in parsing, but will fail to load
	_, err = GetDockerArchiveImage(tarPath, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error loading docker archive")
}

// Tests for GetOCIArchiveImage

func TestGetOCIArchiveImage_NonExistentFile(t *testing.T) {
	_, err := GetOCIArchiveImage("/nonexistent/file.tar", "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error extracting OCI archive")
}

func TestGetOCIArchiveImage_InvalidTarball(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "invalid.tar")

	// Create a file with invalid tar content
	err := os.WriteFile(tarPath, []byte("not a valid tar file"), 0600)
	require.NoError(t, err)

	_, err = GetOCIArchiveImage(tarPath, "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error extracting OCI archive")
}

// Tests for GetImage with different transports

func TestGetImage_InvalidReference(t *testing.T) {
	// Test with completely invalid reference
	_, err := GetImage("")
	require.Error(t, err)
}
