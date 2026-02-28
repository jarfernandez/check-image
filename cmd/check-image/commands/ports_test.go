package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAllowedPorts_CommaSeparated(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []int
		wantErr   bool
		errString string
	}{
		{
			name:    "Empty string",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Single port",
			input:   "80",
			want:    []int{80},
			wantErr: false,
		},
		{
			name:    "Multiple ports",
			input:   "80,443,8080",
			want:    []int{80, 443, 8080},
			wantErr: false,
		},
		{
			name:    "Ports with spaces",
			input:   "80, 443, 8080",
			want:    []int{80, 443, 8080},
			wantErr: false,
		},
		{
			name:    "Ports with extra spaces",
			input:   " 80 , 443 , 8080 ",
			want:    []int{80, 443, 8080},
			wantErr: false,
		},
		{
			name:    "Ports with empty values",
			input:   "80,,443",
			want:    []int{80, 443},
			wantErr: false,
		},
		{
			name:      "Invalid port - non-numeric",
			input:     "80,abc,443",
			wantErr:   true,
			errString: "invalid port",
		},
		{
			name:      "Invalid port - decimal",
			input:     "80.5",
			wantErr:   true,
			errString: "invalid port",
		},
		{
			name:    "Invalid port - negative",
			input:   "-80",
			want:    []int{-80}, // Note: parseAllowedPorts doesn't validate port range
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global variable
			allowedPorts = tt.input

			got, err := parseAllowedPorts()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAllowedPorts_FromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		fileName    string
		want        []int
		wantErr     bool
		errString   string
	}{
		{
			name: "Valid JSON file",
			fileContent: `{
				"allowed-ports": [80, 443, 8080]
			}`,
			fileName: "ports.json",
			want:     []int{80, 443, 8080},
			wantErr:  false,
		},
		{
			name: "Valid YAML file",
			fileContent: `allowed-ports:
  - 80
  - 443
  - 8080`,
			fileName: "ports.yaml",
			want:     []int{80, 443, 8080},
			wantErr:  false,
		},
		{
			name: "Empty ports array",
			fileContent: `{
				"allowed-ports": []
			}`,
			fileName: "ports.json",
			want:     []int{},
			wantErr:  false,
		},
		{
			name:        "Invalid JSON",
			fileContent: `{invalid json}`,
			fileName:    "ports.json",
			wantErr:     true,
			errString:   "invalid JSON",
		},
		{
			name: "Missing allowed-ports field",
			fileContent: `{
				"other-field": [80, 443]
			}`,
			fileName: "ports.json",
			want:     nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, tt.fileName)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0600)
			require.NoError(t, err)

			// Set the global variable with @ prefix
			allowedPorts = "@" + filePath

			got, err := parseAllowedPorts()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAllowedPorts_FileNotFound(t *testing.T) {
	allowedPorts = "@/nonexistent/file.json"

	_, err := parseAllowedPorts()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestPortsCommand(t *testing.T) {
	// Test that ports command exists and has correct properties
	assert.NotNil(t, portsCmd)
	assert.Equal(t, "ports image", portsCmd.Use)
	assert.Contains(t, portsCmd.Short, "unauthorized ports")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, portsCmd.Args)
	err := portsCmd.Args(portsCmd, []string{})
	assert.Error(t, err)

	err = portsCmd.Args(portsCmd, []string{"image"})
	assert.NoError(t, err)

	err = portsCmd.Args(portsCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestPortsCommandFlags(t *testing.T) {
	// Test that allowed-ports flag exists
	flag := portsCmd.Flags().Lookup("allowed-ports")
	assert.NotNil(t, flag)
	assert.Equal(t, "p", flag.Shorthand)
}

func TestRunPorts_NoExposedPorts(t *testing.T) {
	allowedPortsList = []int{80, 443}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: nil,
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed)
}

func TestRunPorts_ExposedPortsWithNoAllowedList(t *testing.T) {
	allowedPortsList = nil

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"80/tcp":   {},
			"443/tcp":  {},
			"8080/tcp": {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when exposed ports exist but no allowed list is provided")
}

func TestRunPorts_AllPortsAllowed(t *testing.T) {
	allowedPortsList = []int{80, 443, 8080}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"80/tcp":   {},
			"443/tcp":  {},
			"8080/tcp": {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when all exposed ports are in allowed list")
}

func TestRunPorts_SomePortsNotAllowed(t *testing.T) {
	allowedPortsList = []int{80, 443}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"80/tcp":   {},
			"443/tcp":  {},
			"8080/tcp": {},
			"9090/tcp": {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when some exposed ports are not in allowed list")
}

func TestRunPorts_NoPortsAllowed(t *testing.T) {
	allowedPortsList = []int{80, 443}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"8080/tcp": {},
			"9090/tcp": {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.False(t, result.Passed, "Should fail when no exposed ports are in allowed list")
}

func TestRunPorts_DifferentProtocols(t *testing.T) {
	allowedPortsList = []int{80, 443, 53}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"80/tcp":  {},
			"443/tcp": {},
			"53/udp":  {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should handle different protocols (tcp/udp)")
}

func TestRunPorts_InvalidImageReference(t *testing.T) {
	allowedPortsList = []int{80, 443}

	_, err := runPorts("oci:/nonexistent/path:latest")
	require.Error(t, err)
}

func TestRunPorts_SubsetOfAllowedPorts(t *testing.T) {
	allowedPortsList = []int{80, 443, 8080, 9090, 3000}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{
			"80/tcp":  {},
			"443/tcp": {},
		},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when exposed ports are a subset of allowed ports")
}

func TestRunPorts_EmptyExposedPortsMap(t *testing.T) {
	allowedPortsList = []int{80, 443}

	imageRef := createTestImage(t, testImageOptions{
		exposedPorts: map[string]struct{}{},
	})

	result, err := runPorts(imageRef)
	require.NoError(t, err)
	assert.True(t, result.Passed, "Should succeed when exposed ports map is empty")
}

func TestParseAllowedPorts_FromStdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		want        []int
	}{
		{
			name:    "JSON from stdin",
			input:   `{"allowed-ports": [80, 443, 8080]}`,
			wantErr: false,
			want:    []int{80, 443, 8080},
		},
		{
			name: "YAML from stdin",
			input: `allowed-ports:
  - 80
  - 443`,
			wantErr: false,
			want:    []int{80, 443},
		},
		{
			name:        "Invalid JSON from stdin",
			input:       `{invalid}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "Empty stdin",
			input:       "",
			wantErr:     true,
			errContains: "stdin is empty",
		},
		{
			name:    "Empty allowed-ports array from stdin",
			input:   `{"allowed-ports": []}`,
			wantErr: false,
			want:    []int{},
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

			// Write test data
			go func() {
				_, _ = w.Write([]byte(tt.input))
				w.Close()
			}()

			// Set the global variable to use stdin
			allowedPorts = "@-"

			got, err := parseAllowedPorts()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
