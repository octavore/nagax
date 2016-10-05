package oauth

import (
	"net/http"
)

func (m *Module) Authenticate(rw http.ResponseWriter, req *http.Request) (bool, *string, error) {
	return m.Sessions.Authenticate(rw, req)
}

func (m *Module) Logout(rw http.ResponseWriter, req *http.Request) {
	m.Sessions.DestroySession(rw, req)
}
