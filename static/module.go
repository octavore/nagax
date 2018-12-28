package static

import (
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/octavore/naga/service"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/router"
)

const defaultStaticBasePath = "/static/"

var defaultStaticDirs = []string{
	"/static/js/",
	"/static/css/",
	"/static/vendor/",
	"/static/images/",
}

type fileSource interface {
	MustBytes(filepath string) ([]byte, error)
}

// Module static serves static files
type Module struct {
	Router *router.Module
	Logger *logger.Module

	handle404      http.HandlerFunc
	handle500      func(rw http.ResponseWriter, req *http.Request, err error)
	staticBasePath string
	staticDirs     []string
	box            fileSource
}

// Init this module
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.Router.HTTPRouter.NotFound = m
		m.staticBasePath = defaultStaticBasePath
		m.staticDirs = defaultStaticDirs
		m.handle404 = m.DefaultHandle404
		m.handle500 = m.DefaultHandle500
		return nil
	}
	c.Start = func() {
		m.Router.Root.Handle(m.staticBasePath, m)
	}
}

// Configure this module with given options
func (m *Module) Configure(opts ...option) {
	for _, opt := range opts {
		opt(m)
	}
}

// DefaultHandle404 default 404 handler
func (m *Module) DefaultHandle404(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusNotFound)
}

// DefaultHandle500 default 500 handler
func (m *Module) DefaultHandle500(rw http.ResponseWriter, req *http.Request, err error) {
	m.Logger.Errorf("%s: %s", req.URL, err)
	http.Error(rw, "internal server error", http.StatusInternalServerError)
}

func (m *Module) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isStaticAsset := false
	for _, dir := range m.staticDirs {
		isStaticAsset = isStaticAsset || strings.HasPrefix(req.URL.Path, dir)
	}
	p := strings.TrimPrefix(req.URL.Path, m.staticBasePath)
	if !isStaticAsset {
		p = "index.html"
	}
	m.ServeAsset(rw, req, p, true)
}

// ServeAsset serves a filepath from the packr box. handle404 and handle500
// should not recurse.
func (m *Module) ServeAsset(rw http.ResponseWriter, req *http.Request, filepath string, customErrHandler bool) {
	ext := path.Ext(filepath)
	b, err := m.box.MustBytes(filepath)
	if err != nil {
		if !customErrHandler {
			m.Logger.Errorf("%s: %s", req.URL, err)
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		switch {
		case strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "no such file or directory"):
			m.handle404(rw, req)
		default:
			m.handle500(rw, req, err)
		}
		return
	}
	rw.Header().Add("Content-Type", mime.TypeByExtension(ext))
	rw.Write(b)
}
