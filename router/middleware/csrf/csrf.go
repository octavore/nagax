package csrf

import (
	"net/http"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	csrf2 "github.com/octavore/nagax/users/csrf"
	"github.com/octavore/nagax/users/session"
	"github.com/octavore/nagax/util/errors"
)

type Module struct {
	Router  *router.Module
	Session *session.Module
	CSRF    *csrf2.Module
	Logger  *logger.Module
}

func (m *Module) Init(c *service.Config) {
}

func (m *Module) New() func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		session, err := m.Session.Verify(req)
		if err != nil {
			// bad request cookie
			m.Logger.Warning(errors.Wrap(err))
			re := router.NewRequestError(req, http.StatusBadRequest, "invalid request")
			m.Router.HandleError(rw, req, re)
			return
		}

		csrfHeader := req.Header.Get("x-csrf-token")
		ok, err := m.CSRF.Verify(session, csrfHeader)
		if err != nil {
			// error decoding the token
			m.Logger.Warning(errors.Wrap(err))
			re := router.NewRequestError(req, http.StatusBadRequest, "invalid request")
			m.Router.HandleError(rw, req, re)
			return
		}

		if !ok {
			// invalid token
			m.Logger.Warning(errors.New("invalid csrf token"))
			re := router.NewRequestError(req, http.StatusBadRequest, "invalid csrf token")
			m.Router.HandleError(rw, req, re)
			return
		}
		next(rw, req)
	}
}
