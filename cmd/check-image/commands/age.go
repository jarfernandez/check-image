package commands

import (
	"check-image/internal/imageutil"
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

var maxAge uint

var ageCmd = &cobra.Command{
	Use:   "age image",
	Short: "Validate container image age",
	Long: `Validate the age of a container image.
The 'image' argument should be the name of a container image.`,
	Example: `  check-image age ubuntu:20.04
  check-image age ubuntu:20.04 --max-age 30`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAge(args[0]); err != nil {
			return fmt.Errorf("check age operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ageCmd)
	ageCmd.Flags().UintVarP(&maxAge, "max-age", "a", 90, "Maximum age in days (optional)")
}

func runAge(imageName string) error {
	fmt.Printf("Checking age of image %s\n", imageName)

	_, config, err := imageutil.GetRemoteImageAndConfig(imageName)
	if err != nil {
		return err
	}

	if config.Created.IsZero() {
		return fmt.Errorf("image creation date is not set")
	}

	age := time.Since(config.Created.Time).Hours() / 24

	fmt.Printf("Image creation date: %s\n", config.Created.Format(time.RFC3339))
	fmt.Printf("Image age: %.0f days\n", age)

	if age > float64(maxAge) {
		fmt.Printf("Image is older than %d days\n", maxAge)
		Result = ValidationFailed
	} else {
		fmt.Printf("Image is less than %d days old\n", maxAge)
		if Result != ValidationFailed {
			Result = ValidationSucceeded
		}
	}

	return nil
}
