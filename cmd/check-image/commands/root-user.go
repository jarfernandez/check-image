package commands

import (
	"check-image/internal/imageutil"
	"fmt"
	"github.com/spf13/cobra"
)

var rootUserCmd = &cobra.Command{
	Use:   "root-user image",
	Short: "Validate that the image is configured to run the container as a non-root user",
	Long: `Validate that the image is configured to run the container as a non-root user.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`,
	Example: `  check-image root-user nginx:latest
  check-image root-user oci:/path/to/layout:1.0
  check-image root-user oci-archive:/path/to/image.tar:latest
  check-image root-user docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runRootUser(args[0]); err != nil {
			return fmt.Errorf("check root-user operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rootUserCmd)
}

func runRootUser(imageName string) error {
	fmt.Printf("Checking if image %s is configured to run as a non-root user\n", imageName)

	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return err
	}

	isNonRoot := config.Config.User != "root" && config.Config.User != ""
	SetValidationResult(
		isNonRoot,
		"Image is configured to run as a non-root user",
		"Image is not configured to run as a non-root user",
	)

	return nil
}
