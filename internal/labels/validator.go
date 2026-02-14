package labels

import (
	"fmt"
	"regexp"
)

// ValidationResult represents the result of label validation
type ValidationResult struct {
	Passed        bool
	MissingLabels []string
	InvalidLabels []InvalidLabel
}

// InvalidLabel represents a label that exists but doesn't meet requirements
type InvalidLabel struct {
	Name            string
	ActualValue     string
	ExpectedValue   string
	ExpectedPattern string
	Reason          string
}

// ValidateLabels checks image labels against policy requirements.
// It returns a comprehensive validation result containing all failures.
// The validation passes only if all required labels exist and meet their requirements.
func ValidateLabels(imageLabels map[string]string, policy *Policy) (*ValidationResult, error) {
	result := &ValidationResult{
		Passed:        true,
		MissingLabels: make([]string, 0),
		InvalidLabels: make([]InvalidLabel, 0),
	}

	// Iterate through all required labels
	for _, req := range policy.RequiredLabels {
		actualValue, exists := imageLabels[req.Name]

		// Check if label exists
		if !exists {
			result.Passed = false
			result.MissingLabels = append(result.MissingLabels, req.Name)
			continue
		}

		// Label exists - validate value if required
		if req.Value != "" {
			// Exact value match required (case-sensitive)
			if actualValue != req.Value {
				result.Passed = false
				result.InvalidLabels = append(result.InvalidLabels, InvalidLabel{
					Name:          req.Name,
					ActualValue:   actualValue,
					ExpectedValue: req.Value,
					Reason:        fmt.Sprintf("label %q has value %q but expected %q", req.Name, actualValue, req.Value),
				})
			}
		} else if req.Pattern != "" {
			// Pattern match required
			re, err := regexp.Compile(req.Pattern)
			if err != nil {
				// This should not happen if policy was validated properly
				return nil, fmt.Errorf("failed to compile pattern for label %q: %w", req.Name, err)
			}

			if !re.MatchString(actualValue) {
				result.Passed = false
				result.InvalidLabels = append(result.InvalidLabels, InvalidLabel{
					Name:            req.Name,
					ActualValue:     actualValue,
					ExpectedPattern: req.Pattern,
					Reason:          fmt.Sprintf("label %q value %q does not match pattern %q", req.Name, actualValue, req.Pattern),
				})
			}
		}
		// If neither value nor pattern is specified, existence check is sufficient (already passed)
	}

	return result, nil
}
