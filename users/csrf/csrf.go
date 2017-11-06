package csrf

import (
	"encoding/json"
	"time"

	jose "github.com/square/go-jose"

	"github.com/octavore/nagax/util/errors"
	"github.com/octavore/nagax/util/token"
)

// UserSession data to be marshalled
type csrfPayload struct {
	State       string
	Token       string
	ExpireAfter time.Time
}

// newSessionCookie creates a new encrypted cookie for the given UserSession
func (m *Module) New(state string) (string, error) {
	b, err := json.Marshal(&csrfPayload{
		State:       state,
		Token:       token.New32(),
		ExpireAfter: time.Now().Add(m.csrfValidityDuration),
	})
	if err != nil {
		return "", err
	}

	obj, err := m.encrypter.Encrypt(b)
	if err != nil {
		return "", err
	}

	msg, err := obj.CompactSerialize()
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (m *Module) Verify(state, token string) (bool, error) {
	obj, err := jose.ParseEncrypted(token)
	if err != nil {
		m.Logger.Error(errors.Wrap(err))
		return false, nil
	}
	b, err := obj.Decrypt(m.decryptionKey)
	csrfPayload := &csrfPayload{}
	if err = json.Unmarshal(b, csrfPayload); err != nil {
		m.Logger.Error(errors.Wrap(err))
		return false, nil
	}
	if state != csrfPayload.State {
		return false, nil
	}
	if time.Now().After(csrfPayload.ExpireAfter) {
		return false, nil
	}
	return true, nil
}
