# Changelog

All notable changes to the Go License Management System will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.1] - 2025-04-29

### Fixed
- Updated module path to follow Go's semantic import versioning rules
- Changed import paths to include `/v2` suffix for compatibility with Go modules
- This is a compatibility fix and contains no functional changes from v2.0.0

## [2.0.0] - 2025-04-29

### Added
- Binary license format for improved security and smaller file size
- Automatic hardware detection with `--auto-hardware` flag
- Backward compatibility for reading and verifying legacy JSON licenses

### Changed
- Licenses are now generated exclusively in binary format
- Updated verification to support both binary and legacy JSON formats
- Improved format detection in the `info` command
- Updated documentation to reflect the new binary format

### Removed
- JSON license generation option (only binary format is now supported)
- Encryption-related code (simplified the codebase)

## [1.1.0] - 2025-04-14

### Added
- Serial number field to license structure for better tracking and identification
- Serial number parameter in all license generation functions
- Serial number display in CLI tools and client examples

### Changed
- Updated license verification to include serial number validation
- Improved documentation to reflect the new serial number field
- Updated all test cases to include serial number

## [1.0.0] - 2025-04-14

### Added
- Initial release of the Go License Management System
- License verification library (`pkg/licverify`)
  - Offline license validation
  - Hardware binding (MAC address, disk ID, hostname)
  - Digital signature verification
  - Expiration date checking
- License generation library (`pkg/licgen`)
  - License creation with configurable parameters
  - RSA key pair generation
  - Digital signing of licenses
- Command-line tools
  - `licforge` - Comprehensive CLI for license management
    - Key generation with configurable key size
    - License generation with hardware binding
    - Interactive license creation mode
    - License verification and information display
  - `client-example` - Example client application demonstrating license verification
- Comprehensive test suite
  - Unit tests for all components
  - Integration tests for the full license workflow
- Documentation
  - Usage instructions for all components
  - Security considerations
  - Examples for client integration
