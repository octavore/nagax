package oauth

import (
	"github.com/octavore/naga/service"
	"golang.org/x/oauth2"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/users/session"
)

// UserStore is an interface for managing users by oauth id
type UserStore interface {
	Create(*oauth2.Token) (id string, err error)
	Get(*oauth2.Token) (id string, err error)
	Save(userToken string, token *oauth2.Token) error
}

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

func getOrCreateUser(store UserStore, accessToken *oauth2.Token) (userToken string, err error) {
	userToken, err = store.Get(accessToken)
	if err != nil {
		return "", err
	}
	if userToken == "" {
		return store.Create(accessToken)
	}
	return userToken, store.Save(userToken, accessToken)
}
