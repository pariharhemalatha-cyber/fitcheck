package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenPostgres connects to Supabase/Postgres and runs migrations.
func OpenPostgres(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := MigratePostgres(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Rebind converts SQLite-style ? placeholders to Postgres $1, $2, ...
func Rebind(q string) string {
	if !strings.Contains(q, "?") {
		return q
	}
	n := 1
	var b strings.Builder
	for _, c := range q {
		if c == '?' {
			b.WriteString(fmt.Sprintf("$%d", n))
			n++
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}
