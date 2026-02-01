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
			remainder:  "/path/to/layout:v1.0",
			wantPath:   "/path/to/layout",
			wantTag:    "v1.0",
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
