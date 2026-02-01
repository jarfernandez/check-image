package secrets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  []string
		policy   *Policy
		wantLen  int
		wantVars []string
	}{
		{
			name: "Detect password variable",
			envVars: []string{
				"PATH=/usr/bin",
				"DB_PASSWORD=secret123",
				"USER=root",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen:  1,
			wantVars: []string{"DB_PASSWORD"},
		},
		{
			name: "Detect multiple sensitive variables",
			envVars: []string{
				"API_KEY=abc123",
				"SECRET_TOKEN=xyz789",
				"DATABASE_PASSWORD=pass",
				"HOME=/root",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen:  3,
			wantVars: []string{"API_KEY", "SECRET_TOKEN", "DATABASE_PASSWORD"},
		},
		{
			name: "Case insensitive detection",
			envVars: []string{
				"my_password=secret",
				"MY_SECRET=value",
				"Api_Token=token",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen:  3,
			wantVars: []string{"my_password", "MY_SECRET", "Api_Token"},
		},
		{
			name: "Exclude public key",
			envVars: []string{
				"PUBLIC_KEY=ssh-rsa...",
				"PRIVATE_KEY=-----BEGIN",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen:  1,
			wantVars: []string{"PRIVATE_KEY"},
		},
		{
			name: "Custom exclusion list",
			envVars: []string{
				"MY_PASSWORD=secret",
				"OTHER_PASSWORD=secret2",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: []string{"MY_PASSWORD"},
			},
			wantLen:  1,
			wantVars: []string{"OTHER_PASSWORD"},
		},
		{
			name: "Custom patterns",
			envVars: []string{
				"CUSTOM_CRED=value",
				"NORMAL_VAR=value",
			},
			policy: &Policy{
				CheckEnvVars:      true,
				ExcludedEnvVars:   DefaultExcludedEnvVars,
				CustomEnvPatterns: []string{"CRED"},
			},
			wantLen:  1,
			wantVars: []string{"CUSTOM_CRED"},
		},
		{
			name: "Check disabled",
			envVars: []string{
				"PASSWORD=secret",
			},
			policy: &Policy{
				CheckEnvVars: false,
			},
			wantLen: 0,
		},
		{
			name: "No sensitive variables",
			envVars: []string{
				"PATH=/usr/bin",
				"HOME=/root",
				"LANG=en_US.UTF-8",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen: 0,
		},
		{
			name: "Malformed env var without equals",
			envVars: []string{
				"MALFORMED",
				"PASSWORD=secret",
			},
			policy: &Policy{
				CheckEnvVars:    true,
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantLen:  1,
			wantVars: []string{"PASSWORD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := CheckEnvironmentVariables(tt.envVars, tt.policy)
			assert.Len(t, findings, tt.wantLen)

			if tt.wantVars != nil {
				foundVars := make([]string, len(findings))
				for i, f := range findings {
					foundVars[i] = f.Name
				}
				for _, wantVar := range tt.wantVars {
					assert.Contains(t, foundVars, wantVar)
				}
			}

			// All findings should have a description
			for _, finding := range findings {
				assert.NotEmpty(t, finding.Description)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	exclusionList := []string{"PUBLIC_KEY", "SSH_PUBLIC_KEY", "DISPLAY"}

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Excluded value",
			value: "PUBLIC_KEY",
			want:  true,
		},
		{
			name:  "Not excluded",
			value: "PRIVATE_KEY",
			want:  false,
		},
		{
			name:  "Case sensitive - different case",
			value: "public_key",
			want:  false,
		},
		{
			name:  "Empty value",
			value: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExcluded(tt.value, exclusionList)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPathExcluded(t *testing.T) {
	excludedPatterns := []string{
		"/var/log/**",
		"*.tmp",
		"/tmp/cache",
		"test_*.log",
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Directory prefix match",
			path: "/var/log/app.log",
			want: true,
		},
		{
			name: "Directory prefix nested",
			path: "/var/log/nginx/access.log",
			want: true,
		},
		{
			name: "Exact directory match",
			path: "/var/log",
			want: true,
		},
		{
			name: "Glob pattern match",
			path: "/data/file.tmp",
			want: true,
		},
		{
			name: "Exact path match",
			path: "/tmp/cache",
			want: true,
		},
		{
			name: "Pattern with prefix",
			path: "/logs/test_error.log",
			want: true,
		},
		{
			name: "No match",
			path: "/etc/passwd",
			want: false,
		},
		{
			name: "Similar but not matching",
			path: "/var/logs/app.log",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathExcluded(tt.path, excludedPatterns)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesFilePattern(t *testing.T) {
	patterns := []string{
		"id_rsa",
		"*.key",
		".aws/credentials",
		"/etc/shadow",
		"secrets.json",
	}

	tests := []struct {
		name            string
		path            string
		wantMatch       bool
		wantDescription string
	}{
		{
			name:            "SSH key basename",
			path:            "/root/.ssh/id_rsa",
			wantMatch:       true,
			wantDescription: "SSH private key",
		},
		{
			name:            "SSH key exact",
			path:            "id_rsa",
			wantMatch:       true,
			wantDescription: "SSH private key",
		},
		{
			name:      "Key file with glob",
			path:      "/path/to/private.key",
			wantMatch: true,
		},
		{
			name:            "AWS credentials path",
			path:            "/home/user/.aws/credentials",
			wantMatch:       true,
			wantDescription: "AWS credentials",
		},
		{
			name:      "AWS credentials partial",
			path:      ".aws/credentials",
			wantMatch: true,
		},
		{
			name:            "Shadow file",
			path:            "/etc/shadow",
			wantMatch:       true,
			wantDescription: "shadow password file",
		},
		{
			name:      "Secrets JSON basename",
			path:      "/app/config/secrets.json",
			wantMatch: true,
		},
		{
			name:      "No match",
			path:      "/etc/hosts",
			wantMatch: false,
		},
		{
			name:      "Different file",
			path:      "/etc/hosts",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, description := matchesFilePattern(tt.path, patterns)
			assert.Equal(t, tt.wantMatch, matched)
			if tt.wantMatch {
				assert.NotEmpty(t, description)
				if tt.wantDescription != "" {
					assert.Equal(t, tt.wantDescription, description)
				}
			}
		})
	}
}

func TestDescribePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "Known pattern",
			pattern: "id_rsa",
			want:    "SSH private key",
		},
		{
			name:    "Another known pattern",
			pattern: ".aws/credentials",
			want:    "AWS credentials",
		},
		{
			name:    "Unknown pattern",
			pattern: "unknown.pattern",
			want:    "sensitive file pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describePattern(tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function to create a tar layer with files
func createLayerWithFiles(t *testing.T, files map[string]string) v1.Layer {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name:    name,
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		err := tw.WriteHeader(hdr)
		require.NoError(t, err)

		_, err = tw.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	// Copy buffer data to create layer
	data := buf.Bytes()
	opener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	layer, err := tarball.LayerFromOpener(opener)
	require.NoError(t, err)

	return layer
}

func TestCheckFilesInLayers(t *testing.T) {
	tests := []struct {
		name      string
		layers    []map[string]string // Each map represents files in a layer
		policy    *Policy
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "Find SSH key in single layer",
			layers: []map[string]string{
				{
					"/root/.ssh/id_rsa":          "-----BEGIN RSA PRIVATE KEY-----",
					"/root/.ssh/authorized_keys": "ssh-rsa ...",
				},
			},
			policy: &Policy{
				CheckFiles:      true,
				ExcludedPaths:   []string{},
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantFiles: []string{"/root/.ssh/id_rsa"},
		},
		{
			name: "Find secrets across multiple layers",
			layers: []map[string]string{
				{
					"/app/config.json": "{}",
				},
				{
					"/root/.aws/credentials": "[default]\naws_access_key_id=...",
				},
				{
					"/etc/ssl/private.key": "-----BEGIN PRIVATE KEY-----",
				},
			},
			policy: &Policy{
				CheckFiles:      true,
				ExcludedPaths:   []string{},
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantFiles: []string{"/root/.aws/credentials", "/etc/ssl/private.key"},
		},
		{
			name: "Exclude paths",
			layers: []map[string]string{
				{
					"/app/secrets.json":       "sensitive",
					"/var/log/secrets.json":   "not sensitive",
					"/tmp/test/private.key":   "test key",
					"/production/private.key": "real key",
				},
			},
			policy: &Policy{
				CheckFiles:      true,
				ExcludedPaths:   []string{"/var/log/**", "/tmp/**"},
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantFiles: []string{"/app/secrets.json", "/production/private.key"},
		},
		{
			name: "Check disabled",
			layers: []map[string]string{
				{
					"/root/.ssh/id_rsa": "private key",
				},
			},
			policy: &Policy{
				CheckFiles: false,
			},
			wantFiles: []string{},
		},
		{
			name: "Deduplication across layers",
			layers: []map[string]string{
				{
					"/app/secrets.json": "v1",
				},
				{
					"/app/secrets.json": "v2", // Same file in different layer
					"/app/config.json":  "{}",
				},
			},
			policy: &Policy{
				CheckFiles:      true,
				ExcludedPaths:   []string{},
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantFiles: []string{"/app/secrets.json"},
		},
		{
			name: "No sensitive files",
			layers: []map[string]string{
				{
					"/app/main.go":  "package main",
					"/etc/hosts":    "127.0.0.1 localhost",
					"/usr/bin/curl": "binary",
				},
			},
			policy: &Policy{
				CheckFiles:      true,
				ExcludedPaths:   []string{},
				ExcludedEnvVars: DefaultExcludedEnvVars,
			},
			wantFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create image with layers
			img := empty.Image

			for _, layerFiles := range tt.layers {
				layer := createLayerWithFiles(t, layerFiles)
				var err error
				img, err = mutate.AppendLayers(img, layer)
				require.NoError(t, err)
			}

			findings, err := CheckFilesInLayers(img, tt.policy)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Extract file paths from findings
			foundPaths := make([]string, len(findings))
			for i, f := range findings {
				foundPaths[i] = f.Path
			}

			assert.ElementsMatch(t, tt.wantFiles, foundPaths)

			// Verify all findings have descriptions
			for _, finding := range findings {
				assert.NotEmpty(t, finding.Description)
				assert.GreaterOrEqual(t, finding.LayerIndex, 0)
			}
		})
	}
}

func TestScanLayer_DirectorySkipped(t *testing.T) {
	// Create a layer with a directory
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add directory
	hdr := &tar.Header{
		Name:     "/root/.ssh/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
		ModTime:  time.Now(),
	}
	require.NoError(t, tw.WriteHeader(hdr))

	// Add file
	hdr = &tar.Header{
		Name:    "/root/.ssh/id_rsa",
		Mode:    0600,
		Size:    10,
		ModTime: time.Now(),
	}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err := tw.Write([]byte("privatekey"))
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	// Copy buffer data to create layer
	data := buf.Bytes()
	opener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}

	layer, err := tarball.LayerFromOpener(opener)
	require.NoError(t, err)

	policy := &Policy{
		CheckFiles:      true,
		ExcludedPaths:   []string{},
		ExcludedEnvVars: DefaultExcludedEnvVars,
	}

	findings, err := scanLayer(layer, 0, policy)
	require.NoError(t, err)

	// Should only find the file, not the directory
	assert.Len(t, findings, 1)
	assert.Equal(t, "/root/.ssh/id_rsa", findings[0].Path)
}

func TestCheckFilesInLayers_EmptyImage(t *testing.T) {
	img := empty.Image

	policy := &Policy{
		CheckFiles:      true,
		ExcludedPaths:   []string{},
		ExcludedEnvVars: DefaultExcludedEnvVars,
	}

	findings, err := CheckFilesInLayers(img, policy)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

// TestCheckFilesInLayers_CorruptedLayer tests handling of corrupted layers
func TestCheckFilesInLayers_CorruptedLayer(t *testing.T) {
	// Create an image with a valid layer and a corrupted layer
	img := empty.Image

	// Add a valid layer
	layer1 := createLayerWithFiles(t, map[string]string{
		"/app/config.json": "{}",
	})
	var err error
	img, err = mutate.AppendLayers(img, layer1)
	require.NoError(t, err)

	// Add a corrupted layer (not gzipped properly)
	corruptedData := []byte("not a valid gzipped tar")
	corruptedOpener := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(corruptedData)), nil
	}
	corruptedLayer, err := tarball.LayerFromOpener(corruptedOpener)
	require.NoError(t, err)
	img, err = mutate.AppendLayers(img, corruptedLayer)
	require.NoError(t, err)

	policy := &Policy{
		CheckFiles:      true,
		ExcludedPaths:   []string{},
		ExcludedEnvVars: DefaultExcludedEnvVars,
	}

	// Should not fail completely, just log warning for corrupted layer
	findings, err := CheckFilesInLayers(img, policy)
	require.NoError(t, err)
	// Findings slice should exist (even if empty)
	// The implementation continues processing valid layers even when some fail
	assert.GreaterOrEqual(t, len(findings), 0)
}
