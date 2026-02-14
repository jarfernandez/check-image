package commands

import (
	"fmt"
	"strings"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/jarfernandez/check-image/internal/registry"
	"github.com/spf13/cobra"
)

var registryPolicy string

var registryCmd = &cobra.Command{
	Use:   "registry image",
	Short: "Validate that the image registry is trusted",
	Long: `Validate that the image registry is trusted.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag

Note: Registry validation is only applicable for registry images and will be skipped for other transports.`,
	Example: `  check-image registry nginx:latest --registry-policy registry-policy.json
  check-image registry docker.io/library/nginx:latest --registry-policy registry-policy.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runRegistry(args[0])
		if err != nil {
			return fmt.Errorf("check registry operation failed: %w", err)
		}

		if err := renderResult(result); err != nil {
			return err
		}

		// Check if skipped (non-registry transport)
		if d, ok := result.Details.(output.RegistryDetails); ok && d.Skipped {
			Result = ValidationSkipped
			return nil
		}

		if result.Passed {
			if Result != ValidationFailed {
				Result = ValidationSucceeded
			}
		} else {
			Result = ValidationFailed
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
	registryCmd.Flags().StringVarP(&registryPolicy, "registry-policy", "r", "", "Registry policy file (JSON or YAML)")
	if err := registryCmd.MarkFlagRequired("registry-policy"); err != nil {
		panic(fmt.Sprintf("failed to mark registry-policy flag as required: %v", err))
	}
}

func runRegistry(imageName string) (*output.CheckResult, error) {
	imageRegistry, err := imageutil.GetImageRegistry(imageName)
	if err != nil {
		// Check if this is a non-registry transport
		if strings.Contains(err.Error(), "not applicable") {
			return &output.CheckResult{
				Check:   "registry",
				Image:   imageName,
				Passed:  true,
				Message: "Registry validation skipped (not applicable for this transport)",
				Details: output.RegistryDetails{Skipped: true},
			}, nil
		}
		return nil, fmt.Errorf("unable to get image registry: %w", err)
	}

	policy, err := registry.LoadRegistryPolicy(registryPolicy)
	if err != nil {
		return nil, fmt.Errorf("unable to load registry policy: %w", err)
	}

	allowed := policy.IsRegistryAllowed(imageRegistry)

	var msg string
	if allowed {
		msg = fmt.Sprintf("Registry %s is trusted", imageRegistry)
	} else {
		msg = fmt.Sprintf("Registry %s is not trusted", imageRegistry)
	}

	return &output.CheckResult{
		Check:   "registry",
		Image:   imageName,
		Passed:  allowed,
		Message: msg,
		Details: output.RegistryDetails{
			Registry: imageRegistry,
		},
	}, nil
}
