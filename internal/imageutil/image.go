package imageutil

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	cr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// GetImageRegistry extracts the registry from the image name
func GetImageRegistry(imageName string) (string, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}

	return ref.Context().RegistryStr(), nil
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

// GetImage retrieves the image (trying local first, then remote) based on the image name
func GetImage(imageName string) (cr.Image, error) {
	image, err := GetLocalImage(imageName)
	if err == nil {
		return image, nil
	}

	image, err = GetRemoteImage(imageName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving the image: %w", err)
	}

	return image, nil
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
