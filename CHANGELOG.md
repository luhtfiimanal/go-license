# Changelog

All notable changes to the Go License Management System will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
