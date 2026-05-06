# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `examples/*/main_test.go` smoke tests for all examples.
- `render` package smoke test.
- Gin baseline benchmarks for performance comparison.

### Changed
- `httperrors.Error.MarshalJSON` now uses a pointer receiver.
- `bind` skips body reads when `Content-Length` is 0 and no transfer encoding is present.
- `DefaultValidator` now reuses the package-level `Validate` instance.
- Documentation now uses the actual `httperrors.Error` fields and JSON response keys.
- Documentation now aligns Go version, security policy, performance reproduction notes, and release history.

### Fixed
- Package-level `logger.Debug`, `Info`, `Warn`, `Error`, `Fatal`, and `Panic` now spread variadic arguments correctly.
- `*httperrors.Error` returned from `IsValid()` is passed through without being wrapped as `BIND_ERROR`, preserving user-defined `Code` and `HTTPCode`.
- `IsValid()` is now called for pointer handler parameters after binding.

### Deprecated
- `ErrInvalidHandlerType`; use `MsgInvalidHandlerType`.

### Removed
- `logger.NewWithContext` no longer reads trace IDs from the `logger.TraceID` string context key. Use `TraceIDKey`.
- `IsValidHandlerFunc` no longer accepts `interface` as the second parameter type.

## [0.0.10] - 2026-04-30

### Added
- Added struct binding examples for URL and query parameters.
- Comprehensive unit tests for `logger` package (90.8% coverage)
- Comprehensive unit tests for `utils` package (100% coverage)
- CHANGELOG.md file for tracking project changes
- Added Codecov integration and expanded test coverage across core packages.

### Changed
- Upgraded the module and GitHub Actions toolchain to Go 1.25.
- Aligned Gin ecosystem dependencies with Gin upstream, including `github.com/gin-gonic/gin` v1.12.0 and `github.com/gin-contrib/cors` v1.7.7.
- Updated `github.com/rs/zerolog` to v1.35.1.
- Updated GitHub Actions dependencies, including `codecov/codecov-action` v6.

### Fixed
- Attached trace IDs to the request context in logger middleware.
- Renamed `reponse_time.go` to `response_time.go` (typo correction)

## [0.1.0-beta] - 2024-12-06

> Note: This pre-release tag has been deprecated. Subsequent releases follow the `0.0.x` track (see `[0.0.10]` above) and will continue until the project graduates to `0.1.0` stable.

### Added
- Smart handler signature support with automatic parameter binding and rendering
- Support for multiple handler signatures:
  - `func()`
  - `func(ctx *Context) T`
  - `func(ctx *Context) (T, error)`
  - `func(ctx *Context, args S) T`
  - `func(ctx *Context, args S) (T, error)`
- Automatic parameter binding from JSON, Form, Query, URI, Header
- Context value binding with `context:"key"` tag
- Custom validator interface `IsValider`
- Automatic response rendering based on return types
- Structured HTTP error handling with `httperrors.Error`
- High-performance logging system based on zerolog
- TraceID/RequestID tracking support
- Multi-domain routing support with regex pattern matching
- X-Response-Time middleware for performance monitoring
- CORS configuration support

### Changed
- Upgraded to Go 1.24.0
- Updated github.com/gin-gonic/gin to v1.11.0
- Updated golang.org/x/crypto to v0.45.0
- Updated github.com/go-playground/validator/v10 to v10.28.0
- Updated github.com/stretchr/testify to v1.11.1
- Updated github.com/gin-contrib/cors to v1.7.6
- Updated github.com/rs/zerolog to v1.34.0
- Enabled ContextWithFallback by default for better context.Context support

### Fixed
- Data race issue when concurrently executing `fox.New()` (#33)
- Context.Request not taking effect in middlewares (#31)
- Context.Context implementation for thread safety (#30)
- Valid handler signature validation (#22)
- Linter errors and code formatting issues (#17)

### Removed
- Custom tag name binding from validator (breaking change) (#20)

## Dependencies

### Core Dependencies
- github.com/gin-gonic/gin v1.12.0
- github.com/gin-contrib/cors v1.7.7
- github.com/go-playground/validator/v10 v10.30.1
- github.com/json-iterator/go v1.1.12
- github.com/mitchellh/mapstructure v1.5.0
- gopkg.in/natefinch/lumberjack.v2 v2.2.1
- github.com/rs/zerolog v1.35.1
- github.com/stretchr/testify v1.11.1

### Indirect Dependencies
- golang.org/x/crypto v0.49.0
- golang.org/x/net v0.52.0
- golang.org/x/sys v0.42.0
- golang.org/x/text v0.35.0
- google.golang.org/protobuf v1.36.11

## Development

### CI/CD
- GitHub Actions for automated testing
- golangci-lint v2.4.0 for code quality checks
- Dependabot for automatic dependency updates (weekly)
- Multi-platform testing (Ubuntu, macOS)
- Race condition detection enabled

### Code Quality
- Test coverage: ~57% (main package)
- Linters: gosec, misspell, revive, testifylint, perfsprint, usestdlibvars
- Code formatting: gofmt, gofumpt, goimports

## Notes

**⚠️ Beta Status**: This project is currently in beta. APIs may change without notice. Not recommended for production use until v1.0.0 release.

### Known Issues
- Content negotiation by `Accept` header is pending and should be tracked in a GitHub issue before release.
- lumberjack dependency uses +incompatible version

### Upcoming Features
- Enhanced content negotiation based on Accept headers
- Additional middleware for common use cases
- Improved documentation and examples
- Performance optimizations
- Security enhancements (rate limiting, request size limits)

## Contributing

Please see the [contributing guidelines](CONTRIBUTING.md) for more information on how to contribute to this project.

## License

See [LICENSE](LICENSE) file for details.

---

For more information, visit the [GitHub repository](https://github.com/fox-gonic/fox).
