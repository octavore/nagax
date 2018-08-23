package migrate

import (
	"fmt"
	"log"
	"time"

	"github.com/octavore/naga/service"
	"gopkg.in/cenkalti/backoff.v2"
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
				log.Println("migrate:", err)
				return
			}
			err = backoff.RetryNotify(b.Migrate, m.backoff, func(err error, duration time.Duration) {
				log.Printf("can't connect to mysql: %s, will retry in %s", err, duration)
			})
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
			b, err := m.GetBackend(dbname)
			if err != nil {
				log.Println("migrate:", err)
				return
			}
			err = b.Reset()
			if err != nil {
				log.Println("migrate:", err)
			}
			err = b.Migrate()
			if err != nil {
				log.Println("migrate:", err)
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
