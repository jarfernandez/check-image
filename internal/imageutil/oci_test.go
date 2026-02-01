package imageutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createOCILayout creates a minimal OCI layout directory for testing
func createOCILayout(t *testing.T, layoutPath string) (v1.Image, v1.Hash) {
	t.Helper()

	// Create random image
	img, err := random.Image(512, 1)
	require.NoError(t, err)

	// Create layout path
	p, err := layout.Write(layoutPath, empty.Index)
	require.NoError(t, err)

	// Append the image
	err = p.AppendImage(img)
	require.NoError(t, err)

	// Get the digest
	digest, err := img.Digest()
	require.NoError(t, err)

	return img, digest
}

// createOCILayoutWithTag creates an OCI layout with a tagged image
func createOCILayoutWithTag(t *testing.T, layoutPath string, tag string) (v1.Image, v1.Hash) {
	t.Helper()

	// Create random image
	img, err := random.Image(512, 1)
	require.NoError(t, err)

	// Create layout path
	p, err := layout.Write(layoutPath, empty.Index)
	require.NoError(t, err)

	// Append the image
	err = p.AppendImage(img)
	require.NoError(t, err)

	// Get the digest
	digest, err := img.Digest()
	require.NoError(t, err)

	// Update index with tag annotation
	idx, err := p.ImageIndex()
	require.NoError(t, err)

	manifest, err := idx.IndexManifest()
	require.NoError(t, err)

	// Add tag annotation to the first manifest
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

	return img, digest
}

func TestGetOCILayoutImage_WithDigest(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	_, digest := createOCILayout(t, layoutPath)

	// Load image by digest
	img, err := GetOCILayoutImage(layoutPath, digest.String())
	require.NoError(t, err)
	require.NotNil(t, img)

	// Verify it's the correct image
	loadedDigest, err := img.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)
}

func TestGetOCILayoutImage_WithTag(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	_, digest := createOCILayoutWithTag(t, layoutPath, "v1.0")

	// Load image by tag
	img, err := GetOCILayoutImage(layoutPath, "v1.0")
	require.NoError(t, err)
	require.NotNil(t, img)

	// Verify it's the correct image
	loadedDigest, err := img.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)
}

func TestGetOCILayoutImage_TagWithColon(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	_, digest := createOCILayoutWithTag(t, layoutPath, ":latest")

	// Load image by tag (with leading colon)
	img, err := GetOCILayoutImage(layoutPath, "latest")
	require.NoError(t, err)
	require.NotNil(t, img)

	loadedDigest, err := img.Digest()
	require.NoError(t, err)
	assert.Equal(t, digest, loadedDigest)
}

func TestGetOCILayoutImage_NonexistentPath(t *testing.T) {
	_, err := GetOCILayoutImage("/nonexistent/oci-layout", "sha256:abc123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error reading OCI layout")
}

func TestGetOCILayoutImage_InvalidDigest(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	createOCILayout(t, layoutPath)

	// Try to load with a digest that doesn't exist
	_, err := GetOCILayoutImage(layoutPath, "sha256:0000000000000000000000000000000000000000000000000000000000000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error retrieving image from layout")
}

func TestGetOCILayoutImage_TagNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	createOCILayoutWithTag(t, layoutPath, "v1.0")

	// Try to load with a tag that doesn't exist
	_, err := GetOCILayoutImage(layoutPath, "v2.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in layout index")
}

func TestGetOCILayoutImage_EmptyLayout(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	// Create empty layout
	_, err := layout.Write(layoutPath, empty.Index)
	require.NoError(t, err)

	// Try to load an image from empty layout
	_, err = GetOCILayoutImage(layoutPath, "latest")
	require.Error(t, err)
}

func TestResolveTagInLayout(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	_, digest := createOCILayoutWithTag(t, layoutPath, "v1.23")

	p, err := layout.FromPath(layoutPath)
	require.NoError(t, err)

	// Test exact tag match
	resolvedDigest, err := resolveTagInLayout(p, "v1.23")
	require.NoError(t, err)
	assert.Equal(t, digest.String(), resolvedDigest)
}

func TestResolveTagInLayout_MultipleImages(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	// Create first image with tag
	_, digest1 := createOCILayoutWithTag(t, layoutPath, "v1.0")

	// Append second image with different tag
	p, err := layout.FromPath(layoutPath)
	require.NoError(t, err)

	img2, err := random.Image(512, 1)
	require.NoError(t, err)

	err = p.AppendImage(img2)
	require.NoError(t, err)

	digest2, err := img2.Digest()
	require.NoError(t, err)

	// Update index with second tag
	idx, err := p.ImageIndex()
	require.NoError(t, err)

	manifest, err := idx.IndexManifest()
	require.NoError(t, err)

	for i := range manifest.Manifests {
		if manifest.Manifests[i].Digest == digest2 {
			if manifest.Manifests[i].Annotations == nil {
				manifest.Manifests[i].Annotations = make(map[string]string)
			}
			manifest.Manifests[i].Annotations["org.opencontainers.image.ref.name"] = "v2.0"
			break
		}
	}

	// Write updated index
	indexBytes, err := json.Marshal(manifest)
	require.NoError(t, err)

	indexPath := filepath.Join(layoutPath, "index.json")
	err = os.WriteFile(indexPath, indexBytes, 0644)
	require.NoError(t, err)

	// Reload path
	p, err = layout.FromPath(layoutPath)
	require.NoError(t, err)

	// Resolve first tag
	resolved1, err := resolveTagInLayout(p, "v1.0")
	require.NoError(t, err)
	assert.Equal(t, digest1.String(), resolved1)

	// Resolve second tag
	resolved2, err := resolveTagInLayout(p, "v2.0")
	require.NoError(t, err)
	assert.Equal(t, digest2.String(), resolved2)

	// Verify images are different
	assert.NotEqual(t, resolved1, resolved2)
}

func TestResolveTagInLayout_NoAnnotations(t *testing.T) {
	tmpDir := t.TempDir()
	layoutPath := filepath.Join(tmpDir, "oci-layout")

	// Create layout without tag annotations
	createOCILayout(t, layoutPath)

	p, err := layout.FromPath(layoutPath)
	require.NoError(t, err)

	// Try to resolve a tag that doesn't exist
	_, err = resolveTagInLayout(p, "latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in layout index")
}
