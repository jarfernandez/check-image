package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/stretchr/testify/require"
)

// testImageOptions holds configuration for creating test images
type testImageOptions struct {
	user         string
	created      time.Time
	exposedPorts map[string]struct{}
	env          []string
	layerCount   int
	layerSizes   []int64             // Optional: specific sizes for each layer in bytes. If nil, default sizes are used.
	layerFiles   []map[string]string // Optional: files to add to each layer. Map of path -> content.
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

	// Add layers if requested
	if opts.layerCount > 0 || len(opts.layerFiles) > 0 {
		var layers []v1.Layer

		// Determine actual number of layers to create
		numLayers := opts.layerCount
		if len(opts.layerFiles) > numLayers {
			numLayers = len(opts.layerFiles)
		}

		for i := 0; i < numLayers; i++ {
			var layer v1.Layer

			// If files are specified for this layer, create layer with files
			if opts.layerFiles != nil && i < len(opts.layerFiles) && len(opts.layerFiles[i]) > 0 {
				layer = createLayerWithFiles(t, opts.layerFiles[i])
			} else {
				// Otherwise create layer with specific size
				var size int64
				if opts.layerSizes != nil && i < len(opts.layerSizes) {
					size = opts.layerSizes[i]
				} else {
					// Default: 1KB per layer
					size = 1024
				}
				layer = createTestLayer(t, size)
			}
			layers = append(layers, layer)
		}
		img, err = mutate.AppendLayers(img, layers...)
		require.NoError(t, err)
	}

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

// createTestLayer creates a test layer with approximately the specified compressed size
// Uses random data to minimize compression and achieve predictable sizes
func createTestLayer(t *testing.T, sizeBytes int64) v1.Layer {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Create a file with random content to reach approximately the desired COMPRESSED size
	// Random data doesn't compress well, so compressed size ≈ uncompressed size
	// Account for tar header overhead (~512 bytes) and minimal gzip overhead (~20 bytes)
	contentSize := sizeBytes - 532
	if contentSize < 0 {
		contentSize = 0
	}

	// Create truly random content that won't compress
	content := make([]byte, contentSize)
	_, err := rand.Read(content)
	require.NoError(t, err)

	// Write tar entry
	hdr := &tar.Header{
		Name:    "/layer-data",
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}
	err = tw.WriteHeader(hdr)
	require.NoError(t, err)

	_, err = tw.Write(content)
	require.NoError(t, err)

	require.NoError(t, tw.Close())

	// Create layer with gzip compression
	data := buf.Bytes()
	opener := func() (io.ReadCloser, error) {
		var gzipped bytes.Buffer
		gw := gzip.NewWriter(&gzipped)
		_, err := gw.Write(data)
		require.NoError(t, err)
		require.NoError(t, gw.Close())
		return io.NopCloser(bytes.NewReader(gzipped.Bytes())), nil
	}

	layer, err := tarball.LayerFromOpener(opener)
	require.NoError(t, err)

	return layer
}

// createLayerWithFiles creates a test layer with specific files and contents
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
