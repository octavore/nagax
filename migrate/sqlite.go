package migrate

import (
	"database/sql"
	"errors"
	"os"

	migrate "github.com/rubenv/sql-migrate"
)

type sqliteBackend struct {
	Datasource
	migrate.MigrationSource
}

func (m *sqliteBackend) connect() (*sql.DB, error) {
	return sql.Open("sqlite", m.Datasource.DSN)
}

// Reset drops and recreates the database
func (m *sqliteBackend) Reset() error {
	if _, err := os.Stat(m.DSN); os.IsNotExist(err) {
		db, err := m.connect()
		if err != nil {
			return err
		}
		defer db.Close()
		_, err = db.Exec(`VACUUM`)
		return err
	}

	return nil
}

// Drop the database `dbname`
func (m *sqliteBackend) Drop() error {
	return errors.New("not implemented; please delete the database manually")
}

func (m *sqliteBackend) Migrate() error {
	return m.Datasource.migrate(m.MigrationSource)
}

func (m *sqliteBackend) UnappliedMigrations() ([]string, error) {
	return m.Datasource.unappliedMigrations(m.MigrationSource)
}

func (m *sqliteBackend) migrations() migrate.MigrationSource {
	return m.MigrationSource
}
