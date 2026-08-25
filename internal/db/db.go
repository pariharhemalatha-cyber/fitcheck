package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const DefaultUserID = "local"

// Open opens (or creates) a SQLite database at path and runs migrations.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := Migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
