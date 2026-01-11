# Check Image

Check Image is a Go-based CLI tool designed for validating container images. It ensures that images meet specific standards such as size, age, ports, and security configurations. This project follows Go conventions for command-line tools and is structured into `cmd` and `internal` directories.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation

To build the project, run the following command:

```bash
go build -o check-image ./cmd/check-image
```

## Usage

After building, you can run the CLI tool:

```bash
./check-image --help
```

## Commands

The CLI supports various commands for validating container images. Each command is defined in the `cmd/check-image/commands` directory.

## Development

### Project Structure

- `cmd/check-image/main.go`: The entry point of the application that initializes the CLI and executes commands.
- `cmd/check-image/commands/`: Contains individual command implementations using the `cobra` library.
- `internal/imageutil/imageutil.go`: Provides utilities for interacting with container images, such as fetching images from local or remote sources and retrieving image configurations.
- `internal/registry/policy.go`: Manages registry policies, including trusted and excluded registries.
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
