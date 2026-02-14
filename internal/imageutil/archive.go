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

// extractOCIArchive extracts an OCI tarball to a temporary directory
// Returns the path to the temporary directory
func extractOCIArchive(tarballPath string) (string, error) {
	// #nosec G304 -- tarballPath is user-provided but validated by caller
	file, err := os.Open(tarballPath)
	if err != nil {
		return "", fmt.Errorf("error opening tarball: %w", err)
	}
	defer func() {
		_ = file.Close() // Best effort cleanup
	}()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "oci-archive-*")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}

	// Determine if the file is gzipped
	var tarReader *tar.Reader
	if strings.HasSuffix(tarballPath, ".gz") || strings.HasSuffix(tarballPath, ".tgz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			_ = os.RemoveAll(tempDir) // Best effort cleanup
			return "", fmt.Errorf("error creating gzip reader: %w", err)
		}
		defer func() {
			_ = gzReader.Close() // Best effort cleanup
		}()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	// Track total extracted size to prevent decompression bombs
	var totalSize int64

	// Extract all files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = os.RemoveAll(tempDir) // Best effort cleanup
			return "", fmt.Errorf("error reading tar header: %w", err)
		}

		// Check decompression bomb protection
		totalSize += header.Size
		if totalSize > maxDecompressedSize {
			_ = os.RemoveAll(tempDir) // Best effort cleanup
			return "", fmt.Errorf("tarball exceeds maximum decompressed size of %d bytes", maxDecompressedSize)
		}

		// #nosec G305 -- Path traversal protection implemented below
		target := filepath.Join(tempDir, header.Name)

		// Ensure target is within tempDir (security check for path traversal)
		if !strings.HasPrefix(target, filepath.Clean(tempDir)+string(os.PathSeparator)) {
			_ = os.RemoveAll(tempDir) // Best effort cleanup
			return "", fmt.Errorf("illegal file path in tarball: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory with secure permissions
			// #nosec G703 -- target validated against path traversal above (HasPrefix check)
			if err := os.MkdirAll(target, 0750); err != nil {
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("error creating directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Create parent directories if needed
			// #nosec G703 -- target validated against path traversal above (HasPrefix check)
			if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("error creating parent directory for %s: %w", target, err)
			}

			// Use safe file mode (limit to standard file permissions)
			// #nosec G115 -- Mode masked to 0777 ensures safe conversion
			fileMode := os.FileMode(header.Mode) & 0777
			// #nosec G304,G703 -- target path validated above for traversal (HasPrefix check)
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, fileMode)
			if err != nil {
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("error creating file %s: %w", target, err)
			}

			// Copy file contents with size limit
			// #nosec G110 -- Size limit enforced above via totalSize tracking
			written, err := io.Copy(outFile, tarReader)
			if err != nil {
				_ = outFile.Close()       // Best effort cleanup
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("error writing file %s: %w", target, err)
			}

			// Verify written size matches header
			if written != header.Size {
				_ = outFile.Close()       // Best effort cleanup
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("size mismatch for %s: expected %d, got %d", target, header.Size, written)
			}

			if err := outFile.Close(); err != nil {
				_ = os.RemoveAll(tempDir) // Best effort cleanup
				return "", fmt.Errorf("error closing file %s: %w", target, err)
			}

		default:
			// Skip other types (symlinks, etc.)
			continue
		}
	}

	return tempDir, nil
}
