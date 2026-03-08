package imageutil

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
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

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic error", errors.New("something went wrong"), false},
		{"net timeout error", &net.DNSError{IsTimeout: true}, true},
		{"net DNS not found", &net.DNSError{Err: "no such host"}, true},
		{"HTTP 503 in message", errors.New("unexpected status code 503 Service Unavailable"), true},
		{"HTTP 429 in message", errors.New("unexpected status code 429 Too Many Requests"), true},
		{"HTTP 500 in message", errors.New("unexpected status code 500 Internal Server Error"), true},
		{"HTTP 502 in message", errors.New("unexpected status code 502 Bad Gateway"), true},
		{"HTTP 504 in message", errors.New("unexpected status code 504 Gateway Timeout"), true},
		{"HTTP 404 not retryable", errors.New("unexpected status code 404 Not Found"), false},
		{"HTTP 401 not retryable", errors.New("unexpected status code 401 Unauthorized"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRetryConstants(t *testing.T) {
	assert.Equal(t, 3, maxRetries)
	assert.Equal(t, 1*time.Second, retryBaseWait)
}

func TestRemoteTransport_IsConfigured(t *testing.T) {
	assert.NotNil(t, remoteTransport, "remoteTransport must be configured")

	// Verify it's an *http.Transport with expected timeouts
	transport, ok := remoteTransport.(*http.Transport)
	require.True(t, ok, "remoteTransport must be an *http.Transport")
	assert.Equal(t, 15*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 30*time.Second, transport.ResponseHeaderTimeout)
}

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

	img, cfg, cleanup, err := GetImageAndConfig(context.Background(), imageName)
	require.NoError(t, err)
	defer cleanup()
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
			img, cleanup, err := GetImage(context.Background(), tt.imageName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			defer cleanup()

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
	_, _, err := GetImage(context.Background(), "invalid reference with spaces")
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
	_, _, _, err := GetImageAndConfig(context.Background(), "invalid reference with spaces")
	require.Error(t, err)
	// The error should come from the name parsing, not from our ParseReference
}

func TestGetImageAndConfig_OCIWithoutTag(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := tmpDir + "/oci-layout"
	createOCILayout(t, layoutPath)

	// OCI transport requires tag or digest
	_, _, _, err := GetImageAndConfig(context.Background(), "oci:"+layoutPath)
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
	loadedImg, cleanup, err := GetImage(context.Background(), "oci-archive:"+tarPath+":test-tag")
	require.NoError(t, err)
	require.NotNil(t, loadedImg)

	// Verify it's the same image by comparing digest
	loadedDigest, err := loadedImg.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)
	cleanup()

	// Test with digest reference
	loadedImg2, cleanup2, err := GetImage(context.Background(), "oci-archive:"+tarPath+"@"+digest.String())
	require.NoError(t, err)
	require.NotNil(t, loadedImg2)
	cleanup2()

	// Test error: missing reference
	_, _, err = GetImage(context.Background(), "oci-archive:"+tarPath)
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
	loadedImg, cleanup, err := GetImage(context.Background(), "docker-archive:"+tarPath)
	require.NoError(t, err)
	defer cleanup()
	require.NotNil(t, loadedImg)

	// Verify it's the same image by comparing digest
	origDigest, err := img.Digest()
	require.NoError(t, err)
	loadedDigest, err := loadedImg.Digest()
	require.NoError(t, err)
	assert.Equal(t, origDigest, loadedDigest)
}

func TestRetryWithBackoff(t *testing.T) {
	const fastWait = time.Microsecond

	t.Run("success on first attempt", func(t *testing.T) {
		calls := 0
		fn := func() (v1.Image, error) {
			calls++
			return nil, nil
		}
		img, err := retryWithBackoff(context.Background(), 3, fastWait, fn)
		require.NoError(t, err)
		assert.Nil(t, img)
		assert.Equal(t, 1, calls)
	})

	t.Run("success after retryable failures", func(t *testing.T) {
		calls := 0
		fn := func() (v1.Image, error) {
			calls++
			if calls < 3 {
				return nil, &net.DNSError{IsTimeout: true}
			}
			return nil, nil
		}
		img, err := retryWithBackoff(context.Background(), 3, fastWait, fn)
		require.NoError(t, err)
		assert.Nil(t, img)
		assert.Equal(t, 3, calls)
	})

	t.Run("non-retryable error fails immediately", func(t *testing.T) {
		calls := 0
		permErr := errors.New("unexpected status code 401 Unauthorized")
		fn := func() (v1.Image, error) {
			calls++
			return nil, permErr
		}
		img, err := retryWithBackoff(context.Background(), 3, fastWait, fn)
		require.Error(t, err)
		assert.Nil(t, img)
		assert.Equal(t, 1, calls)
		assert.ErrorIs(t, err, permErr)
	})

	t.Run("context cancelled during backoff", func(t *testing.T) {
		calls := 0
		ctx, cancel := context.WithCancel(context.Background())
		fn := func() (v1.Image, error) {
			calls++
			cancel() // cancel before returning so ctx.Done() is already closed when select runs
			return nil, &net.DNSError{IsTimeout: true}
		}
		// Use a long backoff so time.After never fires — only ctx.Done() can win the select.
		img, err := retryWithBackoff(ctx, 3, time.Hour, fn)
		require.Error(t, err)
		assert.Nil(t, img)
		assert.Equal(t, 1, calls)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("retries exhausted", func(t *testing.T) {
		calls := 0
		retryableErr := &net.DNSError{IsTimeout: true, Err: "timeout"}
		fn := func() (v1.Image, error) {
			calls++
			return nil, retryableErr
		}
		img, err := retryWithBackoff(context.Background(), 2, fastWait, fn)
		require.Error(t, err)
		assert.Nil(t, img)
		assert.Equal(t, 3, calls)
		assert.ErrorContains(t, err, "after 3 attempts")
		assert.ErrorIs(t, err, retryableErr)
	})
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
	daemonErr := errors.New("daemon unavailable")
	fakeImg, err := random.Image(512, 1)
	require.NoError(t, err)

	tests := []struct {
		name      string
		ctx       context.Context
		localErr  error
		remoteImg v1.Image
		remoteErr error
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "daemon success — no remote fallback",
			ctx:      context.Background(),
			localErr: nil,
			wantErr:  false,
		},
		{
			name:      "daemon fails, remote succeeds",
			ctx:       context.Background(),
			localErr:  daemonErr,
			remoteImg: fakeImg,
			wantErr:   false,
		},
		{
			name:      "daemon fails, remote fails",
			ctx:       context.Background(),
			localErr:  daemonErr,
			remoteErr: errors.New("registry unreachable"),
			wantErr:   true,
		},
		{
			name:      "context cancelled before remote fallback",
			ctx:       func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			localErr:  daemonErr,
			remoteImg: nil, // remote must NOT be called
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteCalled := false

			origLocal := getLocalImageFn
			t.Cleanup(func() { getLocalImageFn = origLocal })
			if tt.localErr == nil {
				getLocalImageFn = func(_ context.Context, _ string) (v1.Image, error) {
					return fakeImg, nil
				}
			} else {
				getLocalImageFn = func(_ context.Context, _ string) (v1.Image, error) {
					return nil, tt.localErr
				}
			}

			origRemote := getRemoteImageFn
			t.Cleanup(func() { getRemoteImageFn = origRemote })
			getRemoteImageFn = func(_ context.Context, _ string) (v1.Image, error) {
				remoteCalled = true
				return tt.remoteImg, tt.remoteErr
			}

			img, cleanup, err := GetImage(tt.ctx, "nginx:latest")
			if err == nil {
				defer cleanup()
			}

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Nil(t, img)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, img)

			if tt.localErr == nil {
				assert.False(t, remoteCalled, "remote should not be called when daemon succeeds")
			}
		})
	}
}

