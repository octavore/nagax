package static

import (
	"net/http"
)

type option func(m *Module)

// WithBox configures the static module with a source
func WithBox(box fileSource) option {
	return func(m *Module) {
		if m.box != nil {
			panic("box already configured for static module")
		}
		m.box = box
	}
}

// WithStaticBase configures static module with a base URL path for
// serving static assets
func WithStaticBase(dir string) option {
	return func(m *Module) {
		m.staticBasePath = dir
	}
}

// WithStaticDirs configures static modules with valid paths to respond to
// (must begin with the staticBasePath)
func WithStaticDirs(dirs ...string) option {
	return func(m *Module) {
		m.staticDirs = dirs
	}
}

// WithHandle404 configures static module with a base URL path for
// serving static assets
func WithHandle404(fn http.HandlerFunc) option {
	return func(m *Module) {
		m.handle404 = fn
	}
}

// WithHandle500 configures static module with a base URL path for
// serving static assets
func WithHandle500(fn func(rw http.ResponseWriter, req *http.Request, err error)) option {
	return func(m *Module) {
		m.handle500 = fn
	}
}
