package commands

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/spf13/cobra"
)

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck image",
	Short: "Validate that the image has a healthcheck defined",
	Long: `Validate that the image has a healthcheck defined.

` + imageArgFormatsDoc,
	Example: `  check-image healthcheck nginx:latest
  check-image healthcheck nginx:latest -o json
  check-image healthcheck oci:/path/to/layout:1.0
  check-image healthcheck oci-archive:/path/to/image.tar:latest
  check-image healthcheck docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runHealthcheck(args[0])
		if err != nil {
			return fmt.Errorf("check healthcheck operation failed: %w", err)
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
	rootCmd.AddCommand(healthcheckCmd)
}

func runHealthcheck(imageName string) (*output.CheckResult, error) {
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

	hasHealthcheck := config.Config.Healthcheck != nil &&
		len(config.Config.Healthcheck.Test) > 0 &&
		config.Config.Healthcheck.Test[0] != "NONE"

	var msg string
	if hasHealthcheck {
		msg = "Image has a healthcheck defined"
	} else {
		msg = "Image does not have a healthcheck defined"
	}

	return &output.CheckResult{
		Check:   "healthcheck",
		Image:   imageName,
		Passed:  hasHealthcheck,
		Message: msg,
		Details: output.HealthcheckDetails{
			HasHealthcheck: hasHealthcheck,
		},
	}, nil
}
