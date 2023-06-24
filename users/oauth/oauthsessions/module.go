package oauthsessions

import (
	"net/http"
	"net/url"

	"github.com/octavore/naga/service"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/users/oauth"
	"github.com/octavore/nagax/users/session"
	"github.com/octavore/nagax/util/errors"
	"golang.org/x/oauth2"
)

type Module struct {
	OAuth   *oauth.Module
	Session *session.Module
	Auth    *users.Module

	BasePath        string
	OAuthConfig     *oauth2.Config
	GetOrCreateUser func(*oauth2.Config, *http.Request, *oauth2.Token, string) (string, error)
	RedirectPath    string
}

func (m *Module) Init(c *service.Config) {
	c.Start = func() {
		m.OAuth.AddProvider(&oauth.Provider{
			Base:         m.BasePath,
			Config:       m.OAuthConfig,
			PostCallback: m.redirectCallback,
			Options:      nil,
		})
		m.Auth.RegisterAuthenticator(m.Session)
	}
}

func (m *Module) redirectCallback(req *http.Request, rw http.ResponseWriter, token *oauth2.Token, state string) error {
	// get the URL to redirect to
	redirectURL := &url.URL{Path: m.RedirectPath}

	if state != "" {
		query := redirectURL.Query()
		query.Set("state", state)
		redirectURL.RawQuery = query.Encode()
	}

	if redirectURL.Host == "" {
		redirectURL.Scheme = req.URL.Scheme
		redirectURL.Host = req.URL.Host
	}

	// convert access token to user
	userToken, err := m.GetOrCreateUser(m.OAuthConfig, req, token, state)
	if err != nil {
		return errors.Wrap(err)
	}

	err = m.Session.CreateSession(userToken, rw)
	if err != nil {
		return errors.Wrap(err)
	}
	http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
	return nil

}
