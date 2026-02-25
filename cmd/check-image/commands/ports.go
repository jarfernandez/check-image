package commands

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jarfernandez/check-image/internal/fileutil"
	"github.com/jarfernandez/check-image/internal/imageutil"
	"github.com/jarfernandez/check-image/internal/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type allowedPortsFile struct {
	AllowedPorts []int `json:"allowed-ports" yaml:"allowed-ports"`
}

var (
	allowedPorts     string
	allowedPortsList []int
)

var portsCmd = &cobra.Command{
	Use:   "ports image",
	Short: "Validate that the image does not expose unauthorized ports",
	Long: `Validate that the image does not expose unauthorized ports.

` + imageArgFormatsDoc,
	Example: `  check-image ports nginx:latest --allowed-ports 80,443
  check-image ports nginx:latest --allowed-ports @allowed-ports.json
  check-image ports nginx:latest --allowed-ports @allowed-ports.yaml
  check-image ports oci:/path/to/layout:1.0 --allowed-ports 8080,8443
  check-image ports oci-archive:/path/to/image.tar:latest --allowed-ports @allowed-ports.json
  check-image ports docker-archive:/path/to/image.tar:tag --allowed-ports 80,443
  cat allowed-ports.json | check-image ports nginx:latest --allowed-ports @-`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		allowedPortsList, err = parseAllowedPorts()
		if err != nil {
			return fmt.Errorf("invalid check ports arguments: %w", err)
		}

		log.Debugln("Allowed ports:", allowedPortsList)

		return runCheckCmd("ports", runPorts, args[0])
	},
}

func init() {
	rootCmd.AddCommand(portsCmd)
	portsCmd.Flags().StringVarP(&allowedPorts, "allowed-ports", "p", "", "Comma-separated list of allowed ports or @<file> with JSON or YAML array (optional)")
}

func parseAllowedPorts() ([]int, error) {
	if allowedPorts == "" {
		return nil, nil
	}

	if after, ok := strings.CutPrefix(allowedPorts, "@"); ok {
		path := after

		// Read file or stdin
		data, err := fileutil.ReadFileOrStdin(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read ports file: %w", err)
		}

		var portsFromFile allowedPortsFile

		// Unmarshal config data (JSON or YAML)
		if err := fileutil.UnmarshalConfigData(data, &portsFromFile, path); err != nil {
			return nil, err
		}

		return portsFromFile.AllowedPorts, nil
	}

	parts := strings.Split(allowedPorts, ",")
	var ports []int
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		port, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid port '%s': %w", trimmed, err)
		}
		ports = append(ports, port)
	}

	return ports, nil
}

func runPorts(imageName string) (*output.CheckResult, error) {
	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return nil, err
	}

	// Extract exposed ports from the image config
	exposedPorts := make([]int, 0)
	for portProtocol := range config.Config.ExposedPorts {
		// Port format is typically "8080/tcp" or "53/udp"
		parts := strings.Split(portProtocol, "/")
		if len(parts) > 0 {
			port, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("error parsing port number '%s': %w", parts[0], err)
			}
			exposedPorts = append(exposedPorts, port)
		}
	}

	details := output.PortsDetails{
		ExposedPorts:      exposedPorts,
		AllowedPorts:      allowedPortsList,
		UnauthorizedPorts: nil,
	}

	if len(exposedPorts) == 0 {
		return &output.CheckResult{
			Check:   "ports",
			Image:   imageName,
			Passed:  true,
			Message: "No ports are exposed in this image",
			Details: details,
		}, nil
	}

	if len(allowedPortsList) == 0 {
		return &output.CheckResult{
			Check:   "ports",
			Image:   imageName,
			Passed:  false,
			Message: "No allowed ports were provided",
			Details: details,
		}, nil
	}

	// Check if all exposed ports are in the allowed list
	unauthorizedPorts := make([]int, 0)
	for _, exposedPort := range exposedPorts {
		isAllowed := slices.Contains(allowedPortsList, exposedPort)
		if !isAllowed {
			unauthorizedPorts = append(unauthorizedPorts, exposedPort)
		}
	}

	details.UnauthorizedPorts = unauthorizedPorts
	passed := len(unauthorizedPorts) == 0

	var msg string
	if passed {
		msg = "All exposed ports are in the allowed list"
	}

	return &output.CheckResult{
		Check:   "ports",
		Image:   imageName,
		Passed:  passed,
		Message: msg,
		Details: details,
	}, nil
}
