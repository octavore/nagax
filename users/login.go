package users

import (
	"encoding/json"
	"net/http"

	"github.com/square/go-jose"
)

// GetLoginCookie returns a session cookie
func (m *Module) GetLoginCookie(userID string) (*http.Cookie, error) {
	b, err := json.Marshal(userSession{ID: userID})
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
		MaxAge:   86400,
		Domain:   m.CookieDomain,
		Secure:   m.SecureCookie,
	}, nil
}

// Authenticate a request with a cookie
func (m *Module) Authenticate(req *http.Request) (string, error) {
	cookie, err := req.Cookie(CookieName)
	if err != nil {
		return "", err
	}

	obj, err := jose.ParseEncrypted(cookie.Value)
	if err != nil {
		return "", err
	}

	b, err := obj.Decrypt(m.decryptionKey)
	session := userSession{}
	if err = json.Unmarshal(b, &session); err != nil {
		return "", err
	}

	return session.ID, nil
}
