package config

import (
	"fmt"

	"github.com/octavore/naga/service"
)

func (m *Module) registerCommands(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword: "config:explain",
		Run: func(ctx *service.CommandContext) {
			m.PrintConsolidatedConfig()
		},
		ShortUsage: "Explain config.json options",
		Usage:      "",
	})

	c.AddCommand(&service.Command{
		Keyword: "config:print",
		Run: func(ctx *service.CommandContext) {
			fmt.Println(string(m.Byte))
		},
		ShortUsage: "Print current config.json",
		Usage:      "",
	})
}
