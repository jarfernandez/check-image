package commands

import (
	"fmt"
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

type ValidationResult int

const (
	ValidationFailed ValidationResult = iota
	ValidationSucceeded
	ValidationSkipped
)

var Result = ValidationSkipped

var logLevel string

var rootCmd = &cobra.Command{
	Use:          "check-image",
	Short:        "Validation of container images",
	Long:         `Validation of container images to ensure they meet certain standards (size, age, ports, security configurations, etc.).`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level %s: %w", logLevel, err)
		}
		log.SetLevel(level)
		log.Debugln("Log level set to", level.String())

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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing check-image: %v", err)
	}
}

// SetValidationResult updates the global validation result based on whether
// validation passed or failed, and prints the appropriate message.
func SetValidationResult(passed bool, successMsg, failureMsg string) {
	if passed {
		fmt.Println(successMsg)
		if Result != ValidationFailed {
			Result = ValidationSucceeded
		}
	} else {
		fmt.Println(failureMsg)
		Result = ValidationFailed
	}
}
