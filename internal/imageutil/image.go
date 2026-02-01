package imageutil

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	cr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// GetImageRegistry extracts the registry from the image name
func GetImageRegistry(imageName string) (string, error) {
	ref, err := ParseReference(imageName)
	if err != nil {
		return "", err
	}

	// Non-registry transports don't have a registry
	if ref.Transport != TransportDaemonRegistry {
		return "", fmt.Errorf("registry not applicable for %s transport", ref.Transport)
	}

	// Parse as regular image reference
	parsedRef, err := name.ParseReference(ref.Path)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}

	return parsedRef.Context().RegistryStr(), nil
}

// GetLocalImage retrieves the local image from a reference name
func GetLocalImage(imageName string) (cr.Image, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("error parsing the reference: %w", err)
	}

	image, err := daemon.Image(ref)
	if err != nil {
		return nil, fmt.Errorf("error retrieving the local image: %w", err)
	}

	return image, nil
}

// GetRemoteImage retrieves the remote image from a reference name
func GetRemoteImage(imageName string) (cr.Image, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("error parsing the reference: %w", err)
	}

	kc := authn.DefaultKeychain
	image, err := remote.Image(ref, remote.WithAuthFromKeychain(kc))
	if err != nil {
		return nil, fmt.Errorf("error retrieving the remote image: %w", err)
	}

	return image, nil
}

// GetDockerArchiveImage retrieves an image from a Docker tarball (docker save format)
func GetDockerArchiveImage(tarballPath string, tag string) (cr.Image, error) {
	// Parse the tag if provided
	var nameTag *name.Tag
	if tag != "" {
		parsedTag, err := name.NewTag(tag)
		if err != nil {
			return nil, fmt.Errorf("error parsing tag %s: %w", tag, err)
		}
		nameTag = &parsedTag
	}

	// Load image from tarball
	image, err := tarball.ImageFromPath(tarballPath, nameTag)
	if err != nil {
		return nil, fmt.Errorf("error loading docker archive from %s: %w", tarballPath, err)
	}

	return image, nil
}

// GetOCIArchiveImage retrieves an image from an OCI tarball
func GetOCIArchiveImage(tarballPath string, reference string) (cr.Image, error) {
	// OCI archives need to be extracted to a temporary directory first
	// then loaded using the OCI layout functions
	tempDir, err := extractOCIArchive(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("error extracting OCI archive: %w", err)
	}
	// Note: Caller should clean up tempDir when done, but for now we'll leave it
	// TODO: Implement proper cleanup mechanism

	// Use the OCI layout loader
	return GetOCILayoutImage(tempDir, reference)
}

// GetImage retrieves the image using transport-aware reference parsing
func GetImage(imageName string) (cr.Image, error) {
	ref, err := ParseReference(imageName)
	if err != nil {
		return nil, err
	}

	switch ref.Transport {
	case TransportOCI:
		// OCI layout directory - no fallback
		reference := ref.Digest
		if reference == "" {
			reference = ref.Tag
		}
		if reference == "" {
			return nil, fmt.Errorf("oci transport requires tag or digest")
		}
		return GetOCILayoutImage(ref.Path, reference)

	case TransportOCIArchive:
		// OCI tarball - extract and load
		reference := ref.Digest
		if reference == "" {
			reference = ref.Tag
		}
		if reference == "" {
			return nil, fmt.Errorf("oci-archive transport requires tag or digest")
		}
		return GetOCIArchiveImage(ref.Path, reference)

	case TransportDockerArchive:
		// Docker tarball - load directly
		return GetDockerArchiveImage(ref.Path, ref.Tag)

	case TransportDaemonRegistry:
		// Default mode: try local daemon, fall back to remote registry
		image, err := GetLocalImage(ref.Path)
		if err == nil {
			return image, nil
		}
		return GetRemoteImage(ref.Path)

	default:
		return nil, fmt.Errorf("unsupported transport: %s", ref.Transport)
	}
}

// GetImageConfig retrieves the configuration file of a given container image
func GetImageConfig(image cr.Image) (*cr.ConfigFile, error) {
	config, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error retrieving the image configuration: %w", err)
	}

	return config, nil
}

// GetImageAndConfig retrieves both the image and its configuration file given an image name
func GetImageAndConfig(imageName string) (cr.Image, *cr.ConfigFile, error) {
	image, err := GetImage(imageName)
	if err != nil {
		return nil, nil, err
	}

	config, err := GetImageConfig(image)
	if err != nil {
		return nil, nil, err
	}

	return image, config, nil
}
