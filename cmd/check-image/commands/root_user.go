package commands

import (
	"context"
	"strings"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/spf13/cobra"
)

var rootUserCmd = &cobra.Command{
	Use:   "root-user image",
	Short: "Validate that the image is configured to run the container as a non-root user",
	Long: `Validate that the image is configured to run the container as a non-root user.

` + imageArgFormatsDoc,
	Example: `  check-image root-user nginx:latest
  check-image root-user oci:/path/to/layout:1.0
  check-image root-user oci-archive:/path/to/image.tar:latest
  check-image root-user docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return runCheckCmd(checkRootUser, runRootUser, ctx, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(rootUserCmd)
}

func runRootUser(ctx context.Context, imageName string) (*output.CheckResult, error) {
	_, config, cleanup, err := imageutil.GetImageAndConfig(ctx, imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	isNonRoot := !isRootUser(config.Config.User)

	var msg string
	if isNonRoot {
		msg = "Image is configured to run as a non-root user"
	} else {
		msg = "Image is not configured to run as a non-root user"
	}

	return &output.CheckResult{
		Check:   checkRootUser,
		Image:   imageName,
		Passed:  isNonRoot,
		Message: msg,
		Details: output.RootUserDetails{
			User: config.Config.User,
		},
	}, nil
}

// isRootUser reports whether the USER value represents the root account.
// It checks for the name "root", empty string (Docker default is root), and UID 0
// in both plain "0" and "0:group" formats, since UID 0 is root regardless of name.
func isRootUser(user string) bool {
	if user == "" || user == "root" {
		return true
	}
	// Check for UID 0 (e.g., "0", "0:0", "0:somegroup")
	userFields := strings.SplitN(user, ":", 2)
	return userFields[0] == "0"
}
