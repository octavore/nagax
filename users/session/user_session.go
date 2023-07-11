package session

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/octavore/nagax/util/errors"

	"github.com/go-jose/go-jose/v3"
)

// UserSession data to be marshalled
type UserSession struct {
	ID        string `json:"user_id"`
	SessionID string `json:"session_id"`
}

func (m *Module) newScopedSessionCookie(u *UserSession, domain string) (*http.Cookie, error) {
	if !strings.HasPrefix(domain, m.CookieDomain) {
		return nil, errors.New("sessions: scoped cookie must have %s as a suffix", m.CookieDomain)
	}

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
		Name:     m.CookieName,
		Value:    msg,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(m.SessionValidityDuration.Seconds()),
		Domain:   domain,
		Secure:   m.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// NewSessionCookie creates a new encrypted cookie for the given UserSession
func (m *Module) NewSessionCookie(u *UserSession) (*http.Cookie, error) {
	return m.newScopedSessionCookie(u, m.CookieDomain)
}

func (m *Module) decodeCookieValue(value string) *UserSession {
	obj, err := jose.ParseEncrypted(value)
	if err != nil {
		m.Logger.Infof("Invalid cookie value: %s.", err)
		return nil
	}

	b, err := obj.Decrypt(m.decryptionKey)
	session := &UserSession{}
	err = json.Unmarshal(b, session)
	if err != nil {
		m.Logger.Infof("Invalid cookie value: %s.", err)
		return nil
	}
	return session
}

// getSessionFromRequest reads the current session from the request,
// and if it is valid, returns the corresponding UserSession.
// No error if there was no cookie, or the cookie was valid.
// If there is an invalid cookie, an error is returned.
func (m *Module) getSessionFromRequest(req *http.Request) (*UserSession, error) {
	cookie, err := req.Cookie(m.CookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}
	if err != nil {
		m.Logger.ErrorCtx(req.Context(), errors.Wrap(err))
		return nil, nil
	}

	session := m.decodeCookieValue(cookie.Value)
	if session == nil {
		return nil, nil
	}
	if m.RevocationStore.IsRevoked(session.SessionID) {
		return nil, nil
	}
	return session, nil
}
