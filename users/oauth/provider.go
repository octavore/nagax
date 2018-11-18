package oauth

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/util/errors"
	"golang.org/x/oauth2"
)

// Provider allows configuration of multiple oauth providers
type Provider struct {
	// required
	Base            string
	Config          *oauth2.Config
	GetOrCreateUser func(*oauth2.Config, *http.Request, *oauth2.Token, string) (id string, err error)

	// optional
	PostOAuthRedirectPath string
	Options               []oauth2.AuthCodeOption
	SetOAuthState         func(req *http.Request) string

	logger *logger.Module
}

func (p *Provider) handleOAuthStart(rw http.ResponseWriter, req *http.Request, _ router.Params) error {
	// TODO: add some kind of verifier thing
	state := ""
	if p.SetOAuthState != nil {
		state += base64.StdEncoding.EncodeToString([]byte(p.SetOAuthState(req)))
	}
	url := p.Config.AuthCodeURL(state, p.Options...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
	return nil
}

func (p *Provider) doCallback(req *http.Request) (string, *url.URL, error) {
	// oauth handshake
	code := req.FormValue("code")
	accessToken, err := p.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", nil, errors.Wrap(err)
	}

	// get the URL to redirect to
	redirectURL := &url.URL{Path: p.PostOAuthRedirectPath}

	// try to read the state from the query parameters
	state := ""
	encState := req.FormValue("state")
	if encState != "" {
		stateByte, err := base64.StdEncoding.DecodeString(encState)
		if err != nil {
			p.logger.Error("error decoding state: ", err)
		} else {
			query := redirectURL.Query()
			state = string(stateByte)
			query.Set("state", state)
			redirectURL.RawQuery = query.Encode()
		}
	}

	if redirectURL.Host == "" {
		redirectURL.Scheme = req.URL.Scheme
		redirectURL.Host = req.URL.Host
	}

	p.logger.Infof("redirecting after oauth: %s", redirectURL.String())

	// convert access token to user
	userToken, err := p.GetOrCreateUser(p.Config, req, accessToken, state)
	if err != nil {
		return "", nil, errors.Wrap(err)
	}

	return userToken, redirectURL, nil
}

func (m *Module) register(p *Provider) {
	if p.PostOAuthRedirectPath == "" {
		p.PostOAuthRedirectPath = "/"
	}

	m.Router.GET(p.Base+"/login", p.handleOAuthStart)
	m.Router.GET(p.Base+"/callback", func(rw http.ResponseWriter, req *http.Request, _ router.Params) error {
		userToken, redirectURL, err := p.doCallback(req)
		if err != nil {
			return errors.Wrap(err)
		}
		err = m.Sessions.CreateSession(userToken, rw)
		if err != nil {
			return errors.Wrap(err)
		}
		http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
		return nil
	})
}
