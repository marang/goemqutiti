# CRUSH Agent Instructions

## Build Commands
- `go build ./...` - Build all packages
- `go build -o emqutiti cmd/emqutiti/main.go` - Build main executable

## Test Commands
- `go test ./...` - Run all tests
- `go test -v ./...` - Run all tests with verbose output
- `go test -run TestName ./path/to/package` - Run a specific test
- `go test -bench=. ./...` - Run all benchmarks
- `go test -cover ./...` - Run tests with coverage
- `go test -race ./...` - Run tests with race detector
- `go test -run ExampleSet_manual -tags manual ./...` - Run manual keyring tests (requires real keyring)

## Lint/Format Commands
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Vet all packages
- `go mod tidy` - Clean up unused dependencies

## References
This file complements AGENTS.md which contains additional repository guidelines, UI guidelines, and form utilities information. Please consult both files when working with this codebase.