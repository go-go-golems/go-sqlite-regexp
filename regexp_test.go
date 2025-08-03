package sqlite_regexp

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRegexpFunction(t *testing.T) {
	tests := []struct {
		pattern  string
		text     string
		expected int
	}{
		{"^hello", "hello world", 1},
		{"^hello", "world hello", 0},
		{"world$", "hello world", 1},
		{"world$", "world hello", 0},
		{"\\d+", "abc123def", 1},
		{"\\d+", "abcdef", 0},
		{"[a-z]+", "ABC", 0},
		{"[a-z]+", "abc", 1},
		{"phone|mobile", "smartphone", 1},
		{"phone|mobile", "tablet", 0},
	}

	for _, test := range tests {
		result, err := regexpFunction(test.pattern, test.text)
		if err != nil {
			t.Errorf("regexpFunction(%q, %q) returned error: %v", test.pattern, test.text, err)
			continue
		}
		if result != test.expected {
			t.Errorf("regexpFunction(%q, %q) = %d, expected %d", test.pattern, test.text, result, test.expected)
		}
	}
}

func TestRegexpFunctionInvalidPattern(t *testing.T) {
	_, err := regexpFunction("[", "test")
	if err == nil {
		t.Error("Expected error for invalid regex pattern, got nil")
	}
}

func TestOpenWithRegexp(t *testing.T) {
	db, err := OpenWithRegexp(":memory:")
	if err != nil {
		t.Fatalf("OpenWithRegexp failed: %v", err)
	}
	defer db.Close()

	// Test that REGEXP function is available
	var result int
	err = db.QueryRow("SELECT 'hello world' REGEXP '^hello'").Scan(&result)
	if err != nil {
		t.Fatalf("REGEXP query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}
}

func TestRegexpJoin(t *testing.T) {
	db, err := OpenWithRegexp(":memory:")
	if err != nil {
		t.Fatalf("OpenWithRegexp failed: %v", err)
	}
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE patterns (pattern TEXT, category TEXT);
		CREATE TABLE items (item TEXT, amount REAL);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO patterns VALUES ('^apple', 'fruits'), ('book$', 'literature');
		INSERT INTO items VALUES ('apple pie', 10.0), ('textbook', 50.0), ('orange juice', 5.0);
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test REGEXP JOIN
	rows, err := db.Query(`
		SELECT p.category, i.item, i.amount 
		FROM patterns p 
		JOIN items i ON i.item REGEXP p.pattern
		ORDER BY p.category, i.item
	`)
	if err != nil {
		t.Fatalf("REGEXP JOIN query failed: %v", err)
	}
	defer rows.Close()

	expected := []struct {
		category string
		item     string
		amount   float64
	}{
		{"fruits", "apple pie", 10.0},
		{"literature", "textbook", 50.0},
	}

	var results []struct {
		category string
		item     string
		amount   float64
	}

	for rows.Next() {
		var category, item string
		var amount float64
		if err := rows.Scan(&category, &item, &amount); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		results = append(results, struct {
			category string
			item     string
			amount   float64
		}{category, item, amount})
	}

	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("Result %d: expected %+v, got %+v", i, expected[i], result)
		}
	}
}

func TestRegexpCache(t *testing.T) {
	// Clear cache first
	ClearRegexpCache()
	if GetCacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", GetCacheSize())
	}

	// Test cache population
	_, err := regexpFunction("test", "test")
	if err != nil {
		t.Fatalf("regexpFunction failed: %v", err)
	}

	if GetCacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", GetCacheSize())
	}

	// Test cache reuse (same pattern)
	_, err = regexpFunction("test", "another test")
	if err != nil {
		t.Fatalf("regexpFunction failed: %v", err)
	}

	if GetCacheSize() != 1 {
		t.Errorf("Expected cache size still 1, got %d", GetCacheSize())
	}

	// Test cache growth (different pattern)
	_, err = regexpFunction("different", "test")
	if err != nil {
		t.Fatalf("regexpFunction failed: %v", err)
	}

	if GetCacheSize() != 2 {
		t.Errorf("Expected cache size 2, got %d", GetCacheSize())
	}
}

func TestRegisterRegexpFunction(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Register REGEXP function
	err = RegisterRegexpFunction(db)
	if err != nil {
		t.Fatalf("RegisterRegexpFunction failed: %v", err)
	}

	// Test that function is registered
	var result int
	err = db.QueryRow("SELECT 'test123' REGEXP '\\d+'").Scan(&result)
	if err != nil {
		t.Fatalf("REGEXP query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}
}

