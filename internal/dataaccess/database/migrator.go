package database

import (
	"context"
	"database/sql"
	"embed"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

var (
	// go:embed migrations/mysql
	migrationDirectoryMySQL embed.FS
)

type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}
type migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) Migrator {
	return &migrator{
		db: db,
	}
}
func (m migrator) migrate(ctx context.Context, direction migrate.MigrationDirection) error {
	_, err := migrate.ExecContext(ctx, m.db, "mysql", migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationDirectoryMySQL,
		Root:       "migrations/mysql",
	}, direction)
	if err != nil {
		log.Printf("failed to execute migration")
		return err
	}
	log.Printf("successfully executed database migrations")
	return nil
}
func (m migrator) Down(ctx context.Context) error {
	return m.migrate(ctx, migrate.Down)
}
func (m migrator) Up(ctx context.Context) error {
	return m.migrate(ctx, migrate.Up)
}
