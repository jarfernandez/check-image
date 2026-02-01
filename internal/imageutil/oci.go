package imageutil

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

// GetOCILayoutImage loads an image from an OCI layout directory
func GetOCILayoutImage(layoutPath, reference string) (v1.Image, error) {
	path, err := layout.FromPath(layoutPath)
	if err != nil {
		return nil, fmt.Errorf("error reading OCI layout: %w", err)
	}

	// Try parsing as digest first
	hash, err := v1.NewHash(reference)
	if err != nil {
		// Not a digest - try tag resolution
		digest, err := resolveTagInLayout(path, reference)
		if err != nil {
			return nil, fmt.Errorf("error resolving tag: %w", err)
		}
		hash, err = v1.NewHash(digest)
		if err != nil {
			return nil, fmt.Errorf("error parsing resolved digest: %w", err)
		}
	}

	image, err := path.Image(hash)
	if err != nil {
		return nil, fmt.Errorf("error retrieving image from layout: %w", err)
	}

	return image, nil
}

// resolveTagInLayout finds a digest for a given tag in the layout's index
func resolveTagInLayout(path layout.Path, tag string) (string, error) {
	index, err := path.ImageIndex()
	if err != nil {
		return "", fmt.Errorf("error reading index: %w", err)
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return "", fmt.Errorf("error reading manifest: %w", err)
	}

	// Search for tag in annotations
	for _, desc := range manifest.Manifests {
		if refName, ok := desc.Annotations["org.opencontainers.image.ref.name"]; ok {
			// Match exact tag or :tag format
			if refName == tag || refName == fmt.Sprintf(":%s", tag) {
				return desc.Digest.String(), nil
			}
		}
	}

	return "", fmt.Errorf("tag %q not found in layout index", tag)
}
