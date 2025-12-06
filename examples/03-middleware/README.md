# Middleware Example

This example demonstrates how to use and create custom middleware in Fox.

## Features

- Built-in middleware (Logger, ResponseTime, Recovery)
- Custom authentication middleware
- Custom rate limiting middleware
- Request ID middleware
- Route-specific middleware
- Middleware groups

## Running

```bash
go run main.go
```

## Testing

### Public Routes

```bash
# Home page
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health
```

### Protected Routes (Require Authentication)

```bash
# Without token (should fail)
curl http://localhost:8080/api/profile

# With invalid token (should fail)
curl -H "Authorization: Bearer invalid-token" \
  http://localhost:8080/api/profile

# With valid token (should succeed)
curl -H "Authorization: Bearer valid-token" \
  http://localhost:8080/api/profile

# Get data with valid token
curl -H "Authorization: Bearer valid-token" \
  http://localhost:8080/api/data
```

### Rate Limited Routes

```bash
# First request (should succeed)
curl http://localhost:8080/limited/resource

# Immediate second request (should fail with 429)
curl http://localhost:8080/limited/resource

# Wait 1 second and try again (should succeed)
sleep 1
curl http://localhost:8080/limited/resource
```

### Route-Specific Middleware

```bash
curl http://localhost:8080/special
```

### Logger Middleware

```bash
curl http://localhost:8080/with-logger
```

## Middleware Execution Order

1. Global middleware execute first (in order of `Use()`)
2. Group middleware execute next
3. Route-specific middleware execute last
4. Handler executes
5. Middleware post-processing (in reverse order)

```
Request → Logger → ResponseTime → RequestID → Recovery →
         → [Group Auth] → [Route Specific] → Handler
```

## Creating Custom Middleware

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before request
        start := time.Now()

        // Process request
        c.Next()

        // After request
        duration := time.Since(start)
        fmt.Printf("Request took %v\n", duration)
    }
}

// Use it
router.Use(MyMiddleware())
```

## Aborting Request

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !isAuthorized(c) {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Next()
    }
}
```

## Built-in Middleware

Fox provides these built-in middleware:

- `fox.Logger()`: Request logging with TraceID
- `fox.NewXResponseTimer()`: Response time tracking
- `gin.Recovery()`: Panic recovery (from Gin)
