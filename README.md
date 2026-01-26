# Check Image

Check Image is a Go-based CLI tool designed for validating container images. It ensures that images meet specific standards such as size, age, ports, and security configurations. This project follows Go conventions for command-line tools and is structured into `cmd` and `internal` directories.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Configuration Files](#configuration-files)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation

To build and install the project, run:

```bash
go install ./cmd/check-image
```

This will install the `check-image` binary to your `GOBIN` directory.

## Usage

After installation, you can run the CLI tool:

```bash
check-image --help
```

## Commands

The CLI supports various commands for validating container images. Each command is defined in the `cmd/check-image/commands` directory.

### Available Commands

#### `size`
Validates that the image size and layer count are within acceptable limits.

```bash
check-image size <image> --max-size <MB> --max-layers <count>
```

Options:
- `--max-size`: Maximum image size in MB (default: 500)
- `--max-layers`: Maximum number of layers (default: 20)

#### `age`
Validates that the image is not older than a specified number of days.

```bash
check-image age <image> --max-age <days>
```

Options:
- `--max-age`: Maximum image age in days (default: 90)

#### `registry`
Validates that the image registry is trusted based on a policy file.

```bash
check-image registry <image> --registry-policy <file>
```

Options:
- `--registry-policy`: Path to registry policy file (JSON or YAML, required)

Policy file supports either:
- `trusted-registries`: Allowlist of trusted registries
- `excluded-registries`: Blocklist of excluded registries

#### `ports`
Validates that the image does not expose unauthorized ports.

```bash
check-image ports <image> --allowed-ports <ports>
```

Options:
- `--allowed-ports`: Comma-separated list of allowed ports or `@<file>` with JSON/YAML array

#### `root-user`
Validates that the image runs as non-root user.

```bash
check-image root-user <image>
```

#### `secrets`
Validates that the image does not contain sensitive data (passwords, tokens, keys).

```bash
check-image secrets <image> [--secrets-policy <file>] [--skip-env-vars] [--skip-files]
```

Options:
- `--secrets-policy`: Path to secrets policy file (JSON or YAML, optional)
- `--skip-env-vars`: Skip environment variable checks
- `--skip-files`: Skip file system checks

The command scans:
- Environment variables for sensitive patterns (password, secret, token, key, etc.)
- Files across all image layers for common secret files (SSH keys, cloud credentials, password files, etc.)

### Global Flags

All commands support:
- `--log-level`: Set log level (trace, debug, info, warn, error, fatal, panic)

## Configuration Files

The `config/` directory contains sample configuration files that can be used as templates:

### Allowed Ports Files
- `config/allowed-ports.json` - Sample allowed ports configuration in JSON format
- `config/allowed-ports.yaml` - Sample allowed ports configuration in YAML format

Example usage:
```bash
check-image ports nginx:latest --allowed-ports @config/allowed-ports.json
```

### Registry Policy Files
- `config/registry-policy.json` - Sample registry trust policy in JSON format
- `config/registry-policy.yaml` - Sample registry trust policy in YAML format

Example usage:
```bash
check-image registry nginx:latest --registry-policy config/registry-policy.json
```

### Secrets Policy Files
- `config/secrets-policy.json` - Sample secrets detection policy in JSON format
- `config/secrets-policy.yaml` - Sample secrets detection policy in YAML format

Example usage:
```bash
check-image secrets nginx:latest --secrets-policy config/secrets-policy.json
```

## Development

### Project Structure

- `cmd/check-image/main.go`: The entry point of the application that initializes the CLI and executes commands.
- `cmd/check-image/commands/`: Contains individual command implementations using the `cobra` library.
- `internal/imageutil/imageutil.go`: Provides utilities for interacting with container images, such as fetching images from local or remote sources and retrieving image configurations.
- `internal/registry/policy.go`: Manages registry policies, including trusted and excluded registries.
- `internal/secrets/`: Handles secrets detection, including policy loading and scanning for sensitive data in environment variables and files.
- `config/`: Contains sample configuration files for registry policies, allowed ports, and secrets detection.
- `go.mod`: Defines the module and its dependencies.
- `go.sum`: Contains the checksums for module dependencies.

### External Dependencies

- `github.com/spf13/cobra`: For CLI command structure.
- `github.com/google/go-containerregistry`: For interacting with container registries.
- `gopkg.in/yaml.v3`: For parsing YAML configuration files.
- `github.com/sirupsen/logrus`: For logging.

## Testing

To run tests, use the following command:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please ensure that:

- All new commands and utilities are covered by tests.
- Code follows Go's idiomatic error handling practices.
- Variable names are meaningful and descriptive.
- Functions are small and focused on a single responsibility.
- Table-driven tests are used for functions with multiple input scenarios.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
