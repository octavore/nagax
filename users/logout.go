package users

import (
	"net/http"
	"time"
)

func (m *Module) HandleLogout(rw http.ResponseWriter, req *http.Request) {
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
