package commands

import (
	"check-image/internal/imageutil"
	"check-image/internal/registry"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var registryPolicy string

var registryCmd = &cobra.Command{
	Use:   "registry image",
	Short: "Validate that the image registry is trusted",
	Long: `Validate that the image registry is trusted.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag

Note: Registry validation is only applicable for registry images and will be skipped for other transports.`,
	Example: `  check-image registry nginx:latest --registry-policy registry-policy.json
  check-image registry docker.io/library/nginx:latest --registry-policy registry-policy.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runRegistry(args[0]); err != nil {
			return fmt.Errorf("check registry operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
	registryCmd.Flags().StringVarP(&registryPolicy, "registry-policy", "r", "", "Registry policy file (JSON or YAML)")
	if err := registryCmd.MarkFlagRequired("registry-policy"); err != nil {
		panic(fmt.Sprintf("failed to mark registry-policy flag as required: %v", err))
	}
}

func runRegistry(imageName string) error {
	fmt.Printf("Checking registry of image %s\n", imageName)

	imageRegistry, err := imageutil.GetImageRegistry(imageName)
	if err != nil {
		// Check if this is a non-registry transport
		if strings.Contains(err.Error(), "not applicable") {
			fmt.Println("Registry validation skipped (not applicable for this transport)")
			Result = ValidationSkipped
			return nil
		}
		return fmt.Errorf("unable to get image registry: %w", err)
	}

	fmt.Printf("Image registry: %s\n", imageRegistry)

	policy, err := registry.LoadRegistryPolicy(registryPolicy)
	if err != nil {
		return fmt.Errorf("unable to load registry policy: %w", err)
	}

	SetValidationResult(
		policy.IsRegistryAllowed(imageRegistry),
		fmt.Sprintf("Registry %s is trusted", imageRegistry),
		fmt.Sprintf("Registry %s is not trusted", imageRegistry),
	)

	return nil
}
