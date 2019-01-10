package session

import (
	"fmt"
	"net/http"
	"time"
)

// CreateSession update the response with a session cookie
func (m *Module) CreateSession(userToken string, rw http.ResponseWriter) error {
	cookie, err := m.newSessionCookie(&UserSession{
		ID:        userToken,
		SessionID: fmt.Sprintf("%s-%s", userToken, time.Now().String()),
	})
	if err != nil {
		return err
	}
	http.SetCookie(rw, cookie)
	return nil
}

// CreateScopedSession update the response with a session cookie
func (m *Module) CreateScopedSession(userToken, domain string, rw http.ResponseWriter) error {
	cookie, err := m.newScopedSessionCookie(&UserSession{
		ID:        userToken,
		SessionID: fmt.Sprintf("%s-%s", userToken, time.Now().String()),
	}, domain)
	if err != nil {
		return err
	}
	http.SetCookie(rw, cookie)
	return nil
}
