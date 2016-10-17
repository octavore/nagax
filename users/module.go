package users

import (
	"net/http"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
)

type Module struct {
	Logger            *logger.Module
	ErrorHandler      func(http.ResponseWriter, *http.Request, error)
	BaseAuthenticator Authenticator // defaults to MultiAuthenticate
	Authenticators    []Authenticator
}

var _ service.Module = &Module{}

// RegisterAuthenticator registers a new authenticator
func (m *Module) RegisterAuthenticator(a Authenticator) {
	m.Authenticators = append(m.Authenticators, a)
}

// Init implements the Module interface method
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.BaseAuthenticator = m
		return nil
	}
}

// AuthError implements Error
// q: who should control the response?
// should it be json? html? what is the status code?
// should this be described in protobuf?
type AuthError struct {
	Response []byte
	Code     int
	JSON     bool // default false
}

// AuthFunc is for users of this library to implement
type AuthFunc func(req *http.Request) (bool, *http.Cookie, error)

type Authenticator interface {
	// Authenticate returns true if the authenticator was used.
	// userToken is empty if not authenticated
	Authenticate(rw http.ResponseWriter, req *http.Request) (handled bool, userToken *string, err error)

	// Logout the given request
	Logout(rw http.ResponseWriter, req *http.Request)
}

// Authenticate: BasicAuth user/password
// Authenticate: token stored in database
// Authenticate: token stored in redis
// Authenticate: token stored in cookie -> how to get a cookie?

// After authentication: implementers store something in the context?
// What if they want to write to a cookie? i.e. modify the response
