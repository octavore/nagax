package auth_router

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/octavore/naga/service"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (m *Module[A]) registerRoutesList(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword: "routes:list",
		Run: func(c *service.CommandContext) {
			for _, r := range m.routeRegistry {
				tHandler := reflect.TypeOf(r.handler)
				tReq := tHandler.In(2)
				tRes := tHandler.Out(0)
				fmt.Printf("%-8s%s\n", r.method, r.path)
				if r.version == "proto" {
					fmt.Println("        " +
						color.GreenString(stringifyParam(tReq)) +
						" -> " +
						color.BlueString(stringifyParam(tRes)),
					)
				}
			}
		},
		ShortUsage: "Print auth_router routes",
		Usage:      "Print routes which were registered with the auth_router module (note: routes registered with a different module will not appear!)",
	})
}

func stringifyParam(r reflect.Type) string {
	if r.ConvertibleTo(reflect.TypeOf(&emptypb.Empty{})) {
		return "<nil>"
	}
	return r.String()
}
