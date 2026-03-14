package commands

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type allowedPlatformsFile struct {
	AllowedPlatforms []string `json:"allowed-platforms" yaml:"allowed-platforms"`
}

var allowedPlatforms string

var platformCmd = &cobra.Command{
	Use:   "platform image",
	Short: "Validate that the image platform is in the allowed list",
	Long: `Validate that the image platform is in the allowed list.

` + imageArgFormatsDoc,
	Example: `  check-image platform nginx:latest --allowed-platforms linux/amd64,linux/arm64
  check-image platform nginx:latest --allowed-platforms @config/allowed-platforms.json
  check-image platform nginx:latest --allowed-platforms @config/allowed-platforms.yaml
  check-image platform oci:/path/to/layout:1.0 --allowed-platforms linux/amd64
  check-image platform oci-archive:/path/to/image.tar:latest --allowed-platforms @config/allowed-platforms.yaml --output json
  check-image platform docker-archive:/path/to/image.tar:tag --allowed-platforms linux/amd64,linux/arm64
  cat config/allowed-platforms.json | check-image platform nginx:latest --allowed-platforms @-`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		platforms, err := parseAllowedPlatforms()
		if err != nil {
			return fmt.Errorf("invalid check platform arguments: %w", err)
		}

		log.Debugln("Allowed platforms:", platforms)

		ctx := cmd.Context()
		return runCheckCmd(checkPlatform, func(ctx context.Context, img string) (*output.CheckResult, error) {
			return runPlatform(ctx, img, platforms)
		}, ctx, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(platformCmd)
	platformCmd.Flags().StringVar(&allowedPlatforms, "allowed-platforms", "", "Comma-separated list of allowed platforms or @<file> with JSON or YAML array")
}

// parseAllowedPlatforms parses the --allowed-platforms flag value into a slice of platform strings.
// The flag is required; an error is returned if it is empty.
func parseAllowedPlatforms() ([]string, error) {
	return parseAllowedPlatformsFrom(allowedPlatforms)
}

func parseAllowedPlatformsFrom(platformsStr string) ([]string, error) {
	if platformsStr == "" {
		return nil, fmt.Errorf("--allowed-platforms is required")
	}

	if after, ok := strings.CutPrefix(platformsStr, "@"); ok {
		var platformsFromFile allowedPlatformsFile
		if err := parseAllowedListFromFile(after, &platformsFromFile); err != nil {
			return nil, err
		}
		for _, p := range platformsFromFile.AllowedPlatforms {
			if err := validatePlatformFormat(p); err != nil {
				return nil, err
			}
		}
		return platformsFromFile.AllowedPlatforms, nil
	}

	parts := strings.Split(platformsStr, ",")
	var platforms []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if err := validatePlatformFormat(trimmed); err != nil {
			return nil, err
		}
		platforms = append(platforms, trimmed)
	}

	return platforms, nil
}

// validatePlatformFormat checks that platform follows the OS/Architecture[/Variant] format.
func validatePlatformFormat(platform string) error {
	segments := strings.Split(platform, "/")
	if len(segments) < 2 || len(segments) > 3 {
		return fmt.Errorf("invalid platform format %q: expected OS/Architecture or OS/Architecture/Variant", platform)
	}
	return nil
}

func runPlatform(ctx context.Context, imageName string, allowedPlatformsList []string) (*output.CheckResult, error) {
	_, config, cleanup, err := imageutil.GetImageAndConfig(ctx, imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	// Build platform string: OS/Architecture[/Variant]
	platform := config.OS + "/" + config.Architecture
	if config.Variant != "" {
		platform += "/" + config.Variant
	}

	log.Debugf("Image platform: %s", platform)

	details := output.PlatformDetails{
		Platform:         platform,
		AllowedPlatforms: allowedPlatformsList,
	}

	if slices.Contains(allowedPlatformsList, platform) {
		return &output.CheckResult{
			Check:   checkPlatform,
			Image:   imageName,
			Passed:  true,
			Message: fmt.Sprintf("Platform %s is in the allowed list", platform),
			Details: details,
		}, nil
	}

	return &output.CheckResult{
		Check:   checkPlatform,
		Image:   imageName,
		Passed:  false,
		Message: fmt.Sprintf("Platform %s is not in the allowed list", platform),
		Details: details,
	}, nil
}
