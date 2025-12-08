.PHONY: all check help gofmt govet golangci-lint test coverage coverage-html benchmark security clean

# Default target
all: check

# Run all checks
check: gofmt govet golangci-lint test

# Help target
help:
	@echo "Available targets:"
	@echo "  make check          - Run all checks (gofmt, govet, golangci-lint, test)"
	@echo "  make test           - Run all tests"
	@echo "  make coverage       - Run tests and show coverage in terminal"
	@echo "  make coverage-html  - Generate HTML coverage report and open it"
	@echo "  make benchmark      - Run all benchmark tests"
	@echo "  make security       - Run security scans (govulncheck)"
	@echo "  make gofmt          - Format check"
	@echo "  make govet          - Run go vet"
	@echo "  make golangci-lint  - Run golangci-lint"
	@echo "  make clean          - Clean generated files"

# Format check
gofmt:
	@echo "Running gofmt..."
	@gofmt -s -l . | tee .gofmt.log
	@test `cat .gofmt.log | wc -l` -eq 0 || (echo "gofmt found issues, run 'gofmt -s -w .' to fix" && exit 1)
	@rm .gofmt.log
	@echo "✓ gofmt passed"

# Go vet
govet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"

# golangci-lint
golangci-lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...
	@echo "✓ golangci-lint passed"

# Run tests
test:
	@echo "Running tests..."
	@FOX_MODE=test GIN_MODE=test go test -v -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Filtering coverage to exclude examples..."
	@grep -v "examples/" coverage.txt > coverage_filtered.txt && mv coverage_filtered.txt coverage.txt || true
	@echo "✓ Tests passed"

# Show coverage in terminal
coverage: test
	@echo "Coverage summary:"
	@go tool cover -func=coverage.txt | tail -1

# Generate HTML coverage report
coverage-html: test
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Please open coverage.html manually"

# Run all benchmarks
benchmark:
	@echo "Running all benchmarks..."
	@FOX_MODE=test GIN_MODE=test go test -bench=. -benchmem -run=^$$ ./... | tee benchmark_results.txt
	@echo "✓ Benchmark results saved to benchmark_results.txt"

# Run security scans
security:
	@echo "Running security scans..."
	@echo "→ Running govulncheck..."
	@govulncheck ./... || echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"
	@echo "✓ Security scan completed"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -f coverage.txt coverage.html
	@rm -f benchmark*.txt benchmark*.out
	@rm -f .gofmt.log
	@echo "✓ Clean completed"
