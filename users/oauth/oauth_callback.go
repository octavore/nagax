package oauth

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

func (m *Module) handleOAuthCallback(rw http.ResponseWriter, req *http.Request) {
	// oauth handshake
	code := req.FormValue("code")
	accessToken, err := m.OAuthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	// convert access token to user
	userToken, err := getOrCreateUser(m.UserStore, accessToken)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	// get the URL to redirect to
	redirectURL := &url.URL{
		Scheme: req.URL.Scheme,
		Host:   req.URL.Host,
		Path:   m.PostOAuthRedirectPath,
	}

	// try to read the state from the query parameters
	state := req.FormValue("state")
	if state != "" {
		stateByte, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			m.Logger.Error("error decoding state: ", err)
		} else {
			query := redirectURL.Query()
			query.Set("state", string(stateByte))
			redirectURL.RawQuery = query.Encode()
		}
	}

	m.Logger.Infof("redirecting after oauth: %s", redirectURL.String())

	err = m.Sessions.CreateSession(userToken, rw)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}
	http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
}