// Tests for GetLocalImage error paths

func TestGetLocalImage_InvalidImageName(t *testing.T) {
	_, err := GetLocalImage(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing the reference")
}

func TestGetLocalImage_InvalidReference(t *testing.T) {
	_, err := GetLocalImage(context.Background(), "INVALID@IMAGE@NAME")
	require.Error(t, err)
}

// Tests for GetRemoteImage error paths

func TestGetRemoteImage_InvalidImageName(t *testing.T) {
	_, err := GetRemoteImage(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing the reference")
}

func TestGetRemoteImage_InvalidReference(t *testing.T) {
	_, err := GetRemoteImage(context.Background(), "INVALID@IMAGE@NAME")
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
	_, _, err := GetOCIArchiveImage("/nonexistent/file.tar", "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error extracting OCI archive")
}

func TestGetOCIArchiveImage_InvalidTarball(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "invalid.tar")

	// Create a file with invalid tar content
	err := os.WriteFile(tarPath, []byte("not a valid tar file"), 0600)
	require.NoError(t, err)

	_, _, err = GetOCIArchiveImage(tarPath, "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error extracting OCI archive")
}

// Tests for GetImage with different transports

func TestGetImage_InvalidReference(t *testing.T) {
	// Test with completely invalid reference
	_, _, err := GetImage(context.Background(), "")
	require.Error(t, err)
}

// TestGetOCIArchiveImage_CleanupRemovesTempDir verifies that the cleanup function
// returned by GetOCIArchiveImage removes the extracted temporary directory.
func TestGetOCIArchiveImage_CleanupRemovesTempDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an OCI layout and pack it into a tarball
	layoutPath := filepath.Join(tmpDir, "layout")
	_, digest := createOCILayoutWithTag(t, layoutPath, "cleanup-test")
	tarPath := filepath.Join(tmpDir, "image.tar")
	createTarballFromOCILayout(t, layoutPath, tarPath)

	// Load the image — this extracts to a new temp dir inside os.TempDir()
	img, cleanup, err := GetOCIArchiveImage(tarPath, digest.String())
	require.NoError(t, err)
	require.NotNil(t, img)

	// Verify the image is usable before cleanup
	_, err = img.Digest()
	require.NoError(t, err)

	// Find the extracted temp dir that was created by GetOCIArchiveImage.
	// We identify it by listing oci-archive-* entries in os.TempDir() before
	// and after cleanup so the test remains independent of implementation details.
	pattern := filepath.Join(os.TempDir(), "oci-archive-*")
	beforeDirs, err := filepath.Glob(pattern)
	require.NoError(t, err)
	require.NotEmpty(t, beforeDirs, "expected at least one oci-archive-* temp dir to exist before cleanup")

	// Call cleanup and verify all previously found dirs are gone
	cleanup()

	afterDirs, err := filepath.Glob(pattern)
	require.NoError(t, err)
	for _, dir := range beforeDirs {
		assert.NotContains(t, afterDirs, dir, "temp dir %s should have been removed by cleanup", dir)
	}
}
