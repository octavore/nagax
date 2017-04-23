package tokenauth

import (
	"strings"
)

type option func(*Module)

func (m *Module) Configure(opts ...option) {
	for _, opt := range opts {
		opt(m)
	}
}

// WithTokenSource sets the backing token source. Required.
func WithTokenSource(ts TokenSource) option {
	return func(m *Module) {
		m.tokenSource = ts
	}
}

// WithHeader sets the http header to use. Defaults to 'Authorization'. Optional.
func WithHeader(header string) option {
	return func(m *Module) {
		m.header = header
	}
}

// WithPrefix sets the prefix to check for in the header. Defaults to 'Token'. Optional.
func WithPrefix(prefix string) option {
	return func(m *Module) {
		m.prefix = strings.ToLower(prefix)
	}
}
