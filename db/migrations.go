package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB) error {
	// Ensure migrations table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Load filenames
	files, err := readMigrationFiles("./db/migrations")
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := applyOne(db, f); err != nil {
			return err
		}
	}

	return nil
}

func readMigrationFiles(dir string) ([]string, error) {
	var out []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".sql") {
			out = append(out, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(out)
	return out, nil
}

func applyOne(db *sql.DB, path string) error {
	id := filepath.Base(path)

	// Has this migration been applied?
	var exists bool
	err := db.QueryRow("SELECT 1 FROM migrations WHERE id = ?", id).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("checking migration %s: %w", id, err)
	}
	if exists {
		return nil
	}

	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("apply migration %s: %w", path, err)
	}

	_, err = db.Exec("INSERT INTO migrations(id) VALUES (?)", id)
	if err != nil {
		return fmt.Errorf("record migration %s: %w", id, err)
	}

	return nil
}
