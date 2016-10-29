package session

import (
	"net/http"
	"time"
)

// DestroySession handles a logout request and attempts to erase the session cookie.
func (m *Module) DestroySession(rw http.ResponseWriter, req *http.Request) {
	session, err := m.getSessionFromRequest(req)
	if session == nil {
		return
	} else if err != nil {
		// TODO: log error
		return
	}

	m.RevocationStore.Revoke(session.SessionID, revocationTrackDuration)
	http.SetCookie(rw, &http.Cookie{
		Name:     CookieName,
		MaxAge:   -1,
		HttpOnly: true,
		Expires:  time.Now().Add(-time.Hour),
		Value:    "deleted",
		Path:     "/",
		Domain:   m.CookieDomain,
		Secure:   m.SecureCookie,
	})
}

// Logout implements github.com/octavore/nagax/users.Logout
// note that there is no redirect
func (m *Module) Logout(rw http.ResponseWriter, req *http.Request) {
	m.DestroySession(rw, req)
}
