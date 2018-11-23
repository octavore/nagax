package oauth

import (
	"context"
	"encoding/base64"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/util/errors"
)

// RedirectCallbackFn is the type of the callback after successfully handling the oauth callback
// token exchange.
type RedirectCallbackFn = func(*http.Request, http.ResponseWriter, *oauth2.Token, string) error

// Provider allows configuration of multiple oauth providers
type Provider struct {
	// required
	Base         string
	Config       *oauth2.Config
	PostCallback RedirectCallbackFn

	// optional
	Options       []oauth2.AuthCodeOption
	SetOAuthState func(*http.Request, router.Params) (string, error)
	NewClient     func(context.Context, *oauth2.Token) *http.Client
}

func (m *Module) register(p *Provider) {
	m.Router.GET(p.Base+"/login", p.HandleOAuthStart)
	m.Router.GET(p.Base+"/callback", p.handleCallback)
}

// HandleOAuthStart is the handler for redirecting to the oauth provider.
func (p *Provider) HandleOAuthStart(rw http.ResponseWriter, req *http.Request, par router.Params) error {
	var state string
	if p.SetOAuthState != nil {
		rawState, err := p.SetOAuthState(req, par)
		if err != nil {
			return errors.Wrap(err)
		}
		state = base64.StdEncoding.EncodeToString([]byte(rawState))
	}
	url := p.Config.AuthCodeURL(state, p.Options...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
	return nil
}

// doCallback parses the oauth callback and state if valid, and then calls PostCallback
func (p *Provider) handleCallback(rw http.ResponseWriter, req *http.Request, _ router.Params) error {
	// oauth handshake
	code := req.FormValue("code")
	accessToken, err := p.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return errors.Wrap(err)
	}
	var state string
	encState := req.FormValue("state")
	if encState != "" {
		stateByte, err := base64.StdEncoding.DecodeString(encState)
		if err != nil {
			return errors.New("error decoding state: %s", err)
		}
		state = string(stateByte)
	}
	err = p.PostCallback(req, rw, accessToken, state)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

// Client returns a new http client for authenticated requests. Provider
// can be configured with NewClient to override the default oauth2.Config.Client.
func (p *Provider) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	if p.NewClient != nil {
		return p.NewClient(ctx, t)
	}
	return p.Config.Client(ctx, t)
}
