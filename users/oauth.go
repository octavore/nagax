package users

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

// URLS for the oauth handlers
const (
	OAuthStartURL    = "/oauth/start"
	OAuthCallbackURL = "/oauth/callback"
)

func (m *Module) setupRoutes() {
	m.Router.HandleFunc(OAuthStartURL, m.handleOAuthStart)
	m.Router.HandleFunc(OAuthCallbackURL, m.handleOAuthCallback)
}

func (m *Module) handleOAuthStart(rw http.ResponseWriter, req *http.Request) {
	// TODO: add some kind of verifier thing
	state := ""
	if m.oauthState != nil {
		state += base64.StdEncoding.EncodeToString([]byte(m.oauthState(req)))
	}

	url := m.oauthConfig.AuthCodeURL(state, m.oauthOptions...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
}

func (m *Module) handleOAuthCallback(rw http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	token, err := m.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	state := req.FormValue("state")
	if state != "" {
		stateByte, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			m.ErrorHandler(rw, req, err)
			return
		}
		state = string(stateByte)
		// m.validateState()
		// m.paramsFromState()
	}

	userID, err := m.UserStore.Get(token)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}
	if userID == "" {
		userID, err = m.UserStore.Create(token)
	} else {
		err = m.UserStore.Save(userID, token)
	}
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	cookie, err := m.NewSessionCookie(userID)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	redirectURL, err := url.Parse(m.OAuthRedirect)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}
	if state != "" {
		query := redirectURL.Query()
		query.Set("state", state)
		redirectURL.RawQuery = query.Encode()
	}

	rw.Header().Add("Set-Cookie", cookie.String())
	log.Printf("redirecting after oauth: %s", redirectURL.String())
	http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
}
