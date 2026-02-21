package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jarfernandez/check-image/internal/fileutil"
	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type allowedPlatformsFile struct {
	AllowedPlatforms []string `json:"allowed-platforms" yaml:"allowed-platforms"`
}

var (
	allowedPlatforms     string
	allowedPlatformsList []string
)

var platformCmd = &cobra.Command{
	Use:   "platform image",
	Short: "Validate that the image platform is in the allowed list",
	Long: `Validate that the image platform is in the allowed list.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`,
	Example: `  check-image platform nginx:latest --allowed-platforms linux/amd64,linux/arm64
  check-image platform nginx:latest --allowed-platforms @config/allowed-platforms.json
  check-image platform nginx:latest --allowed-platforms @config/allowed-platforms.yaml
  check-image platform oci:/path/to/layout:1.0 --allowed-platforms linux/amd64
  check-image platform oci-archive:/path/to/image.tar:latest --allowed-platforms @config/allowed-platforms.yaml --output json
  check-image platform docker-archive:/path/to/image.tar:tag --allowed-platforms linux/amd64,linux/arm64
  cat config/allowed-platforms.json | check-image platform nginx:latest --allowed-platforms @-`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		allowedPlatformsList, err = parseAllowedPlatforms()
		if err != nil {
			return fmt.Errorf("invalid check platform arguments: %w", err)
		}

		log.Debugln("Allowed platforms:", allowedPlatformsList)

		result, err := runPlatform(args[0])
		if err != nil {
			return fmt.Errorf("check platform operation failed: %w", err)
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
	rootCmd.AddCommand(platformCmd)
	platformCmd.Flags().StringVar(&allowedPlatforms, "allowed-platforms", "", "Comma-separated list of allowed platforms or @<file> with JSON or YAML array")
}

// parseAllowedPlatforms parses the --allowed-platforms flag value into a slice of platform strings.
// The flag is required; an error is returned if it is empty.
func parseAllowedPlatforms() ([]string, error) {
	if allowedPlatforms == "" {
		return nil, fmt.Errorf("--allowed-platforms is required")
	}

	if after, ok := strings.CutPrefix(allowedPlatforms, "@"); ok {
		path := after

		data, err := fileutil.ReadFileOrStdin(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read platforms file: %w", err)
		}

		var platformsFromFile allowedPlatformsFile
		if err := fileutil.UnmarshalConfigData(data, &platformsFromFile, path); err != nil {
			return nil, err
		}

		return platformsFromFile.AllowedPlatforms, nil
	}

	parts := strings.Split(allowedPlatforms, ",")
	var platforms []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		platforms = append(platforms, trimmed)
	}

	return platforms, nil
}

func runPlatform(imageName string) (*output.CheckResult, error) {
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

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
			Check:   "platform",
			Image:   imageName,
			Passed:  true,
			Message: fmt.Sprintf("Platform %s is in the allowed list", platform),
			Details: details,
		}, nil
	}

	return &output.CheckResult{
		Check:   "platform",
		Image:   imageName,
		Passed:  false,
		Message: fmt.Sprintf("Platform %s is not in the allowed list", platform),
		Details: details,
	}, nil
}
