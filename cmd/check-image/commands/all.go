package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/jarfernandez/check-image/internal/fileutil"
	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// validCheckNames lists all check names recognized by the all command.
var validCheckNames = []string{"age", "size", "ports", "registry", "root-user", "secrets"}

// allConfig represents the configuration file structure for the all command.
type allConfig struct {
	Checks allChecksConfig `json:"checks" yaml:"checks"`
}

type allChecksConfig struct {
	Age      *ageCheckConfig      `json:"age,omitempty"       yaml:"age,omitempty"`
	Size     *sizeCheckConfig     `json:"size,omitempty"      yaml:"size,omitempty"`
	Ports    *portsCheckConfig    `json:"ports,omitempty"     yaml:"ports,omitempty"`
	Registry *registryCheckConfig `json:"registry,omitempty"  yaml:"registry,omitempty"`
	RootUser *rootUserCheckConfig `json:"root-user,omitempty" yaml:"root-user,omitempty"`
	Secrets  *secretsCheckConfig  `json:"secrets,omitempty"   yaml:"secrets,omitempty"`
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
	RegistryPolicy *string `json:"registry-policy,omitempty" yaml:"registry-policy,omitempty"`
}

type rootUserCheckConfig struct{}

type secretsCheckConfig struct {
	SecretsPolicy *string `json:"secrets-policy,omitempty" yaml:"secrets-policy,omitempty"`
	SkipEnvVars   *bool   `json:"skip-env-vars,omitempty"  yaml:"skip-env-vars,omitempty"`
	SkipFiles     *bool   `json:"skip-files,omitempty"     yaml:"skip-files,omitempty"`
}

// checkRunner represents a single check to be executed.
type checkRunner struct {
	name string
	run  func(imageName string) (*output.CheckResult, error)
}

var configFile string
var skipChecks string
var failFast bool

