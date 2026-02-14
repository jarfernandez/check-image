package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadStdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid JSON input",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "Valid YAML input",
			input:   "key: value\nanother: value2",
			wantErr: false,
		},
		{
			name:        "Empty stdin",
			input:       "",
			wantErr:     true,
			errContains: "stdin is empty",
		},
		{
			name:    "Large input within limit",
			input:   strings.Repeat("a", 1024*1024), // 1MB
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// Create pipe to mock stdin
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdin = r

			// Write test data in goroutine
			go func() {
				_, _ = w.Write([]byte(tt.input))
				w.Close()
			}()

			// Test ReadStdin
			data, err := ReadStdin()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, []byte(tt.input), data)
		})
	}
}

func TestReadStdin_SizeLimit(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe to mock stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	// Write data exceeding the 10MB limit
	largeData := make([]byte, maxStdinSize+1)
	go func() {
		_, _ = w.Write(largeData)
		w.Close()
	}()

	// Test ReadStdin
	_, err = ReadStdin()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum size")
}

func TestReadFileOrStdin_File(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content from file")
	err := os.WriteFile(filePath, content, 0600)
	require.NoError(t, err)

	// Test reading from file
	data, err := ReadFileOrStdin(filePath)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestReadFileOrStdin_Stdin(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe to mock stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	testData := []byte("test content from stdin")
	go func() {
		_, _ = w.Write(testData)
		w.Close()
	}()

	// Test reading from stdin with "-"
	data, err := ReadFileOrStdin("-")
	require.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestReadFileOrStdin_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.txt")

	_, err := ReadFileOrStdin(filePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access file")
}
