all: lint test
PHONY: test coverage lint golint clean vendor
GOOS=linux

.PHONY: test coverage lint golint vendor clean

test: | lint
	@echo Running unit tests...
	@go test -cover -short ./...

coverage:
	@echo Generating coverage report...
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out

lint: golint

golint: | vendor
	@echo Linting Go files...
	@golangci-lint run

clean:
	@echo Cleaning...
	@rm -rf coverage.out
	@go clean -testcache

vendor:
	@go mod download
	@go mod tidy
