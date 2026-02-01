package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/stretchr/testify/require"
)

// testImageOptions holds configuration for creating test images
type testImageOptions struct {
	user         string
	created      time.Time
	exposedPorts map[string]struct{}
	env          []string
	layerCount   int
}

// createTestOCILayout creates an OCI layout in a temporary directory with a test image
// Returns the path to the layout and the tag name
func createTestOCILayout(t *testing.T, tag string, opts testImageOptions) string {
	t.Helper()

	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	// Create base image
	img := empty.Image

	// Create config
	cfg, err := img.ConfigFile()
	require.NoError(t, err)

	// Set config options
	cfg.Config.User = opts.user
	cfg.Created = v1.Time{Time: opts.created}
	cfg.Config.ExposedPorts = opts.exposedPorts
	cfg.Config.Env = opts.env

	// Apply config
	img, err = mutate.ConfigFile(img, cfg)
	require.NoError(t, err)

	// Add layers if requested (skip for now, not critical for command tests)
	_ = opts.layerCount

	// Create layout
	p, err := layout.Write(layoutPath, empty.Index)
	require.NoError(t, err)

	// Append image
	err = p.AppendImage(img)
	require.NoError(t, err)

	// Get digest
	digest, err := img.Digest()
	require.NoError(t, err)

	// Add tag annotation
	idx, err := p.ImageIndex()
	require.NoError(t, err)

	manifest, err := idx.IndexManifest()
	require.NoError(t, err)

	for i := range manifest.Manifests {
		if manifest.Manifests[i].Digest == digest {
			if manifest.Manifests[i].Annotations == nil {
				manifest.Manifests[i].Annotations = make(map[string]string)
			}
			manifest.Manifests[i].Annotations["org.opencontainers.image.ref.name"] = tag
			break
		}
	}

	// Write updated index
	indexBytes, err := json.Marshal(manifest)
	require.NoError(t, err)

	indexPath := filepath.Join(layoutPath, "index.json")
	err = os.WriteFile(indexPath, indexBytes, 0644)
	require.NoError(t, err)

	return layoutPath
}

// createTestImage creates a test image reference with OCI layout
func createTestImage(t *testing.T, opts testImageOptions) string {
	t.Helper()

	layoutPath := createTestOCILayout(t, "latest", opts)
	return "oci:" + layoutPath + ":latest"
}
