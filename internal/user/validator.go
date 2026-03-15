package user

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// UserInfo holds the parsed components of a USER directive.
type UserInfo struct {
	Raw       string
	UserPart  string
	GroupPart string
	IsNumeric bool
	UID       *uint64
}

// Violation represents a single validation failure.
type Violation struct {
	Rule    string
	Message string
}

// Result holds the outcome of user validation.
type Result struct {
	Passed     bool
	Violations []Violation
}

// ParseUser parses the USER directive string into its components.
func ParseUser(user string) UserInfo {
	info := UserInfo{Raw: user}

	if user == "" {
		return info
	}

	parts := strings.SplitN(user, ":", 2)
	info.UserPart = parts[0]
	if len(parts) == 2 {
		info.GroupPart = parts[1]
	}

	if uid, err := strconv.ParseUint(info.UserPart, 10, 64); err == nil {
		info.IsNumeric = true
		info.UID = &uid
	}

	return info
}

// ValidateUser validates the USER directive against an optional policy.
// If policy is nil, only the basic non-root check is performed.
func ValidateUser(user string, policy *Policy) Result {
	info := ParseUser(user)
	var violations []Violation

	// Always enforce: empty user defaults to root
	if user == "" {
		violations = append(violations, Violation{
			Rule:    "non-empty",
			Message: "USER directive is not set (defaults to root)",
		})
		return Result{Passed: false, Violations: violations}
	}

	// Always enforce: explicit "root" username
	if info.UserPart == "root" {
		violations = append(violations, Violation{
			Rule:    "non-root",
			Message: "USER directive must not be root",
		})
		return Result{Passed: false, Violations: violations}
	}

	// Always enforce: UID 0 is root
	if info.IsNumeric && *info.UID == 0 {
		violations = append(violations, Violation{
			Rule:    "non-root-uid",
			Message: "UID 0 is root",
		})
		return Result{Passed: false, Violations: violations}
	}

	// Without policy, basic non-root check passes
	if policy == nil {
		return Result{Passed: true}
	}

	// With policy, check all rules (collect all violations)
	if policy.RequireNumeric != nil && *policy.RequireNumeric && !info.IsNumeric {
		violations = append(violations, Violation{
			Rule:    "require-numeric",
			Message: fmt.Sprintf("USER %q must be a numeric UID", info.UserPart),
		})
	}

	if slices.Contains(policy.BlockedUsers, info.UserPart) {
		violations = append(violations, Violation{
			Rule:    "blocked-user",
			Message: fmt.Sprintf("user %q is in the blocked users list", info.UserPart),
		})
	}

	// UID range checks only apply when userPart is numeric
	if info.IsNumeric {
		if policy.MinUID != nil && *info.UID < uint64(*policy.MinUID) {
			violations = append(violations, Violation{
				Rule:    "min-uid",
				Message: fmt.Sprintf("UID %d is below minimum %d", *info.UID, *policy.MinUID),
			})
		}
		if policy.MaxUID != nil && *info.UID > uint64(*policy.MaxUID) {
			violations = append(violations, Violation{
				Rule:    "max-uid",
				Message: fmt.Sprintf("UID %d is above maximum %d", *info.UID, *policy.MaxUID),
			})
		}
	}

	return Result{Passed: len(violations) == 0, Violations: violations}
}
