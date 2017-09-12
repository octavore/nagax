package router

import (
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/util/memlogger"
)

type TestModule struct {
	*Module
}

func (m *TestModule) Init(c *service.Config) {
	c.Setup = func() error {
		m.Logger.Logger = &memlogger.MemoryLogger{}
		return nil
	}
}

type testEnv struct {
	module *Module
	logger *memlogger.MemoryLogger
	stop   func()
}

func setup() testEnv {
	module := &TestModule{}
	stop := service.New(module).StartForTest()
	return testEnv{
		module: module.Module,
		logger: module.Logger.Logger.(*memlogger.MemoryLogger),
		stop:   stop,
	}
}
