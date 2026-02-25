package commands

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/output"
)

// runCheckCmd is the standard RunE body shared by every single-check command.
// checkName is used only for the error message; run is the check implementation.
func runCheckCmd(checkName string, run func(string) (*output.CheckResult, error), imageName string) error {
	result, err := run(imageName)
	if err != nil {
		return fmt.Errorf("check %s operation failed: %w", checkName, err)
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
}
