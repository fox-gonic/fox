# Parameter Binding Example

This example demonstrates automatic parameter binding from different sources.

## Features

- JSON body binding with validation
- URI parameter binding
- Query parameter binding
- Combined URI and JSON binding
- Custom validation

## Running

```bash
go run main.go
```

## Testing

### Create User (JSON Binding)

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "secret123",
    "age": 25
  }'
```

### Update User (URI + JSON Binding)

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice_updated",
    "email": "alice.new@example.com"
  }'
```

### Query Users (Query Parameters)

```bash
curl "http://localhost:8080/users?page=1&page_size=10&keyword=alice"
```

### Get User by ID

```bash
curl http://localhost:8080/users/1
```

### Validation Test

```bash
# Valid request
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "secret123",
    "age": 25
  }'

# Invalid: reserved username
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "secret123",
    "age": 25
  }'

# Invalid: email format
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "invalid-email",
    "password": "secret123",
    "age": 25
  }'

# Invalid: password too short
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "123",
    "age": 25
  }'
```

## Validation Tags

Fox uses `github.com/go-playground/validator/v10` for validation:

- `required`: Field must not be empty
- `email`: Must be a valid email address
- `min`, `max`: String length or numeric value range
- `gte`, `lte`: Greater/less than or equal
- `gt`, `lt`: Greater/less than
- `omitempty`: Skip validation if empty

For more validation tags, see: https://github.com/go-playground/validator
