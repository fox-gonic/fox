all: gofmt govet golangci-lint golint test

check: gofmt govet golangci-lint golint test

gofmt:
	test `gofmt -s -l . | wc -l` -eq 0

govet:
	go vet ./...

golangci-lint:
	golangci-lint run --go=1.18 ./...

golint:
	test `golint ./... | wc -l` -eq 0

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
