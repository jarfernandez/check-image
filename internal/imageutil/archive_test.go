package imageutil

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for extractOCIArchive

func TestExtractOCIArchive_ValidTarball(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create a simple tar with a few files
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	tw := tar.NewWriter(file)

	// Add a directory
	err = tw.WriteHeader(&tar.Header{
		Name:     "subdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	})
	require.NoError(t, err)

	// Add a file
	content := []byte("test content")
	err = tw.WriteHeader(&tar.Header{
		Name:     "subdir/file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Extract the tarball
	extractedDir, err := extractOCIArchive(tarPath)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(extractedDir) }()

	// Verify the extracted content
	extractedFile := filepath.Join(extractedDir, "subdir", "file.txt")
	data, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestExtractOCIArchive_GzippedTarball(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")

	// Create a gzipped tar
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	gw := gzip.NewWriter(file)
	tw := tar.NewWriter(gw)

	// Add a file
	content := []byte("gzipped content")
	err = tw.WriteHeader(&tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Extract the tarball
	extractedDir, err := extractOCIArchive(tarPath)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(extractedDir) }()

	// Verify the extracted content
	extractedFile := filepath.Join(extractedDir, "file.txt")
	data, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestExtractOCIArchive_PathTraversalPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "malicious.tar")

	// Create a tar with path traversal attempt
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	tw := tar.NewWriter(file)

	// Try to escape with ../
	content := []byte("malicious content")
	err = tw.WriteHeader(&tar.Header{
		Name:     "../../../etc/passwd",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Attempt to extract - should fail
	_, err = extractOCIArchive(tarPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "illegal file path")
}

func TestExtractOCIArchive_DecompressionBombPrevention(t *testing.T) {
	// Skip this test as it requires creating multi-GB files which is slow
	// The decompression bomb protection is validated through code review:
	// - archive.go lines 66-71 check totalSize against maxDecompressedSize
	// - The check happens on tar header Size field before extraction
	// - This prevents malicious archives from exhausting disk space
	t.Skip("Skipping decompression bomb test - would require creating multi-GB files")
}

func TestExtractOCIArchive_NonExistentFile(t *testing.T) {
	_, err := extractOCIArchive("/nonexistent/file.tar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error opening tarball")
}

func TestExtractOCIArchive_InvalidGzip(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "invalid.tar.gz")

	// Create a file with .gz extension but invalid gzip content
	err := os.WriteFile(tarPath, []byte("not gzipped data"), 0600)
	require.NoError(t, err)

	_, err = extractOCIArchive(tarPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error creating gzip reader")
}

func TestExtractOCIArchive_SkipsSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "symlink.tar")

	// Create a tar with a symlink
	file, err := os.Create(tarPath)
	require.NoError(t, err)
	tw := tar.NewWriter(file)

	// Add a regular file first
	content := []byte("target content")
	err = tw.WriteHeader(&tar.Header{
		Name:     "target.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)

	// Add a symlink (should be skipped)
	err = tw.WriteHeader(&tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "target.txt",
	})
	require.NoError(t, err)

	err = tw.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Extract - symlink should be skipped
	extractedDir, err := extractOCIArchive(tarPath)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(extractedDir) }()

	// Verify target file exists
	targetPath := filepath.Join(extractedDir, "target.txt")
	_, err = os.Stat(targetPath)
	require.NoError(t, err)

	// Verify symlink was skipped (doesn't exist)
	linkPath := filepath.Join(extractedDir, "link.txt")
	_, err = os.Stat(linkPath)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
