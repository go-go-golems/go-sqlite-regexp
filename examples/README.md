# Examples

This directory contains examples demonstrating how to use the go-sqlite-regexp package.

## Running the Examples

### Method 1: Using Make (Recommended)

From the root directory:

```bash
# Build and run examples
make run-examples

# Just build examples
make examples
```

### Method 2: Manual Build

From this directory:

```bash
# Initialize module (if not done already)
go mod init examples
go mod edit -replace github.com/go-go-golems/go-sqlite-regexp=../
go get github.com/mattn/go-sqlite3
go get github.com/go-go-golems/go-sqlite-regexp

# Build and run
go build -o example example.go
./example
```

## Example Output

When you run the example, you should see output similar to:

```
=== REGEXP JOIN Example ===
Query: SELECT p.category, i.item, i.amount FROM patterns AS p JOIN items AS i ON i.item REGEXP p.pattern

Category     Item              Amount
----------------------------------------
electronics  mobile phone      299.99
electronics  smartphone        699.99
fruits       apple juice         3.99
fruits       apple pie          12.50
literature   cookbook           24.99
literature   textbook           89.99
numbers      item123            15.75
vehicles     car rental        150.00
vehicles     car wash           25.00
vehicles     carrot cake         8.50
```

## What the Example Demonstrates

The example shows:

1. **Database Setup**: Creating an in-memory SQLite database
2. **Function Registration**: Automatically registering the REGEXP function
3. **Table Creation**: Setting up `patterns` and `items` tables
4. **Data Insertion**: Adding sample patterns and items
5. **REGEXP JOIN**: Using regular expressions in JOIN conditions to categorize items

This demonstrates the core use case: pattern-based categorization using flexible regular expression matching instead of exact string matches.

## Customizing the Example

You can modify the example to test different patterns:

- Add new patterns to the `patterns` array
- Add new items to the `items` array
- Try different regular expressions
- Test with your own data

## Troubleshooting

If you encounter build errors:

1. **CGO Issues**: Ensure you have a C compiler installed
   ```bash
   # Ubuntu/Debian
   sudo apt install build-essential
   
   # macOS
   xcode-select --install
   ```

2. **Module Issues**: Make sure you're in the examples directory and have run the module setup commands

3. **Path Issues**: Ensure the replace directive points to the correct parent directory

