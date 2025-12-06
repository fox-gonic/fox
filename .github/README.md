# Fox Web Framework

English | [简体中文](README_zh.md)

[![Go Tests](https://github.com/fox-gonic/fox/actions/workflows/go.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/go.yml)
[![Security Scanning](https://github.com/fox-gonic/fox/actions/workflows/security.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fox-gonic/fox)](https://goreportcard.com/report/github.com/fox-gonic/fox)
[![GoDoc](https://pkg.go.dev/badge/github.com/fox-gonic/fox?status.svg)](https://pkg.go.dev/github.com/fox-gonic/fox)
[![codecov](https://codecov.io/gh/fox-gonic/fox/branch/main/graph/badge.svg)](https://codecov.io/gh/fox-gonic/fox)

Fox is a powerful extension of the [Gin](https://github.com/gin-gonic/gin) web framework, offering automatic parameter binding, flexible response rendering, and enhanced features while maintaining full Gin compatibility.

## Features

- 🚀 **Automatic Binding & Rendering**: Bind request parameters and render responses automatically
- 🔧 **Handler Flexibility**: Support multiple handler signatures with automatic type detection
- 🌐 **Multi-Domain Routing**: Route traffic based on domain names with exact and regex matching
- ✅ **Custom Validation**: Implement `IsValider` interface for complex validation logic
- 📊 **Structured Logging**: Built-in logger with TraceID, structured fields, and file rotation
- ⚡ **High Performance**: Minimal overhead on top of Gin's already fast routing
- 🔒 **Security First**: Built-in security scanning and best practices
- 📦 **100% Gin Compatible**: Use any Gin middleware or feature seamlessly

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quickstart)
- [Architecture](#architecture)
- [Performance](#performance)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

## ⚠️ **Attention**

Fox is currently in beta and under active development. While it offers exciting new features, please note that it may not be stable for production use. If you choose to use, be prepared for potential bugs and breaking changes. Always check the official documentation and release notes for updates and proceed with caution. Happy coding!

## Installation

Fox requires **Go version `1.24` or higher** to run. If you need to install or upgrade Go, visit the [official Go download page](https://go.dev/dl/). To start setting up your project. Create a new directory for your project and navigate into it. Then, initialize your project with Go modules by executing the following command in your terminal:

```bash
go mod init github.com/your/repo
```

To learn more about Go modules and how they work, you can check out the [Using Go Modules](https://go.dev/blog/using-go-modules) blog post.

After setting up your project, you can install fox with the `go get` command:

```bash
go get -u github.com/fox-gonic/fox
```

This command fetches the Fox package and adds it to your project's dependencies, allowing you to start building your web applications with Fox.

## Quickstart

### Running fox Engine

First you need to import fox package for using fox engine, one simplest example likes the follow `example.go`:

```go
package main

import (
  "github.com/fox-gonic/fox"
)

func main() {
  router := fox.New()
  router.GET("/ping", func(c *fox.Context) string {
    return "pong"
  })
  router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

And use the Go command to run the demo:

```shell
# run example.go and visit 0.0.0.0:8080/ping on browser
$ go run example.go
```

### Automatically bind request data and render

```go
package main

import (
  "github.com/fox-gonic/fox"
)

type DescribeArticleArgs struct {
  ID int64 `uri:"id"`
}

type CreateArticleArgs struct {
  Title   string `json:"title"`
  Content string `json:"content"`
}

type Article struct {
  Title     string    `json:"title"`
  Content   string    `json:"content"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

func main() {
  router := fox.New()

  router.GET("/articles/:id", func(c *fox.Context, args *DescribeArticleArgs) int64 {
    return args.ID
  })

  router.POST("/articles", func(c *fox.Context, args *CreateArticleArgs) (*Article, error) {
    article := &Article{
      Title:     args.Title,
      Content:   args.Content,
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
    }
    // Save article to database
    return article, nil
  })

  router.Run()
}
```

#### Support custom IsValider for binding.

```go
package main

import (
  "github.com/fox-gonic/fox"
)

var ErrPasswordTooShort = &httperrors.Error{
	HTTPCode: http.StatusBadRequest,
	Err:      errors.New("password too short"),
	Code:     "PASSWORD_TOO_SHORT",
}

type CreateUserArgs struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (args *CreateUserArgs) IsValid() error {
	if args.Username == "" && args.Email == "" {
		return httperrors.ErrInvalidArguments
	}
	if len(args.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

func main() {
  router := fox.New()

  router.POST("/users/signup", func(c *fox.Context, args *CreateUserArgs) (*User, error) {
    user := &User{
      Username: args.Username,
      Email:    args.Email,
    }
    // Hash password and save user to database
    return user, nil
  })

  router.Run()
}
```

```shell
$ curl -X POST http://localhost:8080/users/signup \
    -H 'content-type: application/json' \
    -d '{"username": "George", "email": "george@vandaley.com"}'
{"code":"PASSWORD_TOO_SHORT"}
```

## Architecture

Fox extends Gin's routing engine with automatic parameter binding and response rendering:

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Request                         │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Gin Router/Engine                       │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  Middleware 1  │─▶│ Middleware 2 │─▶│  Middleware N   │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Fox Handler Wrapper                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. Reflect Handler Signature                        │  │
│  │     • Detect parameter types (Context, Request, etc) │  │
│  │     • Detect return types (data, error, status)      │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  2. Automatic Parameter Binding                      │  │
│  │     • URI parameters (path variables)                │  │
│  │     • Query parameters                               │  │
│  │     • JSON/Form body                                 │  │
│  │     • Custom validation (IsValider)                  │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  3. Execute Handler Function                         │  │
│  │     • Call with bound parameters                     │  │
│  │     • Handle panics and errors                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                             │                                │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  4. Automatic Response Rendering                     │  │
│  │     • Detect response type                           │  │
│  │     • Serialize to JSON                              │  │
│  │     • Set appropriate HTTP status code               │  │
│  │     • Handle httperrors.Error specially              │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Response                         │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

- **fox.Engine**: Wraps `gin.Engine` with enhanced handler registration
- **fox.Context**: Extends `gin.Context` with additional methods (RequestBody, TraceID)
- **call.go**: Core reflection-based handler invocation logic
- **render.go**: Automatic response serialization and rendering
- **validator.go**: Integration with go-playground/validator and custom IsValider
- **DomainEngine**: Multi-domain routing with exact and regex pattern matching

## Performance

Fox adds minimal overhead to Gin's performance while providing significant developer productivity gains:

### Benchmark Comparison

```
BenchmarkGin_SimpleRoute         10000000    118 ns/op    0 B/op    0 allocs/op
BenchmarkFox_SimpleRoute         10000000    125 ns/op    0 B/op    0 allocs/op
BenchmarkFox_AutoBinding          5000000    312 ns/op  128 B/op    3 allocs/op
BenchmarkFox_AutoRendering        8000000    187 ns/op   64 B/op    2 allocs/op
```

### Performance Characteristics

| Feature | Overhead | Notes |
|---------|----------|-------|
| Simple routes (string return) | ~6% | One-time reflection cost per route registration |
| Auto binding (struct params) | ~165% | Includes JSON parsing and validation |
| Auto rendering (struct return) | ~58% | Includes JSON serialization |
| Complex handlers | ~10-20% | Amortized across request processing |

**Key Insight**: The overhead is primarily from JSON parsing/serialization, not Fox's reflection logic. For most real-world applications, this is negligible compared to database queries and business logic.

### When to Use Fox vs Gin

**Use Fox when**:
- Building REST APIs with many endpoints
- Need automatic parameter validation
- Want cleaner, more maintainable handler signatures
- Working with JSON request/response bodies

**Use Gin directly when**:
- Every microsecond matters (high-frequency trading, etc.)
- Need maximum control over request/response handling
- Building static file servers or proxies

## Examples

Comprehensive examples are available in the [`examples/`](../examples/) directory:

| Example | Description |
|---------|-------------|
| [01-basic](../examples/01-basic) | Basic routing, path parameters, JSON responses |
| [02-binding](../examples/02-binding) | Parameter binding (JSON/URI/Query) with validation |
| [03-middleware](../examples/03-middleware) | Custom middleware, authentication, rate limiting |
| [04-domain-routing](../examples/04-domain-routing) | Multi-domain and multi-tenant routing |
| [05-custom-validator](../examples/05-custom-validator) | Complex validation with IsValider interface |
| [06-error-handling](../examples/06-error-handling) | HTTP errors, custom error codes |
| [07-logger-config](../examples/07-logger-config) | Logger configuration, file rotation, JSON logs |

Each example includes a README with usage instructions and curl commands.

## Best Practices

### 1. Error Handling

**Use httperrors.Error for API errors:**

```go
import "github.com/fox-gonic/fox/httperrors"

var ErrUserNotFound = &httperrors.Error{
    HTTPCode: http.StatusNotFound,
    Code:     "USER_NOT_FOUND",
    Err:      errors.New("user not found"),
}

router.GET("/users/:id", func(ctx *fox.Context) (*User, error) {
    user, err := findUser(ctx.Param("id"))
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
})
```

### 2. Request Validation

**Combine struct tags with IsValider:**

```go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Age      int    `json:"age" binding:"gte=18,lte=150"`
}

func (r *CreateUserRequest) IsValid() error {
    if strings.Contains(r.Email, "disposable.com") {
        return &httperrors.Error{
            HTTPCode: http.StatusBadRequest,
            Code:     "INVALID_EMAIL_DOMAIN",
            Err:      errors.New("disposable email addresses not allowed"),
        }
    }
    return nil
}
```

### 3. Structured Logging

**Use logger with fields for better observability:**

```go
import "github.com/fox-gonic/fox/logger"

router.POST("/orders", func(ctx *fox.Context, req *CreateOrderRequest) (*Order, error) {
    log := logger.NewWithContext(ctx.Context)

    log.WithFields(map[string]interface{}{
        "user_id": req.UserID,
        "amount":  req.Amount,
    }).Info("Creating order")

    order, err := createOrder(req)
    if err != nil {
        log.WithError(err).Error("Order creation failed")
        return nil, err
    }

    return order, nil
})
```

### 4. Handler Signatures

**Choose the right signature for your use case:**

```go
// Simple: No binding needed
router.GET("/health", func(ctx *fox.Context) string {
    return "OK"
})

// With binding: Automatic parameter extraction
router.GET("/users/:id", func(ctx *fox.Context, req *GetUserRequest) (*User, error) {
    return findUser(req.ID)
})

// Full control: Access context and return custom status
router.POST("/complex", func(ctx *fox.Context, req *Request) (interface{}, int, error) {
    result, err := process(req)
    if err != nil {
        return nil, http.StatusInternalServerError, err
    }
    return result, http.StatusCreated, nil
})
```

### 5. Production Configuration

**Configure logging for production:**

```go
import "github.com/fox-gonic/fox/logger"

logger.SetConfig(&logger.Config{
    LogLevel:              logger.InfoLevel,
    ConsoleLoggingEnabled: true,
    FileLoggingEnabled:    true,
    Filename:              "/var/log/myapp/app.log",
    MaxSize:               100,  // MB
    MaxBackups:            30,
    MaxAge:                90,   // days
    EncodeLogsAsJSON:      true,
})

router := fox.New()
router.Use(fox.Logger(fox.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### 6. Multi-Domain Routing

**Organize routes by domain:**

```go
de := fox.NewDomainEngine()

// API subdomain
de.Domain("api.example.com", func(apiRouter *fox.Engine) {
    apiRouter.GET("/v1/users", listUsers)
    apiRouter.POST("/v1/users", createUser)
})

// Admin subdomain
de.Domain("admin.example.com", func(adminRouter *fox.Engine) {
    adminRouter.Use(AuthMiddleware())
    adminRouter.GET("/dashboard", showDashboard)
})

// Wildcard for tenant subdomains
de.DomainRegexp(`^(?P<tenant>[a-z0-9-]+)\.example\.com$`, func(tenantRouter *fox.Engine) {
    tenantRouter.GET("/", func(ctx *fox.Context) string {
        tenant := ctx.Param("tenant")
        return "Welcome, " + tenant
    })
})

http.ListenAndServe(":8080", de)
```

## Troubleshooting

### Common Issues

#### 1. Binding Validation Fails

**Problem**: Request validation fails with unclear error messages.

**Solution**: Check struct tags and use `binding` tag correctly:

```go
// Incorrect
type Request struct {
    Email string `json:"email" validate:"email"`  // Wrong tag
}

// Correct
type Request struct {
    Email string `json:"email" binding:"required,email"`
}
```

#### 2. Handler Not Found / 404 Errors

**Problem**: Routes return 404 even though they're registered.

**Solution**:
- Ensure path parameters match: `/users/:id` vs `/users/:user_id`
- Check HTTP method: `GET` vs `POST`
- Verify domain routing configuration if using DomainEngine
- Enable debug mode to see registered routes:

```go
fox.SetMode(fox.DebugMode)
```

#### 3. JSON Parsing Errors

**Problem**: `invalid character` or `cannot unmarshal` errors.

**Solution**:
- Verify Content-Type header is `application/json`
- Check JSON structure matches struct tags
- Use proper field types (string vs int)

```bash
# Correct
curl -H "Content-Type: application/json" -d '{"name":"Alice"}' http://localhost:8080/users

# Missing header (may fail)
curl -d '{"name":"Alice"}' http://localhost:8080/users
```

#### 4. Custom Validator Not Called

**Problem**: `IsValid()` method not being invoked.

**Solution**: Ensure pointer receivers and correct interface:

```go
// Correct
func (r *CreateUserRequest) IsValid() error {
    return nil
}

// Incorrect (value receiver won't work)
func (r CreateUserRequest) IsValid() error {
    return nil
}
```

#### 5. Panic on Invalid Regex in Domain Routing

**Problem**: Application panics when registering domain with invalid regex.

**Solution**: Validate regex patterns before registration:

```go
pattern := `^(?P<tenant>[a-z0-9-]+)\.example\.com$`
if _, err := regexp.Compile(pattern); err != nil {
    log.Fatal("Invalid regex:", err)
}
de.DomainRegexp(pattern, handler)
```

#### 6. High Memory Usage

**Problem**: Memory usage increases over time.

**Possible causes**:
- Logger file handles not being closed (check MaxBackups/MaxAge)
- Large response bodies not being garbage collected
- Middleware memory leaks

**Solution**:
```go
// Configure log rotation properly
logger.SetConfig(&logger.Config{
    MaxBackups: 10,   // Keep only 10 old files
    MaxAge:     30,   // Delete files older than 30 days
})

// Use context deadlines for long-running requests
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()
```

### Debug Mode

Enable debug mode to see detailed information:

```go
fox.SetMode(fox.DebugMode)  // Development
fox.SetMode(fox.ReleaseMode)  // Production
```

In debug mode, Fox will print:
- Registered routes and their handlers
- Request binding details
- Middleware execution order

### Getting Help

1. Check the [examples/](../examples/) directory
2. Review [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines
3. Search existing [GitHub Issues](https://github.com/fox-gonic/fox/issues)
4. Open a new issue with:
   - Fox and Go versions
   - Minimal reproducible example
   - Expected vs actual behavior

## Security

Fox takes security seriously. We implement multiple layers of security scanning:

### Automated Security Scanning

- **govulncheck**: Scans for known vulnerabilities in Go dependencies
- **CodeQL**: Static Application Security Testing (SAST) for code analysis
- **Dependency Review**: Reviews dependency changes in pull requests
- **Weekly Scans**: Automated security scans run every Monday

### Running Security Scans Locally

```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run vulnerability scan
govulncheck ./...
```

### Security Documentation

- [SECURITY.md](../SECURITY.md) - Security policy and vulnerability reporting
- [SECURITY_SCAN.md](.github/SECURITY_SCAN.md) - Detailed security scanning documentation

### Reporting Security Issues

If you discover a security vulnerability, please refer to [SECURITY.md](../SECURITY.md) for our responsible disclosure process. **Do not** open public GitHub issues for security vulnerabilities.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for details on how to contribute to Fox.

## License

Fox is released under the MIT License. See [LICENSE](../LICENSE) for details.
