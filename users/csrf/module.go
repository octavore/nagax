package csrf

import (
	"crypto/rsa"
	"time"

	"github.com/octavore/naga/service"
	jose "gopkg.in/square/go-jose.v1"

	"github.com/octavore/nagax/keystore"
	"github.com/octavore/nagax/logger"
)

const (
	defaultKeyFile      = "session.key"
	keyAlgorithm        = jose.RSA_OAEP
	contentEncryption   = jose.A128GCM
	defaultCSRFValidity = 12 * time.Hour
)

// KeyStore interface for retrieving keys (used for encrypting session cookie)
type KeyStore interface {
	LoadPrivateKey(string) ([]byte, *rsa.PrivateKey, error)
	LoadPublicKey(string) ([]byte, error)
}

var (
	_ service.Module = &Module{}
)

type Module struct {
	Logger *logger.Module

	csrfValidityDuration time.Duration
	keyStore             KeyStore
	keyFile              string
	decryptionKey        interface{}
	encrypter            jose.Encrypter
}

// Init implements module.Init
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.keyStore = &keystore.KeyStore{}
		m.keyFile = defaultKeyFile
		m.csrfValidityDuration = defaultCSRFValidity
		return nil
	}

	c.Start = func() {
		var err error
		privateKey, publicKey, err := loadKeys(m.keyFile, m.keyStore)
		if err != nil {
			c.Fatal(err)
		}
		m.decryptionKey = privateKey
		m.encrypter, err = jose.NewEncrypter(keyAlgorithm, contentEncryption, publicKey)
		if err != nil {
			c.Fatal(err)
		}
	}
}

// Configure this module with given options
func (m *Module) Configure(opts ...option) {
	for _, opt := range opts {
		opt(m)
	}
}
