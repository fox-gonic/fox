# Contributing to Fox

First off, thank you for considering contributing to Fox! It's people like you that make Fox such a great framework.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Commit Message Guidelines](#commit-message-guidelines)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.24.0 or later
- Git
- golangci-lint v2.7.1 or later (for linting)
- Make (optional, but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/fox.git
cd fox
```

3. Add the upstream repository:

```bash
git remote add upstream https://github.com/fox-gonic/fox.git
```

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the behavior
- **Expected behavior**
- **Actual behavior**
- **Go version** (`go version`)
- **Fox version**
- **Code samples** (if applicable)
- **Error messages** (full stack trace)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use case**: Why would this be useful?
- **Proposed solution**: How should it work?
- **Alternatives considered**: What other solutions did you think about?
- **Additional context**: Any other relevant information

### Your First Code Contribution

Unsure where to begin? You can start by looking through issues labeled:

- `good first issue` - Simple issues suitable for newcomers
- `help wanted` - Issues where we need community help
- Check the [TODO.md](TODO.md) for a list of planned improvements

### Pull Requests

We actively welcome your pull requests! Here's how:

1. Check existing PRs to avoid duplicates
2. Create a new branch for your feature/fix
3. Make your changes following our guidelines
4. Add or update tests
5. Ensure all tests pass
6. Update documentation if needed
7. Submit a pull request

## Development Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 2. Verify Your Setup

```bash
# Run tests
make test

# Run linter
make lint

# Or run everything
make all
```

### 3. Common Development Tasks

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific package tests
go test -v ./logger/

# Run linter
golangci-lint run ./...

# Format code
gofmt -s -w .
gofumpt -w .

# Check formatting
make gofmt
```

## Coding Guidelines

### General Principles

- **Simplicity**: Keep code simple and readable
- **Consistency**: Follow existing code style
- **Documentation**: Document all exported functions and types
- **Testing**: Write tests for new features and bug fixes
- **Performance**: Be mindful of performance, but prioritize correctness

### Go Style

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://go.dev/doc/effective_go).

#### Naming

```go
// Good: Clear, concise names
func GetUser(id string) (*User, error)
type UserService struct{}

// Bad: Unclear abbreviations
func GetU(i string) (*U, error)
type UsrSrv struct{}
```

#### Error Handling

```go
// Good: Return errors, don't panic
func ParseConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    // ...
}

// Bad: Panic on errors (except in init or main)
func ParseConfig(path string) *Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // ❌
    }
    // ...
}
```

#### Comments

```go
// Good: Complete sentences, explain why
// ParseUser extracts user information from the request.
// It returns an error if the request body is malformed or
// if required fields are missing.
func ParseUser(r *http.Request) (*User, error) {
    // ...
}

// Bad: Incomplete, states the obvious
// parse user
func ParseUser(r *http.Request) (*User, error) {
    // ...
}
```

#### Package Documentation

```go
// Package fox provides a lightweight web framework for building web applications.
//
// Fox extends the Gin framework with automatic parameter binding and response rendering.
// It supports flexible handler signatures and provides built-in error handling.
//
// Example usage:
//
//	engine := fox.Default()
//	engine.GET("/hello", func(ctx *fox.Context) string {
//	    return "Hello, World!"
//	})
//	engine.Run(":8080")
package fox
```

### Code Organization

```go
// Order of declarations in a file:
// 1. Package comment
// 2. Package statement
// 3. Imports (grouped: stdlib, external, internal)
// 4. Constants
// 5. Variables
// 6. Types
// 7. Functions

package fox

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/fox-gonic/fox/httperrors"
)

const DefaultTimeout = 30 * time.Second

var ErrInvalidInput = errors.New("invalid input")

type Config struct {
    Port int
}

func New() *Engine {
    // ...
}
```

### Avoid Common Pitfalls

```go
// ❌ Don't use interface{} without reason
func Process(data interface{}) error

// ✅ Use concrete types or type parameters
func Process(data *User) error
func Process[T any](data T) error

// ❌ Don't ignore errors
data, _ := parseData(input)

// ✅ Handle errors properly
data, err := parseData(input)
if err != nil {
    return fmt.Errorf("parse data: %w", err)
}
```

## Testing Guidelines

### Test Coverage

- Aim for **80%+ coverage** for new code
- All new features **must** include tests
- Bug fixes **should** include regression tests
- Test both success and failure paths

### Writing Tests

```go
func TestFunctionName(t *testing.T) {
    // Use table-driven tests when possible
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ToUpper(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Test Organization

```go
// Use testify for assertions
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Use require for critical assertions (stops test on failure)
require.NoError(t, err)
require.NotNil(t, result)

// Use assert for non-critical assertions (continues test)
assert.Equal(t, expected, actual)
assert.True(t, condition)
```

### Benchmark Tests

```go
func BenchmarkFunction(b *testing.B) {
    // Setup code here
    input := generateTestData()

    b.ResetTimer() // Reset timer after setup

    for i := 0; i < b.N; i++ {
        _ = Function(input)
    }
}
```

### Test Naming

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Benchmark functions: `BenchmarkFunctionName`
- Example functions: `ExampleFunctionName`

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest upstream changes:

```bash
git fetch upstream
git rebase upstream/main
```

2. **Run all checks**:

```bash
make all  # Runs format, lint, vet, and tests
```

3. **Check test coverage**:

```bash
go test -cover ./...
```

4. **Update documentation** if needed

### PR Title and Description

**Title Format**: `type: brief description`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Example**:
```
feat: add request size limit middleware

Add middleware to limit request body size and prevent DoS attacks.

## Changes
- Add MaxBytesMiddleware
- Add tests for size limiting
- Update documentation

## Test Plan
- Manual testing with large payloads
- Unit tests for various sizes
- Integration tests with real HTTP requests
```

### Review Process

1. **Automated checks** must pass:
   - All tests pass
   - Code coverage doesn't decrease
   - Linter passes with 0 issues

2. **Code review** by maintainers:
   - At least one approval required
   - Address all review comments

3. **Merge**:
   - Squash and merge preferred for feature branches
   - Merge commit for important releases

### After Your PR is Merged

1. Delete your feature branch
2. Pull the latest changes from upstream
3. Celebrate! 🎉

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>: <subject>

<body>

<footer>
```

### Type

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, semicolons, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance (dependencies, build, etc.)
- `ci`: CI/CD changes

### Examples

```
feat: add context timeout support

Add support for request-level timeouts through context.
This allows users to set custom timeouts for specific routes.

Closes #123
```

```
fix: resolve panic in parameter binding

Fix panic that occurred when binding to nil pointer fields.
Added validation to check for nil pointers before binding.

Fixes #456
```

```
test: increase logger package coverage to 90%

- Add tests for all log levels
- Add tests for configuration
- Add benchmark tests
```

### Commit Best Practices

- **Use the imperative mood**: "Add feature" not "Added feature"
- **Capitalize the subject line**
- **No period at the end of subject line**
- **Limit subject to 50 characters**
- **Wrap body at 72 characters**
- **Separate subject from body with blank line**
- **Explain what and why, not how**

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feat/my-new-feature
# or
git checkout -b fix/bug-description
```

### 2. Make Changes

```bash
# Make your changes
vim file.go

# Run tests frequently
go test ./...

# Format code
gofmt -s -w .
```

### 3. Commit Changes

```bash
git add .
git commit -m "feat: add new feature"
```

### 4. Push and Create PR

```bash
git push origin feat/my-new-feature
# Then create PR on GitHub
```

### 5. Update Based on Review

```bash
# Make requested changes
vim file.go

# Commit
git commit -m "feat: address review comments"

# Push
git push origin feat/my-new-feature
```

## Getting Help

- 📖 Check the [README](.github/README.md)
- 📝 Review [TODO.md](TODO.md) for planned work
- 💬 Open a [Discussion](https://github.com/fox-gonic/fox/discussions)
- 🐛 Open an [Issue](https://github.com/fox-gonic/fox/issues)
- 📧 Email: [miclle.zheng@gmail.com](mailto:miclle.zheng@gmail.com)

## Recognition

Contributors will be recognized in:
- Release notes
- CHANGELOG.md
- Special thanks section (for significant contributions)

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Gin Documentation](https://gin-gonic.com/)
- [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)

---

Thank you for contributing to Fox! 🦊

Last updated: 2025-12-06
