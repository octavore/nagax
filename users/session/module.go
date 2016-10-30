package session

import (
	"time"

	"github.com/octavore/naga/service"
	"github.com/square/go-jose"

	"github.com/octavore/nagax/users/session/keystore"
)

// todo: make these configurable
const (
	CookieName        = "session"
	keyAlgorithm      = jose.RSA_OAEP
	contentEncryption = jose.A128GCM

	defaultRevocationFlush  time.Duration = 15 * time.Minute
	revocationTrackDuration time.Duration = 24 * time.Hour
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
	LoadPrivateKey() ([]byte, error)
	LoadPublicKey() ([]byte, error)
}

var _ service.Module = &Module{}

// Module session is for keeping track of sessions
// See:
// - NewSessionCookie
// - Verify
// - VerifyAndExtend
// - EndSession
type Module struct {
	KeyStore        KeyStore
	RevocationStore RevocationStore

	SecureCookie bool
	CookieDomain string

	decryptionKey           interface{}
	encrypter               jose.Encrypter
	sessionValidityDuration time.Duration
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
		if m.CookieDomain == "" {
			panic("session.CookieDomain is required")
		}

		var err error
		m.encrypter, m.decryptionKey, err = loadKeys(m.KeyStore)
		if err != nil {
			panic(err)
		}
		s, ok := m.RevocationStore.(*InMemoryRevocationStore)
		if ok {
			go s.Start()
		}
	}
}

// load keys from the keystore
func loadKeys(keyStore KeyStore) (jose.Encrypter, interface{}, error) {
	privateKey, err := keyStore.LoadPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	decryptionKey, err := jose.LoadPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	pub, err := keyStore.LoadPublicKey()
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
