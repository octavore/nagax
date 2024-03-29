package migrate

import (
	"database/sql"

	"github.com/octavore/naga/service"
	migrate "github.com/rubenv/sql-migrate"
	backoff "gopkg.in/cenkalti/backoff.v2"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
)

func init() {
	migrate.SetTable("schema_migrations")
}

// Module migrate provides support for migrating postgres databases
// using rubenv/sql-migrate
type Module struct {
	Config *config.Module
	DB     *sql.DB
	Logger *logger.Module

	config          Config
	migrationSource migrate.MigrationSource
	backoff         backoff.BackOff

	suffixForTest string
	env           service.Environment
}

// Config for migrate module
type Config struct {
	Datasources     map[string]Datasource `json:"datasources"`
	MigrationsDir   string                `json:"migrations"`
	MigrationsTable string                `json:"migrations_table"`
}

// Datasource is parsed from the config
type Datasource struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// Init the migrate module
func (m *Module) Init(c *service.Config) {
	m.registerCommands(c)

	// alias sqlite dialect
	migrate.MigrationDialects["sqlite"] = migrate.MigrationDialects["sqlite3"]

	c.Setup = func() error {
		m.env = c.Env()
		m.backoff = &backoff.StopBackOff{}
		err := m.Config.ReadConfig(&m.config)
		if m.config.MigrationsTable != "" {
			migrate.SetTable(m.config.MigrationsTable)
		}
		return err
	}
}

// ConnectDefault to the DB with name specified by env
func (m *Module) ConnectDefault() (*sql.DB, error) {
	ds, err := m.GetBackend(m.env.String())
	if err != nil {
		return nil, err
	}
	return ds.Connect()
}

// Connect is a helper function to connect to this datasource
func (d *Datasource) Connect() (*sql.DB, error) {
	return sql.Open(d.Driver, d.DSN)
}

func (d *Datasource) unappliedMigrations(m migrate.MigrationSource) ([]string, error) {
	db, err := d.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	plans, _, err := migrate.PlanMigration(db, d.Driver, m, migrate.Up, 0)
	if err != nil {
		return nil, err
	}
	migrations := []string{}
	for _, plan := range plans {
		migrations = append(migrations, plan.Id)
	}
	return migrations, nil
}

// Migrate runs migrations in m
func (d *Datasource) migrate(m migrate.MigrationSource) error {
	db, err := d.Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = migrate.Exec(db, d.Driver, m, migrate.Up)
	return err
}
