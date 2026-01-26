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
The 'image' argument should be the name of a container image.`,
	Example: "  check-image root-user nginx:latest",
	Args:    cobra.ExactArgs(1),
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
