package imageutil

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	cr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// GetRemoteImage retrieves the remote image from a reference name
func GetRemoteImage(imageName string) (cr.Image, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("error parsing the reference: %w", err)
	}

	kc := authn.DefaultKeychain
	image, err := remote.Image(ref, remote.WithAuthFromKeychain(kc))
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

// GetRemoteImageAndConfig retrieves both the remote image and its configuration file given an image name
func GetRemoteImageAndConfig(imageName string) (cr.Image, *cr.ConfigFile, error) {
	image, err := GetRemoteImage(imageName)
	if err != nil {
		return nil, nil, err
	}

	config, err := GetImageConfig(image)
	if err != nil {
		return nil, nil, err
	}

	return image, config, nil
}
