package db

import (
	"embed"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func migrate(name string, db *sqlx.DB) error {
	slog.Info("Migrating database", "name", name)

	goose.SetBaseFS(embedMigrations)
	goose.SetTableName("migrations")

	err := goose.SetDialect("sqlite3")
	if err != nil {
		slog.Error("Could not find sqlite dialog for migrations", "error", err)
		return err
	}

	err = goose.Up(db.DB, "migrations")
	if err != nil {
		slog.Error("Failed to migrate database", "name", name, "error", err)
		return err
	}

	slog.Info("Migrated database successfully", "name", name)
	return nil
}
