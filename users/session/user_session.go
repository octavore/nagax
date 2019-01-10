package session

import (
	"encoding/json"
	"net/http"

	"github.com/octavore/nagax/util/errors"

	jose "github.com/square/go-jose"
)

// UserSession data to be marshalled
type UserSession struct {
	ID        string `json:"user_id"`
	SessionID string `json:"session_id"`
}

func (m *Module) newScopedSessionCookie(u *UserSession, domain string) (*http.Cookie, error) {
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
		MaxAge:   int(m.SessionValidityDuration.Seconds()),
		Domain:   domain,
		Secure:   m.SecureCookie,
	}, nil
}

// newSessionCookie creates a new encrypted cookie for the given UserSession
func (m *Module) newSessionCookie(u *UserSession) (*http.Cookie, error) {
	return m.newScopedSessionCookie(u, m.CookieDomain)
}

// getSessionFromRequest reads the current session from the request,
// and if it is valid, returns the corresponding UserSession.
// No error if there was no cookie, or the cookie was valid.
// If there is an invalid cookie, an error is returned.
func (m *Module) getSessionFromRequest(req *http.Request) (*UserSession, error) {
	cookie, err := req.Cookie(CookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	} else if err != nil {
		m.Logger.Error(errors.Wrap(err))
		return nil, nil
	}

	obj, err := jose.ParseEncrypted(cookie.Value)
	if err != nil {
		m.Logger.Error(errors.Wrap(err))
		return nil, nil
	}

	b, err := obj.Decrypt(m.decryptionKey)
	session := &UserSession{}
	if err = json.Unmarshal(b, session); err != nil {
		m.Logger.Error(errors.Wrap(err))
		return nil, nil
	}
	if m.RevocationStore.IsRevoked(session.SessionID) {
		return nil, nil
	}

	return session, nil
}
