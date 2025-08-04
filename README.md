# go-sqlite-regexp

A Go package that adds REGEXP functionality to SQLite databases. This package enables pattern-based matching in SQL queries, allowing for flexible JOINs and WHERE clauses using Go's regular expression syntax.

## Quick Start

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

    // Create test data
    _, err = db.Exec(`
        CREATE TABLE items (name TEXT);
        INSERT INTO items VALUES ('apple'), ('banana'), ('apricot');
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Use REGEXP in queries
    rows, err := db.Query("SELECT name FROM items WHERE name REGEXP '^ap'")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var name string
        rows.Scan(&name)
        fmt.Println(name) // apple, apricot
    }
}
```

## Installation

Install the package with Go modules:

```bash
go get github.com/go-go-golems/go-sqlite-regexp
```

**Requirements:**
- Go 1.21 or later
- CGO enabled (required for go-sqlite3)
- C compiler (GCC, Clang, or equivalent)

On Ubuntu/Debian: `sudo apt install build-essential`
On macOS: `xcode-select --install`

## Usage

### Opening a Database

The simplest approach uses `OpenWithRegexp` which automatically registers the REGEXP function:

```go
db, err := sqlite_regexp.OpenWithRegexp("database.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// REGEXP function is now available in SQL queries
```

### Manual Registration

For existing database connections, register the function manually:

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

### REGEXP Syntax

The REGEXP function uses Go's RE2 regular expression syntax:

```sql
-- Basic usage
SELECT * FROM users WHERE name REGEXP '^John'

-- Common patterns
WHERE email REGEXP '@gmail\.com$'        -- ends with @gmail.com
WHERE code REGEXP '\d{3}-\d{3}-\d{4}'    -- phone number format
WHERE category REGEXP 'fruit|vegetable'  -- contains either word
```

### Pattern-Based JOINs

REGEXP enables flexible data categorization through pattern matching in JOINs:

```sql
-- Join items to categories based on pattern matching
SELECT c.name, i.item, i.amount
FROM categories c
JOIN items i ON i.item REGEXP c.pattern

-- Example data:
-- categories: ('Electronics', '^(phone|laptop|tablet)')
-- items: ('phone-case', 'laptop-bag', 'apple')
-- Result: Electronics matches phone-case and laptop-bag
```

## API Reference

### Core Functions

**`OpenWithRegexp(dataSourceName string) (*sql.DB, error)`**
Opens a SQLite database and registers the REGEXP function.

**`RegisterRegexpFunction(db *sql.DB) error`**  
Registers REGEXP function with an existing database connection.

### Cache Management

**`ClearRegexpCache()`**  
Clears the internal regex cache. Use in long-running applications to manage memory.

**`GetCacheSize() int`**  
Returns the number of cached compiled patterns.

```go
// Monitor cache usage
fmt.Printf("Cache size: %d patterns\n", sqlite_regexp.GetCacheSize())

// Clear cache periodically in long-running services
go func() {
    ticker := time.NewTicker(time.Hour)
    for range ticker.C {
        sqlite_regexp.ClearRegexpCache()
    }
}()
```

## Performance

Regular expressions are automatically cached for performance. First use compiles and caches the pattern; subsequent uses reuse the cached pattern.

**Tips for better performance:**
- Use anchors when possible: `^pattern$` vs `.*pattern.*`
- Avoid complex patterns on large datasets
- Monitor cache size with `GetCacheSize()`
- Create database indexes on columns used in WHERE clauses

## Troubleshooting

### CGO Build Errors

**Error:** `exec: "gcc": executable file not found`

**Solution:** Install a C compiler:
```bash
# Ubuntu/Debian
sudo apt install build-essential

# macOS  
xcode-select --install
```

### Function Not Found

**Error:** `no such function: REGEXP`

**Solution:** Ensure REGEXP function is registered:
```go
// Use either approach:
db, err := sqlite_regexp.OpenWithRegexp(":memory:")
// OR
err = sqlite_regexp.RegisterRegexpFunction(db)
```

### Invalid Patterns

**Error:** `error parsing regexp: missing closing ]`

**Solution:** Validate regex patterns before use:
```go
// Test patterns separately
_, err := regexp.Compile("[a-z]+")  // Valid
_, err := regexp.Compile("[a-z")    // Invalid - missing closing bracket
```

## Building

Standard Go build with CGO enabled:

```bash
go build -o myapp main.go

# For static binaries
CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o myapp main.go
```

### Docker

Use a base image with build tools:

```dockerfile
FROM golang:1.21-alpine AS builder
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

## Testing

```bash
go test -v              # Run tests
go test -race -v        # Test with race detection  
go test -bench=. -v     # Run benchmarks
go test -cover -v       # Test with coverage
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
