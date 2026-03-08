.PHONY: build test test-coverage lint lint-fix

build:
	go build -o koikoi .

test:
	go test -count=1 -timeout 30s ./...

test-coverage:
	go test -coverprofile=coverage.out -count=1 -timeout 30s ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
