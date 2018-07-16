package static

import "github.com/gobuffalo/packr"

type option func(m *Module)

// WithBox configures the static module with a source
func WithBox(box *packr.Box) option {
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
