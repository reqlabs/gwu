# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-07-09

### Added

- Initial release of Gwu (Generic/Go Web Utility) for GoLang.
- Basic functionality for handling HTTP requests using generics (Go >=1.18).
- `gwu.Handle` function to create standard `http.Handler`.
- Utility functions: `IntoJSON`, `CnIn`, `Exec`, `JSON`, `PathVal`, `Empty`, and `ValIn`.
- Example implementation: In-Memory Poem Store with JSON API.
- Postman collection for testing the example API.

[unreleased]: https://github.com/yourusername/gwu/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/gwu/releases/tag/v1.0.0