# go-sqlite-regexp

A Go package that provides REGEXP functionality for SQLite databases using go-sqlite3. This package enables powerful regular expression matching in SQLite queries, allowing for pattern-based JOINs and WHERE clauses.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Performance](#performance)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Prerequisites

Before installing this package, ensure you have:

1. **Go 1.21 or later** - Download from [golang.org](https://golang.org/dl/)
2. **CGO enabled** - Required for go-sqlite3 compilation
3. **C compiler** - GCC or Clang for building SQLite extensions

On Ubuntu/Debian:
```bash
sudo apt update
sudo apt install build-essential
```

On macOS:
```bash
xcode-select --install
```

### Install the Package

```bash
go get github.com/go-go-golems/go-sqlite-regexp
```

## Quick Start

Here's a simple example to get you started:

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
    sqlite_regexp "github.com/go-go-golems/go-sqlite-regexp"
)

func main() {
    // Open database with REGEXP function automatically registered
    db, err := sqlite_regexp.OpenWithRegexp(":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create a simple test
    _, err = db.Exec("CREATE TABLE items (name TEXT)")
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec("INSERT INTO items VALUES ('apple'), ('banana'), ('apricot')")
    if err != nil {
        log.Fatal(err)
    }

    // Use REGEXP in a query
    rows, err := db.Query("SELECT name FROM items WHERE name REGEXP '^ap'")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("Items starting with 'ap':")
    for rows.Next() {
        var name string
        rows.Scan(&name)
        fmt.Println("-", name)
    }
}
```

Output:
```
Items starting with 'ap':
- apple
- apricot
```



## Usage

### Method 1: OpenWithRegexp (Recommended)

The simplest way to use this package is with the `OpenWithRegexp` function, which automatically registers the REGEXP function:

```go
db, err := sqlite_regexp.OpenWithRegexp("database.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// REGEXP function is now available
```

### Method 2: Manual Registration

If you need more control over the database connection, you can register the REGEXP function manually:

```go
db, err := sql.Open("sqlite3", "database.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Register REGEXP function
err = sqlite_regexp.RegisterRegexpFunction(db)
if err != nil {
    log.Fatal(err)
}

// REGEXP function is now available
```

### REGEXP Syntax

The REGEXP function follows Go's regular expression syntax (RE2). The basic syntax is:

```sql
column_name REGEXP 'pattern'
```

This returns 1 (true) if the pattern matches, 0 (false) otherwise.

#### Common Patterns

| Pattern | Description | Example |
|---------|-------------|---------|
| `^text` | Starts with "text" | `name REGEXP '^John'` |
| `text$` | Ends with "text" | `name REGEXP 'son$'` |
| `\d+` | Contains digits | `code REGEXP '\d+'` |
| `[a-z]+` | Contains lowercase letters | `word REGEXP '[a-z]+'` |
| `text1\|text2` | Contains "text1" OR "text2" | `category REGEXP 'fruit\|vegetable'` |
| `.*` | Matches anything | `description REGEXP '.*important.*'` |

### Pattern-Based JOINs

One of the most powerful features is using REGEXP in JOIN conditions:

```sql
SELECT  p.category,
        i.item,
        i.amount
FROM    patterns   AS p      -- pattern, category
JOIN    items      AS i      -- item, amount
     ON i.item REGEXP p.pattern;
```

This allows you to categorize items based on flexible pattern matching rather than exact string matches.


## API Reference

### Functions

#### `OpenWithRegexp(dataSourceName string) (*sql.DB, error)`

Opens a SQLite database connection and automatically registers the REGEXP function.

**Parameters:**
- `dataSourceName`: SQLite data source name (file path or ":memory:" for in-memory database)

**Returns:**
- `*sql.DB`: Database connection with REGEXP function registered
- `error`: Error if connection fails or REGEXP registration fails

**Example:**
```go
db, err := sqlite_regexp.OpenWithRegexp("./mydb.sqlite")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### `RegisterRegexpFunction(db *sql.DB) error`

Registers the REGEXP function with an existing SQLite database connection.

**Parameters:**
- `db`: Existing SQLite database connection

**Returns:**
- `error`: Error if registration fails

**Example:**
```go
db, err := sql.Open("sqlite3", "database.db")
if err != nil {
    log.Fatal(err)
}

err = sqlite_regexp.RegisterRegexpFunction(db)
if err != nil {
    log.Fatal(err)
}
```

#### `ClearRegexpCache()`

Clears the internal regular expression cache. Useful for memory management in long-running applications.

**Example:**
```go
// Clear cache periodically in long-running applications
sqlite_regexp.ClearRegexpCache()
```

#### `GetCacheSize() int`

Returns the number of compiled regular expressions currently in the cache.

**Returns:**
- `int`: Number of cached regular expressions

**Example:**
```go
size := sqlite_regexp.GetCacheSize()
fmt.Printf("Cache contains %d compiled patterns\n", size)
```

### REGEXP SQL Function

Once registered, the REGEXP function is available in SQL queries:

```sql
REGEXP(pattern, text) -> INTEGER
```

**Parameters:**
- `pattern`: Regular expression pattern (string)
- `text`: Text to match against (string)

**Returns:**
- `1`: If pattern matches the text
- `0`: If pattern does not match the text

**Note:** The function can also be used with the infix operator syntax: `text REGEXP pattern`


## Examples

### Example 1: Pattern-Based Categorization with JOINs

This example demonstrates the exact use case mentioned in the package description - using REGEXP for JOINs to categorize items based on patterns:

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
    sqlite_regexp "github.com/go-go-golems/go-sqlite-regexp"
)

func main() {
    // Open database with REGEXP function
    db, err := sqlite_regexp.OpenWithRegexp(":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create tables
    _, err = db.Exec(`
        CREATE TABLE patterns (pattern TEXT, category TEXT);
        CREATE TABLE items (item TEXT, amount REAL);
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Insert patterns for categorization
    patterns := [][]interface{}{
        {"^apple", "fruits"},
        {"^car", "vehicles"},
        {"book$", "literature"},
        {"phone|mobile", "electronics"},
        {"\\d+", "numbers"},
    }

    for _, p := range patterns {
        _, err = db.Exec("INSERT INTO patterns VALUES (?, ?)", p[0], p[1])
        if err != nil {
            log.Fatal(err)
        }
    }

    // Insert items to categorize
    items := [][]interface{}{
        {"apple pie", 12.50},
        {"car wash", 25.00},
        {"textbook", 89.99},
        {"smartphone", 699.99},
        {"item123", 15.75},
        {"mobile phone", 299.99},
    }

    for _, i := range items {
        _, err = db.Exec("INSERT INTO items VALUES (?, ?)", i[0], i[1])
        if err != nil {
            log.Fatal(err)
        }
    }

    // The magic: REGEXP JOIN for pattern-based categorization
    rows, err := db.Query(`
        SELECT  p.category,
                i.item,
                i.amount
        FROM    patterns   AS p
        JOIN    items      AS i
             ON i.item REGEXP p.pattern
        ORDER BY p.category, i.item
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Printf("%-12s %-15s %8s\n", "Category", "Item", "Amount")
    fmt.Println("----------------------------------------")

    for rows.Next() {
        var category, item string
        var amount float64
        rows.Scan(&category, &item, &amount)
        fmt.Printf("%-12s %-15s %8.2f\n", category, item, amount)
    }
}
```

**Output:**
```
Category     Item              Amount
----------------------------------------
electronics  mobile phone      299.99
electronics  smartphone        699.99
fruits       apple pie          12.50
literature   textbook           89.99
numbers      item123            15.75
vehicles     car wash           25.00
```

### Example 2: Data Validation and Filtering

```go
// Validate email addresses
rows, err := db.Query(`
    SELECT email 
    FROM users 
    WHERE email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
`)

// Find phone numbers in text
rows, err := db.Query(`
    SELECT content 
    FROM messages 
    WHERE content REGEXP '\b\d{3}-\d{3}-\d{4}\b'
`)

// Extract URLs from content
rows, err := db.Query(`
    SELECT url 
    FROM links 
    WHERE url REGEXP '^https?://[^\s]+$'
`)
```

### Example 3: Advanced Pattern Matching

```go
// Case-insensitive matching (use (?i) flag)
rows, err := db.Query(`
    SELECT name 
    FROM products 
    WHERE name REGEXP '(?i)^premium'
`)

// Multiple conditions with REGEXP
rows, err := db.Query(`
    SELECT * 
    FROM logs 
    WHERE message REGEXP 'error|warning|critical'
    AND timestamp REGEXP '2024-01-'
`)

// Complex pattern for parsing structured data
rows, err := db.Query(`
    SELECT data 
    FROM records 
    WHERE data REGEXP '^[A-Z]{2}\d{4}-[A-Z]{3}-\d{2}$'
`)
```


## Performance

### Regex Caching

This package automatically caches compiled regular expressions to improve performance:

- **First use**: Pattern is compiled and cached
- **Subsequent uses**: Cached pattern is reused
- **Memory management**: Use `ClearRegexpCache()` in long-running applications

### Performance Tips

1. **Use anchors when possible**: `^pattern$` is faster than `.*pattern.*`
2. **Avoid complex patterns in large datasets**: Consider pre-filtering data
3. **Monitor cache size**: Use `GetCacheSize()` to track memory usage
4. **Index your data**: Create indexes on columns used in WHERE clauses

### Benchmarks

Pattern compilation and caching provide significant performance benefits:

```
BenchmarkRegexpCached-8      1000000    1.2 μs/op
BenchmarkRegexpUncached-8      10000   120.0 μs/op
```

## Troubleshooting

### Common Issues

#### CGO Build Errors

**Problem**: Build fails with CGO-related errors
```
# github.com/mattn/go-sqlite3
exec: "gcc": executable file not found in $PATH
```

**Solution**: Install a C compiler
```bash
# Ubuntu/Debian
sudo apt install build-essential

# macOS
xcode-select --install

# Windows
# Install TDM-GCC or MinGW-w64
```

#### Invalid Regular Expression

**Problem**: SQL query fails with regex compilation error
```
Error: error parsing regexp: missing closing ]: `[a-z`
```

**Solution**: Escape special characters and validate patterns
```go
// Bad
pattern := "[a-z"

// Good
pattern := "[a-z]+"
```

#### Function Not Found

**Problem**: SQL error "no such function: REGEXP"
```
Error: no such function: REGEXP
```

**Solution**: Ensure REGEXP function is registered
```go
// Make sure you call one of these:
db, err := sqlite_regexp.OpenWithRegexp(":memory:")
// OR
err = sqlite_regexp.RegisterRegexpFunction(db)
```

#### Memory Issues in Long-Running Applications

**Problem**: Memory usage grows over time

**Solution**: Periodically clear the regex cache
```go
// Clear cache every hour in long-running services
ticker := time.NewTicker(time.Hour)
go func() {
    for range ticker.C {
        sqlite_regexp.ClearRegexpCache()
    }
}()
```

### Debugging

Enable SQLite debugging to troubleshoot issues:

```go
import "github.com/mattn/go-sqlite3"

// Enable SQLite extended result codes
db, err := sql.Open("sqlite3", "file:test.db?_extended_result_codes=true")
```

### Getting Help

1. **Check the examples**: Review the examples directory
2. **Read Go regexp documentation**: [regexp package](https://pkg.go.dev/regexp)
3. **SQLite REGEXP reference**: [SQLite documentation](https://sqlite.org/lang_expr.html)
4. **File an issue**: [GitHub Issues](https://github.com/go-go-golems/go-sqlite-regexp/issues)


## Compilation Instructions

### Building Your Application

When building applications that use this package, ensure CGO is enabled:

```bash
# Standard build (CGO enabled by default)
go build -o myapp main.go

# Explicitly enable CGO
CGO_ENABLED=1 go build -o myapp main.go

# Cross-compilation (requires appropriate C compiler)
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o myapp-linux main.go
```

### Docker Builds

For Docker builds, use a base image with build tools:

```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o myapp

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/myapp .
CMD ["./myapp"]
```

### Static Linking

For static binaries (useful for deployment):

```bash
CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o myapp main.go
```

## Testing

Run the test suite to verify functionality:

```bash
# Run all tests
go test -v

# Run tests with race detection
go test -race -v

# Run benchmarks
go test -bench=. -v

# Test with coverage
go test -cover -v
```

## Contributing

We welcome contributions! Here's how to get started:

### Development Setup

1. **Fork the repository**
2. **Clone your fork**:
   ```bash
   git clone https://github.com/yourusername/go-sqlite-regexp.git
   cd go-sqlite-regexp
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```
4. **Run tests**:
   ```bash
   go test -v
   ```

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. **Make your changes**
3. **Add tests** for new functionality
4. **Run the test suite**:
   ```bash
   go test -v
   go test -race -v
   ```
5. **Update documentation** if needed
6. **Commit your changes**:
   ```bash
   git commit -m "Add your feature description"
   ```
7. **Push and create a pull request**

### Code Style

- Follow standard Go formatting (`go fmt`)
- Add comments for exported functions
- Include examples in documentation
- Write comprehensive tests

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver for Go
- [Go regexp package](https://pkg.go.dev/regexp) - Regular expression implementation
- [SQLite](https://sqlite.org/) - The database engine

---

**Made with ❤️ by the go-go-golems team**

For more tools and utilities, visit [go-go-golems](https://github.com/go-go-golems).

