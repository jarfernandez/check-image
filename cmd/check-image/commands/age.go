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

` + imageArgFormatsDoc,
	Example: `  check-image age nginx:latest
  check-image age nginx:latest --max-age 30
  check-image age oci:/path/to/layout:1.0
  check-image age oci-archive:/path/to/image.tar:latest
  check-image age docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckCmd(checkAge, func(img string) (*output.CheckResult, error) {
			return runAge(img, maxAge)
		}, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(ageCmd)
	ageCmd.Flags().UintVarP(&maxAge, "max-age", "a", defaultMaxAgeDays, "Maximum age in days (optional)")
}

func runAge(imageName string, maxAgeDays uint) (*output.CheckResult, error) {
	_, config, cleanup, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if config.Created.IsZero() {
		return nil, fmt.Errorf("image creation date is not set")
	}

	age := time.Since(config.Created.Time).Hours() / 24
	passed := age <= float64(maxAgeDays)

	var msg string
	if passed {
		msg = fmt.Sprintf("Image is less than %d days old", maxAgeDays)
	} else {
		msg = fmt.Sprintf("Image is older than %d days", maxAgeDays)
	}

	return &output.CheckResult{
		Check:   checkAge,
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: output.AgeDetails{
			CreatedAt: config.Created.Format(time.RFC3339),
			AgeDays:   age,
			MaxAge:    maxAgeDays,
		},
	}, nil
}
