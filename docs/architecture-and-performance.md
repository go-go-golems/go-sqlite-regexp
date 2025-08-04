---
Title: SQLite REGEXP Architecture and Performance Guide
Slug: architecture-performance
Short: Essential guide to go-sqlite-regexp performance optimization and avoiding N+1 query anti-patterns
Topics:
- architecture
- performance
- optimization
- patterns
- best-practices
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# SQLite REGEXP Architecture and Performance Guide

## Overview

The go-sqlite-regexp library enables pattern-based matching in SQLite through Go's `regexp` package. While powerful, it introduces specific performance characteristics that developers must understand to avoid severe performance degradation, particularly the N+1 query anti-pattern.

## Architecture Basics

The library registers a `REGEXP` function with SQLite that:
1. Accepts text input and regex pattern
2. Compiles patterns using Go's `regexp` package  
3. Caches compiled patterns automatically
4. Returns boolean results (1/0)

```go
// Primary usage
db, err := sqlite_regexp.OpenWithRegexp("database.db")

// SQL usage
SELECT * FROM table WHERE column REGEXP '^pattern.*'
```

**Pattern Caching:**
- First use: compilation + execution (~5ms)
- Subsequent uses: execution only (~0.01ms)
- Thread-safe with automatic caching
- Memory: ~200-500 bytes per cached pattern

## Critical Performance Anti-Pattern: N+1 Queries

The most dangerous performance issue occurs when executing individual REGEXP queries for each row in a dataset.

### Anti-Pattern Example (DANGEROUS)

```go
// PERFORMANCE KILLER: Individual query per transaction
func categorizeTransactions(db *sql.DB, transactions []Transaction) error {
    for _, tx := range transactions {
        if tx.Category == "" {
            // This executes a separate query for EACH transaction
            category, err := applyCategoryPattern(db, tx.Description)
            if err != nil {
                return err
            }
            tx.Category = category
        }
    }
    return nil
}

func applyCategoryPattern(db *sql.DB, description string) (string, error) {
    query := `
        SELECT category 
        FROM patterns 
        WHERE ? REGEXP pattern 
        ORDER BY priority DESC 
        LIMIT 1`
    
    var category string
    err := db.QueryRow(query, description).Scan(&category)
    return category, err
}
```

### Real-World Case Study

**Problem Scenario:**
- 9,576 banking transactions
- 5 category patterns  
- 6,495 uncategorized transactions (68%)
- Individual queries: 6,495 × 5 = 32,475 REGEXP operations
- Result: 5-10 minute processing time, application appears hung

**Root Causes:**
1. N+1 query pattern (individual database query per transaction)
2. Connection pool bottleneck (`SetMaxOpenConns(1)`)
3. Database I/O overhead multiplied by transaction count

## Optimization Strategies

### Strategy 1: Bulk Processing with SQL

Replace individual queries with single bulk operations using correlated subqueries.

```sql
-- Single query handles all pattern matching
UPDATE transactions 
SET category = (
    SELECT p.category 
    FROM patterns p 
    WHERE UPPER(transactions.description) REGEXP UPPER(p.pattern)
    ORDER BY p.priority DESC 
    LIMIT 1
)
WHERE (category = '' OR category IS NULL)
  AND EXISTS (
    SELECT 1 FROM patterns p 
    WHERE UPPER(transactions.description) REGEXP UPPER(p.pattern)
  )
```

**Benefits:**
- 95%+ performance improvement
- Single database round-trip
- Eliminates N+1 pattern
- Atomic operation

### Strategy 2: In-Memory Bulk Processing

For maximum performance, load patterns into memory and process in Go.

```go
type CategoryPattern struct {
    Pattern  *regexp.Regexp
    Category string
    Priority int
}

// Load patterns once
func loadCategoryPatterns(db *sql.DB) ([]CategoryPattern, error) {
    rows, err := db.Query(`
        SELECT pattern, category, priority 
        FROM patterns 
        ORDER BY priority DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var patterns []CategoryPattern
    for rows.Next() {
        var p CategoryPattern
        var patternStr string
        
        err := rows.Scan(&patternStr, &p.Category, &p.Priority)
        if err != nil {
            return nil, err
        }
        
        // Compile with case-insensitive flag
        compiled, err := regexp.Compile("(?i)" + patternStr)
        if err != nil {
            continue // Skip invalid patterns
        }
        
        p.Pattern = compiled
        patterns = append(patterns, p)
    }
    
    return patterns, nil
}

