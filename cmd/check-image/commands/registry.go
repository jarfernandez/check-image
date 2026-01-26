package commands

import (
	"check-image/internal/imageutil"
	"check-image/internal/registry"
	"fmt"
	"github.com/spf13/cobra"
)

var registryPolicy string

var registryCmd = &cobra.Command{
	Use:   "registry image",
	Short: "Validate that the image registry is trusted",
	Long: `Validate that the image registry is trusted.
The 'image' argument should be the name of a container image.`,
	Example: `  check-image registry ghcr.io/kubernetes-sigs/kind/node --registry-policy registry-policy.json
  check-image registry ghcr.io/kubernetes-sigs/kind/node --registry-policy registry-policy.yaml`,
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
