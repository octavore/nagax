package main

import (
	_ "github.com/lib/pq"

	"github.com/octavore/nagax/migrate"
	"github.com/octavore/naga/service"
)

type TestApp struct {
	DB *migrate.Module
}

func (t *TestApp) Init(*service.Config) {
}

func main() {
	service.New(&TestApp{}).Run()
}