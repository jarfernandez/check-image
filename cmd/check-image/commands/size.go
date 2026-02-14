package commands

import (
	"fmt"
	"math"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
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
		result, err := runSize(args[0])
		if err != nil {
			return fmt.Errorf("check size operation failed: %w", err)
		}

		if err := renderResult(result); err != nil {
			return err
		}

		if result.Passed {
			UpdateResult(ValidationSucceeded)
		} else {
			UpdateResult(ValidationFailed)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sizeCmd)
	sizeCmd.Flags().UintVarP(&maxSize, "max-size", "m", 500, "Maximum size in megabytes (optional)")
	sizeCmd.Flags().UintVarP(&maxLayers, "max-layers", "y", 20, "Maximum number of layers (optional)")
}

func runSize(imageName string) (*output.CheckResult, error) {
	image, err := imageutil.GetImage(imageName)
	if err != nil {
		return nil, err
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("error retrieving the layers: %w", err)
	}

	layerInfos := make([]output.LayerInfo, 0, len(layers))
	var totalSize int64
	for i, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return nil, fmt.Errorf("error getting size of layer %d: %w", i+1, err)
		}
		totalSize += size
		layerInfos = append(layerInfos, output.LayerInfo{Index: i + 1, Bytes: size})
	}

	// Validate that maxSize doesn't overflow when converting to int64
	if maxSize > math.MaxInt64/(1024*1024) {
		return nil, fmt.Errorf("max-size value %d is too large", maxSize)
	}
	maxSizeBytes := int64(maxSize) * 1024 * 1024

	layersOK := uint(len(layers)) <= maxLayers
	sizeOK := totalSize <= maxSizeBytes
	passed := layersOK && sizeOK

	var msg string
	switch {
	case !layersOK && !sizeOK:
		msg = fmt.Sprintf("Image has more than %d layers and size exceeds the recommended limit of %d MB", maxLayers, maxSize)
	case !layersOK:
		msg = fmt.Sprintf("Image has more than %d layers", maxLayers)
	case !sizeOK:
		msg = fmt.Sprintf("Image size exceeds the recommended limit of %d MB", maxSize)
	default:
		msg = fmt.Sprintf("Image size is within the allowed limit of %d MB", maxSize)
	}

	return &output.CheckResult{
		Check:   "size",
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: output.SizeDetails{
			TotalBytes: totalSize,
			TotalMB:    float64(totalSize) / 1024 / 1024,
			MaxSizeMB:  maxSize,
			LayerCount: len(layers),
			MaxLayers:  maxLayers,
			Layers:     layerInfos,
		},
	}, nil
}
