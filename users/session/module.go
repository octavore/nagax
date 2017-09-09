package session

import (
	"crypto/rsa"
	"time"

	"github.com/octavore/naga/service"
	"github.com/square/go-jose"

	"github.com/octavore/nagax/keystore"
	"github.com/octavore/nagax/users"
)

// todo: make these configurable
const (
	CookieName        = "session"
	defaultKeyFile    = "session.key"
	keyAlgorithm      = jose.RSA_OAEP
	contentEncryption = jose.A128GCM

	defaultRevocationFlush time.Duration = 15 * time.Minute
)

// RevocationStore is the interface for a store which
// keeps track of revoked sessions. By default it uses
// an in-memory store
type RevocationStore interface {
	Revoke(id string, trackFor time.Duration)
	IsRevoked(id string) bool
}

// KeyStore interface for retrieving keys (used for encrypting session cookie)
type KeyStore interface {
	LoadPrivateKey(string) ([]byte, *rsa.PrivateKey, error)
	LoadPublicKey(string) ([]byte, error)
}

var (
	_ service.Module      = &Module{}
	_ users.Authenticator = &Module{} // this module is an authenticator
)

// Module session is for keeping track of sessions
// See:
// - NewSessionCookie
// - Verify
// - VerifyAndExtend
// - EndSession
type Module struct {
	KeyStore        KeyStore
	RevocationStore RevocationStore

	SecureCookie            bool
	CookieDomain            string
	KeyFile                 string
	SessionValidityDuration time.Duration

	decryptionKey           interface{}
	encrypter               jose.Encrypter
	revocationTrackDuration time.Duration
}

// Init implements module.Init
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.RevocationStore = NewInMemoryRevocationStore(defaultRevocationFlush)
		return nil
	}

	c.Start = func() {
		if m.KeyStore == nil {
			m.KeyStore = &keystore.KeyStore{}
		}
		if m.KeyFile == "" {
			m.KeyFile = defaultKeyFile
		}
		var err error
		m.encrypter, m.decryptionKey, err = loadKeys(m.KeyFile, m.KeyStore)
		if err != nil {
			panic(err)
		}
		m.revocationTrackDuration = m.SessionValidityDuration
		s, ok := m.RevocationStore.(*InMemoryRevocationStore)
		if ok {
			go s.Start()
		}
	}
}

// load keys from the keystore
func loadKeys(keyFile string, keyStore KeyStore) (jose.Encrypter, interface{}, error) {
	privateKey, _, err := keyStore.LoadPrivateKey(keyFile)
	if err != nil {
		return nil, nil, err
	}
	decryptionKey, err := jose.LoadPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	pub, err := keyStore.LoadPublicKey(keyFile)
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := jose.LoadPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}

	encrypter, err := jose.NewEncrypter(keyAlgorithm, contentEncryption, publicKey)
	if err != nil {
		return nil, nil, err
	}

	return encrypter, decryptionKey, err
}
