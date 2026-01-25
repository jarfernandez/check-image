# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Check Image is a Go CLI tool for validating container images against security and operational standards. It validates images for size, age, exposed ports, registry trust, and security configurations (like non-root users).

## Build and Test Commands

### Building
Use `go install` instead of `go build` to install the binary to `GOBIN`:
```bash
go install ./cmd/check-image
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/imageutil
go test ./internal/registry
```

### Running the CLI
```bash
# Show help
./check-image --help

# Set log level (trace, debug, info, warn, error, fatal, panic)
./check-image --log-level debug <command>
```

## Architecture

### Command Pattern
All validation commands follow a consistent pattern:
1. Commands are in `cmd/check-image/commands/` and use Cobra framework
2. Each command sets a global `Result` variable (`ValidationFailed`, `ValidationSucceeded`, or `ValidationSkipped`) defined in `root.go`
3. Commands should update `Result = ValidationFailed` when validation fails, and `Result = ValidationSucceeded` only when passing (preserving failures from previous checks)

### Image Retrieval Strategy
The `imageutil` package implements a fallback strategy:
- `GetImage()` tries local Docker daemon first, then falls back to remote registry
- `GetLocalImage()` retrieves from Docker daemon
- `GetRemoteImage()` fetches from remote registry with keychain authentication
- All functions use `github.com/google/go-containerregistry` for image operations

### Validation Commands

**size**: Validates image size and layer count
- Flags: `--max-size` (MB, default 500), `--max-layers` (default 20)
- Uses `GetRemoteImage()` directly (not the fallback pattern)

**age**: Validates image creation date
- Flags: `--max-age` (days, default 90)
- Reads `config.Created` timestamp from image config

**registry**: Validates image registry against a trust policy
- Flags: `--registry-policy` (required, JSON or YAML file)
- Policy format: specify either `trusted-registries` (allowlist) or `excluded-registries` (blocklist), but not both
- Allowlist mode: only registries in `trusted-registries` are allowed
- Blocklist mode: all registries except those in `excluded-registries` are allowed

**ports**: Validates exposed ports against an allowed list
- Flags: `--allowed-ports` (comma-separated list or `@file.json`/`@file.yaml`)
- File format: `{"allowed-ports": [80, 443]}`
- Parses ports from image config's `ExposedPorts` field (format: "8080/tcp")

**root-user**: Validates that image runs as non-root
- No flags
- Checks if `config.Config.User` is empty or "root"

### Configuration Files
Sample configuration files are in `config/`:
- `registry-policy.yaml` / `registry-policy.json`: Trusted registries list
- `allowed-ports.yaml` / `allowed-ports.json`: Allowed ports list

Both JSON and YAML formats are supported throughout the tool. Format detection is by file extension (`.yaml`, `.yml` for YAML, otherwise JSON).

### Registry Policy Logic
In `internal/registry/policy.go`:
- Policy must specify either `trusted-registries` or `excluded-registries`, not both
- Allowlist mode (trusted-registries): only registries in the list are allowed
- Blocklist mode (excluded-registries): all registries except those in the list are allowed

### Logging
- Uses `logrus` with stderr output
- Timestamps formatted as "2006-01-02 15:04:05"
- Colors disabled when not running in a terminal
- Set level via `--log-level` flag on any command

## Go Project Rules

### GitHub Integration
- Use the GitHub CLI (`gh`) for all interactions with GitHub (issues, pull requests, comments).
- Use Conventional Commits format for all commit messages (e.g., `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Do not add Claude as a co-author in commits (no `Co-Authored-By: Claude` lines).

### Go Best Practices

#### Project Structure
- Follow a standard Go project layout.
- Place the CLI entrypoint in `cmd/check-image/`.
- Keep `main.go` minimal and move logic to packages.
- Use `internal/` for non-public packages.
- Use `pkg/` only for reusable packages.

#### Package Design
- Keep packages small and focused.
- Avoid circular dependencies.
- Prefer flat package structures.

#### Naming
- Use clear and concise names.
- Exported identifiers must start with a capital letter.
- Unexported identifiers must start with a lowercase letter.
- Avoid stuttering in names.

#### Error Handling
- Handle all errors explicitly.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Return errors instead of panicking.
- Write error messages in lowercase and without punctuation.

#### CLI Rules
- Use a single CLI framework consistently.
- Validate user input early and fail fast.
- Print user-facing errors to `stderr`.
- Return proper exit codes (`0` for success, non-zero for failure).

#### Interfaces
- Define interfaces at the point of use.
- Keep interfaces small.
- Accept interfaces and return concrete types.

#### Testing
- Use table-driven tests when appropriate.
- Test behavior, not implementation.
- Use the standard `testing` package.

#### Formatting and Tooling
- Format code with `gofmt`.
- Run `go vet` and `golangci-lint`.
- Keep `go.mod` tidy.

#### Documentation
- Explain *why*, not *what*.
- Comment only exported identifiers and complex logic.
- Keep documentation aligned with the code.

#### Dependencies
- Minimize external dependencies.
- Prefer the Go standard library.
- Avoid unnecessary dependency bloat.