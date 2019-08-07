package csrf

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	csrf2 "github.com/octavore/nagax/users/csrf"
	"github.com/octavore/nagax/users/session"
	"github.com/octavore/nagax/util/errors"
)

const csrfHeaderKey = "x-csrf-token"

var csrfWhitelist = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"OPTIONS": true,
	"TRACE":   true,
}

type Module struct {
	Router  *router.Module
	Session *session.Module
	CSRF    *csrf2.Module
	Logger  *logger.Module
}

func (m *Module) Init(c *service.Config) {
}

func (m *Module) New(ignorePaths ...string) func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ignoreRouter := httprouter.New()
	noopHandler := func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	for _, path := range ignorePaths {
		ignoreRouter.POST(strings.TrimSuffix(path, "/"), noopHandler)
	}

	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		// 1. whitelisted: ignore csrf
		if csrfWhitelist[req.Method] {
			next(rw, req)
			return
		}

		// 2. h != nil: this path is explicitly not checked for csrf
		cleanedPath := strings.TrimSuffix(req.URL.Path, "/")
		h, _, _ := ignoreRouter.Lookup("POST", cleanedPath)
		if h != nil {
			next(rw, req)
			return
		}

		// 3. only check csrf if user is logged in
		// don't check csrf for logged out users right now
		session, err := m.Session.Verify(req)
		if err != nil {
			m.Router.HandleError(rw, req, errors.Wrap(err))
			return
		}
		if session == "" {
			next(rw, req)
			return
		}

		// 4. csrf token is required, so check it
		csrfToken := req.Header.Get(csrfHeaderKey)
		if csrfToken == "" && req.Method == "POST" {
			csrfToken = req.PostFormValue("csrfToken")
		}
		if csrfToken == "" {
			err := router.NewRequestError(req, http.StatusBadRequest, "csrf token missing")
			m.Router.HandleError(rw, req, errors.Wrap(err))
			return
		}
		ok, err := m.CSRF.Verify(session, csrfToken)
		if err != nil {
			ok = false
			m.Logger.Error(errors.Wrap(err))
		}
		if !ok {
			err := router.NewRequestError(req, http.StatusBadRequest, "invalid csrf token")
			m.Router.HandleError(rw, req, errors.Wrap(err))
			return
		}
		next(rw, req)
	}
}
