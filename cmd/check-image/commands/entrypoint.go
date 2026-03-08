package commands

import (
	"slices"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/spf13/cobra"
)

const shellFlagArg = "-c"

var shellInterpreters = []string{"/bin/sh", "/bin/bash"}

var allowShellForm bool

var entrypointCmd = &cobra.Command{
	Use:   "entrypoint image",
	Short: "Validate that the image has an entrypoint defined and uses exec form",
	Long: `Validate that the image has a startup command defined (ENTRYPOINT or CMD) and uses exec form.

By default the check fails if shell form is detected. Use --allow-shell-form to allow it.

` + imageArgFormatsDoc,
	Example: `  check-image entrypoint nginx:latest
  check-image entrypoint nginx:latest -o json
  check-image entrypoint nginx:latest --allow-shell-form
  check-image entrypoint oci:/path/to/layout:1.0
  check-image entrypoint oci-archive:/path/to/image.tar:latest
  check-image entrypoint docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckCmd(checkEntrypoint, func(img string) (*output.CheckResult, error) {
			return runEntrypoint(img, allowShellForm)
		}, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(entrypointCmd)
	entrypointCmd.Flags().BoolVar(&allowShellForm, "allow-shell-form", false,
		"Allow shell form for entrypoint or cmd without failing (optional)")
}

func runEntrypoint(imageName string, shellFormAllowed bool) (*output.CheckResult, error) {
	_, config, cleanup, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	entrypoint := config.Config.Entrypoint
	startCmd := config.Config.Cmd

	hasEntrypoint := len(entrypoint) > 0 || len(startCmd) > 0
	if !hasEntrypoint {
		return &output.CheckResult{
			Check:   checkEntrypoint,
			Image:   imageName,
			Passed:  false,
			Message: "Image has no entrypoint or cmd defined",
			Details: output.EntrypointDetails{
				HasEntrypoint: false,
			},
		}, nil
	}

	shellForm := isShellFormCommand(entrypoint) || isShellFormCommand(startCmd)
	execForm := !shellForm

	var msg string
	var passed bool
	switch {
	case execForm:
		passed, msg = true, "Image has a valid exec-form entrypoint" // #nosec G101 -- false positive: not a credential
	case shellFormAllowed:
		passed, msg = true, "Image uses shell form but it is allowed"
	default:
		passed, msg = false, "Image uses shell form for entrypoint or cmd"
	}

	details := output.EntrypointDetails{
		HasEntrypoint: true,
		ExecForm:      execForm,
		Entrypoint:    entrypoint,
		Cmd:           startCmd,
	}
	if !execForm && shellFormAllowed {
		details.ShellFormAllowed = true
	}

	return &output.CheckResult{
		Check:   checkEntrypoint,
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: details,
	}, nil
}

// isShellFormCommand returns true if the command slice represents shell form,
// i.e., the first element is /bin/sh or /bin/bash and the second is -c.
// This is how Docker stores ENTRYPOINT/CMD when using shell form in a Dockerfile.
func isShellFormCommand(cmd []string) bool {
	return len(cmd) >= 2 &&
		slices.Contains(shellInterpreters, cmd[0]) &&
		cmd[1] == shellFlagArg
}
