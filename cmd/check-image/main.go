package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jarfernandez/check-image/cmd/check-image/commands"
	"github.com/jarfernandez/check-image/internal/output"
)

// exitResult maps an ExecuteResult to an exit code and prints the final
// status message when appropriate.
func exitResult(result commands.ExecuteResult, stdout io.Writer) int {
	// Execution error has the highest priority — exit code 2.
	// The detailed error message is already logged to stderr by Execute().
	if result.Validation == commands.ExecutionError {
		if result.Format != output.FormatJSON {
			if _, err := fmt.Fprintln(stdout, commands.FailStyle.Render("Execution error")); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		return 2
	}

	// In JSON mode, suppress the final text message (already in JSON)
	if result.Format == output.FormatJSON {
		if result.Validation == commands.ValidationFailed {
			return 1
		}
		return 0
	}

	if result.Validation == commands.ValidationFailed {
		if _, err := fmt.Fprintln(stdout, commands.FailStyle.Render("Validation failed")); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		return 1
	}

	if result.Validation == commands.ValidationSucceeded {
		if _, err := fmt.Fprintln(stdout, commands.PassStyle.Render("Validation succeeded")); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
	}

	return 0
}

// run executes the CLI and returns the exit code.
func run(stdout io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	result := commands.Execute(ctx)
	return exitResult(result, stdout)
}

func main() {
	exitCode := run(os.Stdout)
	os.Exit(exitCode)
}
