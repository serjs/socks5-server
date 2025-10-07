# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased - available on :latest tag for docker image]
### Changed
### Added

## [v0.0.4] - 2025-10-07

**Important:** :warning: Starting with this release, the proxy requires authentication by default. The `REQUIRE_AUTH` parameter is now set to `true` by default. Please refer to the documentation for details on this change.

### Changed
- Migrated to a distroless Docker image from scratch.
- Moved go-socks5 to a modified forked repository at https://github.com/serjs/go-socks5 as a dependency.

### Added
- Added `REQUIRE_AUTH` parameter with a default value of `true` to enforce authentication for the proxy.
- Added `ALLOWED_DEST_FQDN` config environment parameter for filtering destination FQDNs based on regex patterns.
- Added `SetIPWhitelist` config environment parameter for setting a whitelist of IP addresses allowed to use the proxy connection.
- Implemented Dependabot version updates automation.

## [v0.0.3] - 2021-07-07
### Added
- TZ env varible support for scratch image

### Changed
- Update golang to 1.16.5
- Migrate to go module

## [v0.0.2] - 2020-03-21
### Added
- PROXY_PORT env parameter for app
- Multiarch support for docker images

### Changed
ADd caarlos0/env lib for working with ENV variables

## [v0.0.1] - 2018-04-24
### Added
- Optional auth

### Changed
- Golang vendoring
- Change Dockerfile for multistage builds with final scratch image

### Removed
- IDE files