// Process in memory
func categorizeInMemory(transactions []Transaction, patterns []CategoryPattern) []TransactionUpdate {
    var updates []TransactionUpdate
    
    for _, tx := range transactions {
        if tx.Category != "" {
            continue
        }
        
        // Find first matching pattern (pre-sorted by priority)
        for _, pattern := range patterns {
            if pattern.Pattern.MatchString(tx.Description) {
                updates = append(updates, TransactionUpdate{
                    ID:       tx.ID,
                    Category: pattern.Category,
                })
                break
            }
        }
    }
    
    return updates
}
```

**Performance:**
- 99% reduction in database queries
- Native Go processing speeds
- Scalable to 100,000+ transactions

### Strategy 3: Connection Pool Optimization

```go
func configureOptimalDatabase(db *sql.DB) error {
    // Connection pool sizing  
    db.SetMaxOpenConns(10)                    // Allow parallelization
    db.SetMaxIdleConns(5)                     // Reduce overhead
    db.SetConnMaxLifetime(time.Hour)
    db.SetConnMaxIdleTime(10 * time.Minute)
    
    // SQLite optimizations
    pragmas := []string{
        "PRAGMA journal_mode = WAL",           // Concurrent reads
        "PRAGMA synchronous = NORMAL",         // Balance performance/durability
        "PRAGMA cache_size = -64000",          // 64MB cache
        "PRAGMA temp_store = memory",
    }
    
    for _, pragma := range pragmas {
        if _, err := db.Exec(pragma); err != nil {
            return fmt.Errorf("configuring %s: %w", pragma, err)
        }
    }
    
    return nil
}
```

**Connection Pool Guidelines:**

| Workload Type | MaxOpenConns | Rationale |
|---------------|--------------|-----------|
| Bulk Processing | 5-10 | Limited by CPU cores |
| Interactive Queries | 10-20 | Support concurrent users |
| Background Jobs | 2-5 | Prevent resource exhaustion |

## Pattern Design Best Practices

### Efficient Pattern Syntax

```go
// EFFICIENT: Anchored patterns
"^SALARY.*"              // Fast: anchored start
"MORTGAGE|HOME.*LOAN"    // Good: alternation

// INEFFICIENT: Greedy patterns  
".*SALARY.*"             // Slow: full text scan
"[A-Za-z0-9]*PAYMENT.*"  // Slow: complex character classes
```

### Case-Insensitive Optimization

```go
// EFFICIENT: Compile with flag
regexp.Compile("(?i)salary|payroll")

// INEFFICIENT: Runtime conversion
// SQL: WHERE UPPER(text) REGEXP UPPER(pattern)
```

### Pattern Priority

```go
// Order by frequency and specificity
patterns := []Pattern{
    {Pattern: "^SALARY.*", Priority: 200},     // Most common first
    {Pattern: "^PAYROLL.*", Priority: 199},    // Specific before general
    {Pattern: "INCOME", Priority: 100},        // General last
}
```

## Production Considerations

### Cache Management

```go
// Monitor cache size in long-running applications
func manageCacheSize() {
    size := sqlite_regexp.GetCacheSize()
    if size > 1000 {  // Configurable threshold
        sqlite_regexp.ClearRegexpCache()
    }
}
```

### Performance Monitoring

```go
// Key metrics to track
type RegexpMetrics struct {
    CacheSize         int           `json:"cache_size"`
    MatchOperations   int64         `json:"match_operations"`
    AverageMatchTime  time.Duration `json:"avg_match_time"`
    CacheHitRate     float64       `json:"cache_hit_rate"`
}
```

**Alerting Thresholds:**
- Cache size: Warning at 1000 patterns, Critical at 5000
- Average match time: Warning at 10ms, Critical at 50ms
- Cache hit rate: Warning below 80%, Critical below 50%

### Error Handling

```go
func safeRegexpMatch(db *sql.DB, text, pattern string) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    var result bool
    err := db.QueryRowContext(ctx, "SELECT ? REGEXP ?", text, pattern).Scan(&result)
    if err != nil {
        // Fallback to Go regexp for critical operations
        if compiled, compileErr := regexp.Compile(pattern); compileErr == nil {
            return compiled.MatchString(text), nil
        }
        return false, err
    }
    
    return result, nil
}
```

## Troubleshooting Common Issues

### Application Hangs During REGEXP Operations

**Symptoms:** Unresponsive application, high CPU, database timeouts

**Solutions:**
1. Check for N+1 query patterns
2. Implement bulk processing
3. Increase connection pool size
4. Add query timeouts

### Memory Usage Growth

**Symptoms:** Increasing memory usage, eventual OOM errors

**Solutions:**
1. Implement periodic cache clearing
2. Limit pattern diversity  
3. Monitor cache growth
4. Use pattern validation

### Inconsistent Results

**Symptoms:** Same pattern produces different results

**Solutions:**
1. Use consistent case handling
2. Validate pattern syntax
3. Test with representative data
4. Use UTF-8 encoding consistently

## Key Takeaways

1. **Avoid N+1 Queries**: Always prefer bulk processing over individual pattern queries
2. **Optimize Connection Pools**: Use appropriate pool sizing (5-10 connections for REGEXP workloads)
3. **Design Efficient Patterns**: Use anchored, specific patterns with case-insensitive compilation
4. **Monitor Cache Usage**: Implement cache management for long-running applications
5. **Test at Scale**: Validate performance with realistic data volumes

**Implementation Priority:**
1. **Immediate**: Fix N+1 query patterns with bulk processing
2. **Short-term**: Optimize connection pool settings
3. **Long-term**: Add monitoring and cache management

Following these guidelines will ensure excellent performance while maintaining the flexibility of pattern-based data processing in production applications.