var allCmd = &cobra.Command{
	Use:   "all image",
	Short: "Run all validation checks on a container image",
	Long: `Run all validation checks on a container image at once.

By default, runs all checks (age, size, ports, registry, root-user, secrets).
Use --config to specify which checks to run and their parameters.
Use --skip to skip specific checks.
Use --fail-fast to stop on the first check failure.

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag

Precedence rules:
  1. Without --config: all 6 checks run with defaults, except those in --skip
  2. With --config: only checks present in the config file run, except those in --skip
  3. CLI flags override config file values
  4. --skip always takes precedence over the config file`,
	Example: `  check-image all nginx:latest
  check-image all nginx:latest --max-age 30 --max-size 200
  check-image all nginx:latest --skip registry,secrets
  check-image all nginx:latest --config config/config.json
  check-image all nginx:latest -c config/config.yaml --max-age 30 --skip secrets
  check-image all oci:/path/to/layout:1.0
  check-image all oci-archive:/path/to/image.tar:latest --skip ports,registry
  check-image all nginx:latest --fail-fast --skip registry`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAll(cmd, args[0]); err != nil {
			return fmt.Errorf("check all operation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(allCmd)

	allCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (JSON or YAML) (optional)")
	allCmd.Flags().StringVar(&skipChecks, "skip", "", "Comma-separated list of checks to skip (age, size, ports, registry, root-user, secrets) (optional)")
	allCmd.Flags().UintVarP(&maxAge, "max-age", "a", 90, "Maximum age in days (optional)")
	allCmd.Flags().UintVarP(&maxSize, "max-size", "m", 500, "Maximum size in megabytes (optional)")
	allCmd.Flags().UintVarP(&maxLayers, "max-layers", "y", 20, "Maximum number of layers (optional)")
	allCmd.Flags().StringVarP(&allowedPorts, "allowed-ports", "p", "", "Comma-separated list of allowed ports or @<file> with JSON or YAML array (optional)")
	allCmd.Flags().StringVarP(&registryPolicy, "registry-policy", "r", "", "Registry policy file (JSON or YAML)")
	allCmd.Flags().StringVarP(&secretsPolicy, "secrets-policy", "s", "", "Path to secrets policy file (JSON or YAML) (optional)")
	allCmd.Flags().BoolVar(&skipEnvVars, "skip-env-vars", false, "Skip environment variable checks in secrets detection (optional)")
	allCmd.Flags().BoolVar(&skipFiles, "skip-files", false, "Skip file system checks in secrets detection (optional)")
	allCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop on first check failure (optional)")
}

func parseSkipList(skip string) (map[string]bool, error) {
	if skip == "" {
		return nil, nil
	}

	validNames := make(map[string]bool)
	for _, name := range validCheckNames {
		validNames[name] = true
	}

	skipMap := make(map[string]bool)
	for part := range strings.SplitSeq(skip, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if !validNames[name] {
			return nil, fmt.Errorf("unknown check name %q, valid names are: %s", name, strings.Join(validCheckNames, ", "))
		}
		skipMap[name] = true
	}

	return skipMap, nil
}

func loadAllConfig(path string) (*allConfig, error) {
	data, err := fileutil.ReadSecureFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg allConfig
	if err := fileutil.UnmarshalConfigFile(data, &cfg, path); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// applyConfigValues applies configuration file values to package-level variables,
// but only for flags that were not explicitly set on the CLI.
func applyConfigValues(cmd *cobra.Command, cfg *allConfig) {
	applyAgeConfig(cmd, cfg.Checks.Age)
	applySizeConfig(cmd, cfg.Checks.Size)
	applyPortsConfig(cmd, cfg.Checks.Ports)
	applyRegistryConfig(cmd, cfg.Checks.Registry)
	applySecretsConfig(cmd, cfg.Checks.Secrets)
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
		allowedPorts = formatAllowedPorts(cfg.AllowedPorts)
	}
}

func applyRegistryConfig(cmd *cobra.Command, cfg *registryCheckConfig) {
	if cfg != nil && cfg.RegistryPolicy != nil && !cmd.Flags().Changed("registry-policy") {
		registryPolicy = *cfg.RegistryPolicy
	}
}

func applySecretsConfig(cmd *cobra.Command, cfg *secretsCheckConfig) {
	if cfg == nil {
		return
	}
	if cfg.SecretsPolicy != nil && !cmd.Flags().Changed("secrets-policy") {
		secretsPolicy = *cfg.SecretsPolicy
	}
	if cfg.SkipEnvVars != nil && !cmd.Flags().Changed("skip-env-vars") {
		skipEnvVars = *cfg.SkipEnvVars
	}
	if cfg.SkipFiles != nil && !cmd.Flags().Changed("skip-files") {
		skipFiles = *cfg.SkipFiles
	}
}

// formatAllowedPorts converts the allowed-ports config value to a comma-separated string.
func formatAllowedPorts(v any) string {
	switch ports := v.(type) {
	case []any:
		parts := make([]string, 0, len(ports))
		for _, p := range ports {
			parts = append(parts, fmt.Sprintf("%v", p))
		}
		return strings.Join(parts, ",")
	case string:
		return ports
	default:
		return fmt.Sprintf("%v", v)
	}
}

// determineChecks decides which checks to run based on config and skip list.
func determineChecks(cfg *allConfig, skipMap map[string]bool) []checkRunner {
	var checks []checkRunner

	type checkDef struct {
		name    string
		enabled bool
		runFunc func(string) (*output.CheckResult, error)
	}

	var defs []checkDef

	if cfg != nil {
		// With config: only run checks present in the config
		defs = []checkDef{
			{"age", cfg.Checks.Age != nil, runAge},
			{"size", cfg.Checks.Size != nil, runSize},
			{"ports", cfg.Checks.Ports != nil, runPortsForAll},
			{"registry", cfg.Checks.Registry != nil, runRegistry},
			{"root-user", cfg.Checks.RootUser != nil, runRootUser},
			{"secrets", cfg.Checks.Secrets != nil, runSecrets},
		}
	} else {
		// Without config: run all checks
		defs = []checkDef{
			{"age", true, runAge},
			{"size", true, runSize},
			{"ports", true, runPortsForAll},
			{"registry", true, runRegistry},
			{"root-user", true, runRootUser},
			{"secrets", true, runSecrets},
		}
	}

	for _, def := range defs {
		if def.enabled && !skipMap[def.name] {
			checks = append(checks, checkRunner{name: def.name, run: def.runFunc})
		}
	}

	return checks
}

// runPortsForAll wraps runPorts to parse allowed ports first.
func runPortsForAll(imageName string) (*output.CheckResult, error) {
	var err error
	allowedPortsList, err = parseAllowedPorts()
	if err != nil {
		return nil, fmt.Errorf("invalid allowed ports: %w", err)
	}
	return runPorts(imageName)
}

func runAll(cmd *cobra.Command, imageName string) error {
	skipMap, err := parseSkipList(skipChecks)
	if err != nil {
		return err
	}

	var cfg *allConfig
	if configFile != "" {
		cfg, err = loadAllConfig(configFile)
		if err != nil {
			return err
		}
		applyConfigValues(cmd, cfg)
	}

	checks := determineChecks(cfg, skipMap)

	// Validate that registry-policy is provided when registry check will run
	for _, c := range checks {
		if c.name == "registry" && registryPolicy == "" {
			return fmt.Errorf("--registry-policy is required when the registry check is enabled")
		}
	}

	if len(checks) == 0 {
		if OutputFmt == output.FormatJSON {
			skipped := skippedCheckNames(skipMap)
			allResult := output.AllResult{
				Image:  imageName,
				Passed: true,
				Checks: []output.CheckResult{},
				Summary: output.Summary{
					Total:   0,
					Skipped: skipped,
				},
			}
			return output.RenderJSON(os.Stdout, allResult)
		}
		fmt.Println("No checks to run")
		return nil
	}

	if OutputFmt == output.FormatText {
		fmt.Printf("Running %d checks on image %s\n\n", len(checks), imageName)
	}

	results := executeChecks(checks, imageName)

	if OutputFmt == output.FormatJSON {
		return renderAllJSON(imageName, results, skipMap)
	}

	return nil
}

// executeChecks runs each check, collects results, and updates the global Result.
func executeChecks(checks []checkRunner, imageName string) []output.CheckResult {
	var results []output.CheckResult

	for _, check := range checks {
		log.Debugf("Running check: %s", check.name)

		if OutputFmt == output.FormatText {
			fmt.Printf("=== %s ===\n", check.name)
		}

		result, err := check.run(imageName)
		if err != nil {
			log.Errorf("Check %s failed with error: %v", check.name, err)
			Result = ValidationFailed

			errResult := output.CheckResult{
				Check:   check.name,
				Image:   imageName,
				Passed:  false,
				Message: fmt.Sprintf("check failed with error: %v", err),
				Error:   err.Error(),
			}
			results = append(results, errResult)
		} else {
			results = append(results, *result)

			if OutputFmt == output.FormatText {
				if renderErr := renderResult(result); renderErr != nil {
					log.Errorf("Failed to render result for %s: %v", check.name, renderErr)
				}
			}

			if result.Passed {
				if Result != ValidationFailed {
					Result = ValidationSucceeded
				}
			} else {
				Result = ValidationFailed
			}
		}

		if OutputFmt == output.FormatText {
			fmt.Println()
		}

		if failFast && Result == ValidationFailed {
			break
		}
	}

	return results
}

// renderAllJSON renders the aggregated results as a single JSON object.
func renderAllJSON(imageName string, results []output.CheckResult, skipMap map[string]bool) error {
	skipped := skippedCheckNames(skipMap)
	var passed, failed, errored int
	for _, r := range results {
		switch {
		case r.Error != "":
			errored++
		case r.Passed:
			passed++
		default:
			failed++
		}
	}

	allResult := output.AllResult{
		Image:  imageName,
		Passed: Result != ValidationFailed,
		Checks: results,
		Summary: output.Summary{
			Total:   len(results),
			Passed:  passed,
			Failed:  failed,
			Errored: errored,
			Skipped: skipped,
		},
	}
	return output.RenderJSON(os.Stdout, allResult)
}

// skippedCheckNames returns the list of skipped check names from the skip map.
func skippedCheckNames(skipMap map[string]bool) []string {
	if len(skipMap) == 0 {
		return nil
	}
	var names []string
	for _, name := range validCheckNames {
		if skipMap[name] {
			names = append(names, name)
		}
	}
	return names
}
