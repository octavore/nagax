package users

import (
	"net/http"

	"github.com/octavore/naga/service"
	"github.com/octavore/nagax/router"
	"github.com/square/go-jose"
	"golang.org/x/oauth2"
)

// default values
var (
	OAuthState        string                 = "state"
	CookieName        string                 = "session"
	KeyAlgorithm      jose.KeyAlgorithm      = jose.RSA_OAEP
	ContentEncryption jose.ContentEncryption = jose.A128GCM
)

type userSession struct {
	ID string `json:"user_id"`
}

type UserStore interface {
	Create(*oauth2.Token) (id string, err error)
	Get(*oauth2.Token) (id string, err error)
	Save(userID string, token *oauth2.Token) error
}

type KeyStore interface {
	LoadPrivateKey() ([]byte, error)
	LoadPublicKey() ([]byte, error)
}

type Module struct {
	Router *router.Module

	decryptionKey interface{}
	encrypter     jose.Encrypter

	oauthConfig  *oauth2.Config
	oauthOptions []oauth2.AuthCodeOption

	ErrorHandler  func(http.ResponseWriter, *http.Request, error)
	OAuthRedirect string
	KeyStore      KeyStore
	UserStore     UserStore

	SecureCookie bool
	CookieDomain string
}

var _ service.Module = &Module{}

// Configure needs to be called in setup step; todo: make this less weird.
func (m *Module) Configure(
	k KeyStore, u UserStore, config *oauth2.Config, redirectURL string,
	errHandler func(http.ResponseWriter, *http.Request, error), options ...oauth2.AuthCodeOption,
) {
	m.oauthConfig = config
	m.oauthOptions = options
	m.OAuthRedirect = redirectURL
	m.KeyStore = k
	m.UserStore = u
	m.ErrorHandler = errHandler
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.setupRoutes()
		return nil
	}

	// todo: move this to Setup after figuring out a way to ensure Configure runs first?
	c.Start = func() {
		privateKey, err := m.KeyStore.LoadPrivateKey()
		if err != nil {
			panic(err)
		}
		m.decryptionKey, err = jose.LoadPrivateKey(privateKey)
		if err != nil {
			panic(err)
		}

		pub, err := m.KeyStore.LoadPublicKey()
		if err != nil {
			panic(err)
		}
		publicKey, err := jose.LoadPublicKey(pub)
		if err != nil {
			panic(err)
		}

		m.encrypter, err = jose.NewEncrypter(KeyAlgorithm, ContentEncryption, publicKey)
		if err != nil {
			panic(err)
		}
	}
}
