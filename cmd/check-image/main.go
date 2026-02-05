package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jarfernandez/check-image/cmd/check-image/commands"
)

// run executes the CLI and returns the exit code
// This function is testable because it doesn't call os.Exit
func run(stdout io.Writer) int {
	commands.Execute()

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
