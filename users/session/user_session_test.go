package session

import (
	"net/http"
	"testing"

	"github.com/octavore/naga/service"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionCookie(t *testing.T) {
	m := &Module{}
	stop := service.New(m).StartForTest()
	defer stop()

	cookie, err := m.NewSessionCookie(&UserSession{
		ID:        "abc",
		SessionID: "123",
	})
	assert.NoError(t, err)
	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "", cookie.Domain)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)

	session := m.decodeCookieValue(cookie.Value)
	assert.Equal(t, &UserSession{ID: "abc", SessionID: "123"}, session)
}
