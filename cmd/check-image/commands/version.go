package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/jarfernandez/check-image/internal/output"
	ver "github.com/jarfernandez/check-image/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show the check-image version",
	Long:    `Show the check-image version.`,
	Example: `  check-image version`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runVersion(); err != nil {
			return fmt.Errorf("version operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion() error {
	version := strings.TrimSpace(ver.Get())
	if version == "" {
		version = "dev"
	}

	if OutputFmt == output.FormatJSON {
		return output.RenderJSON(os.Stdout, output.VersionResult{Version: version})
	}

	fmt.Printf("%s\n", version)
	return nil
}
