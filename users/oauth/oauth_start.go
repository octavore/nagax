package oauth

import (
	"encoding/base64"
	"net/http"
)

func (m *Module) handleOAuthStart(rw http.ResponseWriter, req *http.Request) {
	// TODO: add some kind of verifier thing
	state := ""
	if m.SetOAuthState != nil {
		state += base64.StdEncoding.EncodeToString([]byte(m.SetOAuthState(req)))
	}
	url := m.OAuthConfig.AuthCodeURL(state, m.OAuthOptions...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
}
