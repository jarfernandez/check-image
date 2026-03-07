package commands

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/jarfernandez/check-image/internal/secrets"
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

` + imageArgFormatsDoc,
	Example: `  check-image secrets nginx:latest
  check-image secrets nginx:latest --secrets-policy secrets-policy.json
  check-image secrets nginx:latest --secrets-policy secrets-policy.yaml
  check-image secrets nginx:latest --skip-env-vars
  check-image secrets nginx:latest --skip-files
  check-image secrets oci:/path/to/layout:1.0
  check-image secrets oci-archive:/path/to/image.tar:latest --secrets-policy secrets-policy.json
  check-image secrets docker-archive:/path/to/image.tar:tag --skip-files
  cat secrets-policy.yaml | check-image secrets nginx:latest --secrets-policy -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckCmd(checkSecrets, func(img string) (*output.CheckResult, error) {
			return runSecrets(img, secretsPolicy, skipEnvVars, skipFiles)
		}, args[0])
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.Flags().StringVarP(&secretsPolicy, "secrets-policy", "s", "", "Secrets policy file (JSON or YAML) (optional)")
	secretsCmd.Flags().BoolVar(&skipEnvVars, "skip-env-vars", false, "Skip environment variable checks (optional)")
	secretsCmd.Flags().BoolVar(&skipFiles, "skip-files", false, "Skip file system checks (optional)")
}

func runSecrets(imageName string, policyPath string, noEnvVars bool, noFiles bool) (*output.CheckResult, error) {
	policy, err := secrets.LoadSecretsPolicy(policyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load secrets policy: %w", err)
	}

	// Override policy based on command-line flags
	if noEnvVars {
		policy.CheckEnvVars = false
	}
	if noFiles {
		policy.CheckFiles = false
	}

	image, config, cleanup, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var envFindings []output.EnvVarFinding
	var fileFindings []output.FileFinding

	// Check environment variables
	if policy.CheckEnvVars {
		log.Debug("Checking environment variables for secrets")
		rawEnvFindings := secrets.CheckEnvironmentVariables(config.Config.Env, policy)
		for _, f := range rawEnvFindings {
			envFindings = append(envFindings, output.EnvVarFinding{
				Name:        f.Name,
				Description: f.Description,
			})
		}
	}

	// Check files in layers
	if policy.CheckFiles {
		log.Debug("Checking files in layers for secrets")
		rawFileFindings, err := secrets.CheckFilesInLayers(image, policy)
		if err != nil {
			return nil, fmt.Errorf("error scanning files: %w", err)
		}
		for _, f := range rawFileFindings {
			fileFindings = append(fileFindings, output.FileFinding{
				Path:        f.Path,
				LayerIndex:  f.LayerIndex,
				Description: f.Description,
			})
		}
	}

	envCount := len(envFindings)
	fileCount := len(fileFindings)
	totalFindings := envCount + fileCount
	passed := totalFindings == 0

	var msg string
	if passed {
		msg = "No secrets detected"
	} else {
		msg = "Secrets detected"
	}

	details := output.SecretsDetails{
		EnvVarFindings: envFindings,
		FileFindings:   fileFindings,
		TotalFindings:  totalFindings,
		EnvVarCount:    envCount,
		FileCount:      fileCount,
	}

	return &output.CheckResult{
		Check:   checkSecrets,
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: details,
	}, nil
}
