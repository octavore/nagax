package migrate

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/octavore/naga/service"
	"gopkg.in/cenkalti/backoff.v2"

	"github.com/octavore/nagax/util/errors"
)

func (m *Module) registerCommands(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword:    "db:migrate <db>",
		ShortUsage: "run db migrations",
		Run: func(ctx *service.CommandContext) {
			if len(ctx.Args) != 1 {
				m.printHelp(ctx)
			}
			b, err := m.GetBackend(ctx.Args[0])
			if err != nil {
				m.Logger.Error("migrate:", err)
				return
			}
			err = backoff.RetryNotify(b.Migrate, m.backoff, func(err error, duration time.Duration) {
				m.Logger.Warningf("can't connect to db: %s, will retry in %s", err, duration)
			})
			if err != nil {
				m.Logger.Error("migrate:", err)
			}
		},
	})

	c.AddCommand(&service.Command{
		Keyword:    "db:status <db>",
		ShortUsage: "show migration status for <db>",
		Run: func(ctx *service.CommandContext) {
			if len(ctx.Args) != 1 {
				m.printHelp(ctx)
			}
			b, err := m.GetBackend(ctx.Args[0])
			if err != nil {
				m.Logger.Error("migrate:", err)
				return
			}

			allMigrations, err := b.migrations().FindMigrations()
			if err != nil {
				m.Logger.Error("migrate:", err)
				return
			}
			var unapplied []string
			err = backoff.RetryNotify(func() error {
				unapplied, err = b.UnappliedMigrations()
				if err != nil {
					return errors.Wrap(err)
				}
				return nil
			}, m.backoff, func(err error, duration time.Duration) {
				m.Logger.Warningf("can't connect to db: %s, will retry in %s", err, duration)
			})
			if err != nil {
				m.Logger.Error("migrate:", err)
			}

			unappliedSet := map[string]bool{}
			for _, m := range unapplied {
				unappliedSet[m] = true
			}

			for _, m := range allMigrations {
				status := color.YellowString("pending")
				if !unappliedSet[m.Id] {
					status = color.GreenString("done")
				}
				fmt.Printf("%-18s%s\n", status, m.Id)
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
			b, err := m.GetBackend(dbname)
			if err != nil {
				m.Logger.Error("migrate:", err)
				return
			}
			err = b.Reset()
			if err != nil {
				m.Logger.Error("migrate:", err)
			}
			err = b.Migrate()
			if err != nil {
				m.Logger.Error("migrate:", err)
			}
		},
	})
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
