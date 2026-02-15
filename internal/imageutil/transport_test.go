package imageutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantTransport Transport
		wantPath      string
		wantTag       string
		wantDigest    string
		wantErr       bool
	}{
		{
			name:          "OCI layout with tag",
			input:         "oci:/path/to/layout:latest",
			wantTransport: TransportOCI,
			wantPath:      "/path/to/layout",
			wantTag:       "latest",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "OCI layout with digest",
			input:         "oci:/path/to/layout@sha256:abc123",
			wantTransport: TransportOCI,
			wantPath:      "/path/to/layout",
			wantTag:       "",
			wantDigest:    "sha256:abc123",
			wantErr:       false,
		},
		{
			name:          "OCI layout relative path with tag",
			input:         "oci:./nginx-layout:v1.23",
			wantTransport: TransportOCI,
			wantPath:      "./nginx-layout",
			wantTag:       "v1.23",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "OCI layout without tag or digest",
			input:         "oci:/path/to/layout",
			wantTransport: TransportOCI,
			wantPath:      "/path/to/layout",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "OCI archive with tag",
			input:         "oci-archive:/path/to/image.tar:latest",
			wantTransport: TransportOCIArchive,
			wantPath:      "/path/to/image.tar",
			wantTag:       "latest",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "OCI archive with digest",
			input:         "oci-archive:./image.tar@sha256:def456",
			wantTransport: TransportOCIArchive,
			wantPath:      "./image.tar",
			wantTag:       "",
			wantDigest:    "sha256:def456",
			wantErr:       false,
		},
		{
			name:          "Docker archive with tag",
			input:         "docker-archive:/path/to/saved.tar:nginx:latest",
			wantTransport: TransportDockerArchive,
			wantPath:      "/path/to/saved.tar",
			wantTag:       "nginx:latest",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "Regular registry reference",
			input:         "nginx:latest",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "nginx:latest",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "Registry reference with registry",
			input:         "docker.io/nginx:latest",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "docker.io/nginx:latest",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "Registry reference with digest",
			input:         "nginx@sha256:xyz789",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "nginx@sha256:xyz789",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "Simple image name without tag (defaults to latest)",
			input:         "nginx",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "nginx:latest",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "Registry reference without tag (defaults to latest)",
			input:         "docker.io/library/nginx",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "docker.io/library/nginx:latest",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
		{
			name:          "GHCR reference without tag (defaults to latest)",
			input:         "ghcr.io/kubernetes-sigs/kind/node",
			wantTransport: TransportDaemonRegistry,
			wantPath:      "ghcr.io/kubernetes-sigs/kind/node:latest",
			wantTag:       "",
			wantDigest:    "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReference(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantTransport, got.Transport)
			assert.Equal(t, tt.wantPath, got.Path)
			assert.Equal(t, tt.wantTag, got.Tag)
			assert.Equal(t, tt.wantDigest, got.Digest)
		})
	}
}

func TestParseTransportReference(t *testing.T) {
	tests := []struct {
		name       string
		transport  Transport
		remainder  string
		wantPath   string
		wantTag    string
		wantDigest string
	}{
		{
			name:       "Path with tag",
			transport:  TransportOCI,
			remainder:  "/path/to/layout:1.0",
			wantPath:   "/path/to/layout",
			wantTag:    "1.0",
			wantDigest: "",
		},
		{
			name:       "Path with digest",
			transport:  TransportOCI,
			remainder:  "/path/to/layout@sha256:abc",
			wantPath:   "/path/to/layout",
			wantTag:    "",
			wantDigest: "sha256:abc",
		},
		{
			name:       "Path without tag or digest",
			transport:  TransportOCI,
			remainder:  "/path/to/layout",
			wantPath:   "/path/to/layout",
			wantTag:    "",
			wantDigest: "",
		},
		{
			name:       "Relative path with tag",
			transport:  TransportOCI,
			remainder:  "./layout:latest",
			wantPath:   "./layout",
			wantTag:    "latest",
			wantDigest: "",
		},
		{
			name:       "Complex path with multiple colons",
			transport:  TransportDockerArchive,
			remainder:  "/path/to/file.tar:nginx:v1.23",
			wantPath:   "/path/to/file.tar",
			wantTag:    "nginx:v1.23",
			wantDigest: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTransportReference(tt.transport, tt.remainder)
			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tt.wantPath, got.Path)
			assert.Equal(t, tt.wantTag, got.Tag)
			assert.Equal(t, tt.wantDigest, got.Digest)
		})
	}
}

func TestFindPathTagSeparator_WindowsPaths(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantIndex int
	}{
		{
			name:      "Windows path with backslash and tag",
			input:     `C:\path\to\image.tar:latest`,
			wantIndex: 20, // Index of ':' before 'latest'
		},
		{
			name:      "Windows path with forward slash and tag",
			input:     "C:/path/to/image.tar:v1.0",
			wantIndex: 20, // Index of ':' before 'v1.0'
		},
		{
			name:      "Unix path with tag",
			input:     "/path/to/image.tar:tag",
			wantIndex: 18, // Index of ':' before 'tag'
		},
		{
			name:      "No tag separator",
			input:     "C:/path/to/image.tar",
			wantIndex: -1, // No separator found
		},
		{
			name:      "Windows drive letter skipped",
			input:     "C:image.tar:tag",
			wantIndex: 11, // Should skip 'C:' and find ':' before 'tag'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := findPathTagSeparator(tt.input)
			assert.Equal(t, tt.wantIndex, index)
		})
	}
}

func TestGetImageRegistry_WithPort(t *testing.T) {
	tests := []struct {
		name         string
		imageName    string
		wantRegistry string
		wantErr      bool
	}{
		{
			name:         "Registry with standard port",
			imageName:    "registry.example.com:5000/myimage:latest",
			wantRegistry: "registry.example.com:5000",
			wantErr:      false,
		},
		{
			name:         "Registry with custom port",
			imageName:    "localhost:8080/app:v1",
			wantRegistry: "localhost:8080",
			wantErr:      false,
		},
		{
			name:         "Default registry with port",
			imageName:    "docker.io:443/library/nginx:latest",
			wantRegistry: "docker.io:443",
			wantErr:      false,
		},
		{
			name:         "IP address with port",
			imageName:    "192.168.1.100:5000/image:tag",
			wantRegistry: "192.168.1.100:5000",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := GetImageRegistry(tt.imageName)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRegistry, registry)
		})
	}
}
