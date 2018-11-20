package oauth

import (
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/router"
)

var _ service.Module = &Module{}

// Module oauth implements oauth start and callback
type Module struct {
	Router       *router.Module
	oauthConfigs []*Provider
}

// Init this module.
func (m *Module) Init(c *service.Config) {
	c.Start = func() {
		for _, p := range m.oauthConfigs {
			m.register(p)
		}
	}
}

// AddProvider adds a new provider to the oauth module. Provider are registered during
// the Start phase.
func (m *Module) AddProvider(p *Provider) {
	m.oauthConfigs = append(m.oauthConfigs, p)
}
