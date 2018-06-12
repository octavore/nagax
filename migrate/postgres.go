package migrate

import (
	"database/sql"
	"net/url"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
)

type postgresBackend struct {
	Datasource
	migrate.MigrationSource
}

// safeConnect connects to template1 so we can create/drop the desired database.
func (p *postgresBackend) safeConnect() (string, *sql.DB, error) {
	u, err := url.Parse(p.Datasource.DSN)
	if err != nil {
		return "", nil, err
	}

	database := strings.Trim(u.Path, "/")
	u.Path = "template1"
	u.RawPath = "template1"

	db, err := sql.Open(p.Datasource.Driver, u.String())
	if err != nil {
		return "", nil, err
	}
	return database, db, nil
}

// Reset drops and recreates the database
func (p *postgresBackend) Reset() error {
	err := p.Drop()
	if err != nil {
		return err
	}

	databaseName, db, err := p.safeConnect()
	if err != nil {
		return err
	}
	defer db.Close()

	// using template0 in order to support test parallelism
	// cf http://stackoverflow.com/questions/4977171/pgerror-error-source-database-template1-is-being-accessed-by-other-users
	// you may be able to hack around by creating some kind of global lock to protect
	// connections to the template1 database?
	// or maybe drop the connection as soon as possible?
	_, err = db.Exec(`CREATE DATABASE ` + databaseName + ` TEMPLATE template0`)
	return err
}

// Drop the database `dbname`
func (p *postgresBackend) Drop() error {
	databaseName, db, err := p.safeConnect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DROP DATABASE IF EXISTS ` + databaseName)
	return err
}

// Migrate the given db

func (p *postgresBackend) Migrate() error {
	return p.Datasource.migrate(p.MigrationSource)
}
