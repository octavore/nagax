package migrate

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
)

type backend interface {
	Connect() (*sql.DB, error)
	Reset() error
	Drop() error
	Migrate() error
	UnappliedMigrations() ([]string, error)
	migrations() migrate.MigrationSource
}

// GetBackend selects the datasource identified by dbname in the config file and
// initializes the correct type of backend.
func (m *Module) GetBackend(dbname string) (backend, error) {
	ds, ok := m.config.Datasources[dbname]
	if !ok {
		return nil, fmt.Errorf("migrate: %q not configured", dbname)
	}

	migrations, err := m.getMigrationSource()
	if err != nil {
		return nil, err
	}

	// special case for parallelizing tests: add a suffix to the dbname
	if dbname == "test" {
		u, err := url.Parse(ds.DSN)
		if err != nil {
			return nil, err
		}
		database := strings.Trim(u.Path, "/")
		if m.suffixForTest == "" {
			m.suffixForTest = randomToken()
		}
		u.Path = database + "_" + m.suffixForTest
		ds.DSN = u.String()
	}
	switch ds.Driver {
	case "postgres":
		return &postgresBackend{ds, migrations}, nil
	case "mysql":
		return &mysqlBackend{ds, migrations}, nil
	}
	return nil, fmt.Errorf("migrate: unsupported driver %s", ds.Driver)
}
