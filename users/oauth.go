package users

import (
	"net/http"

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
	// TODO: store something in the state for next url
	url := m.oauthConfig.AuthCodeURL(OAuthState, m.oauthOptions...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
}

func (m *Module) handleOAuthCallback(rw http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	token, err := m.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
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

	rw.Header().Add("Set-Cookie", cookie.String())
	http.Redirect(rw, req, m.OAuthRedirect, http.StatusTemporaryRedirect)
}
