.PHONY: test test-unit test-integration test-coverage

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./bin/ ./...
run: build
	./bin/server

test: test-unit test-integration

test-unit:
	go test -v ./internal/...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-race:
	go test -race -v ./...

