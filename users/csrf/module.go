package csrf

import (
	"crypto/rsa"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/octavore/naga/service"

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
	KeyFile              string
	decryptionKey        *rsa.PrivateKey
	encrypter            jose.Encrypter
}

// Init implements module.Init
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.keyStore = &keystore.KeyStore{}
		m.KeyFile = defaultKeyFile
		m.csrfValidityDuration = defaultCSRFValidity
		return nil
	}

	c.Start = func() {
		var err error
		_, privateKey, err := m.keyStore.LoadPrivateKey(m.KeyFile)
		if err != nil {
			c.Fatal(err)
		}
		m.decryptionKey = privateKey
		m.encrypter, err = jose.NewEncrypter(contentEncryption, jose.Recipient{
			Algorithm: keyAlgorithm,
			Key:       &privateKey.PublicKey,
		}, nil)
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
