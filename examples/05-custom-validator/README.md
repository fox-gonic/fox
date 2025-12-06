# Custom Validator Example

This example demonstrates how to implement custom validation logic using the `IsValider` interface.

## Features

- Custom password strength validation
- Username format and reserved words validation
- Email domain whitelist validation
- Content profanity checking
- Tag format validation
- Custom error messages with error codes

## Running

```bash
go run main.go
```

## Testing

### Password Validation

```bash
# Too short
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "Abc123"}'

# Missing uppercase
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "abc12345!"}'

# Missing lowercase
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "ABC12345!"}'

# Missing digit
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "Abcdefgh!"}'

# Missing special character
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "Abcd12345"}'

# Valid strong password
curl -X POST http://localhost:8080/validate-password \
  -H "Content-Type: application/json" \
  -d '{"password": "Abc123!@#"}'
```

### Signup Validation

```bash
# Valid signup
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_smith",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'

# Invalid: reserved username
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "SecurePass123!"
  }'

# Invalid: username format
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice@smith",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'

# Invalid: email domain not allowed
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@gmail.com",
    "password": "SecurePass123!"
  }'

# Invalid: weak password
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "weak"
  }'
```

### Post Creation Validation

```bash
# Valid post
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Getting Started with Fox Framework",
    "content": "Fox is an amazing Go web framework that extends Gin...",
    "tags": ["golang", "web-framework", "fox"]
  }'

# Invalid: title too short
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hi",
    "content": "This is a test post content",
    "tags": ["test"]
  }'

# Invalid: no tags
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Post",
    "content": "This is a test post content",
    "tags": []
  }'

# Invalid: tag too short
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Getting Started",
    "content": "This is a test post content",
    "tags": ["a", "golang"]
  }'

# Invalid: tag format
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Getting Started",
    "content": "This is a test post content",
    "tags": ["golang!", "web framework"]
  }'
```

## Implementation

### IsValider Interface

Fox provides the `IsValider` interface for custom validation:

```go
type IsValider interface {
    IsValid() error
}
```

Any struct that implements this interface will have its `IsValid()` method called after standard validation passes.

### Example

```go
type SignupRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func (sr *SignupRequest) IsValid() error {
    // Custom validation logic
    if sr.Username == "admin" {
        return &httperrors.Error{
            HTTPCode: http.StatusBadRequest,
            Code:     "RESERVED_USERNAME",
            Message:  "Username is reserved",
        }
    }
    return nil
}
```

### Error Responses

Custom validators should return `*httperrors.Error` for consistent error handling:

```go
return &httperrors.Error{
    HTTPCode: http.StatusBadRequest,  // HTTP status code
    Code:     "ERROR_CODE",            // Machine-readable code
    Message:  "Human-readable message", // User-friendly message
    Err:      errors.New("internal error"), // Internal error (optional)
}
```

## Validation Flow

1. **Struct tag validation** (binding tags)
2. **IsValid() method** (if implemented)
3. **Handler execution** (if all validations pass)

```
Request → Parse JSON → Validate Tags → IsValid() → Handler
```

## Best Practices

1. **Fail Fast**: Return error as soon as validation fails
2. **Clear Messages**: Provide helpful error messages
3. **Consistent Codes**: Use consistent error codes across your API
4. **Security**: Don't expose sensitive validation logic in error messages
5. **Performance**: Cache compiled regexes or expensive validations
6. **Reusability**: Create reusable validator structs for common patterns

## Common Validation Patterns

- Password strength
- Username/email uniqueness (requires database check)
- Format validation (phone numbers, postal codes)
- Business rule validation
- Cross-field validation
- Conditional validation
- Rate limiting
- Profanity filtering
- File size/type validation
