package users

import (
	"net/http"
	"time"
)

// HandleLogout handles a logout request and attempts to erase the session cookie
func (m *Module) HandleLogout(rw http.ResponseWriter, req *http.Request) {
	session, err := m.getSessionFromRequest(req)
	if session == nil || err != nil {
		// TODO: log error
		return
	}
	m.RevocationStore.Revoke(session.ID, m.revocationTrackDuration)
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
