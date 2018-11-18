package oauth

import (
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/users/session"
)

var _ service.Module = &Module{}
var _ users.Authenticator = &Module{}

// Module oauth implements oauth start and callback
type Module struct {
	Router   *router.Module
	Sessions *session.Module // todo: make this an interface
	Logger   *logger.Module

	oauthConfigs []*Provider
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		return nil
	}

	c.Start = func() {
		for _, p := range m.oauthConfigs {
			m.register(p)
		}
	}
}

func (m *Module) AddProvider(p *Provider) {
	m.oauthConfigs = append(m.oauthConfigs, p)
}
