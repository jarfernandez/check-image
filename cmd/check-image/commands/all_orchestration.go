package commands

import (
	"fmt"
	"os"

	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// checkRunner represents a single check to be executed.
type checkRunner struct {
	name   string
	run    func(imageName string) (*output.CheckResult, error)
	render func(r *output.CheckResult) // text renderer; paired with run at definition time
}

var configFile string
var skipChecks string
var includeChecks string
var failFast bool

var allCmd = &cobra.Command{
	Use:   "all image",
	Short: "Run all validation checks on a container image",
	Long: `Run all validation checks on a container image at once.

By default, runs all checks (age, size, ports, registry, root-user, secrets, healthcheck, labels, entrypoint, platform).
Use --config to specify which checks to run and their parameters.
Use --include to run only specific checks.
Use --skip to skip specific checks.
Use --fail-fast to stop on the first check failure.

Note: --include and --skip are mutually exclusive.

Some checks require additional configuration: registry needs --registry-policy,
labels needs --labels-policy, and platform needs --allowed-platforms. These can
be provided via CLI flags or the --config file. If enabled but not configured,
they fail with an execution error.

Precedence rules:
  1. Without --config: all checks run with defaults, except those in --skip
  2. With --config: only checks present in the config file run, except those in --skip
  3. --include overrides config file check selection (runs only specified checks)
  4. CLI flags override config file values
  5. --include and --skip always take precedence over the config file

The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`,
	Example: `  check-image all nginx:latest --include age,size,root-user --max-age 30 --max-size 200
  check-image all nginx:latest --skip registry,secrets,labels,platform
  check-image all nginx:latest --allowed-platforms linux/amd64,linux/arm64 --skip registry,labels
  check-image all nginx:latest --config config/config.json
  check-image all nginx:latest -c config/config.yaml --max-age 20 --skip secrets
  check-image all oci:/path/to/layout:1.0 --include age,size,root-user,ports,healthcheck
  check-image all oci-archive:/path/to/image.tar:latest --skip ports,registry,secrets,labels,platform
  check-image all nginx:latest --fail-fast --skip registry --config config/config.yaml --output json
  cat config/config.json | check-image all nginx:latest --config -`,
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

	allCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file (JSON or YAML) (optional)")
	allCmd.Flags().StringVar(&skipChecks, "skip", "", "Comma-separated list of checks to skip (age, size, ports, registry, root-user, secrets, healthcheck, labels, entrypoint, platform) (optional)")
	allCmd.Flags().StringVar(&includeChecks, "include", "", "Comma-separated list of checks to run (age, size, ports, registry, root-user, secrets, healthcheck, labels, entrypoint, platform) (optional)")
	allCmd.Flags().UintVarP(&maxAge, "max-age", "a", 90, "Maximum age in days (optional)")
	allCmd.Flags().UintVarP(&maxSize, "max-size", "m", 500, "Maximum size in megabytes (optional)")
	allCmd.Flags().UintVarP(&maxLayers, "max-layers", "y", 20, "Maximum number of layers (optional)")
	allCmd.Flags().StringVarP(&allowedPorts, "allowed-ports", "p", "", "Comma-separated list of allowed ports or @<file> with JSON or YAML array (optional)")
	allCmd.Flags().StringVarP(&registryPolicy, "registry-policy", "r", "", "Registry policy file (JSON or YAML)")
	allCmd.Flags().StringVarP(&secretsPolicy, "secrets-policy", "s", "", "Secrets policy file (JSON or YAML) (optional)")
	allCmd.Flags().BoolVar(&skipEnvVars, "skip-env-vars", false, "Skip environment variable checks in secrets detection (optional)")
	allCmd.Flags().BoolVar(&skipFiles, "skip-files", false, "Skip file system checks in secrets detection (optional)")
	allCmd.Flags().StringVar(&labelsPolicy, "labels-policy", "", "Labels policy file (JSON or YAML)")
	allCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop on first check failure (optional)")
	allCmd.Flags().BoolVar(&allowShellForm, "allow-shell-form", false, "Allow shell form for entrypoint or cmd (optional)")
	allCmd.Flags().StringVar(&allowedPlatforms, "allowed-platforms", "", "Comma-separated list of allowed platforms or @<file> with JSON or YAML array")
}

// determineChecks decides which checks to run based on config, skip list, and include list.
func determineChecks(cfg *allConfig, skipMap map[string]bool, includeMap map[string]bool) []checkRunner {
	var checks []checkRunner

	type checkDef struct {
		name       string
		enabled    bool
		runFunc    func(string) (*output.CheckResult, error)
		renderFunc func(*output.CheckResult)
	}

	var defs []checkDef

	if cfg != nil {
		// With config: only run checks present in the config
		defs = []checkDef{
			{"age", cfg.Checks.Age != nil, runAge, renderAgeText},
			{"size", cfg.Checks.Size != nil, runSize, renderSizeText},
			{"ports", cfg.Checks.Ports != nil, runPortsForAll, renderPortsText},
			{"registry", cfg.Checks.Registry != nil, runRegistry, renderRegistryText},
			{"root-user", cfg.Checks.RootUser != nil, runRootUser, renderRootUserText},
			{"secrets", cfg.Checks.Secrets != nil, runSecrets, renderSecretsText},
			{"healthcheck", cfg.Checks.Healthcheck != nil, runHealthcheck, renderHealthcheckText},
			{"labels", cfg.Checks.Labels != nil, runLabels, renderLabelsText},
			{"entrypoint", cfg.Checks.Entrypoint != nil, runEntrypoint, renderEntrypointText},
			{"platform", cfg.Checks.Platform != nil, runPlatformForAll, renderPlatformText},
		}
	} else {
		// Without config: run all checks
		defs = []checkDef{
			{"age", true, runAge, renderAgeText},
			{"size", true, runSize, renderSizeText},
			{"ports", true, runPortsForAll, renderPortsText},
			{"registry", true, runRegistry, renderRegistryText},
			{"root-user", true, runRootUser, renderRootUserText},
			{"secrets", true, runSecrets, renderSecretsText},
			{"healthcheck", true, runHealthcheck, renderHealthcheckText},
			{"labels", true, runLabels, renderLabelsText},
			{"entrypoint", true, runEntrypoint, renderEntrypointText},
			{"platform", true, runPlatformForAll, renderPlatformText},
		}
	}

	for _, def := range defs {
		if includeMap != nil {
			// --include mode: only run checks explicitly listed
			if includeMap[def.name] {
				checks = append(checks, checkRunner{name: def.name, run: def.runFunc, render: def.renderFunc})
			}
		} else {
			// Default/skip mode: use config/default enablement minus skip
			if def.enabled && !skipMap[def.name] {
				checks = append(checks, checkRunner{name: def.name, run: def.runFunc, render: def.renderFunc})
			}
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

// runPlatformForAll wraps runPlatform to parse allowed platforms first.
func runPlatformForAll(imageName string) (*output.CheckResult, error) {
	var err error
	allowedPlatformsList, err = parseAllowedPlatforms()
	if err != nil {
		return nil, fmt.Errorf("invalid allowed platforms: %w", err)
	}
	return runPlatform(imageName)
}

func runAll(cmd *cobra.Command, imageName string) error {
	skipMap, err := parseCheckNameList(skipChecks)
	if err != nil {
		return err
	}

	includeMap, err := parseCheckNameList(includeChecks)
	if err != nil {
		return err
	}

	if skipMap != nil && includeMap != nil {
		return fmt.Errorf("--include and --skip are mutually exclusive, use only one")
	}

	var cfg *allConfig
	if configFile != "" {
		cfg, err = loadAllConfig(configFile)
		if err != nil {
			return err
		}
		cleanup := applyConfigValues(cmd, cfg)
		defer cleanup()
	}

	checks := determineChecks(cfg, skipMap, includeMap)

	if err := validateRequiredFlags(checks); err != nil {
		return err
	}

	if len(checks) == 0 {
		if OutputFmt == output.FormatJSON {
			skipped := nonRunningCheckNames(skipMap, includeMap)
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
		fmt.Println(headerStyle.Render(fmt.Sprintf("Running %d checks on image %s", len(checks), imageName)))
		fmt.Println()
	}

	results := executeChecks(checks, imageName)

	if OutputFmt == output.FormatJSON {
		return renderAllJSON(imageName, results, skipMap, includeMap)
	}

	return nil
}

// validateRequiredFlags checks that required flags are provided when their checks will run.
func validateRequiredFlags(checks []checkRunner) error {
	for _, c := range checks {
		if c.name == "registry" && registryPolicy == "" {
			return fmt.Errorf("--registry-policy is required when the registry check is enabled")
		}
		if c.name == "labels" && labelsPolicy == "" {
			return fmt.Errorf("--labels-policy is required when the labels check is enabled")
		}
		if c.name == "platform" && allowedPlatforms == "" {
			return fmt.Errorf("--allowed-platforms is required when the platform check is enabled")
		}
	}
	return nil
}

// executeChecks runs each check, collects results, and updates the global Result.
func executeChecks(checks []checkRunner, imageName string) []output.CheckResult {
	var results []output.CheckResult

	for _, check := range checks {
		log.Debugf("Running check: %s", check.name)

		if OutputFmt == output.FormatText {
			fmt.Println(sectionHeader(check.name))
		}

		result, err := check.run(imageName)
		if err != nil {
			log.Errorf("Check %s failed with error: %v", check.name, err)
			UpdateResult(ExecutionError)

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

			if OutputFmt == output.FormatText && check.render != nil {
				check.render(result)
			}

			if result.Passed {
				UpdateResult(ValidationSucceeded)
			} else {
				UpdateResult(ValidationFailed)
			}
		}

		if OutputFmt == output.FormatText {
			fmt.Println()
		}

		if failFast && (Result == ValidationFailed || Result == ExecutionError) {
			break
		}
	}

	return results
}

// renderAllJSON renders the aggregated results as a single JSON object.
func renderAllJSON(imageName string, results []output.CheckResult, skipMap map[string]bool, includeMap map[string]bool) error {
	skipped := nonRunningCheckNames(skipMap, includeMap)
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
		Passed: Result != ValidationFailed && Result != ExecutionError,
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

// nonRunningCheckNames returns the list of check names that did not run.
func nonRunningCheckNames(skipMap map[string]bool, includeMap map[string]bool) []string {
	if includeMap != nil {
		var names []string
		for _, name := range validCheckNames {
			if !includeMap[name] {
				names = append(names, name)
			}
		}
		if len(names) == 0 {
			return nil
		}
		return names
	}

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
