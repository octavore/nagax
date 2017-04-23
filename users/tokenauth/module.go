package tokenauth

import (
	"net/http"

	"strings"

	"github.com/octavore/naga/service"
	"github.com/octavore/nagax/users"
)

const (
	defaultHTTPHeader  = "Authorization"
	defaultTokenPrefix = "token"
)

var (
	_ service.Module      = &Module{}
	_ users.Authenticator = &Module{}
)

type TokenSource interface {
	Get(token string) *string
}

type Module struct {
	tokenSource TokenSource
	header      string
	prefix      string
}

func (m *Module) Init(c *service.Config) {
	c.Start = func() {
		m.header = defaultHTTPHeader
		m.prefix = defaultTokenPrefix
	}
}

func (m *Module) Authenticate(rw http.ResponseWriter, req *http.Request) (bool, *string, error) {
	val := req.Header.Get(m.header)
	if val == "" {
		return false, nil, nil
	}
	parts := strings.SplitN(val, " ", 2)
	if len(parts) != 2 {
		return false, nil, nil
	}
	prefix := strings.ToLower(parts[0])
	if prefix != m.prefix {
		return false, nil, nil
	}
	token := parts[1]
	userID := m.tokenSource.Get(token)
	if userID == nil {
		return false, nil, users.ErrNotAuthorized
	}
	return true, userID, nil
}

func (m *Module) Logout(rw http.ResponseWriter, req *http.Request) {
	// noop
}
