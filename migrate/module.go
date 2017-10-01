package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/octavore/naga/service"
	"github.com/rubenv/sql-migrate"

	"github.com/octavore/nagax/config"
)

func init() {
	migrate.SetTable("schema_migrations")
}

// Module migrate provides support for migrating postgres databases
// using rubenv/sql-migrate
type Module struct {
	Config *config.Module
	DB     *sql.DB

	config          Config
	migrationSource migrate.MigrationSource

	suffixForTest string
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

func (m *Module) printHelp(ctx *service.CommandContext) {
	if len(ctx.Args) != 1 {
		fmt.Println("Please specify a db:")
		if len(m.config.Datasources) == 0 {
			fmt.Println("  no databases found!")
		} else {
			for ds := range m.config.Datasources {
				fmt.Println("  " + ds)
			}
		}
		ctx.UsageExit()
	}
}

func (m *Module) Init(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword:    "db:migrate <db>",
		ShortUsage: "run db migrations",
		Run: func(ctx *service.CommandContext) {
			if len(ctx.Args) != 1 {
				m.printHelp(ctx)
			}
			err := m.Migrate(ctx.Args[0])
			if err != nil {
				log.Println("migrate:", err)
			}
		},
	})

	c.AddCommand(&service.Command{
		Keyword:    "db:reset <db>",
		ShortUsage: "reset database",
		Run: func(ctx *service.CommandContext) {
			if len(ctx.Args) != 1 {
				m.printHelp(ctx)
			}
			dbname := ctx.Args[0]
			err := m.Reset(dbname)
			if err != nil {
				log.Println("migrate:", err)
			}
			err = m.Migrate(dbname)
			if err != nil {
				log.Println("migrate:", err)
			}
		},
	})

	c.Setup = func() error {
		err := m.Config.ReadConfig(&m.config)
		if m.config.MigrationsTable != "" {
			migrate.SetTable(m.config.MigrationsTable)
		}
		return err
	}

	c.SetupTest = func() {
	}
}

func (m *Module) getConfig(dbname string) (*Datasource, error) {
	ds, ok := m.config.Datasources[dbname]
	if !ok {
		return nil, fmt.Errorf("migrate: %q not configured", dbname)
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
	return &ds, nil
}

// Connect to the given DB
func (m *Module) Connect(dbname string) (*sql.DB, error) {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return nil, err
	}
	return sql.Open(ds.Driver, ds.DSN)
}

func (m *Module) AddMigrations(migrationsDir string) {
	panic("todo")
}

type (
	assetFunc    func(path string) ([]byte, error)
	assetDirFunc func(path string) ([]string, error)
)

// SetMigrationSource sets the migration source, for compatibility with
// embedded file assets.
func (m *Module) SetMigrationSource(asset assetFunc, assetDir assetDirFunc, dir string) {
	m.migrationSource = &migrate.AssetMigrationSource{
		Asset:    asset,
		AssetDir: assetDir,
		Dir:      dir,
	}
}

// safeConnect connects to template1 so we can create/drop the desired database.
func (m *Module) safeConnect(dbname string) (string, *sql.DB, error) {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return "", nil, err
	}

	u, err := url.Parse(ds.DSN)
	if err != nil {
		return "", nil, err
	}

	database := strings.Trim(u.Path, "/")
	u.Path = "template1"
	u.RawPath = "template1"

	db, err := sql.Open(ds.Driver, u.String())
	if err != nil {
		return "", nil, err
	}
	return database, db, nil
}

// Reset drops and recreates the database
func (m *Module) Reset(dbname string) error {
	err := m.Drop(dbname)
	if err != nil {
		return err
	}

	databaseName, db, err := m.safeConnect(dbname)
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
func (m *Module) Drop(dbname string) error {
	databaseName, db, err := m.safeConnect(dbname)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DROP DATABASE IF EXISTS ` + databaseName)
	return err
}

// getMigrationSource returns the m.migrationSource if set, otherwise
// it defaults by reading from the MigrationsDir specified in
func (m *Module) getMigrationSource() (migrate.MigrationSource, error) {
	if m.migrationSource != nil {
		return m.migrationSource, nil
	}
	configPath, err := filepath.Abs(m.Config.ConfigPath)
	if err != nil {
		return nil, err
	}
	migrationPath := filepath.Join(filepath.Dir(configPath), m.config.MigrationsDir)
	return migrate.FileMigrationSource{Dir: migrationPath}, nil
}

// Migrate the given db
func (m *Module) Migrate(dbname string) error {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return err
	}

	db, err := sql.Open(ds.Driver, ds.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	migrations, err := m.getMigrationSource()
	if err != nil {
		return err
	}

	_, err = migrate.Exec(db, ds.Driver, migrations, migrate.Up)
	return err
}
