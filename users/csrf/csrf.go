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

// New creates a new encrypted token for the given UserSession
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

// Decode an encrypted token
func (m *Module) Decode(token string) (*csrfPayload, error) {
	obj, err := jose.ParseEncrypted(token)
	if err != nil {
		return nil, err
	}
	b, err := obj.Decrypt(m.decryptionKey)
	csrfPayload := &csrfPayload{}
	if err = json.Unmarshal(b, csrfPayload); err != nil {
		return nil, err
	}
	if time.Now().After(csrfPayload.ExpireAfter) {
		return nil, errors.New("csrf token expired")
	}
	return csrfPayload, nil
}

// Verify an encrypted token
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
