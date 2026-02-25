package commands

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/labels"
	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var labelsPolicy string

var labelsCmd = &cobra.Command{
	Use:   "labels image",
	Short: "Validate that the image has required labels with correct values",
	Long: `Validate that the image has required labels with correct values.

The labels validation checks that required OCI/Docker labels exist in the image
and optionally validates their values against exact matches or regex patterns.

` + imageArgFormatsDoc,
	Example: `  check-image labels nginx:latest --labels-policy labels-policy.json
  check-image labels nginx:latest --labels-policy labels-policy.yaml
  check-image labels nginx:latest --labels-policy labels-policy.yaml -o json
  check-image labels oci:/path/to/layout:1.0 --labels-policy labels-policy.json
  cat labels-policy.yaml | check-image labels nginx:latest --labels-policy -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheckCmd("labels", runLabels, args[0])
	},
}

func init() {
	rootCmd.AddCommand(labelsCmd)
	labelsCmd.Flags().StringVar(&labelsPolicy, "labels-policy", "", "Labels policy file (JSON or YAML)")
	if err := labelsCmd.MarkFlagRequired("labels-policy"); err != nil {
		panic(fmt.Sprintf("failed to mark labels-policy flag as required: %v", err))
	}
}

func runLabels(imageName string) (*output.CheckResult, error) {
	// Load policy
	policy, err := labels.LoadLabelsPolicy(labelsPolicy)
	if err != nil {
		return nil, fmt.Errorf("unable to load labels policy: %w", err)
	}

	log.Debugf("Loaded policy with %d required labels", len(policy.RequiredLabels))

	// Get image config
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

	// Get labels from image (may be nil)
	imageLabels := config.Config.Labels
	if imageLabels == nil {
		imageLabels = make(map[string]string)
	}

	log.Debugf("Image has %d labels", len(imageLabels))

	// Validate labels
	validationResult, err := labels.ValidateLabels(imageLabels, policy)
	if err != nil {
		return nil, fmt.Errorf("label validation failed: %w", err)
	}

	// Build details for output
	reqLabels := make([]output.RequiredLabelCheck, len(policy.RequiredLabels))
	for i, req := range policy.RequiredLabels {
		reqLabels[i] = output.RequiredLabelCheck{
			Name:    req.Name,
			Value:   req.Value,
			Pattern: req.Pattern,
		}
	}

	invalidDetails := make([]output.InvalidLabelDetail, len(validationResult.InvalidLabels))
	for i, inv := range validationResult.InvalidLabels {
		invalidDetails[i] = output.InvalidLabelDetail{
			Name:            inv.Name,
			ActualValue:     inv.ActualValue,
			ExpectedValue:   inv.ExpectedValue,
			ExpectedPattern: inv.ExpectedPattern,
			Reason:          inv.Reason,
		}
	}

	details := output.LabelsDetails{
		RequiredLabels: reqLabels,
		ActualLabels:   imageLabels,
		MissingLabels:  validationResult.MissingLabels,
		InvalidLabels:  invalidDetails,
	}

	// Build message
	var msg string
	if validationResult.Passed {
		msg = "All required labels are present and valid"
	} else {
		msg = "Image does not meet label requirements"
	}

	return &output.CheckResult{
		Check:   "labels",
		Image:   imageName,
		Passed:  validationResult.Passed,
		Message: msg,
		Details: details,
	}, nil
}
