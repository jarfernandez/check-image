package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jarfernandez/check-image/internal/fileutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// validCheckNames lists all check names recognized by the all command.
var validCheckNames = []string{"age", "size", "ports", "registry", "root-user", "secrets", "healthcheck", "labels", "entrypoint", "platform"}

// allConfig represents the configuration file structure for the all command.
type allConfig struct {
	Checks allChecksConfig `json:"checks" yaml:"checks"`
}

type allChecksConfig struct {
	Age         *ageCheckConfig         `json:"age,omitempty"       yaml:"age,omitempty"`
	Size        *sizeCheckConfig        `json:"size,omitempty"      yaml:"size,omitempty"`
	Ports       *portsCheckConfig       `json:"ports,omitempty"     yaml:"ports,omitempty"`
	Registry    *registryCheckConfig    `json:"registry,omitempty"  yaml:"registry,omitempty"`
	RootUser    *rootUserCheckConfig    `json:"root-user,omitempty" yaml:"root-user,omitempty"`
	Secrets     *secretsCheckConfig     `json:"secrets,omitempty"   yaml:"secrets,omitempty"`
	Healthcheck *healthcheckCheckConfig `json:"healthcheck,omitempty"  yaml:"healthcheck,omitempty"`
	Labels      *labelsCheckConfig      `json:"labels,omitempty"       yaml:"labels,omitempty"`
	Entrypoint  *entrypointCheckConfig  `json:"entrypoint,omitempty"   yaml:"entrypoint,omitempty"`
	Platform    *platformCheckConfig    `json:"platform,omitempty"     yaml:"platform,omitempty"`
}

type ageCheckConfig struct {
	MaxAge *uint `json:"max-age,omitempty" yaml:"max-age,omitempty"`
}

type sizeCheckConfig struct {
	MaxSize   *uint `json:"max-size,omitempty"   yaml:"max-size,omitempty"`
	MaxLayers *uint `json:"max-layers,omitempty" yaml:"max-layers,omitempty"`
}

type portsCheckConfig struct {
	AllowedPorts any `json:"allowed-ports,omitempty" yaml:"allowed-ports,omitempty"`
}

type registryCheckConfig struct {
	RegistryPolicy any `json:"registry-policy,omitempty" yaml:"registry-policy,omitempty"`
}

type rootUserCheckConfig struct{}

type healthcheckCheckConfig struct{}

type entrypointCheckConfig struct {
	AllowShellForm *bool `json:"allow-shell-form,omitempty" yaml:"allow-shell-form,omitempty"`
}

type secretsCheckConfig struct {
	SecretsPolicy any   `json:"secrets-policy,omitempty" yaml:"secrets-policy,omitempty"`
	SkipEnvVars   *bool `json:"skip-env-vars,omitempty"  yaml:"skip-env-vars,omitempty"`
	SkipFiles     *bool `json:"skip-files,omitempty"     yaml:"skip-files,omitempty"`
}

type labelsCheckConfig struct {
	LabelsPolicy any `json:"labels-policy,omitempty" yaml:"labels-policy,omitempty"`
}

type platformCheckConfig struct {
	AllowedPlatforms any `json:"allowed-platforms,omitempty" yaml:"allowed-platforms,omitempty"`
}

// parseCheckNameList parses a comma-separated list of check names and validates
// each name against validCheckNames. Returns a map of valid check names.
func parseCheckNameList(list string) (map[string]bool, error) {
	if list == "" {
		return nil, nil
	}

	validNames := make(map[string]bool)
	for _, name := range validCheckNames {
		validNames[name] = true
	}

	nameMap := make(map[string]bool)
	for part := range strings.SplitSeq(list, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if !validNames[name] {
			return nil, fmt.Errorf("unknown check name %q, valid names are: %s", name, strings.Join(validCheckNames, ", "))
		}
		nameMap[name] = true
	}

	return nameMap, nil
}

