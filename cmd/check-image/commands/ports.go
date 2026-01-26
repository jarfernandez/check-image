package commands

import (
	"check-image/internal/fileutil"
	"check-image/internal/imageutil"
	"fmt"
	"strconv"
	"strings"

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
The 'image' argument should be the name of a container image.`,
	Example: `  check-image ports nginx:latest --allowed-ports 80,443
  check-image ports nginx:latest --allowed-ports @allowed-ports.json
  check-image ports nginx:latest --allowed-ports @allowed-ports.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		allowedPortsList, err = parseAllowedPorts()
		if err != nil {
			return fmt.Errorf("invalid check ports arguments: %w", err)
		}

		log.Debugln("Allowed ports:", allowedPortsList)

		if err := runPorts(args[0]); err != nil {
			return fmt.Errorf("check ports operation failed: %w", err)
		}

		return nil
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

	if strings.HasPrefix(allowedPorts, "@") {
		path := strings.TrimPrefix(allowedPorts, "@")

		// Read file securely
		data, err := fileutil.ReadSecureFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read ports file: %w", err)
		}

		var portsFromFile allowedPortsFile

		// Unmarshal config file (JSON or YAML)
		if err := fileutil.UnmarshalConfigFile(data, &portsFromFile, path); err != nil {
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

func runPorts(imageName string) error {
	fmt.Printf("Checking ports of image %s\n", imageName)

	_, config, err := imageutil.GetImageAndConfig(imageName)
	if err != nil {
		return err
	}

	// Extract exposed ports from the image config
	exposedPorts := make([]int, 0)
	for portProtocol := range config.Config.ExposedPorts {
		// Port format is typically "8080/tcp" or "53/udp"
		parts := strings.Split(portProtocol, "/")
		if len(parts) > 0 {
			port, err := strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("error parsing port number '%s': %w", parts[0], err)
			}
			exposedPorts = append(exposedPorts, port)
		}
	}

	if len(exposedPorts) == 0 {
		fmt.Println("No ports are exposed in this image")
		if Result != ValidationFailed {
			Result = ValidationSucceeded
		}
		return nil
	}

	// Display exposed ports
	fmt.Println("Exposed ports:")
	for _, port := range exposedPorts {
		fmt.Printf("  - %d\n", port)
	}

	if len(allowedPortsList) == 0 {
		fmt.Println("No allowed ports were provided")
		Result = ValidationFailed
		return nil
	}

	// Check if all exposed ports are in the allowed list
	unauthorizedPorts := make([]int, 0)
	for _, exposedPort := range exposedPorts {
		isAllowed := false
		for _, allowedPort := range allowedPortsList {
			if exposedPort == allowedPort {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			unauthorizedPorts = append(unauthorizedPorts, exposedPort)
		}
	}

	if len(unauthorizedPorts) > 0 {
		fmt.Println("The following ports are not in the allowed list:")
		for _, port := range unauthorizedPorts {
			fmt.Printf("  - %d\n", port)
		}
	}

	SetValidationResult(
		len(unauthorizedPorts) == 0,
		"All exposed ports are in the allowed list",
		"", // Empty because we already printed the unauthorized ports above
	)

	return nil
}
