package database

import (
	"database/sql"
	"embed"
	"fmt"

	goose "github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Fascade, error) {
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on&_journal_mode=WAL&_cache_size=-64000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewFascade(&DB{db}), nil
}

type Fascade struct {
	innerDb *DB
}

func NewFascade(db *DB) *Fascade {
	return &Fascade{
		innerDb: db,
	}
}

func (df *Fascade) QueryRow(query string, args ...any) *sql.Row {
	return df.innerDb.QueryRow(query, args...)
}

func (df *Fascade) Query(query string, args ...any) (*sql.Rows, error) {
	return df.innerDb.Query(query, args...)
}

func (df *Fascade) Exec(query string, args ...any) (sql.Result, error) {
	return df.innerDb.Exec(query, args...)
}

func (df *Fascade) BeginCommit(vFunc func(*sql.Tx) error) error {
	tx, err := df.innerDb.Begin()
	if err != nil {
		return err
	}
	return vFunc(tx)
}

func (df *Fascade) Close() error {
	return df.innerDb.Close()
}

// MigrateUp runs all available migrations
func (df *Fascade) MigrateUp() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(df.innerDb.DB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back one migration
func (df *Fascade) MigrateDown() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Down(df.innerDb.DB, "migrations"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func (df *Fascade) GetMigrationStatus() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Status(df.innerDb.DB, "migrations"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}
