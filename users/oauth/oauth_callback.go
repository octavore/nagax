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
	accessToken, err := m.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		m.errorHandler(rw, req, err)
		return
	}

	// convert access token to user
	userToken, err := getOrCreateUser(m.userStore, accessToken)
	if err != nil {
		m.errorHandler(rw, req, err)
		return
	}

	// get the URL to redirect to
	redirectURL := &url.URL{
		Path: m.postOAuthRedirectPath,
	}

	// try to read the state from the query parameters
	state := ""
	encState := req.FormValue("state")
	if encState != "" {
		stateByte, err := base64.StdEncoding.DecodeString(encState)
		if err != nil {
			m.Logger.Error("error decoding state: ", err)
		} else {
			query := redirectURL.Query()
			state = string(stateByte)
			query.Set("state", state)
			redirectURL.RawQuery = query.Encode()
		}
	}

	if m.getCallbackRedirectPath != nil {
		altRedirectURL := m.getCallbackRedirectPath(userToken, state)
		if altRedirectURL != nil {
			redirectURL = altRedirectURL
		}
	}

	if redirectURL.Host == "" {
		redirectURL.Scheme = req.URL.Scheme
		redirectURL.Host = req.URL.Host
	}

	m.Logger.Infof("redirecting after oauth: %s", redirectURL.String())

	err = m.Sessions.CreateSession(userToken, rw)
	if err != nil {
		m.errorHandler(rw, req, err)
		return
	}
	http.Redirect(rw, req, redirectURL.String(), http.StatusTemporaryRedirect)
}
