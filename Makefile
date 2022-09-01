all: gofmt govet golangci-lint golint test

check: gofmt govet golangci-lint golint test

gofmt:
	gofmt -s -l . | tee .gofmt.log
	test `cat .gofmt.log | wc -l` -eq 0
	rm .gofmt.log

govet:
	go vet ./...

golangci-lint:
	golangci-lint run --go=1.18 ./...

golint:
	golint ./... | tee .golint.log
	test `cat .golint.log | wc -l` -eq 0
	rm .golint.log

test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test
	go tool cover -html=coverage.txt -o coverage.html
	open coverage.html