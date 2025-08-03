package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sqlite_regexp "github.com/go-go-golems/go-sqlite-regexp"
)

func main() {
	// Open database with REGEXP function registered
	db, err := sqlite_regexp.OpenWithRegexp(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	if err := createTables(db); err != nil {
		log.Fatal(err)
	}

	// Insert sample data
	if err := insertSampleData(db); err != nil {
		log.Fatal(err)
	}

	// Demonstrate REGEXP JOIN
	if err := demonstrateRegexpJoin(db); err != nil {
		log.Fatal(err)
	}
}

func createTables(db *sql.DB) error {
	// Create patterns table
	_, err := db.Exec(`
		CREATE TABLE patterns (
			pattern TEXT,
			category TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create items table
	_, err = db.Exec(`
		CREATE TABLE items (
			item TEXT,
			amount REAL
		)
	`)
	return err
}

func insertSampleData(db *sql.DB) error {
	// Insert patterns
	patterns := []struct {
		pattern  string
		category string
	}{
		{"^apple", "fruits"},
		{"^car", "vehicles"},
		{"book$", "literature"},
		{"phone|mobile", "electronics"},
		{"\\d+", "numbers"},
	}

	for _, p := range patterns {
		_, err := db.Exec("INSERT INTO patterns (pattern, category) VALUES (?, ?)", p.pattern, p.category)
		if err != nil {
			return err
		}
	}

	// Insert items
	items := []struct {
		item   string
		amount float64
	}{
		{"apple pie", 12.50},
		{"car wash", 25.00},
		{"textbook", 89.99},
		{"smartphone", 699.99},
		{"item123", 15.75},
		{"carrot cake", 8.50},
		{"mobile phone", 299.99},
		{"cookbook", 24.99},
		{"apple juice", 3.99},
		{"car rental", 150.00},
	}

	for _, i := range items {
		_, err := db.Exec("INSERT INTO items (item, amount) VALUES (?, ?)", i.item, i.amount)
		if err != nil {
			return err
		}
	}

	return nil
}

func demonstrateRegexpJoin(db *sql.DB) error {
	fmt.Println("=== REGEXP JOIN Example ===")
	fmt.Println("Query: SELECT p.category, i.item, i.amount FROM patterns AS p JOIN items AS i ON i.item REGEXP p.pattern")
	fmt.Println()

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
		return err
	}
	defer rows.Close()

	fmt.Printf("%-12s %-15s %8s\n", "Category", "Item", "Amount")
	fmt.Println("----------------------------------------")

	for rows.Next() {
		var category, item string
		var amount float64
		if err := rows.Scan(&category, &item, &amount); err != nil {
			return err
		}
		fmt.Printf("%-12s %-15s %8.2f\n", category, item, amount)
	}

	return rows.Err()
}

