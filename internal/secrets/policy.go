package secrets

import (
	"fmt"

	"github.com/jarfernandez/check-image/internal/fileutil"
)

// Policy defines configuration for secrets detection
type Policy struct {
	CheckEnvVars       bool     `yaml:"check-env-vars" json:"check-env-vars"`
	CheckFiles         bool     `yaml:"check-files" json:"check-files"`
	ExcludedPaths      []string `yaml:"excluded-paths" json:"excluded-paths"`
	ExcludedEnvVars    []string `yaml:"excluded-env-vars" json:"excluded-env-vars"`
	CustomEnvPatterns  []string `yaml:"custom-env-patterns" json:"custom-env-patterns"`
	CustomFilePatterns []string `yaml:"custom-file-patterns" json:"custom-file-patterns"`
}

// Default patterns for detection
var (
	// DefaultEnvPatterns are case-insensitive keywords that indicate sensitive environment variables
	DefaultEnvPatterns = []string{
		"password",
		"passwd",
		"secret",
		"token",
		"key",
		"credential",
		"auth",
		"api",
	}

	// DefaultExcludedEnvVars are environment variables that match patterns but aren't actually secrets
	DefaultExcludedEnvVars = []string{
		"PUBLIC_KEY",
		"SSH_PUBLIC_KEY",
	}

	// DefaultFilePatterns maps file patterns to their descriptions.
	// This is the single source of truth for file pattern detection.
	// #nosec G101 -- These are pattern names for detecting secrets, not actual credentials
	DefaultFilePatterns = map[string]string{
		// SSH keys
		"id_rsa":     "SSH private key",
		"id_dsa":     "SSH private key",
		"id_ecdsa":   "SSH private key",
		"id_ed25519": "SSH private key",
		"*.ppk":      "PuTTY private key",

		// Cloud credentials
		".aws/credentials": "AWS credentials",
		".kube/config":     "Kubernetes config",

		// Keys
		"*.key": "private key file",

		// Password files
		"/etc/shadow": "shadow password file",
		".pgpass":     "PostgreSQL password file",
		".my.cnf":     "MySQL credentials",
		".netrc":      "authentication credentials",

		// Others
		".npmrc":           "NPM credentials",
		".git-credentials": "Git credentials",
		"secrets.json":     "secrets file",
		"secrets.yaml":     "secrets file",
		"secrets.yml":      "secrets file",
		"wallet.dat":       "cryptocurrency wallet",
	}
)

// LoadSecretsPolicy loads a secrets policy from a file or stdin (if path is "-")
// which can be in either YAML or JSON format.
// If path is empty, returns a default policy.
func LoadSecretsPolicy(path string) (*Policy, error) {
	// Return default policy if no path provided
	if path == "" {
		return &Policy{
			CheckEnvVars:      true,
			CheckFiles:        true,
			ExcludedPaths:     []string{},
			ExcludedEnvVars:   DefaultExcludedEnvVars,
			CustomEnvPatterns: []string{},
		}, nil
	}

	// Read file or stdin
	data, err := fileutil.ReadFileOrStdin(path)
	if err != nil {
		return nil, fmt.Errorf("error reading secrets policy: %w", err)
	}

	var policy Policy

	// Unmarshal config data (JSON or YAML)
	if err := fileutil.UnmarshalConfigData(data, &policy, path); err != nil {
		return nil, err
	}

	// Set defaults for excluded env vars if not specified
	if len(policy.ExcludedEnvVars) == 0 {
		policy.ExcludedEnvVars = DefaultExcludedEnvVars
	}

	return &policy, nil
}

// GetEnvPatterns returns all environment variable patterns (default + custom)
func (p *Policy) GetEnvPatterns() []string {
	patterns := make([]string, len(DefaultEnvPatterns))
	copy(patterns, DefaultEnvPatterns)
	patterns = append(patterns, p.CustomEnvPatterns...)
	return patterns
}

// GetFilePatterns returns all file patterns (default + custom)
func (p *Policy) GetFilePatterns() []string {
	patterns := make([]string, 0, len(DefaultFilePatterns)+len(p.CustomFilePatterns))
	for pattern := range DefaultFilePatterns {
		patterns = append(patterns, pattern)
	}
	patterns = append(patterns, p.CustomFilePatterns...)
	return patterns
}
