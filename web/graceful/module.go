package graceful

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
)

type Module struct {
	Logger         *logger.Module
	done           chan bool
	ctx            context.Context
	cancel         context.CancelFunc
	timeoutSeconds int

	started bool
}

// Init the graceful module
// Usage: Wait should be called in the main function after Run
//
//	func main() {
//	  app := &App{}
//	  service.New(app).Run()
//	  app.Graceful.Wait()
//	}
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		ctx := context.Background()
		m.ctx, m.cancel = context.WithCancel(ctx)
		m.timeoutSeconds = 30
		if c.Env().IsDevelopment() {
			m.timeoutSeconds = 2
		}
		return nil
	}

	c.Start = func() {
		m.done = make(chan bool, 1)
		m.started = true
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigs
			m.Logger.Infof("graceful: got %s signal, shutting down in %d seconds...",
				strings.ToUpper(sig.String()),
				m.timeoutSeconds)
			m.cancel()
			time.Sleep(time.Duration(m.timeoutSeconds) * time.Second)
			close(m.done)
		}()
	}

	c.Stop = func() {
		// todo: cancel the signal handler?
		m.cancel()
	}
}

func (m *Module) GetGracefulShutdownContext() context.Context {
	return m.ctx
}

// Wait should be called at the top level of your app to block until the timeout completes.
func (m *Module) Wait() {
	m.Logger.Infof("graceful: waiting for shutdown...")
	if m.done != nil {
		<-m.done
		m.Logger.Infof("graceful: shutting down now, done channel is closed.")
	}
}

// OnShutdown is a helper to run a shutdownFunc when m.ctx is cancelled
func (m *Module) OnShutdown(shutdownFunc func(ctx context.Context)) {
	go func() {
		<-m.ctx.Done()
		// give a bit of wiggle room for all the OnShutdown goroutines to fire
		shutdownSeconds := time.Duration(m.timeoutSeconds - 1)
		ctx, cancel := context.WithTimeout(context.Background(), shutdownSeconds*time.Second)
		go func() {
			// if the m.done is closed then the app is shutdown
			<-m.done
			m.Logger.Infof("graceful: done channel was closed, cancelling OnShutdown ctx")
			cancel()
		}()
		shutdownFunc(ctx)
		cancel()
	}()
}

// Loop executes fn every `interval` until m.ctx is done (triggered by receiving a SIGTERM or SIGINT)
func (m *Module) Loop(id string, interval time.Duration, fn func(time.Time)) {
	m.Logger.Infof("graceful: [%s] starting loop...", id)
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-m.ctx.Done():
				m.Logger.Infof("graceful: [%s] stopping polling (shutdown)...", id)
				return

			case t := <-ticker.C:
				m.Logger.Infof("graceful: [%s] tick %s", id, t)
				fn(t)
			}
		}
	}()
}
