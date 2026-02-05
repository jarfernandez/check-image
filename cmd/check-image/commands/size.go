package commands

import (
	"fmt"
	"math"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/spf13/cobra"
)

var (
	maxSize   uint
	maxLayers uint
)

var sizeCmd = &cobra.Command{
	Use:   "size image",
	Short: "Validate container image size and number of layers",
	Long: `Validate the size and number of layers of a container image.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`,
	Example: `  check-image size nginx:latest
  check-image size nginx:latest --max-size 300 --max-layers 15
  check-image size oci:/path/to/layout:1.0
  check-image size oci-archive:/path/to/image.tar:latest
  check-image size docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runSize(args[0]); err != nil {
			return fmt.Errorf("check size operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sizeCmd)
	sizeCmd.Flags().UintVarP(&maxSize, "max-size", "s", 500, "Maximum size in megabytes (optional)")
	sizeCmd.Flags().UintVarP(&maxLayers, "max-layers", "y", 20, "Maximum number of layers (optional)")
}

func runSize(imageName string) error {
	fmt.Printf("Checking size and layers of image %s\n", imageName)

	image, err := imageutil.GetImage(imageName)
	if err != nil {
		return err
	}

	layers, err := image.Layers()
	if err != nil {
		return fmt.Errorf("error retrieving the layers: %w", err)
	}

	fmt.Printf("Number of layers: %d\n", len(layers))
	if uint(len(layers)) > maxLayers {
		SetValidationResult(false, "", fmt.Sprintf("Image has more than %d layers", maxLayers))
	}

	var totalSize int64
	for i, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return fmt.Errorf("error getting size of layer %d: %w", i+1, err)
		}
		totalSize += size

		fmt.Printf("  Layer %d: %d bytes\n", i+1, size)
	}
	fmt.Printf("Total size: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/1024/1024)

	// Validate that maxSize doesn't overflow when converting to int64
	if maxSize > math.MaxInt64/(1024*1024) {
		return fmt.Errorf("max-size value %d is too large", maxSize)
	}
	maxSizeBytes := int64(maxSize) * 1024 * 1024
	SetValidationResult(
		totalSize <= maxSizeBytes,
		fmt.Sprintf("Image size is within the allowed limit of %d MB", maxSize),
		fmt.Sprintf("Image size exceeds the recommended limit of %d MB", maxSize),
	)

	return nil
}
