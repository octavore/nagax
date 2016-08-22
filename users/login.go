package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/square/go-jose"
)

// NewSessionCookie returns a new session cookie
func (m *Module) NewSessionCookie(userID string) (*http.Cookie, error) {
	return m.newSessionCookie(&userSession{
		ID:        userID,
		SessionID: fmt.Sprintf("%s-%s", userID, time.Now().String()),
	})
}

func (m *Module) newSessionCookie(u *userSession) (*http.Cookie, error) {
	b, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	obj, err := m.encrypter.Encrypt(b)
	if err != nil {
		return nil, err
	}

	msg, err := obj.CompactSerialize()
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     CookieName,
		Value:    msg,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(m.sessionValidityDuration.Seconds()),
		Domain:   m.CookieDomain,
		Secure:   m.SecureCookie,
	}, nil
}

func (m *Module) getSessionFromRequest(req *http.Request) (*userSession, error) {
	cookie, err := req.Cookie(CookieName)
	if err != nil {
		return nil, err
	}

	obj, err := jose.ParseEncrypted(cookie.Value)
	if err != nil {
		return nil, err
	}

	b, err := obj.Decrypt(m.decryptionKey)
	session := &userSession{}
	if err = json.Unmarshal(b, session); err != nil {
		return nil, err
	}
	if m.RevocationStore.IsRevoked(session.SessionID) {
		return nil, errors.New("invalid session")
	}

	return session, nil
}

// Authenticate a request with a cookie.
func (m *Module) Authenticate(req *http.Request) (string, error) {
	session, err := m.getSessionFromRequest(req)
	if err != nil {
		return "", err
	}
	return session.ID, nil
}

// AuthenticateAndExtend authenticates a cookie based session and
// refreshes the validity period.
func (m *Module) AuthenticateAndExtend(rw http.ResponseWriter, req *http.Request) (string, error) {
	userSession, err := m.getSessionFromRequest(req)
	if err != nil {
		return "", err
	}

	cookie, err := m.newSessionCookie(userSession)
	if err == nil {
		rw.Header().Add("Set-Cookie", cookie.String())
	}
	return userSession.ID, nil
}
