package oauth

import (
	"encoding/base64"
	"net/http"

	"github.com/octavore/nagax/router"
)

func (m *Module) handleOAuthStart(rw http.ResponseWriter, req *http.Request, _ router.Params) error {
	// TODO: add some kind of verifier thing
	state := ""
	if m.setOAuthState != nil {
		state += base64.StdEncoding.EncodeToString([]byte(m.setOAuthState(req)))
	}
	url := m.oauthConfig.AuthCodeURL(state, m.oauthOptions...)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
	return nil
}
