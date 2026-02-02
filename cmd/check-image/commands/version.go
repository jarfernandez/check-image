package commands

import (
	ver "check-image/internal/version"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
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
	fmt.Printf("%s\n", version)
	return nil
}