func loadAllConfig(path string) (*allConfig, error) {
	data, err := fileutil.ReadFileOrStdin(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg allConfig
	if err := fileutil.UnmarshalConfigData(data, &cfg, path); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// applyConfigValues applies configuration file values to package-level variables,
// but only for flags that were not explicitly set on the CLI.
// It returns a cleanup function that removes any temporary files created for
// inline policy objects; the caller must defer the returned function.
func applyConfigValues(cmd *cobra.Command, cfg *allConfig) func() {
	applyAgeConfig(cmd, cfg.Checks.Age)
	applySizeConfig(cmd, cfg.Checks.Size)
	applyPortsConfig(cmd, cfg.Checks.Ports)
	cleanupRegistry := applyRegistryConfig(cmd, cfg.Checks.Registry)
	cleanupSecrets := applySecretsConfig(cmd, cfg.Checks.Secrets)
	cleanupLabels := applyLabelsConfig(cmd, cfg.Checks.Labels)
	applyEntrypointConfig(cmd, cfg.Checks.Entrypoint)
	applyPlatformConfig(cmd, cfg.Checks.Platform)
	return func() {
		cleanupRegistry()
		cleanupSecrets()
		cleanupLabels()
	}
}

func applyAgeConfig(cmd *cobra.Command, cfg *ageCheckConfig) {
	if cfg != nil && cfg.MaxAge != nil && !cmd.Flags().Changed("max-age") {
		maxAge = *cfg.MaxAge
	}
}

func applySizeConfig(cmd *cobra.Command, cfg *sizeCheckConfig) {
	if cfg == nil {
		return
	}
	if cfg.MaxSize != nil && !cmd.Flags().Changed("max-size") {
		maxSize = *cfg.MaxSize
	}
	if cfg.MaxLayers != nil && !cmd.Flags().Changed("max-layers") {
		maxLayers = *cfg.MaxLayers
	}
}

func applyPortsConfig(cmd *cobra.Command, cfg *portsCheckConfig) {
	if cfg != nil && cfg.AllowedPorts != nil && !cmd.Flags().Changed("allowed-ports") {
		allowedPorts = formatAllowedList(cfg.AllowedPorts)
	}
}

// applyInlinePolicy resolves an inline policy value to a temp-file path and sets
// *target. Returns a cleanup func. If policyVal is nil or flagName was explicitly
// set on the CLI, it is a no-op.
func applyInlinePolicy(cmd *cobra.Command, flagName string, policyVal any, target *string) func() {
	if policyVal == nil || cmd.Flags().Changed(flagName) {
		return func() {}
	}
	path, cleanup, err := inlinePolicyToTempFile(flagName, policyVal)
	if err != nil {
		log.Errorf("Failed to format %s: %v", flagName, err)
		return func() {}
	}
	*target = path
	return cleanup
}

func applyRegistryConfig(cmd *cobra.Command, cfg *registryCheckConfig) func() {
	if cfg == nil {
		return func() {}
	}
	return applyInlinePolicy(cmd, "registry-policy", cfg.RegistryPolicy, &registryPolicy)
}

func applySecretsConfig(cmd *cobra.Command, cfg *secretsCheckConfig) func() {
	if cfg == nil {
		return func() {}
	}
	cleanup := func() {}
	if cfg.SecretsPolicy != nil && !cmd.Flags().Changed("secrets-policy") {
		path, cl, err := inlinePolicyToTempFile("secrets-policy", cfg.SecretsPolicy)
		if err != nil {
			log.Errorf("Failed to format secrets policy: %v", err)
		} else {
			secretsPolicy = path
			cleanup = cl
		}
	}
	if cfg.SkipEnvVars != nil && !cmd.Flags().Changed("skip-env-vars") {
		skipEnvVars = *cfg.SkipEnvVars
	}
	if cfg.SkipFiles != nil && !cmd.Flags().Changed("skip-files") {
		skipFiles = *cfg.SkipFiles
	}
	return cleanup
}

func applyLabelsConfig(cmd *cobra.Command, cfg *labelsCheckConfig) func() {
	if cfg == nil {
		return func() {}
	}
	return applyInlinePolicy(cmd, "labels-policy", cfg.LabelsPolicy, &labelsPolicy)
}

func applyEntrypointConfig(cmd *cobra.Command, cfg *entrypointCheckConfig) {
	if cfg != nil && cfg.AllowShellForm != nil && !cmd.Flags().Changed("allow-shell-form") {
		allowShellForm = *cfg.AllowShellForm
	}
}

func applyPlatformConfig(cmd *cobra.Command, cfg *platformCheckConfig) {
	if cfg != nil && cfg.AllowedPlatforms != nil && !cmd.Flags().Changed("allowed-platforms") {
		allowedPlatforms = formatAllowedList(cfg.AllowedPlatforms)
	}
}

// formatAllowedList converts a config value ([]any or string) to a comma-separated string.
// It is used for allowed-ports and allowed-platforms config fields, which can be specified
// as either a slice (e.g. [80, 443]) or a pre-joined string (e.g. "80,443").
func formatAllowedList(v any) string {
	switch items := v.(type) {
	case []any:
		parts := make([]string, 0, len(items))
		for _, p := range items {
			parts = append(parts, fmt.Sprintf("%v", p))
		}
		return strings.Join(parts, ",")
	case string:
		return items
	default:
		return fmt.Sprintf("%v", v)
	}
}

// parseAllowedListFromFile reads data from path (may be "-" for stdin) and
// unmarshals it into dest. It is the shared @file implementation for
// parseAllowedPorts and parseAllowedPlatforms.
func parseAllowedListFromFile(path string, dest any) error {
	data, err := fileutil.ReadFileOrStdin(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return fileutil.UnmarshalConfigData(data, dest, path)
}

// inlinePolicyToTempFile converts an inline policy object (map[string]any) to a
// temporary JSON file so it can be passed to the policy loaders that expect a
// file path. If v is already a string it is returned as-is with a no-op cleanup.
// The caller must invoke the returned cleanup function when done (typically via defer).
func inlinePolicyToTempFile(prefix string, v any) (path string, cleanup func(), err error) {
	switch policy := v.(type) {
	case string:
		return policy, func() {}, nil
	case map[string]any:
		data, err := json.Marshal(policy)
		if err != nil {
			return "", func() {}, fmt.Errorf("failed to marshal inline %s: %w", prefix, err)
		}
		tmpFile, err := os.CreateTemp("", prefix+"-*.json")
		if err != nil {
			return "", func() {}, fmt.Errorf("failed to create temp file for inline %s: %w", prefix, err)
		}
		name := tmpFile.Name()
		if _, err := tmpFile.Write(data); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(name) // #nosec G703 -- name comes from os.CreateTemp, not user input
			return "", func() {}, fmt.Errorf("failed to write inline %s to temp file: %w", prefix, err)
		}
		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(name) // #nosec G703 -- name comes from os.CreateTemp, not user input
			return "", func() {}, fmt.Errorf("failed to close temp file for inline %s: %w", prefix, err)
		}
		return name, func() {
			if removeErr := os.Remove(name); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Warnf("failed to remove temp file %s: %v", name, removeErr)
			}
		}, nil
	default:
		return "", func() {}, fmt.Errorf("%s must be either a string (file path) or an object (inline policy), got %T", prefix, v)
	}
}
