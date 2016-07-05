package users

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	OAuthStartURL    = "/oauth/start"
	OAuthCallbackURL = "/oauth/callback"
)

func (m *Module) setupRoutes() {
	m.Router.HandleFunc(OAuthStartURL, m.handleOAuthStart)
	m.Router.HandleFunc(OAuthCallbackURL, m.handleOAuthCallback)
}

func (m *Module) handleOAuthStart(rw http.ResponseWriter, req *http.Request) {
	log.Println("oauth start")
	url := m.oauthConfig.AuthCodeURL(OAuthState, m.oauthOptions...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
}

func (m *Module) handleOAuthCallback(rw http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	log.Printf("oauth callback %s", code)
	token, err := m.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}

	userID, err := m.UserStore.Get(token)
	if err != nil {
		log.Println("user store get error:", err)
		m.ErrorHandler(rw, req, err)
		return
	}
	if userID == "" {
		userID, err = m.UserStore.Create(token)
	} else {
		err = m.UserStore.Save(userID, token)
	}
	if err != nil {
		log.Println("user store error:", err)
		m.ErrorHandler(rw, req, err)
		return
	}

	cookie, err := m.GetLoginCookie(userID)
	if err != nil {
		log.Println("login error:", err)
		m.ErrorHandler(rw, req, err)
		return
	}

	rw.Header().Add("Set-Cookie", cookie.String())
	http.Redirect(rw, req, m.OAuthRedirect, http.StatusTemporaryRedirect)
}
