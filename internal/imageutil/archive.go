package imageutil

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// maxDecompressedSize limits extraction to prevent decompression bombs (5GB)
	maxDecompressedSize = 5 * 1024 * 1024 * 1024
)

// extractOCIArchive extracts an OCI tarball to a temporary directory.
// Returns the path to the temporary directory, which the caller is responsible for removing.
func extractOCIArchive(tarballPath string) (tempDir string, err error) {
	// #nosec G304 -- tarballPath is user-provided but validated by caller
	file, err := os.Open(tarballPath)
	if err != nil {
		return "", fmt.Errorf("error opening tarball: %w", err)
	}
	defer func() { _ = file.Close() }()

	tempDir, err = os.MkdirTemp("", "oci-archive-*")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	// Single cleanup point: on any error, remove the temp dir so the caller
	// never receives a path to a partially-extracted or empty directory.
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
			tempDir = ""
		}
	}()

	tarReader, closer, err := newTarReader(file, tarballPath)
	if err != nil {
		// Use bare return so the deferred cleanup sees the real tempDir value.
		// Writing return "", err here would zero out the named tempDir before the
		// defer runs, causing os.RemoveAll("") instead of the actual directory.
		return
	}
	defer closer()

	err = extractEntries(tarReader, tempDir)
	return tempDir, err
}

// newTarReader wraps r in a tar.Reader, adding a gzip layer when the path
// ends in .gz or .tgz. The returned closer must be called when done.
func newTarReader(r io.Reader, path string) (*tar.Reader, func(), error) {
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, func() {}, fmt.Errorf("error creating gzip reader: %w", err)
		}
		return tar.NewReader(gz), func() { _ = gz.Close() }, nil
	}
	return tar.NewReader(r), func() {}, nil
}

// extractEntries iterates over all tar entries and extracts them into tempDir.
// Cleanup of tempDir on error is the caller's responsibility.
func extractEntries(tarReader *tar.Reader, tempDir string) error {
	var totalSize int64

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		// Check decompression bomb protection
		totalSize += header.Size
		if totalSize > maxDecompressedSize {
			return fmt.Errorf("tarball exceeds maximum decompressed size of %d bytes", maxDecompressedSize)
		}

		// #nosec G305 -- Path traversal protection implemented below
		target := filepath.Join(tempDir, header.Name)

		// Ensure target is within tempDir (security check for path traversal)
		if !strings.HasPrefix(target, filepath.Clean(tempDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in tarball: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory with secure permissions
			// #nosec G703 -- target validated against path traversal above (HasPrefix check)
			if err := os.MkdirAll(target, 0750); err != nil {
				return fmt.Errorf("error creating directory %s: %w", target, err)
			}

		case tar.TypeReg:
			if err := extractRegularFile(tarReader, target, header); err != nil {
				return err
			}

		default:
			// Skip other types (symlinks, etc.)
			continue
		}
	}

	return nil
}

// extractRegularFile extracts a single regular file from a tar archive
func extractRegularFile(tarReader *tar.Reader, target string, header *tar.Header) error {
	// Create parent directories if needed
	// #nosec G703 -- target validated against path traversal by caller (HasPrefix check)
	if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		return fmt.Errorf("error creating parent directory for %s: %w", target, err)
	}

	// Use safe file mode (limit to standard file permissions)
	// #nosec G115 -- Mode masked to 0777 ensures safe conversion
	fileMode := os.FileMode(header.Mode) & 0777
	// #nosec G304,G703 -- target path validated by caller for traversal (HasPrefix check)
	outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", target, err)
	}

	// Copy file contents with size limit
	// #nosec G110 -- Size limit enforced by caller via totalSize tracking
	written, err := io.Copy(outFile, tarReader)
	if err != nil {
		_ = outFile.Close() // Best effort cleanup
		return fmt.Errorf("error writing file %s: %w", target, err)
	}

	// Verify written size matches header
	if written != header.Size {
		_ = outFile.Close() // Best effort cleanup
		return fmt.Errorf("size mismatch for %s: expected %d, got %d", target, header.Size, written)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("error closing file %s: %w", target, err)
	}

	return nil
}
