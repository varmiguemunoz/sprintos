package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

type migrator struct {
	db *sql.DB
}

func newMigrator(gormDB *gorm.DB) (*migrator, error) {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}
	return &migrator{db: sqlDB}, nil
}

func (m *migrator) Run() error {
	if err := m.ensureTrackingTable(); err != nil {
		return err
	}
	pending, err := m.pendingMigrations()
	if err != nil {
		return err
	}
	for _, name := range pending {
		if err := m.apply(name); err != nil {
			return fmt.Errorf("migration %s failed: %w", name, err)
		}
	}
	return nil
}

func (m *migrator) ensureTrackingTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id         TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (m *migrator) pendingMigrations() ([]string, error) {
	entries, err := fs.Glob(migrationFiles, "migrations/*.up.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)

	rows, err := m.db.Query("SELECT id FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		applied[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var pending []string
	for _, path := range entries {
		name := strings.TrimPrefix(path, "migrations/")
		if !applied[name] {
			pending = append(pending, name)
		}
	}
	return pending, nil
}

func (m *migrator) apply(name string) error {
	content, err := migrationFiles.ReadFile("migrations/" + name)
	if err != nil {
		return err
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return err
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (id, applied_at) VALUES ($1, NOW())",
		name,
	); err != nil {
		return err
	}

	return tx.Commit()
}
