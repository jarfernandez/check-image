package commands

import (
	"check-image/internal/imageutil"
	"check-image/internal/secrets"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	secretsPolicy string
	skipEnvVars   bool
	skipFiles     bool
)

var secretsCmd = &cobra.Command{
	Use:   "secrets image",
	Short: "Validate that the image does not contain sensitive data",
	Long: `Validate that the image does not contain sensitive data (passwords, tokens, keys).
Scans both environment variables and files across all image layers.
The 'image' argument should be the name of a container image.`,
	Example: `  check-image secrets nginx:latest
  check-image secrets ubuntu:latest --secrets-policy secrets-policy.json
  check-image secrets ubuntu:latest --secrets-policy secrets-policy.yaml
  check-image secrets alpine:latest --skip-env-vars
  check-image secrets redis:7.4 --skip-files`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runSecrets(args[0]); err != nil {
			return fmt.Errorf("check secrets operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().StringVarP(&secretsPolicy, "secrets-policy", "p", "", "Path to secrets policy file (JSON or YAML)")
	secretsCmd.Flags().BoolVar(&skipEnvVars, "skip-env-vars", false, "Skip environment variable checks")
	secretsCmd.Flags().BoolVar(&skipFiles, "skip-files", false, "Skip file system checks")
}

func runSecrets(imageName string) error {
	fmt.Printf("Checking secrets in image %s\n", imageName)

	// Load policy from file or use defaults
	policy, err := secrets.LoadSecretsPolicy(secretsPolicy)
	if err != nil {
		return fmt.Errorf("unable to load secrets policy: %w", err)
	}

	// Override policy based on command-line flags
	if skipEnvVars {
		policy.CheckEnvVars = false
	}
	if skipFiles {
		policy.CheckFiles = false
	}

	// Get image and config
	image, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return err
	}

	var envCount, fileCount int

	// Check environment variables
	if policy.CheckEnvVars {
		log.Debug("Checking environment variables for secrets")
		envFindings := secrets.CheckEnvironmentVariables(config.Config.Env, policy)
		envCount = len(envFindings)

		if len(envFindings) > 0 {
			fmt.Println("\nEnvironment Variables:")
			for _, finding := range envFindings {
				fmt.Printf("  - %s (%s)\n", finding.Name, finding.Description)
			}
		}
	}

	// Check files in layers
	if policy.CheckFiles {
		log.Debug("Checking files in layers for secrets")
		fileFindings, err := secrets.CheckFilesInLayers(image, policy)
		if err != nil {
			return fmt.Errorf("error scanning files: %w", err)
		}
		fileCount = len(fileFindings)

		if len(fileFindings) > 0 {
			fmt.Println("\nFiles with Sensitive Patterns:")

			// Group findings by layer for better readability
			layerMap := make(map[int][]secrets.FileFinding)
			for _, finding := range fileFindings {
				layerMap[finding.LayerIndex] = append(layerMap[finding.LayerIndex], finding)
			}

			// Display findings grouped by layer
			for layerIdx := 0; layerIdx < len(layerMap)+10; layerIdx++ { // +10 to ensure we get all layers
				if findings, ok := layerMap[layerIdx]; ok {
					fmt.Printf("  Layer %d:\n", layerIdx+1)
					for _, finding := range findings {
						fmt.Printf("    - %s (%s)\n", finding.Path, finding.Description)
					}
				}
			}
		}
	}

	// Display summary
	totalFindings := envCount + fileCount
	fmt.Printf("\nTotal findings: %d", totalFindings)
	if policy.CheckEnvVars && policy.CheckFiles {
		fmt.Printf(" (%d environment variables, %d files)\n", envCount, fileCount)
	} else {
		fmt.Println()
	}

	// Set validation result
	if totalFindings > 0 {
	    fmt.Println("Secrets detected")
		Result = ValidationFailed
	} else {
		fmt.Println("No secrets detected")
		if Result != ValidationFailed {
			Result = ValidationSucceeded
		}
	}

	return nil
}
