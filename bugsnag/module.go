/*
bugsnag module configures bugsnag and modifies the logger so logger.Error logs to bugsnag as well.
*/
package bugsnag

import (
	"fmt"
	"net/http"

	bugsnagGo "github.com/bugsnag/bugsnag-go/v2"
	goerrors "github.com/go-errors/errors"
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/util/errors"
)

type Config struct {
	Bugsnag *struct {
		APIKey string `json:"api_key"`
	} `json:"bugsnag"`
}

type Module struct {
	Logger *logger.Module
	Config *config.Module

	ProjectPackages []string
	config          Config
	bugsnagEnabled  bool
	originalErrorf  func(format string, args ...any)
}

var _ service.Module = &Module{}

// AppVersion is set via a build flag
var AppVersion = ""

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		err := m.Config.ReadConfig(&m.config)
		if err != nil {
			return err
		}
		m.originalErrorf = m.Logger.Logger.Errorf
		if m.config.Bugsnag != nil {
			m.Logger.Info("bugsnag enabled: ", AppVersion)
			// note: this forces the app to restart
			bugsnagGo.Configure(bugsnagGo.Configuration{
				ReleaseStage:    c.Env().String(),
				ProjectPackages: append(m.ProjectPackages, "github.com/octavore/*"),
				APIKey:          m.config.Bugsnag.APIKey,
				Logger:          Printfer(m.Logger.Infof), // may be redundant with our own logging
				AppVersion:      AppVersion,
			})
			m.bugsnagEnabled = true
		}

		m.Logger.Logger = &bugsnagLogger{Logger: m.Logger.Logger, Notify: m.Notify}
		return nil
	}
}

type Printfer func(fmt string, args ...interface{})

func (p Printfer) Printf(fmt string, args ...interface{}) {
	p(fmt, args...)
}

type GetRequestable interface {
	GetRequest() *http.Request
}

type ErrorStackable interface {
	ErrorStack() string
}

// Notify bugsnag, note that m.Logger.Error calls Notify so Notify musn't call m.Logger.Error
func (m *Module) Notify(err error, rawData ...any) {
	rawData = append(rawData, bugsnagGo.SeverityError)
	if re, ok := err.(GetRequestable); ok && re.GetRequest() != nil {
		rawData = append(rawData, re.GetRequest())
	}
	errType := "error"
	if err2, ok := err.(*goerrors.Error); ok {
		errType = fmt.Sprintf("%T", err2.Err)
		rawData = append(rawData, bugsnagGo.ErrorClass{Name: errType})
	}

	if !m.bugsnagEnabled {
		if re, ok := err.(ErrorStackable); ok {
			m.originalErrorf("bugsnag fake: %T %v %#v\n%s", errType, err, rawData, re.ErrorStack())
		} else {
			m.originalErrorf("bugsnag fake: %T %v %#v", errType, err, rawData)
		}
		return
	}
	e := bugsnagGo.Notify(err, rawData...)
	if e != nil {
		m.originalErrorf("%v", errors.WrapS(e, 2))
	}
	// note: bugsnag does its own logging
}
