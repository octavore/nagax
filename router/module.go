package router

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router/middleware"
	"github.com/octavore/nagax/util/errors"
)

type Params = httprouter.Params

type (
	Handle      func(rw http.ResponseWriter, req *http.Request, par httprouter.Params) error
	HandleError func(rw http.ResponseWriter, req *http.Request, err error)
)

// Config for the router module
type Config struct {
	Port         int  `json:"port"`
	BindExternal bool `json:"bindext"`
}

// Module router implements basic routing with helpers for protobuf-rootd responses.
type Module struct {
	Logger *logger.Module
	Config *config.Module

	Root         *http.ServeMux
	HTTPRouter   *httprouter.Router
	ErrorHandler HandleError
	Middleware   *middleware.MiddlewareServer

	ErrorPage func(rw http.ResponseWriter, req *http.Request, status int)
	config    Config
	server    *http.Server
}

// Init implements service.Init
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.HTTPRouter = httprouter.New()
		m.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			_ = m.HandleError(rw, req, err)
		}
		m.ErrorPage = func(rw http.ResponseWriter, req *http.Request, status int) {
			http.Error(rw, fmt.Sprint(status), status)
		}

		// root handler
		m.Root = http.NewServeMux()
		m.Root.Handle("/", m.HTTPRouter)
		m.Middleware = middleware.NewServer(m.Root.ServeHTTP)
		m.Config.ReadConfig(&m.config)
		return nil
	}

	c.Start = func() {
		laddr := m.laddr()
		m.Logger.Infof("listening on %s", laddr)
		m.server = &http.Server{Addr: laddr, Handler: m.Middleware}
		go m.server.ListenAndServe()
	}

	c.Stop = func() {
		err := m.server.Close()
		if err != nil {
			m.Logger.Error(errors.Wrap(err))
		}
	}
}

func (m *Module) laddr() string {
	iface := "127.0.0.1"
	port := 8000
	if m.config.Port != 0 {
		port = m.config.Port
		if port == 80 || port == 443 {
			iface = "0.0.0.0"
		}
	}
	if m.config.BindExternal {
		iface = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", iface, port)
}

// POST is a shortcut for m.HTTPRouter.POST
func (m *Module) POST(path string, h Handle) {
	m.HTTPRouter.POST(path, m.wrap(h))
}

// GET is a shortcut for m.HTTPRouter.GET
func (m *Module) GET(path string, h Handle) {
	m.HTTPRouter.GET(path, m.wrap(h))
}

// PUT is a shortcut for m.HTTPRouter.PUT
func (m *Module) PUT(path string, h Handle) {
	m.HTTPRouter.PUT(path, m.wrap(h))
}

// PATCH is a shortcut for m.HTTPRouter.PATCH
func (m *Module) PATCH(path string, h Handle) {
	m.HTTPRouter.PATCH(path, m.wrap(h))
}

// DELETE is a shortcut for m.HTTPRouter.DELETE
func (m *Module) DELETE(path string, h Handle) {
	m.HTTPRouter.DELETE(path, m.wrap(h))
}

// Handle is a shortcut for m.HTTPRouter.Handle
func (m *Module) Handle(method, path string, h http.HandlerFunc) {
	m.HTTPRouter.Handle(method, path, func(rw http.ResponseWriter, req *http.Request, _ Params) {
		h(rw, req)
	})
}

// WrappedHandle is a shortcut for m.HTTPRouter.Handle
func (m *Module) WrappedHandle(method, path string, h Handle) {
	m.HTTPRouter.Handle(method, path, m.wrap(h))
}

// Subrouter creates a new router rooted at path
func (m *Module) Subrouter(path string) *httprouter.Router {
	r := httprouter.New()
	m.Root.Handle(path, r)
	return r
}

// wrap the given handler to handle errors
func (m *Module) wrap(h Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, req *http.Request, par Params) {
		err := h(rw, req, par)
		if err != nil && m.ErrorHandler != nil {
			m.ErrorHandler(rw, req, err)
		}
	}
}
