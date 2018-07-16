package static

import (
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gobuffalo/packr"

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

// Module static serves static files
type Module struct {
	Router *router.Module
	Logger *logger.Module

	staticBasePath string
	staticDirs     []string
	box            *packr.Box
}

// Init this module
func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		m.Router.HTTPRouter.NotFound = m
		m.staticBasePath = defaultStaticBasePath
		m.staticDirs = defaultStaticDirs
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

func (m *Module) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isStaticAsset := false
	for _, dir := range m.staticDirs {
		isStaticAsset = isStaticAsset || strings.HasPrefix(req.URL.Path, dir)
	}

	p := strings.TrimPrefix(req.URL.Path, m.staticBasePath)
	ext := path.Ext(p)

	if !isStaticAsset {
		p = "index.html"
		ext = path.Ext(p)
	}

	b, err := m.box.MustBytes(p)
	if err != nil && strings.Contains(err.Error(), "not found") {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		m.Logger.Errorf("%s: %s", req.URL, err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Add("Content-Type", mime.TypeByExtension(ext))
	rw.Write(b)
}
