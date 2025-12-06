# Basic Usage Example

This example demonstrates the most basic usage of Fox framework.

## Features

- Simple GET endpoint
- Path parameters
- Automatic JSON response rendering

## Running

```bash
go run main.go
```

## Testing

```bash
# Test ping endpoint
curl http://localhost:8080/ping

# Test parameterized endpoint
curl http://localhost:8080/hello/world

# Test POST endpoint
curl -X POST http://localhost:8080/echo
```

## Expected Output

```bash
# /ping
pong

# /hello/world
Hello, world!

# /echo
{"message":"Echo service is working"}
```
