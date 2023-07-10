package users

import "net/http"

func (m *Module) Authenticate(rw http.ResponseWriter, req *http.Request) (handled bool, userToken *string, err error) {
	return m.AuthenticateWithList(m.Authenticators, rw, req)
}

func (m *Module) AuthenticateWithList(authenticators []Authenticator, rw http.ResponseWriter, req *http.Request) (handled bool, userToken *string, err error) {
	for _, auth := range authenticators {
		handled, userToken, err := auth.Authenticate(rw, req)
		if !handled {
			// if not handled but we have an error, log it separately
			if err != nil {
				m.Logger.ErrorCtx(req.Context(), err)
			}
			continue
		}
		return true, userToken, err
	}
	// no authenticator handled the request
	return false, nil, nil
}

func (m *Module) Logout(rw http.ResponseWriter, req *http.Request) {
	for _, auth := range m.Authenticators {
		handled, _, _ := auth.Authenticate(rw, req)
		if handled {
			auth.Logout(rw, req)
			return
		}
	}
}
