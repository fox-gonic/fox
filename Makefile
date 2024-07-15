all: gofmt govet golangci-lint test

check: gofmt govet golangci-lint test

gofmt:
	gofmt -s -l . | tee .gofmt.log
	test `cat .gofmt.log | wc -l` -eq 0
	rm .gofmt.log

govet:
	go vet ./...

golangci-lint:
	golangci-lint run --go=1.18 ./...

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test
	go tool cover -html=coverage.txt -o coverage.html
	open coverage.html
