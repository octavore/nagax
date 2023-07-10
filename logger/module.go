package logger

import (
	"context"
	"fmt"
	"log"

	"github.com/octavore/naga/service"
)

type Logger interface {
	Info(args ...any)
	Infof(format string, args ...any)
	InfoCtx(ctx context.Context, args ...any)

	Warning(args ...any)
	Warningf(format string, args ...any)
	WarningCtx(ctx context.Context, args ...any)

	Error(args ...any)
	Errorf(format string, args ...any)
	ErrorCtx(ctx context.Context, args ...any)
}

type DefaultLogger struct{}

func (d *DefaultLogger) Info(args ...any) {
	log.Println("[INFO]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Infof(format string, args ...any) {
	log.Printf("[INFO] "+format, args...)
}

func (d *DefaultLogger) InfoCtx(ctx context.Context, args ...any) {
	d.Info(args...)
}

func (d *DefaultLogger) Warning(args ...any) {
	log.Println("[WARN]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Warningf(format string, args ...any) {
	log.Printf("[WARN] "+format, args...)
}

func (d *DefaultLogger) WarningCtx(ctx context.Context, args ...any) {
	d.Warning(args...)
}

func (d *DefaultLogger) Error(args ...any) {
	log.Println("[ERROR]", fmt.Sprint(args...))
}

func (d *DefaultLogger) Errorf(format string, args ...any) {
	log.Printf("[ERROR] "+format, args...)
}

func (d *DefaultLogger) ErrorCtx(ctx context.Context, args ...any) {
	d.Error(args...)
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
