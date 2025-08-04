# CRUSH.md - Development Guidelines for go-sqlite-regexp

## Build Commands
```bash
make build          # Build the package
make test           # Run all tests
make test-race      # Run tests with race detection
make test-cover     # Run tests with coverage
make bench          # Run benchmarks
make examples       # Build examples
make run-examples   # Run examples
```

## Linting and Formatting
```bash
make fmt    # Format code
make lint   # Lint code (requires golangci-lint)
make tidy   # Tidy dependencies
```

## Running Individual Tests
```bash
# Run a specific test
go test -v -run TestRegexpFunction

# Run tests with race detection
go test -race -v -run TestRegexpFunction

# Run tests with coverage
go test -cover -v -run TestRegexpFunction
```

## Code Style Guidelines

### Imports
- Standard library imports come first
- Third-party imports follow
- Local package imports last
- Group imports with blank lines

### Formatting
- Use `go fmt` for all formatting
- Line length: 100 characters max
- Indentation: tabs for Go files

### Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for exported functions
- Use descriptive names over abbreviations

### Error Handling
- Always handle errors explicitly
- Use `fmt.Errorf` with context when wrapping errors
- Return early on errors to reduce nesting

### Documentation
- Comment all exported functions
- Follow GoDoc conventions
- Include examples in comments when helpful

## Testing Guidelines
- Write table-driven tests when possible
- Test both success and failure cases
- Clean up resources in test helpers
- Use `:memory:` databases for SQLite tests