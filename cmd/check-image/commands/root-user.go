package commands

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
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
		result, err := runRootUser(args[0])
		if err != nil {
			return fmt.Errorf("check root-user operation failed: %w", err)
		}

		if err := renderResult(result); err != nil {
			return err
		}

		if result.Passed {
			if Result != ValidationFailed {
				Result = ValidationSucceeded
			}
		} else {
			Result = ValidationFailed
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rootUserCmd)
}

func runRootUser(imageName string) (*output.CheckResult, error) {
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

	isNonRoot := config.Config.User != "root" && config.Config.User != ""

	var msg string
	if isNonRoot {
		msg = "Image is configured to run as a non-root user"
	} else {
		msg = "Image is not configured to run as a non-root user"
	}

	return &output.CheckResult{
		Check:   "root-user",
		Image:   imageName,
		Passed:  isNonRoot,
		Message: msg,
		Details: output.RootUserDetails{
			User: config.Config.User,
		},
	}, nil
}
