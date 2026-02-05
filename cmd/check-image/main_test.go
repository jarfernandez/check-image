package main

import (
	"bytes"
	"testing"

	"github.com/jarfernandez/check-image/cmd/check-image/commands"
	"github.com/stretchr/testify/assert"
)

func TestRun_ValidationSucceeded(t *testing.T) {
	// Setup - simulate a successful validation
	commands.Result = commands.ValidationSucceeded
	var buf bytes.Buffer

	// Execute
	exitCode := run(&buf)

	// Verify
	assert.Equal(t, 0, exitCode, "Should return exit code 0 for success")
	assert.Contains(t, buf.String(), "Validation succeeded", "Should print success message")
}

func TestRun_ValidationFailed(t *testing.T) {
	// Setup - simulate a failed validation
	commands.Result = commands.ValidationFailed
	var buf bytes.Buffer

	// Execute
	exitCode := run(&buf)

	// Verify
	assert.Equal(t, 1, exitCode, "Should return exit code 1 for failure")
	assert.Contains(t, buf.String(), "Validation failed", "Should print failure message")
}

func TestRun_ValidationSkipped(t *testing.T) {
	// Setup - simulate a skipped validation
	commands.Result = commands.ValidationSkipped
	var buf bytes.Buffer

	// Execute
	exitCode := run(&buf)

	// Verify
	assert.Equal(t, 0, exitCode, "Should return exit code 0 for skipped validation")
	assert.NotContains(t, buf.String(), "Validation succeeded", "Should not print success message")
	assert.NotContains(t, buf.String(), "Validation failed", "Should not print failure message")
}

func TestRun_OutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		result         commands.ValidationResult
		expectedExit   int
		expectedOutput string
		shouldContain  bool
	}{
		{
			name:           "Success prints to stdout",
			result:         commands.ValidationSucceeded,
			expectedExit:   0,
			expectedOutput: "Validation succeeded\n",
			shouldContain:  true,
		},
		{
			name:           "Failure prints to stdout",
			result:         commands.ValidationFailed,
			expectedExit:   1,
			expectedOutput: "Validation failed\n",
			shouldContain:  true,
		},
		{
			name:           "Skipped produces no output",
			result:         commands.ValidationSkipped,
			expectedExit:   0,
			expectedOutput: "",
			shouldContain:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			commands.Result = tt.result
			var buf bytes.Buffer

			// Execute
			exitCode := run(&buf)

			// Verify
			assert.Equal(t, tt.expectedExit, exitCode)
			if tt.shouldContain {
				assert.Equal(t, tt.expectedOutput, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestRun_PreservesState(t *testing.T) {
	// This test verifies that run() doesn't modify global state unexpectedly
	// Setup
	initialResult := commands.ValidationSucceeded
	commands.Result = initialResult
	var buf bytes.Buffer

	// Execute
	_ = run(&buf)

	// Verify - Result should still be the same after run()
	assert.Equal(t, initialResult, commands.Result, "run() should not modify the Result")
}
