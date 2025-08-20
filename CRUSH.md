# CRUSH Agent Instructions

## Build Commands
- `go build ./...` - Build all packages
- `go build -trimpath -ldflags="-s -w" -o emqutiti ./cmd/emqutiti` - Build main executable (optimized)
- `make build` - Build with Makefile (preferred)

## Test Commands
- `go test ./...` - Run all tests
- `go test -v ./...` - Run all tests with verbose output
- `go test -run TestName ./path/to/package` - Run a specific test
- `go test -run ^TestFunctionName$ ./path/to/package -v` - Run a single test function
- `go test -bench=. ./...` - Run all benchmarks
- `go test -cover ./...` - Run tests with coverage
- `go test -race ./...` - Run tests with race detector
- `go test -run ExampleSet_manual -tags manual ./...` - Run manual keyring tests (requires real keyring)
- `make test` - Run tests with vet (preferred)

## Lint/Format Commands
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Vet all packages
- `go mod tidy` - Clean up unused dependencies
- `make vet` - Vet code (preferred)

## Code Style Guidelines
- Use `gofmt` for formatting (configured in Makefile)
- Follow idiomatic Go patterns
- Use short variable names for receivers (e.g., `r` for `Repository`)
- Use descriptive names for exported functions/variables
- Handle errors explicitly - don't ignore them
- Use `fmt.Errorf` with `%w` verb to wrap errors
- Group imports in stdlib, third-party, and local packages
- Comment exported functions with Godoc style comments
- Keep functions small and focused

## References
This file complements AGENTS.md which contains additional repository guidelines, UI guidelines, and form utilities information. Please consult both files when working with this codebase.