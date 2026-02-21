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
go test ./internal/secrets
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
2. Each `runX()` function returns `(*output.CheckResult, error)` — it never prints directly
3. The `RunE` handler in each command calls `renderResult()` to output text or JSON, then updates the global `Result` variable based on `result.Passed`
4. `Result` (`ValidationSkipped`, `ValidationSucceeded`, `ValidationFailed`, or `ExecutionError`) is defined in `root.go` and drives the exit code in `main.go`

### Exit Codes
- **Exit 0**: Validation succeeded (`ValidationSucceeded`) or no checks ran (`ValidationSkipped`)
- **Exit 1**: Validation failed (`ValidationFailed`) — the image did not pass one or more checks
- **Exit 2**: Execution error (`ExecutionError`) — the tool could not run properly (bad config, image not found, invalid arguments, etc.)

Priority ordering: `ExecutionError` > `ValidationFailed` > `ValidationSucceeded` > `ValidationSkipped`. If multiple results occur (e.g., in the `all` command), the highest-priority result determines the exit code.

The `UpdateResult()` helper in `root.go` enforces this precedence. The iota ordering of `ValidationResult` constants matches the priority ordering (higher value = higher priority).

### Output Format
- Controlled by the `--output`/`-o` global flag (values: `text` default, `json`)
- `internal/output/format.go`: Defines `Format` type, `ParseFormat()`, and `RenderJSON()` helper
- `internal/output/results.go`: Result structs (`CheckResult`, `AgeDetails`, `SizeDetails`, `PortsDetails`, `RegistryDetails`, `RootUserDetails`, `HealthcheckDetails`, `SecretsDetails`, `LabelsDetails`, `AllResult`, `Summary`, `VersionResult`)
- `cmd/check-image/commands/render.go`: Text renderers for each check; `renderResult()` dispatches to JSON or text based on `OutputFmt`
- In JSON mode, `main.go` suppresses the final "Validation succeeded/failed" text message (it's already in the JSON)

### Image Retrieval Strategy
The `imageutil` package implements a transport-aware retrieval strategy with fallback support:
- **Transport Detection**: `ParseReference()` detects transport prefix (e.g., `oci:`, `oci-archive:`, `docker-archive:`)
- **OCI Layout Support**: `GetOCILayoutImage()` loads images from OCI layout directories (supports both tag and digest references)
- **OCI Archive Support**: `GetOCIArchiveImage()` loads images from OCI tarball archives
  - Extracts tarball to temporary directory using `extractOCIArchive()`
  - Validates paths to prevent path traversal attacks
  - Enforces 5GB decompression limit to prevent decompression bombs
  - Supports gzipped (.gz, .tgz) and uncompressed tarballs
  - Then uses OCI layout loader on extracted content
- **Docker Archive Support**: `GetDockerArchiveImage()` loads images from Docker tarball archives created with `docker save`
  - Uses `tarball.ImageFromPath()` from go-containerregistry
  - Supports tag-based image selection within multi-image archives
- **Default Behavior** (no transport prefix): `GetImage()` tries local Docker daemon first, then falls back to remote registry
- **Explicit Transports**: When a transport prefix is specified, only that source is attempted (no fallback)
- `GetLocalImage()` retrieves from Docker daemon
- `GetRemoteImage()` fetches from remote registry with keychain authentication
- All functions use `github.com/google/go-containerregistry` for image operations

**Supported Transport Syntax** (Skopeo-compatible):
- `oci:/path/to/layout:tag` - OCI layout directory with tag
- `oci:/path/to/layout@sha256:abc...` - OCI layout directory with digest
- `oci-archive:/path.tar:tag` - OCI tarball archive with tag
- `oci-archive:/path.tar@sha256:abc...` - OCI tarball archive with digest
- `docker-archive:/path.tar:tag` - Docker tarball archive (saved with `docker save`)
- `nginx:latest` or `docker.io/nginx:latest` - Default behavior (daemon → registry)

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

**healthcheck**: Validates that the image has a healthcheck defined
- No flags
- Checks if `config.Config.Healthcheck` is not nil, has a non-empty test command, and test is not `["NONE"]` (explicitly disabled)
- Returns `HealthcheckDetails` with `HasHealthcheck` boolean

**secrets**: Validates that image does not contain sensitive data (passwords, tokens, keys)
- Flags: `--secrets-policy` (optional, JSON or YAML file), `--skip-env-vars`, `--skip-files`
- Scans environment variables for sensitive patterns (case-insensitive matching for keywords like password, secret, token, key, etc.)
- Scans all image layers for files matching sensitive patterns (SSH keys, cloud credentials, password files, etc.)
- Uses `DefaultFilePatterns` map in `internal/secrets/policy.go` as single source of truth for patterns and descriptions
- Policy supports `excluded-paths`, `excluded-env-vars`, and custom patterns
- Works out-of-the-box with sensible defaults when no policy file is provided

**entrypoint**: Validates that image has a startup command defined and uses exec form
- Flags: `--allow-shell-form` (allow shell form without failing; default: exec form required)
- Checks `config.Config.Entrypoint` and `config.Config.Cmd` — at least one must be non-empty
- Shell form detection: `Entrypoint[0]` or `Cmd[0]` is `/bin/sh` or `/bin/bash` and index 1 is `-c`
- Without `--allow-shell-form`: shell form causes FAIL
- With `--allow-shell-form`: shell form detected but PASS; `shell-form-allowed: true` in details, `exec-form: false`
- Returns `EntrypointDetails` with `has-entrypoint`, `exec-form`, `shell-form-allowed` (omitempty), `entrypoint` (omitempty), `cmd` (omitempty)
- `isShellFormCommand()` is the helper function for detecting shell form (used for both Entrypoint and Cmd)

**labels**: Validates that image has required labels (OCI annotations) with correct values
- Flags: `--labels-policy` (required, JSON or YAML file)
- Policy format: defines `required-labels` array with validation rules
- Three validation modes:
  - Existence check: label must be present (only `name` specified)
  - Exact value match: label value must exactly match (specify `name` and `value`)
  - Pattern match: label value must match regex (specify `name` and `pattern`)
- Reports missing labels and invalid label values with detailed error messages
- Supports both file paths and stdin input (`-`) for dynamic policy generation
- Inline policy support: policy can be embedded as object in all-checks config file

**all**: Runs all validation checks on a container image at once
- Flags: `--config` (`-c`, config file), `--include` (comma-separated checks to run), `--skip` (comma-separated checks to skip), `--fail-fast` (stop on first failure), plus all individual check flags (`--max-age`, `--max-size`, `--max-layers`, `--allowed-ports`, `--registry-policy`, `--labels-policy`, `--secrets-policy`, `--skip-env-vars`, `--skip-files`, `--allow-shell-form`)
- `--include` and `--skip` are mutually exclusive
- Precedence: CLI flags > config file values > defaults; `--include` and `--skip` always take precedence over config file check selection
- Without `--config`: runs all 9 checks with defaults (except skipped, or only included)
- With `--config`: only runs checks present in the config file (except skipped); `--include` overrides config check selection
- Uses `applyConfigValues()` with `cmd.Flags().Changed()` to respect CLI overrides
- Wrappers: `runPortsForAll()` calls `parseAllowedPorts()` before `runPorts()`; `runRegistryForAll()` skips gracefully when no `--registry-policy` is provided
- Continue-on-error (default): if a check returns an error, logs it, sets `Result = ValidationFailed`, and continues with the next check
- Fail-fast (`--fail-fast`): stops execution on the first check that fails (validation failure or execution error)

**version**: Shows the check-image version with full build information
- Flags: `--short` (print only the version number)
- Uses global `--output` flag for JSON support
- Calls `ver.GetBuildInfo()` which returns version, commit, build date, Go version, and platform
- `Version`, `Commit`, `BuildDate` are injected at build time via ldflags; `GoVersion` and `Platform` are read from the `runtime` package (no ldflags injection needed)
- Default (no `--short`) text output is a multi-line block; JSON uses `output.BuildInfoResult`
- With `--short`: text outputs just the version string; JSON uses `output.VersionResult` (single `version` field)
- GoReleaser template variables: `{{.ShortCommit}}` (7-char hash) and `{{.Date}}` (RFC3339 UTC) for `Commit` and `BuildDate`
- Docker build args: `VERSION`, `COMMIT`, `BUILD_DATE` (defaults: `dev`, `none`, `unknown`)

### Configuration Files
Sample configuration files are in `config/`:
- `allowed-ports.yaml` / `allowed-ports.json`: Allowed ports list
- `registry-policy.yaml` / `registry-policy.json`: Trusted registries list
- `labels-policy.yaml` / `labels-policy.json`: Required labels validation policy
- `config.yaml` / `config.json`: All-checks configuration (defines which checks to run and their parameters for the `all` command)
- `secrets-policy.yaml` / `secrets-policy.json`: Secrets detection policy with exclusions

Both JSON and YAML formats are supported throughout the tool. Format detection is by file extension (`.yaml`, `.yml` for YAML, otherwise JSON).

#### Stdin Support
All file arguments support reading from stdin using `-` as the path, enabling dynamic configuration from pipelines:
- `--registry-policy -` - Read registry policy from stdin
- `--labels-policy -` - Read labels policy from stdin
- `--secrets-policy -` - Read secrets policy from stdin
- `--allowed-ports @-` - Read allowed ports from stdin
- `--config -` - Read all-checks config from stdin

When reading from stdin, format is auto-detected by content (JSON starts with `{` or `[`, otherwise treated as YAML). The 10MB size limit prevents memory exhaustion.

Example usage:
```bash
# Registry policy from stdin
echo '{"trusted-registries": ["docker.io", "ghcr.io"]}' | \
  check-image registry nginx:latest --registry-policy -

# Secrets policy from pipeline
cat secrets-policy.yaml | check-image secrets nginx:latest --secrets-policy -

# All config from stdin
cat config.json | check-image all nginx:latest --config -
```

#### Inline Config
The `all` command config file supports embedding policies directly as objects instead of file paths:

```json
{
  "checks": {
    "registry": {
      "registry-policy": {
        "trusted-registries": ["docker.io", "ghcr.io"]
      }
    },
    "labels": {
      "labels-policy": {
        "required-labels": [
          {"name": "maintainer"},
          {"name": "org.opencontainers.image.version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"}
        ]
      }
    },
    "secrets": {
      "secrets-policy": {
        "check-env-vars": true,
        "check-files": false,
        "excluded-paths": ["/usr/share/**"]
      }
    },
    "ports": {
      "allowed-ports": [80, 443]
    }
  }
}
```

Both file paths (strings) and inline objects are supported. Inline objects are converted to temporary JSON files internally before being loaded by the policy loaders.

### Registry Policy Logic
In `internal/registry/policy.go`:
- Policy must specify either `trusted-registries` or `excluded-registries`, not both
- Allowlist mode (trusted-registries): only registries in the list are allowed
- Blocklist mode (excluded-registries): all registries except those in the list are allowed

### Secrets Detection Logic
In `internal/secrets/`:
- `policy.go`: Defines `DefaultFilePatterns` map as single source of truth for patterns and descriptions
- `detector.go`: Implements `CheckEnvironmentVariables()` and `CheckFilesInLayers()`
- Environment variable detection uses case-insensitive pattern matching against variable names
- File detection scans all layers (secrets in earlier layers remain in image history)
- Supports exclusion lists for both paths and environment variables to handle false positives
- Pattern descriptions consolidated in `DefaultFilePatterns` map to avoid duplication

### Logging
- Uses `logrus` with stderr output
- Timestamps formatted as "2006-01-02 15:04:05"
- Colors disabled when not running in a terminal
- Set level via `--log-level` flag on any command

### Docker Image
- **Dockerfile**: Multi-stage build in `Dockerfile` (root of repo)
- **Base image**: `gcr.io/distroless/static-debian12:nonroot` (provides CA certs for registry TLS, timezone data, non-root user UID 65532)
- **Build**: `CGO_ENABLED=0`, static binary, `-s -w` stripped, version injection via `ARG VERSION`
- **Registry**: `ghcr.io/jarfernandez/check-image`
- **Platforms**: linux/amd64, linux/arm64
- **Security**: Non-root user, no shell, no package manager, read-only filesystem compatible
- **Local build**: `docker build --build-arg VERSION=dev -t check-image .`
- **Container behavior**: Without Docker socket, `GetLocalImage()` fails silently and falls back to remote registry. This is the expected and recommended mode. Docker socket mounting is possible but grants host-level daemon access.

### GitHub Action
- **Type**: Composite action that downloads the check-image binary from GitHub Releases (like Trivy), giving native Docker daemon access
- **Files**: `action.yml` (action definition), `entrypoint.sh` (binary download + input-to-CLI mapping script)
- **How it works**: Downloads the check-image binary for the runner's OS/arch from GitHub Releases, then runs `check-image all` directly on the runner. Maps `RUNNER_OS`/`RUNNER_ARCH` to goreleaser archive names (e.g., `Linux`/`X64` → `linux`/`amd64`)
- **Command**: Always runs `check-image all` — individual check selection is done via the `checks` input (passed directly as `--include`) or the `skip` input (passed as `--skip`). The two inputs are mutually exclusive
- **Output capture**: stdout (JSON) is captured separately from stderr (logs). JSON goes to the `json` output, logs go to the workflow log
- **Step summary**: Generates `$GITHUB_STEP_SUMMARY` with results table, failed check details, and collapsible full JSON (uses `jq`, pre-installed on GitHub runners)
- **Exit codes**: Propagated directly — 0 (passed), 1 (validation failed), 2 (execution error)
- **Version sync**: The `version` input default in `action.yml` uses the `x-release-please-version` marker. Release-please's `extra-files` config (in `.github/release-please-config.json`) auto-updates this value on each release. README.md version references are also auto-updated via the same mechanism
- **Dogfooding**: The release workflow's docker job uses `uses: ./` to validate `check-image:scan` after Trivy. The docker job depends on goreleaser (`needs: [release-please, goreleaser]`) so the binary is available for download
- **Testing**: `.github/workflows/test-action.yml` tests the action using `uses: ./` against real images

### Release Pipeline
Single workflow in `.github/workflows/release-please.yml` with three chained jobs:

1. **release-please job**: Runs on every push to `main`. Creates/updates a release PR with changelog and version bump. When the release PR is merged, creates the git tag and GitHub release. Exports `releases_created` and `tag_name` as outputs.
2. **goreleaser job**: Depends on release-please. Only runs when `releases_created == 'true'`. Checks out the tag, builds binaries for linux/darwin/windows (amd64/arm64), and uploads them to the GitHub release via `mode: append`.
3. **docker job**: Depends on both release-please and goreleaser (needs binaries available for the check-image action). Only runs when `releases_created == 'true'`. Lints Dockerfile with hadolint, builds single-arch image for Trivy security scanning (CRITICAL/HIGH), validates image with check-image (dogfooding: size, root-user, ports, secrets), then builds and pushes multi-arch image (linux/amd64, linux/arm64) to GHCR with semver tags via `docker/metadata-action`.

All jobs must be in the same workflow because tags created by `GITHUB_TOKEN` do not trigger other workflows (GitHub limitation to prevent infinite loops).

**Permissions**: `contents: write`, `pull-requests: write`, `packages: write` (packages required for GHCR push).

**Configuration files**:
- `.github/release-please-config.json`: Release-please settings (release type, changelog sections)
- `.github/release-please-manifest.json`: Tracks the current version
- `.goreleaser.yml`: GoReleaser build configuration (platforms, archives). Changelog is disabled here; release-please handles it.

**Important**: Release-please tracks release PRs by the `autorelease: pending` label, not by title. When a release PR is successfully released, the label changes to `autorelease: tagged`. If this label gets stuck, release-please will abort with "untagged, merged release PRs outstanding".

**Commit types and releases**: Only `feat:`, `fix:`, `perf:`, and `refactor:` commits trigger version bumps. Use `ci:`, `chore:`, or `docs:` for changes that should not trigger releases (these are configured as `hidden: true` in the changelog sections).

## Go Project Rules

### GitHub Integration
- Use the GitHub CLI (`gh`) for all interactions with GitHub (issues, pull requests, comments).
- Always create a feature branch for new changes. Never commit directly to `main`.
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
- Use the standard `testing` package with `testify` for assertions.
- All tests must be deterministic, fast, and isolated (no Docker daemon, registry, or network access required).
- Use in-memory images and temporary directories for testing.
- Comprehensive unit tests cover all commands and internal packages with 90.6% overall coverage.

#### Formatting and Tooling
- Format code with `gofmt`.
- Run `go vet` and `golangci-lint`.
- Keep `go.mod` tidy.
- Pre-commit hooks automatically enforce these requirements before each commit.
- See `.golangci.yml` for linter configuration (balanced settings).
- Install hooks with: `pre-commit install && pre-commit install --hook-type commit-msg`.

#### Documentation
- Explain *why*, not *what*.
- Comment only exported identifiers and complex logic.
- Keep documentation aligned with the code.

#### Dependencies
- Minimize external dependencies.
- Prefer the Go standard library.
- Avoid unnecessary dependency bloat.
