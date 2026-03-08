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

` + imageArgFormatsDoc,
	Example: `  check-image size nginx:latest
  check-image size nginx:latest --max-size 300 --max-layers 15
  check-image size oci:/path/to/layout:1.0
  check-image size oci-archive:/path/to/image.tar:latest
  check-image size docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckCmd(checkSize, func(img string) (*output.CheckResult, error) {
			return runSize(img, maxSize, maxLayers)
		}, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(sizeCmd)
	sizeCmd.Flags().UintVarP(&maxSize, "max-size", "m", defaultMaxSizeMB, "Maximum size in megabytes (optional)")
	sizeCmd.Flags().UintVarP(&maxLayers, "max-layers", "y", defaultMaxLayerCount, "Maximum number of layers (optional)")
}

func runSize(imageName string, maxSizeMB uint, maxLayerCount uint) (*output.CheckResult, error) {
	image, cleanup, err := imageutil.GetImage(imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

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

	// Validate that maxSizeMB doesn't overflow when converting to int64
	if maxSizeMB > math.MaxInt64/(1024*1024) {
		return nil, fmt.Errorf("max-size value %d is too large", maxSizeMB)
	}
	maxSizeBytes := int64(maxSizeMB) * 1024 * 1024

	layersOK := uint(len(layers)) <= maxLayerCount
	sizeOK := totalSize <= maxSizeBytes
	passed := layersOK && sizeOK

	var msg string
	switch {
	case !layersOK && !sizeOK:
		msg = fmt.Sprintf("Image has more than %d layers and size exceeds the recommended limit of %d MB", maxLayerCount, maxSizeMB)
	case !layersOK:
		msg = fmt.Sprintf("Image has more than %d layers", maxLayerCount)
	case !sizeOK:
		msg = fmt.Sprintf("Image size exceeds the recommended limit of %d MB", maxSizeMB)
	default:
		msg = fmt.Sprintf("Image size is within the allowed limit of %d MB", maxSizeMB)
	}

	return &output.CheckResult{
		Check:   checkSize,
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: output.SizeDetails{
			TotalBytes: totalSize,
			TotalMB:    float64(totalSize) / 1024 / 1024,
			MaxSizeMB:  maxSizeMB,
			LayerCount: len(layers),
			MaxLayers:  maxLayerCount,
			Layers:     layerInfos,
		},
	}, nil
}
