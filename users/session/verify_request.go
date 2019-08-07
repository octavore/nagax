package session

import "net/http"

// Verify a request with a cookie
func (m *Module) Verify(req *http.Request) (string, error) {
	session, err := m.getSessionFromRequest(req)
	if err != nil {
		return "", err
	} else if session == nil {
		return "", nil
	}
	return session.ID, nil
}

// VerifyAndExtend authenticates a cookie based session and
// refreshes the validity period. Returns an error if there
// was a cookie but it was invalid
func (m *Module) VerifyAndExtend(rw http.ResponseWriter, req *http.Request) (string, error) {
	session, err := m.getSessionFromRequest(req)
	if err != nil {
		return "", err
	} else if session == nil {
		return "", nil
	}

	cookie, err := m.NewSessionCookie(session)
	if err == nil {
		rw.Header().Add("Set-Cookie", cookie.String())
	}
	// handle error?
	return session.ID, nil
}

// Authenticate implements github.com/octavore/nagax/users.Authenticate
// todo: store session.ID in the context?
func (m *Module) Authenticate(rw http.ResponseWriter, req *http.Request) (handled bool, userToken *string, err error) {
	session, err := m.getSessionFromRequest(req)
	// getSessionFromRequest returns an error if there was a cookie
	// but it was invalid, so consider the request handled
	if err != nil {
		return true, nil, err
	}
	// no error and no session, so no cookie
	if session == nil {
		return false, nil, nil
	}
	return true, &session.ID, nil
}
