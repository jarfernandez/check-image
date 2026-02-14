package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jarfernandez/check-image/cmd/check-image/commands"
	"github.com/jarfernandez/check-image/internal/output"
)

// run executes the CLI and returns the exit code
// This function is testable because it doesn't call os.Exit
func run(stdout io.Writer) int {
	commands.Execute()

	// Execution error has the highest priority — exit code 2.
	// The detailed error message is already logged to stderr by Execute().
	if commands.Result == commands.ExecutionError {
		if commands.OutputFmt != output.FormatJSON {
			if _, err := fmt.Fprintln(stdout, "Execution error"); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		return 2
	}

	// In JSON mode, suppress the final text message (already in JSON)
	if commands.OutputFmt == output.FormatJSON {
		if commands.Result == commands.ValidationFailed {
			return 1
		}
		return 0
	}

	if commands.Result == commands.ValidationFailed {
		if _, err := fmt.Fprintln(stdout, "Validation failed"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		return 1
	}

	if commands.Result == commands.ValidationSucceeded {
		if _, err := fmt.Fprintln(stdout, "Validation succeeded"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
	}

	return 0
}

func main() {
	exitCode := run(os.Stdout)
	os.Exit(exitCode)
}
