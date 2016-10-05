package oauth

import (
	"net/http"

	"github.com/octavore/naga/service"
	"golang.org/x/oauth2"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
	"github.com/octavore/nagax/users/session"
)

const (
	LoginURL                     = "/login" // make this configurable
	OAuthCallbackPath            = "/oauth/callback"
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
	Router    *router.Module
	Sessions  *session.Module // todo: make this an interface
	Logger    *logger.Module
	UserStore UserStore

	ErrorHandler  func(http.ResponseWriter, *http.Request, error)
	SetOAuthState func(req *http.Request) string

	PostOAuthRedirectPath string
	OAuthConfig           *oauth2.Config
	OAuthOptions          []oauth2.AuthCodeOption
}

func (m *Module) Init(c *service.Config) {
	m.PostOAuthRedirectPath = defaultPostOAuthRedirectPath

	c.Setup = func() error {
		m.Router.HandleFunc(LoginURL, m.handleOAuthStart)
		m.Router.HandleFunc(OAuthCallbackPath, m.handleOAuthCallback)
		return nil
	}

	c.Start = func() {
		if m.UserStore == nil {
			panic("oauth.UserStore is required")
		}
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
