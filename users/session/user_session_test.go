package session

import (
	"net/http"
	"testing"

	"github.com/octavore/naga/service"
	"github.com/shoenig/test"
)

func TestNewSessionCookie(t *testing.T) {
	m, stop := service.New(&Module{}).StartForTest()
	defer stop()

	cookie, err := m.NewSessionCookie(&UserSession{
		ID:        "abc",
		SessionID: "123",
	})
	test.NoError(t, err)
	test.Eq(t, "session", cookie.Name)
	test.Eq(t, "", cookie.Domain)
	test.Eq(t, http.SameSiteLaxMode, cookie.SameSite)
	test.Eq(t, "/", cookie.Path)

	session := m.decodeCookieValue(cookie.Value)
	test.Eq(t, &UserSession{ID: "abc", SessionID: "123"}, session)
}
