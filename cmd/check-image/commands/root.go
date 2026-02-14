package commands

import (
	"fmt"
	"os"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ValidationResult int

const (
	ValidationSkipped   ValidationResult = iota // 0 - no checks ran
	ValidationSucceeded                         // 1 - all checks passed
	ValidationFailed                            // 2 - one or more checks failed
	ExecutionError                              // 3 - tool could not run properly
)

var Result = ValidationSkipped

var logLevel string
var outputFormat string

// OutputFmt holds the parsed output format after PersistentPreRunE.
var OutputFmt output.Format

var rootCmd = &cobra.Command{
	Use:           "check-image",
	Short:         "Validation of container images",
	Long:          `Validation of container images to ensure they meet certain standards (size, age, ports, security configurations, etc.).`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level %s: %w", logLevel, err)
		}
		log.SetLevel(level)
		log.Debugln("Log level set to", level.String())

		f, err := output.ParseFormat(outputFormat)
		if err != nil {
			return err
		}
		OutputFmt = f

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   !isatty.IsTerminal(os.Stderr.Fd()),
	})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Sets the log level (trace, debug, info, warn, error, fatal, panic) (optional)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json (optional)")
}

// UpdateResult updates the global Result with proper precedence.
// Priority ordering: ValidationSkipped(0) < ValidationSucceeded(1) < ValidationFailed(2) < ExecutionError(3).
func UpdateResult(new ValidationResult) {
	if new > Result {
		Result = new
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error executing check-image: %v", err)
		Result = ExecutionError
	}
}
