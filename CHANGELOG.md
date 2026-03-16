# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0](https://github.com/jarfernandez/check-image/compare/v1.0.0...v1.0.0) (2026-03-16)


### ⚠ BREAKING CHANGES

* The `root-user` command has been removed. Use `user` instead, which provides the same basic non-root check plus additional policy-based validation (UID ranges, blocked users, numeric UID requirements).

### Features

* Add --include flag to the all command ([#70](https://github.com/jarfernandez/check-image/issues/70)) ([eb3e239](https://github.com/jarfernandez/check-image/commit/eb3e2394805422eed8c646eb36fa7339ecb4b047))
* Add `--color` flag with terminal color support via Lip Gloss ([#96](https://github.com/jarfernandez/check-image/issues/96)) ([e03eaa6](https://github.com/jarfernandez/check-image/commit/e03eaa653d5a1be0b21767298900ed0ef53180c0))
* add `--fail-fast` flag to `all` command ([#43](https://github.com/jarfernandez/check-image/issues/43)) ([52e4863](https://github.com/jarfernandez/check-image/commit/52e4863afb5d95509d132e73c2c3f1c938e51aa1))
* Add `--output`/`-o` flag with JSON support ([#45](https://github.com/jarfernandez/check-image/issues/45)) ([436389b](https://github.com/jarfernandez/check-image/commit/436389b62d673df874b0400987a2f173b7a5607d))
* add `all` command to run all validation checks at once ([#41](https://github.com/jarfernandez/check-image/issues/41)) ([8fac20e](https://github.com/jarfernandez/check-image/commit/8fac20e26aae87e4d76cf4d06d27b51a36d64a3e))
* Add `platform` validation command ([#84](https://github.com/jarfernandez/check-image/issues/84)) ([7e75ae3](https://github.com/jarfernandez/check-image/commit/7e75ae3e6026a60efe4368e8dd348e053326e645))
* Add `user` command with policy-based validation ([#197](https://github.com/jarfernandez/check-image/issues/197)) ([f37ae30](https://github.com/jarfernandez/check-image/commit/f37ae30e5a5359ff444d3b8e9326ef23be80370f))
* Add colored section separators to the `all` command output ([#103](https://github.com/jarfernandez/check-image/issues/103)) ([6e49bc5](https://github.com/jarfernandez/check-image/commit/6e49bc50f524571fad8a02b27948cbd42422eaa8))
* Add context cancellation checks in secrets layer scanning ([c2af061](https://github.com/jarfernandez/check-image/commit/c2af06143b77ee86b10d050828b376c11b885e0d))
* Add Docker support with multi-arch images and GHCR publishing ([#54](https://github.com/jarfernandez/check-image/issues/54)) ([0551aca](https://github.com/jarfernandez/check-image/commit/0551aca68ed08c746f51ff78b753ea85831a9663))
* Add entrypoint validation command ([#81](https://github.com/jarfernandez/check-image/issues/81)) ([55824d0](https://github.com/jarfernandez/check-image/commit/55824d086b38b99288b97a778ca388afc86b15a8))
* Add GitHub Action for container image validation ([#64](https://github.com/jarfernandez/check-image/issues/64)) ([35d719e](https://github.com/jarfernandez/check-image/commit/35d719ed57463048dd463a7ba80dfe0bc69ab57e))
* Add granular exit codes to distinguish validation failures from execution errors ([#49](https://github.com/jarfernandez/check-image/issues/49)) ([052b655](https://github.com/jarfernandez/check-image/commit/052b655861b5ed08ef3b8c98f7e47946cd80c5ad))
* Add healthcheck validation command ([#61](https://github.com/jarfernandez/check-image/issues/61)) ([6c8ab45](https://github.com/jarfernandez/check-image/commit/6c8ab454a4baa487a52a02cdfb0535a12dd2fcc4))
* Add Homebrew tap distribution ([#91](https://github.com/jarfernandez/check-image/issues/91)) ([95456fe](https://github.com/jarfernandez/check-image/commit/95456fe7e3a52469f46dadb49b5cef018e2ef842))
* Add labels validation command ([#57](https://github.com/jarfernandez/check-image/issues/57)) ([ea47eb8](https://github.com/jarfernandez/check-image/commit/ea47eb83ef0c8b144a32614e9bcacc71d0218725))
* Add private registry authentication support ([#88](https://github.com/jarfernandez/check-image/issues/88)) ([65c2e7f](https://github.com/jarfernandez/check-image/commit/65c2e7fd2f86810c5f36667d0bd8e3e295c44671))
* Add retry with exponential backoff for remote registry calls ([5488754](https://github.com/jarfernandez/check-image/commit/5488754c18c285cfeb9f70dac8af64f7b419aa73))
* Add stdin support and inline configuration for policies ([#51](https://github.com/jarfernandez/check-image/issues/51)) ([e3c5df8](https://github.com/jarfernandez/check-image/commit/e3c5df8e18f8aece1600d0d8c3e56ca7eb23e995))
* add version command ([64ecdb8](https://github.com/jarfernandez/check-image/commit/64ecdb8852dc171755f8b4ff72129b3af7f74395))
* **imageutil:** add local image retrieval ([010016c](https://github.com/jarfernandez/check-image/commit/010016c195a292fbda85329f1460f2d83ec07f40))
* **ports:** change allowed_ports to allowed-ports in config schema ([283df03](https://github.com/jarfernandez/check-image/commit/283df0351b4c2ed04a70d04c2c456912336b905d))
* **registry:** add command to validate trusted image registries ([0ca34c9](https://github.com/jarfernandez/check-image/commit/0ca34c9a4ccc4f53f7926387ff8cbff9934e3bb9))
* **registry:** change trusted_registries to trusted-registries and excluded_registries to excluded-registries in config schema ([eb8a9d4](https://github.com/jarfernandez/check-image/commit/eb8a9d404004e4244162037e6b82a68bb3608871))
* **registry:** enhance registry policy validation ([be77ebe](https://github.com/jarfernandez/check-image/commit/be77ebe8e6787411820bbdef6a99748f6f41efa4))
* Remove `root-user` command in favor of `user` command ([#201](https://github.com/jarfernandez/check-image/issues/201)) ([43ca533](https://github.com/jarfernandez/check-image/commit/43ca533e9323ddca6bc94ef9b3edf5d43a0824f9))
* **secrets:** add command to detect sensitive data in container images ([ad027cb](https://github.com/jarfernandez/check-image/commit/ad027cbc2d45152e18bb156a1a24f648088ac99b))
* support multiple image formats ([8e09297](https://github.com/jarfernandez/check-image/commit/8e0929776374536c435e4c400d4ed40f0e63c429))
* Update module path to use full GitHub URL ([#3](https://github.com/jarfernandez/check-image/issues/3)) ([9abb173](https://github.com/jarfernandez/check-image/commit/9abb173b1191f9e06b5865d53f8886fb6dc052ba))
* **version:** Add build info and --short flag ([#78](https://github.com/jarfernandez/check-image/issues/78)) ([988f982](https://github.com/jarfernandez/check-image/commit/988f982949caf395e41ac41a868ec0d3eda5748a))


### Bug Fixes

* Add file size limit to `ReadSecureFile` ([#186](https://github.com/jarfernandez/check-image/issues/186)) ([1ad5942](https://github.com/jarfernandez/check-image/commit/1ad5942b0d7efbac3d68a2676bb6928d7f4e805d))
* Add HTTP timeouts to remote registry transport ([657cf3f](https://github.com/jarfernandez/check-image/commit/657cf3f4dc74b17ca24c16a420b4f1ec86b11690))
* Add size limit on `--password-stdin` read ([a7b4801](https://github.com/jarfernandez/check-image/commit/a7b4801f47bd5fae6dcc50c3e9d6c04187a65065))
* add v prefix to version output in binaries ([324f08f](https://github.com/jarfernandez/check-image/commit/324f08f48630a89b265cb3562f85fe7c590bff88))
* Cap per-file `io.Copy` with `LimitReader` in `extractRegularFile` to prevent unbounded disk writes from lying tar headers ([#170](https://github.com/jarfernandez/check-image/issues/170)) ([5bda17b](https://github.com/jarfernandez/check-image/commit/5bda17bc5a5b47701336f5c43d4b93c8452c41f4))
* Chain GoReleaser as job within release-please workflow ([9e4dc84](https://github.com/jarfernandez/check-image/commit/9e4dc84587bc618b9738712ba3b61effe6505e25))
* Configure release-please to use simple tag format without component prefix ([cb3fa85](https://github.com/jarfernandez/check-image/commit/cb3fa856c0c04fcc7eb5aa956f2f1ffddba46735))
* Correct GoReleaser ldflags to use correct module path for version injection ([7d9bd49](https://github.com/jarfernandez/check-image/commit/7d9bd4929b5b0852dbe936b48a7fae1db0ff9a2c))
* Detect UID 0 as root in `root-user` check ([#168](https://github.com/jarfernandez/check-image/issues/168)) ([c67638f](https://github.com/jarfernandez/check-image/commit/c67638f68c5b50ac4a8d31291f3ff56ca520f3cf))
* Eliminate temp file leak and deduplicate inline policy formatters ([#98](https://github.com/jarfernandez/check-image/issues/98)) ([7976514](https://github.com/jarfernandez/check-image/commit/797651432b3f086270a1420d815533b40442d7ad))
* Embed render function in checkRunner and add default branch to renderResult ([#104](https://github.com/jarfernandez/check-image/issues/104)) ([07e6b04](https://github.com/jarfernandez/check-image/commit/07e6b04326b3f5035901f03cf81ae13a577e7455))
* Fix GoReleaser brews config and archive format deprecations ([#93](https://github.com/jarfernandez/check-image/issues/93)) ([0a1e71b](https://github.com/jarfernandez/check-image/commit/0a1e71bb223b4058d3b975654a748f5e0f0b1e73))
* Guard renderResult against nil Details panic on error results ([#100](https://github.com/jarfernandez/check-image/issues/100)) ([dbb5076](https://github.com/jarfernandez/check-image/commit/dbb5076ad6a94c0d17eb8e06f8b7b212d1f8a6d5))
* Improve error handling and remove dead code ([#127](https://github.com/jarfernandez/check-image/issues/127)) ([e3cc365](https://github.com/jarfernandez/check-image/commit/e3cc365d446d58eef4c78bdf3304f9daaf733d2a))
* Include labels in default checks help text ([#59](https://github.com/jarfernandez/check-image/issues/59)) ([5f85f52](https://github.com/jarfernandez/check-image/commit/5f85f524ab75d638316e6add88283c478a3bed49))
* Input validation improvements (registry transport, docker-archive tag, platform format) ([#187](https://github.com/jarfernandez/check-image/issues/187)) ([62491e3](https://github.com/jarfernandez/check-image/commit/62491e31abdb61500709faa2253d4f00380cf381))
* Pre-initialize Lip Gloss styles so `FailStyle` renders color even when `PersistentPreRunE` does not run ([#181](https://github.com/jarfernandez/check-image/issues/181)) ([d89d5a9](https://github.com/jarfernandez/check-image/commit/d89d5a95100d7f6e12e90e6a997530776dcef7d1))
* Prevent false positives in path-based secrets file pattern matching ([#139](https://github.com/jarfernandez/check-image/issues/139)) ([9822379](https://github.com/jarfernandez/check-image/commit/982237985cb3be0a37d138c7a7a331785a166c7b))
* Remove `oci-archive:` temp directory leak via cleanup function pattern ([#124](https://github.com/jarfernandez/check-image/issues/124)) ([51bc6e8](https://github.com/jarfernandez/check-image/commit/51bc6e8e8dc65223fec3d40fb98882a70c40445c))
* Remove push-to-main trigger from test-action workflow ([#66](https://github.com/jarfernandez/check-image/issues/66)) ([383defe](https://github.com/jarfernandez/check-image/commit/383defef610d2b45e1066527d1d0d5d9591f6fde))
* Reorder help text in all command for logical flow ([#72](https://github.com/jarfernandez/check-image/issues/72)) ([5e55d2c](https://github.com/jarfernandez/check-image/commit/5e55d2cb44a23f948fc52ebfbf3146a43c94d866))
* Replace magic +10 layer loop with sorted key iteration in secrets renderer ([#101](https://github.com/jarfernandez/check-image/issues/101)) ([15413c6](https://github.com/jarfernandez/check-image/commit/15413c6cde792b275d9dd125cab821f042e871c6))
* Replace string-matching HTTP status detection in `isRetryableError` with typed `transport.Error` assertion ([#177](https://github.com/jarfernandez/check-image/issues/177)) ([3cb73de](https://github.com/jarfernandez/check-image/commit/3cb73de48ab9a4dcfc501937d24e77cf776686c6))
* resolve security warnings detected by gosec ([54d7ae4](https://github.com/jarfernandez/check-image/commit/54d7ae45db268edbca3be3f4ce006167cdcff401))
* Sanitize image-controlled strings in debug log output to prevent log injection ([#189](https://github.com/jarfernandez/check-image/issues/189)) ([8f0dce0](https://github.com/jarfernandez/check-image/commit/8f0dce020808ef7038d9739bee54f6b81faf0d05))
* Scope `staticKeychain` credentials to target registry hostname ([#193](https://github.com/jarfernandez/check-image/issues/193)) ([8d595e1](https://github.com/jarfernandez/check-image/commit/8d595e1c5aee19383aa2aa69cc873b9f5187c435))
* **size:** retrieve local image with remote fallback ([6b8151f](https://github.com/jarfernandez/check-image/commit/6b8151f3598251e72f5553dac85232aec8ffea01))
* Skip `platform` check in Test Action workflow jobs that lack config ([8c227ab](https://github.com/jarfernandez/check-image/commit/8c227ab196b8b59403aa4e7e360ee36ce27f4973))
* Update release-please config to v4 manifest format ([#5](https://github.com/jarfernandez/check-image/issues/5)) ([dc54f9b](https://github.com/jarfernandez/check-image/commit/dc54f9b6a146dbde19b011abe2ff9e5ea328e198))
* Use `release-as` to target v1.0.0 instead of editing manifest ([#203](https://github.com/jarfernandez/check-image/issues/203)) ([e24f54a](https://github.com/jarfernandez/check-image/commit/e24f54a1585f037f0784d26801f780cdde461e50))
* Validate port range 1-65535 in `--allowed-ports` parsing ([#179](https://github.com/jarfernandez/check-image/issues/179)) ([37811d5](https://github.com/jarfernandez/check-image/commit/37811d539aa5fb7b911b6aa5a7cd253bedc2627d))
* Verify SHA-256 checksum of downloaded binary in GitHub Action ([a55bf13](https://github.com/jarfernandez/check-image/commit/a55bf1371affdc29ac42f1537432013e29c4e8a2))
* Warn at runtime when `--password` is used on the command line ([#183](https://github.com/jarfernandez/check-image/issues/183)) ([3236975](https://github.com/jarfernandez/check-image/commit/3236975f48021f9cbbef4f87dd2afaeffb30e9ea))
* Write to stdin pipe in goroutine to prevent deadlock on Windows ([abec527](https://github.com/jarfernandez/check-image/commit/abec52766525a4aa5737df0377db2264163d0230))


### Code Refactoring

* Add `context.Context` propagation and signal handling ([3bdd305](https://github.com/jarfernandez/check-image/commit/3bdd305ea6fb70cef1583eecd79128c34aee96fc))
* Adopt structured logging with `log.WithFields()` at high-value log sites ([#191](https://github.com/jarfernandez/check-image/issues/191)) ([05cbd5f](https://github.com/jarfernandez/check-image/commit/05cbd5f91c6d53af8f3bc7a04920f42c5e914d02))
* Centralise cleanup in `extractOCIArchive` via named return defer ([#137](https://github.com/jarfernandez/check-image/issues/137)) ([49eec76](https://github.com/jarfernandez/check-image/commit/49eec76046aa2d62c32b330c9e28ed7c43830d23))
* Collapse `checkDef` and `checkRunner` into a single type ([#132](https://github.com/jarfernandez/check-image/issues/132)) ([0cf265f](https://github.com/jarfernandez/check-image/commit/0cf265f57c16cb9a008f57a728d7e2e9a3ea7f18))
* eliminate duplicate code across commands and policies ([399e980](https://github.com/jarfernandez/check-image/commit/399e980fbce6bb1aba3abac3c6aa3f05e24ceacf))
* Extract `isDirectoryPattern` and `isGlobPattern` from `isPathExcluded` ([#140](https://github.com/jarfernandez/check-image/issues/140)) ([780c278](https://github.com/jarfernandez/check-image/commit/780c278ca9f2637c73b8444da484435d70b6015e))
* Extract `printSectionHeader`, `runSingleCheck`, `printSectionFooter` from `executeChecks` ([#138](https://github.com/jarfernandez/check-image/issues/138)) ([1b493a8](https://github.com/jarfernandez/check-image/commit/1b493a8de5519c3076bfff2776f0d7a16dd1522a))
* Extract `renderEmptyResult` from `runAll` ([#143](https://github.com/jarfernandez/check-image/issues/143)) ([7b16672](https://github.com/jarfernandez/check-image/commit/7b166725da6b44145b49ac20d287e3f0c8d8b4b5))
* Extract `resolveRegistryCredentials` from `PersistentPreRunE` for direct testability ([#135](https://github.com/jarfernandez/check-image/issues/135)) ([cd5d2ee](https://github.com/jarfernandez/check-image/commit/cd5d2ee7e523914453c00a3451bff1876640c0a5))
* Extract default flag values into named constants ([#112](https://github.com/jarfernandez/check-image/issues/112)) ([f0e3ce1](https://github.com/jarfernandez/check-image/commit/f0e3ce14f01ec4a37c765bf35cad892b4c705fa9))
* Extract repeated image-transport help into `imageArgFormatsDoc` constant ([#115](https://github.com/jarfernandez/check-image/issues/115)) ([0619227](https://github.com/jarfernandez/check-image/commit/06192276c4579b3622f3b05204ccdd8be9219596))
* Extract shared `applyInlinePolicy` helper from `applyRegistryConfig` and `applyLabelsConfig` ([#120](https://github.com/jarfernandez/check-image/issues/120)) ([45352f8](https://github.com/jarfernandez/check-image/commit/45352f804d1ba8ecdb86d8849915ead8badc9930))
* Extract shared `parseAllowedListFromFile` helper from `ports` and `platform` ([#117](https://github.com/jarfernandez/check-image/issues/117)) ([8afbb2a](https://github.com/jarfernandez/check-image/commit/8afbb2a3cfb2f0f27d33959af49b2800ec66dc77))
* Extract shared `RunE` body into `runCheckCmd` helper ([#116](https://github.com/jarfernandez/check-image/issues/116)) ([e16d64c](https://github.com/jarfernandez/check-image/commit/e16d64c358d0eb57bc87521a237d26f39f509da9))
* Extract shell interpreter literals as named constants in `entrypoint.go` ([#131](https://github.com/jarfernandez/check-image/issues/131)) ([990f9ac](https://github.com/jarfernandez/check-image/commit/990f9ac8edb41bada46bb73687e19fba67fc5173))
* Fix three naming readability issues in commands and imageutil ([#128](https://github.com/jarfernandez/check-image/issues/128)) ([9d2d084](https://github.com/jarfernandez/check-image/commit/9d2d0849376453bcb2061e9218ea4d38909231c0))
* Make error precedence in `applyConfigValues` explicit ([#141](https://github.com/jarfernandez/check-image/issues/141)) ([de14be4](https://github.com/jarfernandez/check-image/commit/de14be4c13b293537c642cc2148c6222c0faea76))
* Merge `formatAllowedPorts` and `formatAllowedPlatforms` into `formatAllowedList` ([#113](https://github.com/jarfernandez/check-image/issues/113)) ([c2debd6](https://github.com/jarfernandez/check-image/commit/c2debd6f63a082f22eca6a45535be8a34cc51a68))
* Pass output format as explicit parameter to render functions ([#150](https://github.com/jarfernandez/check-image/issues/150)) ([e040b5c](https://github.com/jarfernandez/check-image/commit/e040b5c31f8390f6de4f9df5a113f006d4b8bc12))
* Reduce gocyclo min-complexity and fix extractOCIArchive ([#47](https://github.com/jarfernandez/check-image/issues/47)) ([bc8a446](https://github.com/jarfernandez/check-image/commit/bc8a44618089e4a143b8ca7e5f66d0e73447b23f))
* Remove `keyStyle` and align text output across commands ([#199](https://github.com/jarfernandez/check-image/issues/199)) ([75df619](https://github.com/jarfernandez/check-image/commit/75df619bb5449ebe4d55393f01b45010de3c2746))
* Remove dead `LoadXPolicyFromObject` functions and their tests ([#121](https://github.com/jarfernandez/check-image/issues/121)) ([122c07f](https://github.com/jarfernandez/check-image/commit/122c07f5c43a0df0a7188e24c7f25fc5cbafbf60))
* Remove dead `UnmarshalConfigFile` function ([#110](https://github.com/jarfernandez/check-image/issues/110)) ([3d09679](https://github.com/jarfernandez/check-image/commit/3d09679364ded786cb0511a03ba765e8d4e0c6db))
* Remove redundant comments and fix import grouping ([#130](https://github.com/jarfernandez/check-image/issues/130)) ([7c9f1ca](https://github.com/jarfernandez/check-image/commit/7c9f1ca97b5cbf84978c2b2380fd80ab58f07a52))
* Rename `UpdateResult` parameter from `new` to `result` ([#125](https://github.com/jarfernandez/check-image/issues/125)) ([97e05c4](https://github.com/jarfernandez/check-image/commit/97e05c4d2b2c837d75f5427d03edf551b52d6649))
* Replace `renderResult` switch with map-based dispatch ([#146](https://github.com/jarfernandez/check-image/issues/146)) ([b22e35c](https://github.com/jarfernandez/check-image/commit/b22e35c2616eb3f4a1fba57a74a1c4caff6e71d9))
* Replace bare type assertions in render functions with `mustDetails` helper ([#126](https://github.com/jarfernandez/check-image/issues/126)) ([739bb99](https://github.com/jarfernandez/check-image/commit/739bb998b48ea6d7d8891a3686bbd17e9a7f03f1))
* Replace check name string literals with package-level constants ([#122](https://github.com/jarfernandez/check-image/issues/122)) ([8fd78ca](https://github.com/jarfernandez/check-image/commit/8fd78caa68010d0df9170b1e20299bee6656cc25))
* Return result struct from `Execute()` instead of reading globals ([#151](https://github.com/jarfernandez/check-image/issues/151)) ([a80080d](https://github.com/jarfernandez/check-image/commit/a80080d6e8f7ae06bc5f6639d31c23acc04c1d71))
* Sort default file patterns in `GetFilePatterns` for deterministic output ([#133](https://github.com/jarfernandez/check-image/issues/133)) ([67a2d9f](https://github.com/jarfernandez/check-image/commit/67a2d9f735ba2ff9eb459c0dc4fc39673588a7a7))
* Split all.go into all_config.go and all_orchestration.go ([#109](https://github.com/jarfernandez/check-image/issues/109)) ([e9a53e5](https://github.com/jarfernandez/check-image/commit/e9a53e548028286ab6b6f81a85528b48ae0a0d1a))
* Split pattern-matching helpers into `patterns.go` ([#147](https://github.com/jarfernandez/check-image/issues/147)) ([0a0387b](https://github.com/jarfernandez/check-image/commit/0a0387b2ec11f7f4dba3a923cc80136a299d71da))
* Unify duplicate `defs` slices in `determineChecks` ([#114](https://github.com/jarfernandez/check-image/issues/114)) ([69c760d](https://github.com/jarfernandez/check-image/commit/69c760d48d605bf15d16b1c371e1319c99563e13))
* Unify finding types between secrets and output packages ([#149](https://github.com/jarfernandez/check-image/issues/149)) ([c6bd8af](https://github.com/jarfernandez/check-image/commit/c6bd8afbd433438c1211f93d301e4b8d4b1ed0cd))
* Use `t.Cleanup` in `resetAllGlobals` for guaranteed test state cleanup ([#156](https://github.com/jarfernandez/check-image/issues/156)) ([4ac3075](https://github.com/jarfernandez/check-image/commit/4ac3075ecb45d3eb4946da2ade86cc2ed07e7ffd))
* Use explicit parameter struct in `buildCheckDefs` instead of package-level globals ([#153](https://github.com/jarfernandez/check-image/issues/153)) ([7ef98be](https://github.com/jarfernandez/check-image/commit/7ef98be2a68a802528abc40763d8d8a421f407b2))
* Use explicit parameters in `runX` functions instead of package-level globals ([#145](https://github.com/jarfernandez/check-image/issues/145)) ([100bcc2](https://github.com/jarfernandez/check-image/commit/100bcc24260c8d5b3a98cc978958ac0407c5cbf0))
* Use named cleanup variable in `inlinePolicyToTempFile` ([#144](https://github.com/jarfernandez/check-image/issues/144)) ([adffbe4](https://github.com/jarfernandez/check-image/commit/adffbe4da03be9cdb8cff47d34615dc3489fbd67))

## [1.0.0](https://github.com/jarfernandez/check-image/compare/v0.21.1...v1.0.0) (2026-03-16)


### ⚠ BREAKING CHANGES

* The `root-user` command has been removed. Use `user` instead, which provides the same basic non-root check plus additional policy-based validation (UID ranges, blocked users, numeric UID requirements).

### Features

* Remove `root-user` command in favor of `user` command ([#201](https://github.com/jarfernandez/check-image/issues/201)) ([43ca533](https://github.com/jarfernandez/check-image/commit/43ca533e9323ddca6bc94ef9b3edf5d43a0824f9))

## [0.21.1](https://github.com/jarfernandez/check-image/compare/v0.21.0...v0.21.1) (2026-03-15)


### Code Refactoring

* Remove `keyStyle` and align text output across commands ([#199](https://github.com/jarfernandez/check-image/issues/199)) ([75df619](https://github.com/jarfernandez/check-image/commit/75df619bb5449ebe4d55393f01b45010de3c2746))

## [0.21.0](https://github.com/jarfernandez/check-image/compare/v0.20.10...v0.21.0) (2026-03-15)


### Features

* Add `user` command with policy-based validation ([#197](https://github.com/jarfernandez/check-image/issues/197)) ([f37ae30](https://github.com/jarfernandez/check-image/commit/f37ae30e5a5359ff444d3b8e9326ef23be80370f))

## [0.20.10](https://github.com/jarfernandez/check-image/compare/v0.20.9...v0.20.10) (2026-03-15)


### Bug Fixes

* Scope `staticKeychain` credentials to target registry hostname ([#193](https://github.com/jarfernandez/check-image/issues/193)) ([8d595e1](https://github.com/jarfernandez/check-image/commit/8d595e1c5aee19383aa2aa69cc873b9f5187c435))
* Skip `platform` check in Test Action workflow jobs that lack config ([8c227ab](https://github.com/jarfernandez/check-image/commit/8c227ab196b8b59403aa4e7e360ee36ce27f4973))
* Verify SHA-256 checksum of downloaded binary in GitHub Action ([a55bf13](https://github.com/jarfernandez/check-image/commit/a55bf1371affdc29ac42f1537432013e29c4e8a2))

## [0.20.9](https://github.com/jarfernandez/check-image/compare/v0.20.8...v0.20.9) (2026-03-14)


### Bug Fixes

* Sanitize image-controlled strings in debug log output to prevent log injection ([#189](https://github.com/jarfernandez/check-image/issues/189)) ([8f0dce0](https://github.com/jarfernandez/check-image/commit/8f0dce020808ef7038d9739bee54f6b81faf0d05))


### Code Refactoring

* Adopt structured logging with `log.WithFields()` at high-value log sites ([#191](https://github.com/jarfernandez/check-image/issues/191)) ([05cbd5f](https://github.com/jarfernandez/check-image/commit/05cbd5f91c6d53af8f3bc7a04920f42c5e914d02))

## [0.20.8](https://github.com/jarfernandez/check-image/compare/v0.20.7...v0.20.8) (2026-03-14)


### Bug Fixes

* Input validation improvements (registry transport, docker-archive tag, platform format) ([#187](https://github.com/jarfernandez/check-image/issues/187)) ([62491e3](https://github.com/jarfernandez/check-image/commit/62491e31abdb61500709faa2253d4f00380cf381))

## [0.20.7](https://github.com/jarfernandez/check-image/compare/v0.20.6...v0.20.7) (2026-03-14)


### Bug Fixes

* Add file size limit to `ReadSecureFile` ([#186](https://github.com/jarfernandez/check-image/issues/186)) ([1ad5942](https://github.com/jarfernandez/check-image/commit/1ad5942b0d7efbac3d68a2676bb6928d7f4e805d))
* Warn at runtime when `--password` is used on the command line ([#183](https://github.com/jarfernandez/check-image/issues/183)) ([3236975](https://github.com/jarfernandez/check-image/commit/3236975f48021f9cbbef4f87dd2afaeffb30e9ea))

## [0.20.6](https://github.com/jarfernandez/check-image/compare/v0.20.5...v0.20.6) (2026-03-10)


### Bug Fixes

* Pre-initialize Lip Gloss styles so `FailStyle` renders color even when `PersistentPreRunE` does not run ([#181](https://github.com/jarfernandez/check-image/issues/181)) ([d89d5a9](https://github.com/jarfernandez/check-image/commit/d89d5a95100d7f6e12e90e6a997530776dcef7d1))

## [0.20.5](https://github.com/jarfernandez/check-image/compare/v0.20.4...v0.20.5) (2026-03-10)


### Bug Fixes

* Validate port range 1-65535 in `--allowed-ports` parsing ([#179](https://github.com/jarfernandez/check-image/issues/179)) ([37811d5](https://github.com/jarfernandez/check-image/commit/37811d539aa5fb7b911b6aa5a7cd253bedc2627d))

## [0.20.4](https://github.com/jarfernandez/check-image/compare/v0.20.3...v0.20.4) (2026-03-09)


### Bug Fixes

* Replace string-matching HTTP status detection in `isRetryableError` with typed `transport.Error` assertion ([#177](https://github.com/jarfernandez/check-image/issues/177)) ([3cb73de](https://github.com/jarfernandez/check-image/commit/3cb73de48ab9a4dcfc501937d24e77cf776686c6))

## [0.20.3](https://github.com/jarfernandez/check-image/compare/v0.20.2...v0.20.3) (2026-03-09)


### Bug Fixes

* Cap per-file `io.Copy` with `LimitReader` in `extractRegularFile` to prevent unbounded disk writes from lying tar headers ([#170](https://github.com/jarfernandez/check-image/issues/170)) ([5bda17b](https://github.com/jarfernandez/check-image/commit/5bda17bc5a5b47701336f5c43d4b93c8452c41f4))

## [0.20.2](https://github.com/jarfernandez/check-image/compare/v0.20.1...v0.20.2) (2026-03-09)


### Bug Fixes

* Detect UID 0 as root in `root-user` check ([#168](https://github.com/jarfernandez/check-image/issues/168)) ([c67638f](https://github.com/jarfernandez/check-image/commit/c67638f68c5b50ac4a8d31291f3ff56ca520f3cf))

## [0.20.1](https://github.com/jarfernandez/check-image/compare/v0.20.0...v0.20.1) (2026-03-08)


### Code Refactoring

* Use `t.Cleanup` in `resetAllGlobals` for guaranteed test state cleanup ([#156](https://github.com/jarfernandez/check-image/issues/156)) ([4ac3075](https://github.com/jarfernandez/check-image/commit/4ac3075ecb45d3eb4946da2ade86cc2ed07e7ffd))

## [0.20.0](https://github.com/jarfernandez/check-image/compare/v0.19.8...v0.20.0) (2026-03-08)


### Features

* Add context cancellation checks in secrets layer scanning ([c2af061](https://github.com/jarfernandez/check-image/commit/c2af06143b77ee86b10d050828b376c11b885e0d))
* Add retry with exponential backoff for remote registry calls ([5488754](https://github.com/jarfernandez/check-image/commit/5488754c18c285cfeb9f70dac8af64f7b419aa73))


### Bug Fixes

* Add HTTP timeouts to remote registry transport ([657cf3f](https://github.com/jarfernandez/check-image/commit/657cf3f4dc74b17ca24c16a420b4f1ec86b11690))
* Add size limit on `--password-stdin` read ([a7b4801](https://github.com/jarfernandez/check-image/commit/a7b4801f47bd5fae6dcc50c3e9d6c04187a65065))
* Write to stdin pipe in goroutine to prevent deadlock on Windows ([abec527](https://github.com/jarfernandez/check-image/commit/abec52766525a4aa5737df0377db2264163d0230))


### Code Refactoring

* Add `context.Context` propagation and signal handling ([3bdd305](https://github.com/jarfernandez/check-image/commit/3bdd305ea6fb70cef1583eecd79128c34aee96fc))

## [0.19.8](https://github.com/jarfernandez/check-image/compare/v0.19.7...v0.19.8) (2026-03-08)


### Code Refactoring

* Pass output format as explicit parameter to render functions ([#150](https://github.com/jarfernandez/check-image/issues/150)) ([e040b5c](https://github.com/jarfernandez/check-image/commit/e040b5c31f8390f6de4f9df5a113f006d4b8bc12))
* Return result struct from `Execute()` instead of reading globals ([#151](https://github.com/jarfernandez/check-image/issues/151)) ([a80080d](https://github.com/jarfernandez/check-image/commit/a80080d6e8f7ae06bc5f6639d31c23acc04c1d71))
* Unify finding types between secrets and output packages ([#149](https://github.com/jarfernandez/check-image/issues/149)) ([c6bd8af](https://github.com/jarfernandez/check-image/commit/c6bd8afbd433438c1211f93d301e4b8d4b1ed0cd))
* Use explicit parameter struct in `buildCheckDefs` instead of package-level globals ([#153](https://github.com/jarfernandez/check-image/issues/153)) ([7ef98be](https://github.com/jarfernandez/check-image/commit/7ef98be2a68a802528abc40763d8d8a421f407b2))

## [0.19.7](https://github.com/jarfernandez/check-image/compare/v0.19.6...v0.19.7) (2026-03-08)


### Code Refactoring

* Replace `renderResult` switch with map-based dispatch ([#146](https://github.com/jarfernandez/check-image/issues/146)) ([b22e35c](https://github.com/jarfernandez/check-image/commit/b22e35c2616eb3f4a1fba57a74a1c4caff6e71d9))
* Split pattern-matching helpers into `patterns.go` ([#147](https://github.com/jarfernandez/check-image/issues/147)) ([0a0387b](https://github.com/jarfernandez/check-image/commit/0a0387b2ec11f7f4dba3a923cc80136a299d71da))

## [0.19.6](https://github.com/jarfernandez/check-image/compare/v0.19.5...v0.19.6) (2026-03-07)


### Code Refactoring

* Extract `renderEmptyResult` from `runAll` ([#143](https://github.com/jarfernandez/check-image/issues/143)) ([7b16672](https://github.com/jarfernandez/check-image/commit/7b166725da6b44145b49ac20d287e3f0c8d8b4b5))
* Make error precedence in `applyConfigValues` explicit ([#141](https://github.com/jarfernandez/check-image/issues/141)) ([de14be4](https://github.com/jarfernandez/check-image/commit/de14be4c13b293537c642cc2148c6222c0faea76))
* Use explicit parameters in `runX` functions instead of package-level globals ([#145](https://github.com/jarfernandez/check-image/issues/145)) ([100bcc2](https://github.com/jarfernandez/check-image/commit/100bcc24260c8d5b3a98cc978958ac0407c5cbf0))
* Use named cleanup variable in `inlinePolicyToTempFile` ([#144](https://github.com/jarfernandez/check-image/issues/144)) ([adffbe4](https://github.com/jarfernandez/check-image/commit/adffbe4da03be9cdb8cff47d34615dc3489fbd67))

## [0.19.5](https://github.com/jarfernandez/check-image/compare/v0.19.4...v0.19.5) (2026-03-06)


### Bug Fixes

* Prevent false positives in path-based secrets file pattern matching ([#139](https://github.com/jarfernandez/check-image/issues/139)) ([9822379](https://github.com/jarfernandez/check-image/commit/982237985cb3be0a37d138c7a7a331785a166c7b))


### Code Refactoring

* Centralise cleanup in `extractOCIArchive` via named return defer ([#137](https://github.com/jarfernandez/check-image/issues/137)) ([49eec76](https://github.com/jarfernandez/check-image/commit/49eec76046aa2d62c32b330c9e28ed7c43830d23))
* Extract `isDirectoryPattern` and `isGlobPattern` from `isPathExcluded` ([#140](https://github.com/jarfernandez/check-image/issues/140)) ([780c278](https://github.com/jarfernandez/check-image/commit/780c278ca9f2637c73b8444da484435d70b6015e))
* Extract `printSectionHeader`, `runSingleCheck`, `printSectionFooter` from `executeChecks` ([#138](https://github.com/jarfernandez/check-image/issues/138)) ([1b493a8](https://github.com/jarfernandez/check-image/commit/1b493a8de5519c3076bfff2776f0d7a16dd1522a))
* Extract `resolveRegistryCredentials` from `PersistentPreRunE` for direct testability ([#135](https://github.com/jarfernandez/check-image/issues/135)) ([cd5d2ee](https://github.com/jarfernandez/check-image/commit/cd5d2ee7e523914453c00a3451bff1876640c0a5))

## [0.19.4](https://github.com/jarfernandez/check-image/compare/v0.19.3...v0.19.4) (2026-03-01)


### Bug Fixes

* Improve error handling and remove dead code ([#127](https://github.com/jarfernandez/check-image/issues/127)) ([e3cc365](https://github.com/jarfernandez/check-image/commit/e3cc365d446d58eef4c78bdf3304f9daaf733d2a))
* Remove `oci-archive:` temp directory leak via cleanup function pattern ([#124](https://github.com/jarfernandez/check-image/issues/124)) ([51bc6e8](https://github.com/jarfernandez/check-image/commit/51bc6e8e8dc65223fec3d40fb98882a70c40445c))


### Code Refactoring

* Collapse `checkDef` and `checkRunner` into a single type ([#132](https://github.com/jarfernandez/check-image/issues/132)) ([0cf265f](https://github.com/jarfernandez/check-image/commit/0cf265f57c16cb9a008f57a728d7e2e9a3ea7f18))
* Extract shell interpreter literals as named constants in `entrypoint.go` ([#131](https://github.com/jarfernandez/check-image/issues/131)) ([990f9ac](https://github.com/jarfernandez/check-image/commit/990f9ac8edb41bada46bb73687e19fba67fc5173))
* Fix three naming readability issues in commands and imageutil ([#128](https://github.com/jarfernandez/check-image/issues/128)) ([9d2d084](https://github.com/jarfernandez/check-image/commit/9d2d0849376453bcb2061e9218ea4d38909231c0))
* Remove redundant comments and fix import grouping ([#130](https://github.com/jarfernandez/check-image/issues/130)) ([7c9f1ca](https://github.com/jarfernandez/check-image/commit/7c9f1ca97b5cbf84978c2b2380fd80ab58f07a52))
* Rename `UpdateResult` parameter from `new` to `result` ([#125](https://github.com/jarfernandez/check-image/issues/125)) ([97e05c4](https://github.com/jarfernandez/check-image/commit/97e05c4d2b2c837d75f5427d03edf551b52d6649))
* Replace bare type assertions in render functions with `mustDetails` helper ([#126](https://github.com/jarfernandez/check-image/issues/126)) ([739bb99](https://github.com/jarfernandez/check-image/commit/739bb998b48ea6d7d8891a3686bbd17e9a7f03f1))
* Replace check name string literals with package-level constants ([#122](https://github.com/jarfernandez/check-image/issues/122)) ([8fd78ca](https://github.com/jarfernandez/check-image/commit/8fd78caa68010d0df9170b1e20299bee6656cc25))
* Sort default file patterns in `GetFilePatterns` for deterministic output ([#133](https://github.com/jarfernandez/check-image/issues/133)) ([67a2d9f](https://github.com/jarfernandez/check-image/commit/67a2d9f735ba2ff9eb459c0dc4fc39673588a7a7))

## [0.19.3](https://github.com/jarfernandez/check-image/compare/v0.19.2...v0.19.3) (2026-02-28)


### Code Refactoring

* Extract shared `applyInlinePolicy` helper from `applyRegistryConfig` and `applyLabelsConfig` ([#120](https://github.com/jarfernandez/check-image/issues/120)) ([45352f8](https://github.com/jarfernandez/check-image/commit/45352f804d1ba8ecdb86d8849915ead8badc9930))
* Extract shared `parseAllowedListFromFile` helper from `ports` and `platform` ([#117](https://github.com/jarfernandez/check-image/issues/117)) ([8afbb2a](https://github.com/jarfernandez/check-image/commit/8afbb2a3cfb2f0f27d33959af49b2800ec66dc77))
* Remove dead `LoadXPolicyFromObject` functions and their tests ([#121](https://github.com/jarfernandez/check-image/issues/121)) ([122c07f](https://github.com/jarfernandez/check-image/commit/122c07f5c43a0df0a7188e24c7f25fc5cbafbf60))

## [0.19.2](https://github.com/jarfernandez/check-image/compare/v0.19.1...v0.19.2) (2026-02-25)


### Code Refactoring

* Extract default flag values into named constants ([#112](https://github.com/jarfernandez/check-image/issues/112)) ([f0e3ce1](https://github.com/jarfernandez/check-image/commit/f0e3ce14f01ec4a37c765bf35cad892b4c705fa9))
* Extract repeated image-transport help into `imageArgFormatsDoc` constant ([#115](https://github.com/jarfernandez/check-image/issues/115)) ([0619227](https://github.com/jarfernandez/check-image/commit/06192276c4579b3622f3b05204ccdd8be9219596))
* Extract shared `RunE` body into `runCheckCmd` helper ([#116](https://github.com/jarfernandez/check-image/issues/116)) ([e16d64c](https://github.com/jarfernandez/check-image/commit/e16d64c358d0eb57bc87521a237d26f39f509da9))
* Merge `formatAllowedPorts` and `formatAllowedPlatforms` into `formatAllowedList` ([#113](https://github.com/jarfernandez/check-image/issues/113)) ([c2debd6](https://github.com/jarfernandez/check-image/commit/c2debd6f63a082f22eca6a45535be8a34cc51a68))
* Remove dead `UnmarshalConfigFile` function ([#110](https://github.com/jarfernandez/check-image/issues/110)) ([3d09679](https://github.com/jarfernandez/check-image/commit/3d09679364ded786cb0511a03ba765e8d4e0c6db))
* Unify duplicate `defs` slices in `determineChecks` ([#114](https://github.com/jarfernandez/check-image/issues/114)) ([69c760d](https://github.com/jarfernandez/check-image/commit/69c760d48d605bf15d16b1c371e1319c99563e13))

## [0.19.1](https://github.com/jarfernandez/check-image/compare/v0.19.0...v0.19.1) (2026-02-24)


### Bug Fixes

* Embed `render` function in `checkRunner` and add `default` branch to `renderResult` ([#104](https://github.com/jarfernandez/check-image/issues/104)) ([07e6b04](https://github.com/jarfernandez/check-image/commit/07e6b04326b3f5035901f03cf81ae13a577e7455))


### Code Refactoring

* Split `all.go` into `all_config.go` and `all_orchestration.go` ([#109](https://github.com/jarfernandez/check-image/issues/109)) ([e9a53e5](https://github.com/jarfernandez/check-image/commit/e9a53e548028286ab6b6f81a85528b48ae0a0d1a))

## [0.19.0](https://github.com/jarfernandez/check-image/compare/v0.18.0...v0.19.0) (2026-02-23)


### Features

* Add colored section separators to the `all` command output ([#103](https://github.com/jarfernandez/check-image/issues/103)) ([6e49bc5](https://github.com/jarfernandez/check-image/commit/6e49bc50f524571fad8a02b27948cbd42422eaa8))


### Bug Fixes

* Eliminate temp file leak and deduplicate inline policy formatters ([#98](https://github.com/jarfernandez/check-image/issues/98)) ([7976514](https://github.com/jarfernandez/check-image/commit/797651432b3f086270a1420d815533b40442d7ad))
* Guard renderResult against nil Details panic on error results ([#100](https://github.com/jarfernandez/check-image/issues/100)) ([dbb5076](https://github.com/jarfernandez/check-image/commit/dbb5076ad6a94c0d17eb8e06f8b7b212d1f8a6d5))
* Replace magic +10 layer loop with sorted key iteration in secrets renderer ([#101](https://github.com/jarfernandez/check-image/issues/101)) ([15413c6](https://github.com/jarfernandez/check-image/commit/15413c6cde792b275d9dd125cab821f042e871c6))

## [0.18.0](https://github.com/jarfernandez/check-image/compare/v0.17.1...v0.18.0) (2026-02-22)


### Features

* Add `--color` flag with terminal color support via Lip Gloss ([#96](https://github.com/jarfernandez/check-image/issues/96)) ([e03eaa6](https://github.com/jarfernandez/check-image/commit/e03eaa653d5a1be0b21767298900ed0ef53180c0))

## [0.17.1](https://github.com/jarfernandez/check-image/compare/v0.17.0...v0.17.1) (2026-02-22)


### Bug Fixes

* Fix GoReleaser brews config and archive format deprecations ([#93](https://github.com/jarfernandez/check-image/issues/93)) ([0a1e71b](https://github.com/jarfernandez/check-image/commit/0a1e71bb223b4058d3b975654a748f5e0f0b1e73))

## [0.17.0](https://github.com/jarfernandez/check-image/compare/v0.16.0...v0.17.0) (2026-02-21)


### Features

* Add Homebrew tap distribution ([#91](https://github.com/jarfernandez/check-image/issues/91)) ([95456fe](https://github.com/jarfernandez/check-image/commit/95456fe7e3a52469f46dadb49b5cef018e2ef842))

## [0.16.0](https://github.com/jarfernandez/check-image/compare/v0.15.0...v0.16.0) (2026-02-21)


### Features

* Add private registry authentication support ([#88](https://github.com/jarfernandez/check-image/issues/88)) ([65c2e7f](https://github.com/jarfernandez/check-image/commit/65c2e7fd2f86810c5f36667d0bd8e3e295c44671))

## [0.15.0](https://github.com/jarfernandez/check-image/compare/v0.14.0...v0.15.0) (2026-02-21)


### Features

* Add `platform` validation command ([#84](https://github.com/jarfernandez/check-image/issues/84)) ([7e75ae3](https://github.com/jarfernandez/check-image/commit/7e75ae3e6026a60efe4368e8dd348e053326e645))

## [0.14.0](https://github.com/jarfernandez/check-image/compare/v0.13.0...v0.14.0) (2026-02-21)


### Features

* Add `entrypoint` validation command ([#81](https://github.com/jarfernandez/check-image/issues/81)) ([55824d0](https://github.com/jarfernandez/check-image/commit/55824d086b38b99288b97a778ca388afc86b15a8))

## [0.13.0](https://github.com/jarfernandez/check-image/compare/v0.12.1...v0.13.0) (2026-02-21)


### Features

* **version:** Add build info and `--short` flag ([#78](https://github.com/jarfernandez/check-image/issues/78)) ([988f982](https://github.com/jarfernandez/check-image/commit/988f982949caf395e41ac41a868ec0d3eda5748a))

## [0.12.1](https://github.com/jarfernandez/check-image/compare/v0.12.0...v0.12.1) (2026-02-16)


### Bug Fixes

* **all:** Reorder help text ([#72](https://github.com/jarfernandez/check-image/issues/72)) ([5e55d2c](https://github.com/jarfernandez/check-image/commit/5e55d2cb44a23f948fc52ebfbf3146a43c94d866))

## [0.12.0](https://github.com/jarfernandez/check-image/compare/v0.11.1...v0.12.0) (2026-02-16)


### Features

* **all:** Add `--include` flag ([#70](https://github.com/jarfernandez/check-image/issues/70)) ([eb3e239](https://github.com/jarfernandez/check-image/commit/eb3e2394805422eed8c646eb36fa7339ecb4b047))

## [0.11.1](https://github.com/jarfernandez/check-image/compare/v0.11.0...v0.11.1) (2026-02-15)


### Bug Fixes

* Remove `push-to-main` trigger from `test-action` workflow ([#66](https://github.com/jarfernandez/check-image/issues/66)) ([383defe](https://github.com/jarfernandez/check-image/commit/383defef610d2b45e1066527d1d0d5d9591f6fde))

## [0.11.0](https://github.com/jarfernandez/check-image/compare/v0.10.0...v0.11.0) (2026-02-15)


### Features

* Add GitHub Action for container image validation ([#64](https://github.com/jarfernandez/check-image/issues/64)) ([35d719e](https://github.com/jarfernandez/check-image/commit/35d719ed57463048dd463a7ba80dfe0bc69ab57e))

## [0.10.0](https://github.com/jarfernandez/check-image/compare/v0.9.1...v0.10.0) (2026-02-15)


### Features

* Add `healthcheck` validation command ([#61](https://github.com/jarfernandez/check-image/issues/61)) ([6c8ab45](https://github.com/jarfernandez/check-image/commit/6c8ab454a4baa487a52a02cdfb0535a12dd2fcc4))

## [0.9.1](https://github.com/jarfernandez/check-image/compare/v0.9.0...v0.9.1) (2026-02-15)


### Bug Fixes

* **all:** Include `labels` in default checks help text ([#59](https://github.com/jarfernandez/check-image/issues/59)) ([5f85f52](https://github.com/jarfernandez/check-image/commit/5f85f524ab75d638316e6add88283c478a3bed49))

## [0.9.0](https://github.com/jarfernandez/check-image/compare/v0.8.0...v0.9.0) (2026-02-15)


### Features

* Add `labels` validation command ([#57](https://github.com/jarfernandez/check-image/issues/57)) ([ea47eb8](https://github.com/jarfernandez/check-image/commit/ea47eb83ef0c8b144a32614e9bcacc71d0218725))

## [0.8.0](https://github.com/jarfernandez/check-image/compare/v0.7.0...v0.8.0) (2026-02-14)


### Features

* Add Docker support with multi-arch images and GHCR publishing ([#54](https://github.com/jarfernandez/check-image/issues/54)) ([0551aca](https://github.com/jarfernandez/check-image/commit/0551aca68ed08c746f51ff78b753ea85831a9663))

## [0.7.0](https://github.com/jarfernandez/check-image/compare/v0.6.0...v0.7.0) (2026-02-14)


### Features

* Add stdin support and inline configuration for policies ([#51](https://github.com/jarfernandez/check-image/issues/51)) ([e3c5df8](https://github.com/jarfernandez/check-image/commit/e3c5df8e18f8aece1600d0d8c3e56ca7eb23e995))

## [0.6.0](https://github.com/jarfernandez/check-image/compare/v0.5.1...v0.6.0) (2026-02-14)


### Features

* Add granular exit codes to distinguish validation failures from execution errors ([#49](https://github.com/jarfernandez/check-image/issues/49)) ([052b655](https://github.com/jarfernandez/check-image/commit/052b655861b5ed08ef3b8c98f7e47946cd80c5ad))

## [0.5.1](https://github.com/jarfernandez/check-image/compare/v0.5.0...v0.5.1) (2026-02-14)


### Code Refactoring

* Reduce gocyclo min-complexity and fix extractOCIArchive ([#47](https://github.com/jarfernandez/check-image/issues/47)) ([bc8a446](https://github.com/jarfernandez/check-image/commit/bc8a44618089e4a143b8ca7e5f66d0e73447b23f))

## [0.5.0](https://github.com/jarfernandez/check-image/compare/v0.4.0...v0.5.0) (2026-02-14)


### Features

* Add `--output`/`-o` flag with JSON support ([#45](https://github.com/jarfernandez/check-image/issues/45)) ([436389b](https://github.com/jarfernandez/check-image/commit/436389b62d673df874b0400987a2f173b7a5607d))

## [0.4.0](https://github.com/jarfernandez/check-image/compare/v0.3.0...v0.4.0) (2026-02-11)


### Features

* **all:** Add `--fail-fast` flag ([#43](https://github.com/jarfernandez/check-image/issues/43)) ([52e4863](https://github.com/jarfernandez/check-image/commit/52e4863afb5d95509d132e73c2c3f1c938e51aa1))

## [0.3.0](https://github.com/jarfernandez/check-image/compare/v0.2.0...v0.3.0) (2026-02-10)


### Features

* Add `all` command to run all validation checks at once ([#41](https://github.com/jarfernandez/check-image/issues/41)) ([8fac20e](https://github.com/jarfernandez/check-image/commit/8fac20e26aae87e4d76cf4d06d27b51a36d64a3e))

## [0.2.0](https://github.com/jarfernandez/check-image/compare/v0.1.1...v0.2.0) (2026-02-09)


### Features

* Update module path to use full GitHub URL ([#3](https://github.com/jarfernandez/check-image/issues/3)) ([42a9f22](https://github.com/jarfernandez/check-image/commit/42a9f2202d8971634c1004ce79ad94f41aaee1c0))


### Bug Fixes

* Chain GoReleaser as job within release-please workflow ([2a5e383](https://github.com/jarfernandez/check-image/commit/2a5e383dfb0ce497fb804604b1cb366b33db1e62))
* Configure release-please to use simple tag format without component prefix ([24b7a0c](https://github.com/jarfernandez/check-image/commit/24b7a0c6936ff932bc0522b256588cb238bb748d))
* Correct GoReleaser ldflags to use correct module path for version injection ([56e44ed](https://github.com/jarfernandez/check-image/commit/56e44edf9e15f2664af987ca4ed6ddbbcf1b4350))
* Update release-please config to v4 manifest format ([#5](https://github.com/jarfernandez/check-image/issues/5)) ([ea52291](https://github.com/jarfernandez/check-image/commit/ea522915a5d7871bedc3eeff45ff5e6e49b4db5f))


### Documentation

* Add pre-built binary installation instructions and clarify version behavior ([12737b3](https://github.com/jarfernandez/check-image/commit/12737b361473e7f278f2d079d3af54a9fa1f875f))
* Document release pipeline architecture in CLAUDE.md ([f17ffad](https://github.com/jarfernandez/check-image/commit/f17ffad008ec6e8e572ea489940e7ce8dfcfa64a))
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
