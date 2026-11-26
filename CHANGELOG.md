# [3.0.0](https://github.com/Jerome1337/ccc/compare/v2.1.0...v3.0.0) (2026-11-26)

### Breaking changes
* a14f4c2 feat!: drop the built-in config
### Features
* a8735ff feat: install a config repo with --from and create one with export

# [2.1.0](https://github.com/Jerome1337/ccc/compare/v2.0.0...v2.1.0) (2026-11-26)

### Features
* ee2e9e2 feat: add --auto mode to pr-review
* 6a9b339 feat: grade pr files by review confidence
* 38a4d34 feat: speed up pr-review and skip config files
### Fixes
* d35b052 fix: address pr-review review comments
* 92bbc27 fix: close pr-review --auto gaps
* 8dd8345 fix: close pr-review carry-over and grade 5 gaps
* aa525c9 fix: close pr-review triage gaps
* 58993f6 fix: harden pr-review grading and path handling
* b365414 fix: keep pr-review carry-over and summary size consistent
* 9d369a6 fix: make pr-review notes name a concrete risk
* 6e15d49 fix: tighten pr-review triage checks

# [2.0.0](https://github.com/Jerome1337/ccc/compare/v1.3.1...v2.0.0) (2026-11-25)


### Features

* rewrite ccc as a go binary that updates itself ([1aa88f0](https://github.com/Jerome1337/ccc/commit/1aa88f0e6a8a7aae8f37559c0e5c4934af1c9ed5))


### BREAKING CHANGES

* the npm package now only prints how to install the binary.
Install it with `curl -fsSL https://raw.githubusercontent.com/Jerome1337/ccc/main/install.sh | sh`,
then run `ccc update --global`.

## [1.3.1](https://github.com/Jerome1337/ccc/compare/v1.3.0...v1.3.1) (2026-11-25)


### Bug Fixes

* match the github repository url case so npm accepts the provenance ([b5e4ce0](https://github.com/Jerome1337/ccc/commit/b5e4ce085b88d46e80031a596f601e8356b34a9a))

# [1.3.0](https://github.com/jerome1337/ccc/compare/v1.2.1...v1.3.0) (2026-11-25)


### Features

* install the humanizer plugin by default with ccc --global ([1de8fb1](https://github.com/jerome1337/ccc/commit/1de8fb18b8ac250de613b83d4a5c14d4e0dabe65))
* install your mcp servers and plugins with ccc init --global and ccc update --global ([40d405c](https://github.com/jerome1337/ccc/commit/40d405c9e4af3e0ec27a947677f9b4db4e984109))
* read the mcp servers and plugins to install from an editable ccc.yml, and install stripe and context7 as plugins ([e8208a4](https://github.com/jerome1337/ccc/commit/e8208a4dd6d70b54dcdd58b66ecaf2e9c4604b03))
* remind you to log in with /mcp and set up plugins with /plugin after ccc --global installs them ([c2a0dba](https://github.com/jerome1337/ccc/commit/c2a0dba66c4f349037d1404c4fe687da7713febc))

## [1.2.1](https://github.com/jerome1337/ccc/compare/v1.2.0...v1.2.1) (2026-11-22)


### Bug Fixes

* publish the skills folder so ccc init copies pr-review.md when installed from npm ([7670c5c](https://github.com/jerome1337/ccc/commit/7670c5c5881d3a59a6fc8db1095eceaaaa7a865b))

# [1.2.0](https://github.com/jerome1337/ccc/compare/v1.1.0...v1.2.0) (2026-11-22)


### Features

* add a reuse rule so existing code is searched and reused before writing new code ([a4faacb](https://github.com/jerome1337/ccc/commit/a4faacb7b64660d7051295cb6b32e444e576093d))
* add ccc update and --global so an existing install can be brought up to date, and add git and .env rules to the coding style ([1c39b37](https://github.com/jerome1337/ccc/commit/1c39b37501398eb2ab968593023e64777a6bc5fb))
* add framework-agnostic coding style rules ([d54065d](https://github.com/jerome1337/ccc/commit/d54065d5ee71370c2d0049419eaafb7a661a2b57))
* allow heroicons or another icon set when the project already uses one, with lucide as the default ([a366bac](https://github.com/jerome1337/ccc/commit/a366bacac27efe844ed05554b95a0843f406cd4a))

# [1.1.0](https://github.com/jerome1337/ccc/compare/v1.0.0...v1.1.0) (2026-03-12)


### Features

* add pr review skill ([a4a4f66](https://github.com/jerome1337/ccc/commit/a4a4f669f99643500517de4b2519f8aede32e5d7))

# 1.0.0 (2026-03-01)


### Features

* make it reusable ([5cc0a1d](https://github.com/jerome1337/ccc/commit/5cc0a1d5c53f5c240f3930b2c3b3624b68942e2a))
