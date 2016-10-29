package databaseauth

import (
	"errors"
	"net/http"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users/session"
)

const (
	defaultLoginPath            = "/login"
	defaultPostAuthRedirectPath = "/"
)

// Module databaseauth provides login via a database
// or database-like backend
type Module struct {
	Router    *router.Module
	Sessions  *session.Module
	Logger    *logger.Module
	UserStore UserStore

	LoginPath            string
	PostAuthRedirectPath string // defaults to "/"
	ErrorHandler         func(http.ResponseWriter, *http.Request, error)
}

var _ service.Module = &Module{}

func (m *Module) Init(c *service.Config) {
	c.Start = func() {
		if m.UserStore == nil {
			panic("databaseauth: UserStore not configured")
		}
		if m.ErrorHandler == nil {
			panic("databaseauth: ErrorHandler not configured")
		}
		if m.PostAuthRedirectPath == "" {
			m.PostAuthRedirectPath = defaultPostAuthRedirectPath
		}

		loginPath := defaultLoginPath
		if m.LoginPath != "" {
			loginPath = m.LoginPath
		}

		m.Router.HandleFunc(loginPath, m.handleLogin)
	}
}

// Login with email and password, returns user id if valid
func (m *Module) Login(email, password string) (string, bool, error) {
	userID, hashedPassword, err := m.UserStore.Get(email)
	if err != nil {
		return "", false, err
	}
	return userID, AuthenticatePassword(password, hashedPassword), nil
}

func (m *Module) handleLogin(rw http.ResponseWriter, req *http.Request) {
	email := req.PostFormValue("email")
	password := req.PostFormValue("password")
	userID, valid, err := m.Login(email, password)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}
	if !valid {
		m.ErrorHandler(rw, req, errors.New("invalid user"))
		return
	}
	err = m.Sessions.CreateSession(userID, rw)
	if err != nil {
		m.ErrorHandler(rw, req, err)
		return
	}
	// todo: allow whitelist of parameters to pass through to redirect
	http.Redirect(rw, req, m.PostAuthRedirectPath, http.StatusFound)
}
