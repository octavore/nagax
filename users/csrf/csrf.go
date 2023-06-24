package csrf

import (
	"encoding/json"
	"time"

	"github.com/go-jose/go-jose/v3"

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
		return "", errors.Wrap(err)
	}

	obj, err := m.encrypter.Encrypt(b)
	if err != nil {
		return "", errors.Wrap(err)
	}

	msg, err := obj.CompactSerialize()
	if err != nil {
		return "", errors.Wrap(err)
	}

	return msg, nil
}

// Decode an encrypted token
func (m *Module) Decode(token string) (*csrfPayload, error) {
	obj, err := jose.ParseEncrypted(token)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	b, err := obj.Decrypt(m.decryptionKey)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	csrfPayload := &csrfPayload{}
	if err = json.Unmarshal(b, csrfPayload); err != nil {
		return nil, errors.Wrap(err)
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
		return false, errors.Wrap(err)
	}
	b, err := obj.Decrypt(m.decryptionKey)
	if err != nil {
		return false, errors.Wrap(err)
	}
	csrfPayload := &csrfPayload{}
	if err = json.Unmarshal(b, csrfPayload); err != nil {
		return false, errors.Wrap(err)
	}
	if state != csrfPayload.State {
		return false, nil
	}
	if time.Now().After(csrfPayload.ExpireAfter) {
		return false, nil
	}
	return true, nil
}
