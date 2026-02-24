package commands

import (
	"fmt"
	"time"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/spf13/cobra"
)

var maxAge uint

var ageCmd = &cobra.Command{
	Use:   "age image",
	Short: "Validate container image age",
	Long: `Validate the age of a container image.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`,
	Example: `  check-image age nginx:latest
  check-image age nginx:latest --max-age 30
  check-image age oci:/path/to/layout:1.0
  check-image age oci-archive:/path/to/image.tar:latest
  check-image age docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runAge(args[0])
		if err != nil {
			return fmt.Errorf("check age operation failed: %w", err)
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
	rootCmd.AddCommand(ageCmd)
	ageCmd.Flags().UintVarP(&maxAge, "max-age", "a", defaultMaxAgeDays, "Maximum age in days (optional)")
}

func runAge(imageName string) (*output.CheckResult, error) {
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

	if config.Created.IsZero() {
		return nil, fmt.Errorf("image creation date is not set")
	}

	age := time.Since(config.Created.Time).Hours() / 24
	passed := age <= float64(maxAge)

	var msg string
	if passed {
		msg = fmt.Sprintf("Image is less than %d days old", maxAge)
	} else {
		msg = fmt.Sprintf("Image is older than %d days", maxAge)
	}

	return &output.CheckResult{
		Check:   "age",
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: output.AgeDetails{
			CreatedAt: config.Created.Format(time.RFC3339),
			AgeDays:   age,
			MaxAge:    maxAge,
		},
	}, nil
}
