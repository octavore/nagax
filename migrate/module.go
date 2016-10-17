package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/octavore/naga/service"
	"github.com/rubenv/sql-migrate"

	"github.com/octavore/nagax/config"
)

func init() {
	migrate.SetTable("schema_migrations")
}

type Module struct {
	Config *config.Module
	DB     *sql.DB

	config Config
}

type Config struct {
	Datasources   map[string]Datasource `json:"datasources"`
	MigrationsDir string                `json:"migrations"`
}

type Datasource struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

func (m *Module) PrintHelp(ctx *service.CommandContext) {
	if len(ctx.Args) != 1 {
		fmt.Println("Please specify a db:")
		if len(m.config.Datasources) == 0 {
			fmt.Println("  no databases found!")
		} else {
			for ds, _ := range m.config.Datasources {
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
				m.PrintHelp(ctx)
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
				m.PrintHelp(ctx)
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
	return &ds, nil
}

func (m *Module) Connect(dbname string) (*sql.DB, error) {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return nil, err
	}
	return sql.Open(ds.Driver, ds.DSN)
}

func (m *Module) Reset(dbname string) error {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return err
	}

	u, err := url.Parse(ds.DSN)
	if err != nil {
		return err
	}

	database := strings.Trim(u.Path, "/")
	u.Path = "template1"
	u.RawPath = "template1"

	db, err := sql.Open(ds.Driver, u.String())
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`DROP DATABASE IF EXISTS ` + database)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE DATABASE ` + database)
	return err
}

func (m *Module) Migrate(dbname string) error {
	ds, err := m.getConfig(dbname)
	if err != nil {
		return err
	}

	db, err := sql.Open(ds.Driver, ds.DSN)
	if err != nil {
		return err
	}

	migrations := migrate.FileMigrationSource{
		Dir: m.config.MigrationsDir,
	}

	_, err = migrate.Exec(db, ds.Driver, migrations, migrate.Up)
	return err
}
