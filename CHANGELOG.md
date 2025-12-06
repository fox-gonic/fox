# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive unit tests for `logger` package (90.8% coverage)
- Comprehensive unit tests for `utils` package (100% coverage)
- CHANGELOG.md file for tracking project changes

### Fixed
- Renamed `reponse_time.go` to `response_time.go` (typo correction)

## [0.1.0-beta] - 2024-12-06

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
- github.com/gin-gonic/gin v1.11.0
- github.com/gin-contrib/cors v1.7.6
- github.com/go-playground/validator/v10 v10.28.0
- github.com/json-iterator/go v1.1.12
- github.com/mitchellh/mapstructure v1.5.0
- github.com/natefinch/lumberjack v2.0.0+incompatible
- github.com/rs/zerolog v1.34.0
- github.com/stretchr/testify v1.11.1

### Indirect Dependencies
- golang.org/x/crypto v0.45.0
- golang.org/x/net v0.47.0
- golang.org/x/sys v0.38.0
- golang.org/x/text v0.31.0
- golang.org/x/tools v0.38.0
- google.golang.org/protobuf v1.36.10

## Development

### CI/CD
- GitHub Actions for automated testing
- golangci-lint v2.1.6 for code quality checks
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
- TODO: Implement render by writer content-type (render.go:39, render.go:63)
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
