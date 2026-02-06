# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3](https://github.com/jarfernandez/check-image/compare/v0.2.2...v0.2.3) (2026-02-06)


### Bug Fixes

* Configure release-please to skip GitHub release creation ([#17](https://github.com/jarfernandez/check-image/issues/17)) ([6efecd9](https://github.com/jarfernandez/check-image/commit/6efecd95a0ee5b2e0ed7a7f3389f16bf094e3d1e))

## [0.2.2](https://github.com/jarfernandez/check-image/compare/v0.2.1...v0.2.2) (2026-02-06)


### Documentation

* Fix installation URLs to use version variable and update to v0.2.1 ([#15](https://github.com/jarfernandez/check-image/issues/15)) ([1edd80c](https://github.com/jarfernandez/check-image/commit/1edd80ce1e36089ae6de5151c319d7ba4c4b6ba0))

## [0.2.1](https://github.com/jarfernandez/check-image/compare/v0.2.0...v0.2.1) (2026-02-06)


### Bug Fixes

* Configure release-please to use simple tag format without component prefix ([#13](https://github.com/jarfernandez/check-image/issues/13)) ([f947711](https://github.com/jarfernandez/check-image/commit/f947711ff053131532a25d8f67ce91f8c259eaac))


### Documentation

* Add pre-built binary installation instructions and clarify version behavior ([#11](https://github.com/jarfernandez/check-image/issues/11)) ([c9fa85b](https://github.com/jarfernandez/check-image/commit/c9fa85bfd90db4d6812ea344b9829b6f2cdc17e9))

## [0.2.0](https://github.com/jarfernandez/check-image/compare/v0.1.1...v0.2.0) (2026-02-06)


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
