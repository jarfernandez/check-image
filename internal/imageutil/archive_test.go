package imageutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
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

// TestExtractEntries_DecompressionBombPrevention verifies that extractEntries
// rejects a tar whose entries collectively declare a size exceeding the 5 GB
// limit, without actually writing any large amount of data. The check fires on
// the tar header Size field, so a single header claiming >5 GB is sufficient to
// trigger the protection immediately — no real data needs to be stored on disk.
func TestExtractEntries_DecompressionBombPrevention(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Declare a single entry whose Size alone exceeds the 5 GB limit.
	const overLimit = maxDecompressedSize + 1
	err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "bomb.txt",
		Size:     overLimit,
		Mode:     0644,
	})
	require.NoError(t, err)
	// Do not flush or close tw — doing so would validate that overLimit bytes
	// were written for the entry and return an error. The header bytes are
	// already in buf, which is all tar.Reader.Next() needs to read Size.

	tr := tar.NewReader(&buf)
	err = extractEntries(tr, t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
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

// TestExtractRegularFile_SizeMismatch verifies the size mismatch error path in
// extractRegularFile. We pass a fake tar.Header claiming 100 bytes while the
// tarReader only has 5 bytes of actual content for the entry.
func TestExtractRegularFile_SizeMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar")

	// Create a tar with a 5-byte file
	content := []byte("hello")
	f, err := os.Create(tarPath)
	require.NoError(t, err)
	tw := tar.NewWriter(f)
	err = tw.WriteHeader(&tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	})
	require.NoError(t, err)
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, f.Close())

	// Open the tar and advance to the file entry
	tarFile, err := os.Open(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	_, err = tarReader.Next()
	require.NoError(t, err)

	// Pass a fake header that claims the file is 100 bytes (larger than actual 5 bytes).
	// io.Copy will read 5 bytes from the tar reader (its actual limit) and return
	// written=5. The mismatch with fakeHeader.Size=100 triggers the size mismatch error.
	fakeHeader := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     100,
	}

	targetPath := filepath.Join(tmpDir, "extracted.txt")
	err = extractRegularFile(tarReader, targetPath, fakeHeader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "size mismatch")
}

// TestExtractRegularFile_OversizedContent verifies that LimitReader prevents
// unbounded disk writes when a tar entry's actual content is larger than the
// declared header.Size. A lying header could bypass the aggregate size check
// (which trusts header.Size) while io.Copy writes unlimited data to disk.
// With the LimitReader fix, the write is capped at header.Size+1 bytes, and
// the existing size-mismatch check then rejects the entry with an error.
// This test confirms both: (1) an error is returned, and (2) the extracted
// file is at most header.Size+1 bytes — not the full 1000 bytes of actual data.
func TestExtractRegularFile_OversizedContent(t *testing.T) {
	// Build an in-memory tar with 1000 bytes of actual content.
	actualContent := bytes.Repeat([]byte("A"), 1000)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(actualContent)),
	})
	require.NoError(t, err)
	_, err = tw.Write(actualContent)
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	// Advance the tar reader to the entry.
	tarReader := tar.NewReader(&buf)
	_, err = tarReader.Next()
	require.NoError(t, err)

	// Pass a fake header claiming only 10 bytes — the lying-header attack.
	fakeHeader := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     10,
	}

	targetPath := filepath.Join(t.TempDir(), "extracted.txt")
	err = extractRegularFile(tarReader, targetPath, fakeHeader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "size mismatch")

	// Key assertion: LimitReader must have stopped the write at header.Size+1=11 bytes.
	// Without the LimitReader fix the file would contain all 1000 bytes of actual data.
	info, statErr := os.Stat(targetPath)
	require.NoError(t, statErr)
	assert.LessOrEqual(t, info.Size(), int64(fakeHeader.Size+1),
		"file on disk must be capped at header.Size+1 bytes, not the full actual content")
}

// errCloseWriter is a WriteCloser whose writes succeed (in-memory buffer)
// but whose Close always returns an error. Used to test the close-error path
// in extractRegularFile without requiring OS-level fault injection.
type errCloseWriter struct {
	buf bytes.Buffer
}

func (w *errCloseWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *errCloseWriter) Close() error                { return errors.New("mock close error") }

// TestExtractRegularFile_CloseError verifies the close-error path in
// extractRegularFile by injecting a WriteCloser whose Close() always fails.
func TestExtractRegularFile_CloseError(t *testing.T) {
	orig := openFileFn
	t.Cleanup(func() { openFileFn = orig })
	openFileFn = func(_ string, _ int, _ os.FileMode) (io.WriteCloser, error) {
		return &errCloseWriter{}, nil
	}

	// Build a minimal in-memory tar with a 5-byte entry
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     5,
	}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err := tw.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	tarReader := tar.NewReader(&buf)
	_, err = tarReader.Next()
	require.NoError(t, err)

	err = extractRegularFile(tarReader, filepath.Join(t.TempDir(), "file.txt"), hdr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error closing file")
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
