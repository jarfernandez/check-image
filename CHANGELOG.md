# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/jarfernandez/check-image/compare/v0.1.1...v0.2.0) (2026-02-05)


### Features

* Update module path to use full GitHub URL ([#3](https://github.com/jarfernandez/check-image/issues/3)) ([42a9f22](https://github.com/jarfernandez/check-image/commit/42a9f2202d8971634c1004ce79ad94f41aaee1c0))


### Bug Fixes

* Update release-please config to v4 manifest format ([#5](https://github.com/jarfernandez/check-image/issues/5)) ([ea52291](https://github.com/jarfernandez/check-image/commit/ea522915a5d7871bedc3eeff45ff5e6e49b4db5f))


### Documentation

* Update installation instructions with GitHub install method ([#4](https://github.com/jarfernandez/check-image/issues/4)) ([434d3ec](https://github.com/jarfernandez/check-image/commit/434d3ec5f1904767b715d5bc94e14e35b2648227))

## [0.1.1](https://github.com/jarfernandez/check-image/releases/tag/v0.1.1) (2026-02-02)

### Bug Fixes

* Add v prefix to version output in binaries ([324f08f](https://github.com/jarfernandez/check-image/commit/324f08f48630a89b265cb3562f85fe7c590bff88))

## [0.1.0](https://github.com/jarfernandez/check-image/releases/tag/v0.1.0) (2026-02-02)

### Features

* Initial release of check-image CLI tool
* Image validation commands: size, age, registry, ports, root-user, secrets
* Support for multiple image sources (OCI layout, OCI archive, Docker archive, registry)
* Comprehensive test coverage (87.6%)
* CI/CD pipeline with automated releases

### Build System

* Add GitHub Actions workflows for automated CI/CD
* Multi-platform builds (Linux, macOS, Windows on amd64/arm64)
* Automated releases with semantic versioning using release-please
* GoReleaser configuration for multi-platform binary distribution
