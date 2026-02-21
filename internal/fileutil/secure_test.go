package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSecureFile_Success(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	// Read file
	data, err := ReadSecureFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestReadSecureFile_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")

	_, err := ReadSecureFile(filePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access file")
}

func TestReadSecureFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to read a directory
	_, err := ReadSecureFile(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path is a directory")
}

func TestReadSecureFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(filePath, []byte{}, 0600)
	require.NoError(t, err)

	data, err := ReadSecureFile(filePath)
	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestReadSecureFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.txt")

	// Create a 1MB file
	content := make([]byte, 1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	err := os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	data, err := ReadSecureFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, len(content), len(data))
}

func TestReadSecureFile_WithDotInPath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	filePath := filepath.Join(subDir, "test.txt")
	content := []byte("test")
	err = os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	// Use path with ./
	pathWithDot := filepath.Join(subDir, ".", "test.txt")
	data, err := ReadSecureFile(pathWithDot)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestReadSecureFile_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	// Change to temp directory and use relative path
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	data, err := ReadSecureFile("test.txt")
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

// TestReadSecureFile_NonexistentParentDirectory verifies that ReadSecureFile returns
// an appropriate error when the parent directory of the given path does not exist.
// This exercises the os.OpenRoot error path in ReadSecureFile.
func TestReadSecureFile_NonexistentParentDirectory(t *testing.T) {
	_, err := ReadSecureFile("/nonexistent-parent-dir-xyz-abc/file.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create root for directory")
}

func TestReadSecureFile_DifferentPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	// Write with different permission modes
	permissions := []os.FileMode{0644, 0600, 0400}

	for _, perm := range permissions {
		err := os.WriteFile(filePath, content, perm)
		require.NoError(t, err)

		data, err := ReadSecureFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, data)
	}
}
