package commands

import (
	"check-image/internal/imageutil"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

// readSecureFile reads a file securely using os.OpenRoot to prevent directory traversal
func readSecureFile(path string) ([]byte, error) {
	// Clean the path to remove any .. or . elements
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// Get the directory containing the file
	dir := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	// Create a root-scoped filesystem
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot create root for directory: %w", err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil {
			log.Warnf("failed to close root: %v", closeErr)
		}
	}()

	// Check if file exists and is a regular file
	info, err := root.Stat(fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Open and read file using the scoped root
	file, err := root.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Warnf("failed to close file: %v", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return data, nil
}

func parseAllowedPorts() ([]int, error) {
	if allowedPorts == "" {
		return nil, nil
	}

	if strings.HasPrefix(allowedPorts, "@") {
		path := strings.TrimPrefix(allowedPorts, "@")

		// Read file securely
		data, err := readSecureFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read ports file: %w", err)
		}

		var portsFromFile allowedPortsFile

		// First, try to unmarshal as JSON
		if err := json.Unmarshal(data, &portsFromFile); err == nil {
			return portsFromFile.AllowedPorts, nil
		}

		// If JSON fails, try to unmarshal as YAML
		if err := yaml.Unmarshal(data, &portsFromFile); err == nil {
			return portsFromFile.AllowedPorts, nil
		}

		return nil, errors.New("invalid file format: must be valid JSON or YAML")
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
		Result = ValidationFailed
	} else {
		fmt.Println("All exposed ports are in the allowed list")
		if Result != ValidationFailed {
			Result = ValidationSucceeded
		}
	}

	return nil
}
