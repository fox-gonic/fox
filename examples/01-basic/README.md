# Basic Usage Example

This example demonstrates the most basic usage of Fox framework.

## Features

- Simple GET endpoint
- Path parameters (two methods: `ctx.Param()` and struct binding)
- Automatic JSON response rendering

## Running

```bash
go run main.go
```

## Testing

```bash
# Test ping endpoint
curl http://localhost:8080/ping

# Test parameterized endpoint - using ctx.Param()
curl http://localhost:8080/hello/world

# Test parameterized endpoint - using struct binding
curl http://localhost:8080/greet/alice

# Test POST endpoint
curl -X POST http://localhost:8080/echo
```

## Expected Output

```bash
# /ping
pong

# /hello/world
Hello, world!

# /greet/alice
Greetings, alice!

# /echo
{"message":"Echo service is working"}
```
