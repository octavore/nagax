package auth_router

import (
	"net/http"

	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
	"github.com/octavore/nagax/users"
)

// Module router implements basic routing with helpers for protobuf-rootd responses.
type Module[AuthSession any] struct {
	Logger *logger.Module
	Config *config.Module
	Router *router.Module
	Auth   *users.Module

	// RequireAuth runs before all Register routes in this module (not Public routes) Use this to
	// parse an auth session from the request. Override authenticators may be passed via the
	// variadic authenticators argument
	RequireAuth func(router.Handle, ...users.Authenticator) router.Handle

	// GetAuthSession is use to pull an AuthSession (ideally provided by RequireAuth above) out of
	// the request context. Note that nil auth sessions are okay; if you want to require an auth
	// session, implement GetAuthSession to throw an ErrNotAuthorized when your auth session is nil
	GetAuthSession func(req *http.Request) (*AuthSession, error)

	routeRegistry []*Route
}

// Init implements service.Init
func (m *Module[A]) Init(c *service.Config) {
	m.registerRoutesList(c)
	c.Setup = func() error {
		// default handlers
		m.RequireAuth = func(h router.Handle, a ...users.Authenticator) router.Handle { return h }
		m.GetAuthSession = func(req *http.Request) (*A, error) { return nil, nil }
		return nil
	}
}

// POST is a shortcut for m.Router.POST. NOT authed by default
func (m *Module[A]) Public(pathSpec string, h router.Handle) {
	method, path := parsePathSpec(pathSpec)
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  method,
		path:    path,
		handler: h,
		version: "http",
	})

	switch method {
	case http.MethodGet:
		m.Router.GET(path, h)
	case http.MethodPost:
		m.Router.POST(path, h)
	case http.MethodDelete:
		m.Router.DELETE(path, h)
	case http.MethodPut:
		m.Router.PUT(path, h)
	case http.MethodPatch:
		m.Router.PATCH(path, h)
	default:
		panic("Unsupported method: " + method)
	}
}

// POST is a shortcut for m.Router.POST. NOT authed
func (m *Module[A]) PublicPOST(path string, h router.Handle) {
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  http.MethodPost,
		path:    path,
		handler: h,
		version: "http",
	})
	m.Router.POST(path, h)
}

// GET is a shortcut for m.Router.GET. NOT authed
func (m *Module[A]) PublicGET(path string, h router.Handle) {
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  http.MethodGet,
		path:    path,
		handler: h,
		version: "http",
	})
	m.Router.GET(path, h)
}

// PUT is a shortcut for m.Router.PUT. NOT authed
func (m *Module[A]) PublicPUT(path string, h router.Handle) {
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  http.MethodPut,
		path:    path,
		handler: h,
		version: "http",
	})
	m.Router.PUT(path, h)
}

// PATCH is a shortcut for m.Router.PATCH. NOT authed
func (m *Module[A]) PublicPATCH(path string, h router.Handle) {
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  http.MethodPatch,
		path:    path,
		handler: h,
		version: "http",
	})
	m.Router.PATCH(path, h)
}

// DELETE is a shortcut for m.Router.DELETE. NOT authed
func (m *Module[A]) PublicDELETE(path string, h router.Handle) {
	m.routeRegistry = append(m.routeRegistry, &Route{
		method:  http.MethodDelete,
		path:    path,
		handler: h,
		version: "http",
	})
	m.Router.DELETE(path, h)
}
