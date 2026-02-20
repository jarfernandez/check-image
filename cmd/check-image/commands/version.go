package commands

import (
	"fmt"
	"os"

	"github.com/jarfernandez/check-image/internal/output"
	ver "github.com/jarfernandez/check-image/internal/version"
	"github.com/spf13/cobra"
)

var shortVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the check-image version",
	Long:  `Show the check-image version with full build information.`,
	Example: `  check-image version
  check-image version --short
  check-image version -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runVersion(); err != nil {
			return fmt.Errorf("version operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print only the version number")
}

func runVersion() error {
	info := ver.GetBuildInfo()

	if shortVersion {
		if OutputFmt == output.FormatJSON {
			return output.RenderJSON(os.Stdout, output.VersionResult{Version: info.Version})
		}
		fmt.Printf("%s\n", info.Version)
		return nil
	}

	if OutputFmt == output.FormatJSON {
		return output.RenderJSON(os.Stdout, output.BuildInfoResult{
			Version:   info.Version,
			Commit:    info.Commit,
			BuiltAt:   info.BuildDate,
			GoVersion: info.GoVersion,
			Platform:  info.Platform,
		})
	}

	fmt.Printf("check-image version %s\n", info.Version)
	fmt.Printf("commit:     %s\n", info.Commit)
	fmt.Printf("built at:   %s\n", info.BuildDate)
	fmt.Printf("go version: %s\n", info.GoVersion)
	fmt.Printf("platform:   %s\n", info.Platform)
	return nil
}
