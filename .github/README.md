# Fox Web Framework

[![Go Tests](https://github.com/fox-gonic/fox/actions/workflows/go.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/go.yml)
[![Security Scanning](https://github.com/fox-gonic/fox/actions/workflows/security.yml/badge.svg)](https://github.com/fox-gonic/fox/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fox-gonic/fox)](https://goreportcard.com/report/github.com/fox-gonic/fox)
[![GoDoc](https://pkg.go.dev/badge/github.com/fox-gonic/fox?status.svg)](https://pkg.go.dev/github.com/fox-gonic/fox)

The fox is an extension of the [gin](https://github.com/gin-gonic/gin) framework

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
    var article = &Article{
      Title:   args.Title,
      Content: args.Content,
    }
    // TODO: do something ...
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
    var user = &User{
      Username: args.Username,
      Email:    args.Email,
    }
    // TODO: do something ...
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
