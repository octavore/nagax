package logger

import (
	"fmt"
	"log"

	"github.com/octavore/naga/service"
)

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warning(args ...interface{})
	Warningf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type DefaultLogger struct{}

func (d *DefaultLogger) Info(args ...interface{}) {
	log.Println("[INFO]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (d *DefaultLogger) Warning(args ...interface{}) {
	log.Println("[WARN]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Warningf(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

func (d *DefaultLogger) Error(args ...interface{}) {
	log.Println("[ERROR]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

var _ service.Module = &Module{}

type Module struct {
	Logger
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.Logger = &DefaultLogger{}
		return nil
	}
}
