# Error Handling Example

This example demonstrates various error handling patterns in Fox.

## Features

- Simple error returns
- HTTP errors with status codes
- Custom error definitions
- Conditional error handling
- Error with additional details
- Panic recovery
- Consistent error responses

## Running

```bash
go run main.go
```

## Testing

### Simple Error

```bash
curl http://localhost:8080/error/simple
```

Expected response:
```json
{
  "error": "something went wrong"
}
```

### HTTP Error with Code

```bash
curl http://localhost:8080/error/http
```

Expected response:
```json
{
  "code": "BAD_REQUEST",
  "message": "The request was invalid"
}
```

### User Not Found (404)

```bash
# User exists
curl http://localhost:8080/user/1

# User not found
curl http://localhost:8080/user/999
```

### Transfer with Balance Check

```bash
# Successful transfer
curl -X POST http://localhost:8080/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 100
  }'

# Insufficient balance
curl -X POST http://localhost:8080/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 10000
  }'

# Same account transfer
curl -X POST http://localhost:8080/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_user_id": 1,
    "to_user_id": 1,
    "amount": 100
  }'
```

### Login with Invalid Credentials

```bash
# Valid credentials
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }'

# Invalid credentials
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "wrongpassword"
  }'
```

### Signup with Duplicate Email

```bash
# New email
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "password123",
    "name": "New User"
  }'

# Duplicate email
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123",
    "name": "Alice"
  }'
```

### Delete User with Restrictions

```bash
# Normal user (success)
curl -X DELETE http://localhost:8080/user/2

# User not found
curl -X DELETE http://localhost:8080/user/999

# Admin user (forbidden)
curl -X DELETE http://localhost:8080/user/1
```

### Error with Details

```bash
curl http://localhost:8080/detailed-error
```

Expected response:
```json
{
  "code": "VALIDATION_FAILED",
  "message": "Request validation failed",
  "details": {
    "field": "email",
    "reason": "invalid format",
    "value": "not-an-email"
  }
}
```

### Panic Recovery

```bash
curl http://localhost:8080/panic
```

The server will recover from the panic and return a 500 error.

## Error Types

### 1. Simple Error

```go
return "", errors.New("something went wrong")
```

Response:
```json
{"error": "something went wrong"}
```

### 2. HTTP Error

```go
return "", &httperrors.Error{
    HTTPCode: http.StatusBadRequest,
    Code:     "ERROR_CODE",
    Message:  "Error message",
}
```

Response (400):
```json
{
  "code": "ERROR_CODE",
  "message": "Error message"
}
```

### 3. HTTP Error with Details

```go
return "", &httperrors.Error{
    HTTPCode: http.StatusBadRequest,
    Code:     "ERROR_CODE",
    Message:  "Error message",
    Details:  map[string]interface{}{"field": "value"},
}
```

Response (400):
```json
{
  "code": "ERROR_CODE",
  "message": "Error message",
  "details": {"field": "value"}
}
```

## Pre-defined Errors

Define common errors as package-level variables:

```go
var ErrUserNotFound = &httperrors.Error{
    HTTPCode: http.StatusNotFound,
    Err:      errors.New("user not found"),
    Code:     "USER_NOT_FOUND",
    Message:  "The requested user does not exist",
}

// Use in handler
func getUser(ctx *fox.Context) (*User, error) {
    if !userExists {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

## Best Practices

1. **Use Consistent Error Codes**
   - Define error codes as constants
   - Use UPPER_SNAKE_CASE format
   - Make them descriptive and unique

2. **Provide Helpful Messages**
   - User-friendly error messages
   - Don't expose internal implementation details
   - Include actionable information when possible

3. **Use Appropriate HTTP Status Codes**
   - 400: Bad Request (client error)
   - 401: Unauthorized (authentication required)
   - 403: Forbidden (insufficient permissions)
   - 404: Not Found
   - 409: Conflict (duplicate resource)
   - 422: Unprocessable Entity (validation error)
   - 500: Internal Server Error

4. **Log Errors Properly**
   - Log stack traces for unexpected errors
   - Include context (user ID, request ID, etc.)
   - Don't log sensitive information

5. **Handle Errors Early**
   - Validate input early
   - Return errors as soon as detected
   - Avoid nested error handling

6. **Wrap Errors for Context**
   ```go
   if err != nil {
       return nil, fmt.Errorf("failed to create user: %w", err)
   }
   ```

## Common HTTP Status Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| 400 | Bad Request | Invalid input, malformed request |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server error |
| 503 | Service Unavailable | Temporary service outage |
