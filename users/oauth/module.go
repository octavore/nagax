package oauth

import (
	"net/http"
	"net/url"

	"github.com/octavore/naga/service"
	"golang.org/x/oauth2"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/users/session"
)

const (
	defaultLoginURL              = "/login"
	defaultOAuthCallbackPath     = "/oauth/callback"
	defaultPostOAuthRedirectPath = "/"
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

	postOAuthRedirectPath   string
	oauthCallbackPath       string
	loginURL                string
	userStore               UserStore
	setOAuthState           func(req *http.Request) string
	oauthConfig             *oauth2.Config
	oauthOptions            []oauth2.AuthCodeOption
	getCallbackRedirectPath func(userToken string, state string) *url.URL
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.postOAuthRedirectPath = defaultPostOAuthRedirectPath
		m.oauthCallbackPath = defaultOAuthCallbackPath
		m.loginURL = defaultLoginURL
		return nil
	}

	c.Start = func() {
		m.Router.GET(m.loginURL, m.handleOAuthStart)
		m.Router.GET(m.oauthCallbackPath, m.handleOAuthCallback)
		if m.userStore == nil {
			panic("oauth.UserStore is required")
		}
	}
}

func (m *Module) Configure(opts ...option) {
	for _, opt := range opts {
		opt(m)
	}
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
