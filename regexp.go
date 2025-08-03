// Package sqlite_regexp provides a REGEXP function for SQLite databases using go-sqlite3.
// This package allows you to perform regular expression matching in SQLite queries,
// enabling powerful pattern-based JOINs and WHERE clauses.
package sqlite_regexp

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"sync"

	"github.com/mattn/go-sqlite3"
)

// regexpCache caches compiled regular expressions to improve performance
var regexpCache = struct {
	sync.RWMutex
	cache map[string]*regexp.Regexp
}{
	cache: make(map[string]*regexp.Regexp),
}

// regexpFunction implements the REGEXP function for SQLite.
// It takes two arguments: the text to match and the pattern.
// Returns 1 if the pattern matches, 0 otherwise.
func regexpFunction(pattern, text string) (int, error) {
	// Check cache first
	regexpCache.RLock()
	re, exists := regexpCache.cache[pattern]
	regexpCache.RUnlock()

	if !exists {
		// Compile the regex and cache it
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return 0, err
		}

		regexpCache.Lock()
		regexpCache.cache[pattern] = re
		regexpCache.Unlock()
	}

	if re.MatchString(text) {
		return 1, nil
	}
	return 0, nil
}

// RegisterRegexpFunction registers the REGEXP function with a SQLite connection.
// This function should be called after opening a database connection but before
// executing any queries that use REGEXP.
func RegisterRegexpFunction(db *sql.DB) error {
	// Get the underlying SQLite connection with proper context
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Raw(func(driverConn interface{}) error {
		sqliteConn, ok := driverConn.(*sqlite3.SQLiteConn)
		if !ok {
			return driver.ErrBadConn
		}

		// Register the REGEXP function
		return sqliteConn.RegisterFunc("regexp", regexpFunction, true)
	})
}

// OpenWithRegexp opens a SQLite database connection and automatically registers
// the REGEXP function. This is a convenience function that combines sql.Open
// with RegisterRegexpFunction.
func OpenWithRegexp(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := RegisterRegexpFunction(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// ClearRegexpCache clears the internal regexp cache. This can be useful
// for memory management in long-running applications.
func ClearRegexpCache() {
	regexpCache.Lock()
	regexpCache.cache = make(map[string]*regexp.Regexp)
	regexpCache.Unlock()
}

// GetCacheSize returns the number of compiled regular expressions in the cache.
func GetCacheSize() int {
	regexpCache.RLock()
	size := len(regexpCache.cache)
	regexpCache.RUnlock()
	return size
}

