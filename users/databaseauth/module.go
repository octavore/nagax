package databaseauth

import (
	"errors"
	"net/http"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users/session"
	"github.com/octavore/nagax/util/token"
)

const (
	defaultLoginPath            = "/login"
	defaultPostAuthRedirectPath = ""
)

var (
	_ service.Module = &Module{}
)

// Module databaseauth provides login via a database
// or database-like backend. It uses the session module.
// user authenticates via this module, then is a given a cookie
// which authenticates future requests. It registers m.loginPath for
// handling the login POST request.
type Module struct {
	Router   *router.Module
	Sessions *session.Module
	Logger   *logger.Module

	userStore            UserStore
	loginPath            string
	postAuthRedirectPath string // defaults to "/"
}

var _ service.Module = &Module{}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.postAuthRedirectPath = defaultPostAuthRedirectPath
		m.loginPath = defaultLoginPath
		return nil
	}
	c.Start = func() {
		if m.userStore == nil {
			panic("databaseauth: UserStore not configured")
		}
		m.Router.POST(m.loginPath, m.handleLogin)
	}
}

// Create a new user
func (m *Module) Create(email, password string) (string, error) {
	return m.userStore.Create(email, HashPassword(password, token.New32()))
}

// Login with email and password, returns user id if valid
func (m *Module) Login(email, password string) (string, bool, error) {
	userID, hashedPassword, err := m.userStore.Get(email)
	if err != nil {
		return "", false, err
	}
	return userID, AuthenticatePassword(password, hashedPassword), nil
}

func (m *Module) handleLogin(rw http.ResponseWriter, req *http.Request, par router.Params) error {
	email := req.PostFormValue("email")
	password := req.PostFormValue("password")
	userID, valid, err := m.Login(email, password)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid user")
	}
	err = m.Sessions.CreateSession(userID, rw)
	if err != nil {
		return err
	}
	// todo: allow whitelist of parameters to pass through to redirect
	if m.postAuthRedirectPath != "" {
		http.Redirect(rw, req, m.postAuthRedirectPath, http.StatusFound)
	}
	return nil
}
