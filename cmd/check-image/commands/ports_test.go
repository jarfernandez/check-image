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
	assert.Contains(t, err.Error(), "failed to read ports file")
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
