package migrate

import (
	"database/sql"
	"fmt"
	"regexp"

	migrate "github.com/rubenv/sql-migrate"
)

type mysqlBackend struct {
	Datasource
	migrate.MigrationSource
}

// connectWithoutDB connects without a db so we can create/drop the desired database.
func (m *mysqlBackend) connectWithoutDB() (string, *sql.DB, error) {
	re := regexp.MustCompile("(.*)/([^?]+)")
	matches := re.FindStringSubmatch(m.Datasource.DSN)
	if len(matches) != 3 {
		return "", nil, fmt.Errorf("migrate: error parsing mysql dsn")
	}
	dsn := matches[1] + "/"
	database := matches[2]

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", nil, err
	}
	return database, db, err
}

// Reset drops and recreates the database
func (m *mysqlBackend) Reset() error {
	err := m.Drop()
	if err != nil {
		return err
	}
	databaseName, db, err := m.connectWithoutDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE DATABASE ` + databaseName)
	return err
}

// Drop the database `dbname`
func (m *mysqlBackend) Drop() error {
	databaseName, db, err := m.connectWithoutDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DROP DATABASE IF EXISTS ` + databaseName)
	return err
}

func (m *mysqlBackend) Migrate() error {
	return m.Datasource.migrate(m.MigrationSource)
}
