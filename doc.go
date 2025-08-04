// Package sqlite_regexp provides REGEXP functionality for SQLite databases using go-sqlite3.
//
// This package enables powerful regular expression matching in SQLite queries,
// allowing for pattern-based JOINs and WHERE clauses. It automatically caches
// compiled regular expressions for improved performance.
//
// # Quick Start
//
// The simplest way to use this package is with OpenWithRegexp:
//
//	db, err := sqlite_regexp.OpenWithRegexp(":memory:")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Now you can use REGEXP in SQL queries
//	rows, err := db.Query("SELECT name FROM users WHERE name REGEXP '^John'")
//
// # Pattern-Based JOINs
//
// One of the most powerful features is using REGEXP in JOIN conditions:
//
//	SELECT  p.category,
//	        i.item,
//	        i.amount
//	FROM    patterns   AS p
//	JOIN    items      AS i
//	     ON i.item REGEXP p.pattern;
//
// This allows you to categorize items based on flexible pattern matching
// rather than exact string matches.
//
// # Performance
//
// The package automatically caches compiled regular expressions to improve
// performance. For long-running applications, you can manage the cache:
//
//	// Check cache size
//	size := sqlite_regexp.GetCacheSize()
//
//	// Clear cache to free memory
//	sqlite_regexp.ClearRegexpCache()
//
// # Regular Expression Syntax
//
// The package uses Go's regexp package (RE2 syntax). Common patterns include:
//
//	^text     - Starts with "text"
//	text$     - Ends with "text"
//	\d+       - Contains digits
//	[a-z]+    - Contains lowercase letters
//	text1|text2 - Contains "text1" OR "text2"
//
// For complete syntax reference, see: https://pkg.go.dev/regexp/syntax
package sqlite_regexp
