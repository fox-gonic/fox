# Domain-Based Routing Example

This example demonstrates multi-domain routing with Fox's DomainEngine.

## Features

- Exact domain matching
- Regex domain pattern matching
- Wildcard subdomain routing
- Default/fallback routing
- Multi-tenant applications support

## Running

```bash
go run main.go
```

## Testing

Since you're testing locally, you'll need to either:

### Option 1: Use /etc/hosts (Recommended for local testing)

Add these lines to `/etc/hosts`:

```
127.0.0.1 api.example.com
127.0.0.1 admin.example.com
127.0.0.1 app1.staging.example.com
127.0.0.1 app2.staging.example.com
127.0.0.1 tenant1.app.example.com
127.0.0.1 tenant2.app.example.com
127.0.0.1 www.example.com
127.0.0.1 example.com
```

Then test:

```bash
# API domain
curl http://api.example.com:8080/
curl http://api.example.com:8080/users
curl http://api.example.com:8080/status

# Admin domain
curl http://admin.example.com:8080/
curl http://admin.example.com:8080/dashboard
curl http://admin.example.com:8080/settings

# Staging subdomains (regex match)
curl http://app1.staging.example.com:8080/
curl http://app2.staging.example.com:8080/info

# Tenant apps (regex match)
curl http://tenant1.app.example.com:8080/
curl http://tenant2.app.example.com:8080/tenant-info

# Default domain
curl http://www.example.com:8080/
curl http://example.com:8080/about
curl http://example.com:8080/contact
```

### Option 2: Use Host Header

```bash
# API domain
curl -H "Host: api.example.com" http://localhost:8080/
curl -H "Host: api.example.com" http://localhost:8080/users

# Admin domain
curl -H "Host: admin.example.com" http://localhost:8080/
curl -H "Host: admin.example.com" http://localhost:8080/dashboard

# Staging subdomains
curl -H "Host: app1.staging.example.com" http://localhost:8080/
curl -H "Host: app2.staging.example.com" http://localhost:8080/info

# Tenant apps
curl -H "Host: tenant1.app.example.com" http://localhost:8080/
curl -H "Host: tenant2.app.example.com" http://localhost:8080/tenant-info

# Default domain
curl -H "Host: www.example.com" http://localhost:8080/
curl http://localhost:8080/about
```

## Domain Matching Priority

1. **Exact domain match** (first registered)
2. **Regex domain match** (first registered that matches)
3. **Default/fallback routes**

Example:
- `api.example.com` → API routes (exact match)
- `*.staging.example.com` → Staging routes (regex match)
- `anything-else.com` → Default routes (fallback)

## Use Cases

### 1. API Gateway

Route different services based on subdomain:
- `api.example.com` → API service
- `admin.example.com` → Admin panel
- `cdn.example.com` → CDN service

### 2. Multi-tenant SaaS

Each tenant gets their own subdomain:
- `tenant1.app.example.com` → Tenant 1's application
- `tenant2.app.example.com` → Tenant 2's application

### 3. Environment Separation

- `*.prod.example.com` → Production services
- `*.staging.example.com` → Staging services
- `*.dev.example.com` → Development services

### 4. Microservices

Different microservices on different domains:
- `users.api.example.com` → User service
- `orders.api.example.com` → Order service
- `payments.api.example.com` → Payment service

## Implementation Details

```go
// Create domain engine
de := fox.NewDomainEngine()

// Exact domain
de.Domain("api.example.com", func(router *fox.Engine) {
    // API routes
})

// Regex pattern
de.DomainRegexp(`^.*\.staging\.example\.com$`, func(router *fox.Engine) {
    // Staging routes
})

// Default routes
de.GET("/", handler)
```

## Tips

- Domain matching is case-insensitive
- Port numbers in Host header are automatically stripped
- Regex patterns must match the entire domain name
- First matching domain handler is used (be careful with order)
- Use exact matches for better performance when possible
