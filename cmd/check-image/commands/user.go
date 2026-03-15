package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/jarfernandez/check-image/internal/user"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var userPolicy string
var userMinUID uint
var userMaxUID uint
var blockedUsers string
var requireNumeric bool

var userCmd = &cobra.Command{
	Use:   "user image",
	Short: "Validate that the image user meets security requirements",
	Long: `Validate that the image user meets security requirements.

Without any flags or policy file, performs a basic non-root check (same as root-user).
With flags or a policy file, enforces UID ranges, blocked usernames, and numeric UID requirements.

` + imageArgFormatsDoc,
	Example: `  check-image user nginx:latest
  check-image user nginx:latest --min-uid 1000
  check-image user nginx:latest --min-uid 1000 --max-uid 65534
  check-image user nginx:latest --blocked-users daemon,nobody,www-data
  check-image user nginx:latest --require-numeric
  check-image user nginx:latest --user-policy config/user-policy.yaml
  check-image user nginx:latest --user-policy config/user-policy.yaml --min-uid 500
  cat user-policy.yaml | check-image user nginx:latest --user-policy -
  check-image user oci:/path/to/layout:1.0 --min-uid 1000
  check-image user oci-archive:/path/to/image.tar:latest --user-policy policy.yaml
  check-image user docker-archive:/path/to/image.tar:tag`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		policy, err := resolveUserPolicy(cmd)
		if err != nil {
			return err
		}
		return runCheckCmd(checkUser, func(ctx context.Context, img string) (*output.CheckResult, error) {
			return runUser(ctx, img, policy)
		}, ctx, args[0], OutputFmt)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.Flags().StringVar(&userPolicy, "user-policy", "", "User policy file (JSON or YAML) (optional)")
	userCmd.Flags().UintVar(&userMinUID, "min-uid", 0, "Minimum allowed UID (optional)")
	userCmd.Flags().UintVar(&userMaxUID, "max-uid", 0, "Maximum allowed UID (optional)")
	userCmd.Flags().StringVar(&blockedUsers, "blocked-users", "", "Comma-separated list of blocked usernames (optional)")
	userCmd.Flags().BoolVar(&requireNumeric, "require-numeric", false, "Require user to be a numeric UID (optional)")
}

// resolveUserPolicy builds a *user.Policy from the combination of --user-policy file
// and individual CLI flags. CLI flags override policy file values.
// Returns nil when no flags or policy are provided (basic non-root check).
func resolveUserPolicy(cmd *cobra.Command) (*user.Policy, error) {
	var policy *user.Policy

	// Load policy file if provided
	if userPolicy != "" {
		p, err := user.LoadUserPolicy(userPolicy)
		if err != nil {
			return nil, fmt.Errorf("unable to load user policy: %w", err)
		}
		policy = p
	}

	// Check if any individual flags were set
	hasFlags := cmd.Flags().Changed("min-uid") ||
		cmd.Flags().Changed("max-uid") ||
		cmd.Flags().Changed("blocked-users") ||
		cmd.Flags().Changed("require-numeric")

	if !hasFlags {
		return policy, nil
	}

	// Create policy if not loaded from file
	if policy == nil {
		policy = &user.Policy{}
	}

	// Overlay CLI flags on top of policy
	if cmd.Flags().Changed("min-uid") {
		policy.MinUID = &userMinUID
	}
	if cmd.Flags().Changed("max-uid") {
		policy.MaxUID = &userMaxUID
	}
	if cmd.Flags().Changed("blocked-users") {
		policy.BlockedUsers = parseBlockedUsers(blockedUsers)
	}
	if cmd.Flags().Changed("require-numeric") {
		policy.RequireNumeric = &requireNumeric
	}

	// Validate the final combined policy
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	return policy, nil
}

// parseBlockedUsers splits a comma-separated list of usernames, trimming whitespace.
func parseBlockedUsers(s string) []string {
	if s == "" {
		return nil
	}
	var users []string
	for part := range strings.SplitSeq(s, ",") {
		name := strings.TrimSpace(part)
		if name != "" {
			users = append(users, name)
		}
	}
	return users
}

func runUser(ctx context.Context, imageName string, policy *user.Policy) (*output.CheckResult, error) {
	_, config, cleanup, err := imageutil.GetImageAndConfig(ctx, imageName)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	userValue := config.Config.User
	result := user.ValidateUser(userValue, policy)
	info := user.ParseUser(userValue)

	log.Debugf("USER directive: %q, is-numeric: %v, passed: %v, violations: %d",
		userValue, info.IsNumeric, result.Passed, len(result.Violations))

	var msg string
	if result.Passed {
		msg = "Image user meets all requirements"
	} else {
		msg = "Image user does not meet requirements"
	}

	// Build violations for output
	var violations []output.UserViolation
	for _, v := range result.Violations {
		violations = append(violations, output.UserViolation{
			Rule:    v.Rule,
			Message: v.Message,
		})
	}

	details := output.UserDetails{
		User:       userValue,
		IsNumeric:  info.IsNumeric,
		UID:        info.UID,
		Violations: violations,
	}

	// Include policy constraints in output when a policy was provided
	if policy != nil {
		details.MinUID = policy.MinUID
		details.MaxUID = policy.MaxUID
		details.BlockedUsers = policy.BlockedUsers
		if policy.RequireNumeric != nil {
			details.RequireNumeric = *policy.RequireNumeric
		}
	}

	return &output.CheckResult{
		Check:   checkUser,
		Image:   imageName,
		Passed:  result.Passed,
		Message: msg,
		Details: details,
	}, nil
}
