# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently, the following versions are supported:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x (beta)   | :white_check_mark: |
| < 0.1.0 | :x:                |

**Note**: Fox is currently in beta. While we take security seriously, please be aware that the API may change and the framework is not yet recommended for production use.

## Reporting a Vulnerability

We take the security of Fox seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do NOT:

- **DO NOT** open a public GitHub issue for security vulnerabilities
- **DO NOT** disclose the vulnerability publicly until we've had a chance to address it
- **DO NOT** exploit the vulnerability for malicious purposes

### Please DO:

**Report security vulnerabilities by email to: [miclle.zheng@gmail.com](mailto:miclle.zheng@gmail.com)**

Please include the following information in your report:

- Type of vulnerability (e.g., XSS, SQL Injection, DoS, etc.)
- Full paths of source file(s) related to the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability, including how an attacker might exploit it

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity within 5 business days
- **Updates**: We will keep you informed of our progress
- **Fix**: We aim to release a fix within:
  - Critical vulnerabilities: 7 days
  - High severity: 14 days
  - Medium/Low severity: 30 days
- **Credit**: With your permission, we will credit you in the release notes and CHANGELOG

## Security Best Practices

When using Fox in your applications, we recommend following these security best practices:

### 1. Input Validation

Always validate and sanitize user input:

```go
type UserInput struct {
    Username string `json:"username" binding:"required,alphanum,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
}

// Use custom validators for complex rules
func (u *UserInput) IsValid() error {
    // Additional validation logic
    return nil
}
```

### 2. Use HTTPS

Always use HTTPS in production:

```go
// Use TLS configuration
server := &http.Server{
    Addr:    ":443",
    Handler: engine,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
}
server.ListenAndServeTLS("cert.pem", "key.pem")
```

### 3. Configure CORS Properly

Don't use wildcard origins in production:

```go
// BAD - Don't do this in production
engine.CORS(cors.Config{
    AllowOrigins: []string{"*"},
})

// GOOD - Specify allowed origins
engine.CORS(cors.Config{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST"},
    AllowHeaders: []string{"Origin", "Content-Type"},
})
```

### 4. Set Security Headers

Use security headers to protect against common attacks:

```go
engine.Use(func(ctx *fox.Context) {
    ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
    ctx.Writer.Header().Set("X-Frame-Options", "DENY")
    ctx.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
    ctx.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000")
    ctx.Next()
})
```

### 5. Limit Request Size

Protect against large payload attacks:

```go
// Set max request body size
http.MaxBytesReader(w, r.Body, 10*1024*1024) // 10MB limit
```

### 6. Rate Limiting

Implement rate limiting to prevent abuse:

```go
// TODO: Fox will provide built-in rate limiting middleware
// For now, use external packages or implement custom middleware
```

### 7. Error Handling

Don't expose sensitive information in error messages:

```go
// BAD - Exposes internal details
return httperrors.InternalServerError(
    "Database connection failed: " + err.Error(),
)

// GOOD - Generic error message for client
return httperrors.InternalServerError(
    "An error occurred processing your request",
)
```

### 8. Authentication & Authorization

- Always use strong password hashing (e.g., bcrypt, argon2)
- Implement proper session management
- Use CSRF tokens for state-changing operations
- Validate permissions on every protected endpoint

### 9. Logging

- Log security-relevant events
- Don't log sensitive information (passwords, tokens, PII)
- Monitor logs for suspicious activity

```go
// Use structured logging
logger.WithFields(map[string]any{
    "user_id": userID,
    "action": "login_attempt",
    "ip": ctx.ClientIP(),
}).Info("User login")
```

### 10. Dependencies

- Regularly update dependencies
- Monitor for security advisories
- Use `go list -m all | go run golang.org/x/vuln/cmd/govulncheck@latest`

## Known Security Limitations

### Beta Status

Fox is currently in beta. While we follow security best practices, the framework has not yet undergone:

- Independent security audit
- Extensive production testing
- Formal penetration testing

### Current Limitations

1. **No built-in rate limiting**: You must implement your own or use third-party middleware
2. **No built-in CSRF protection**: Implement your own CSRF middleware
3. **No built-in request size limits**: Configure at the HTTP server level
4. **Limited security middleware**: We're working on expanding security features

## Security Features Roadmap

We're actively working on improving security features:

- [ ] Built-in rate limiting middleware
- [ ] CSRF protection middleware
- [ ] Security headers middleware (Helmet-style)
- [ ] Request size limiting middleware
- [ ] Formal security audit (planned for v1.0)
- [ ] Security documentation and examples
- [ ] Integration with security scanning tools

## Security Hall of Fame

We'd like to thank the following individuals for responsibly disclosing security issues:

- (No reports yet)

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/doc/security)
- [Gin Security Best Practices](https://github.com/gin-gonic/gin#security)

## Questions?

If you have questions about security that are not sensitive in nature, feel free to open a GitHub issue or discussion.

For security-sensitive questions or concerns, please email [miclle.zheng@gmail.com](mailto:miclle.zheng@gmail.com).

---

Last updated: 2025-12-06
