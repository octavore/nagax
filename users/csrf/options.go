package csrf

import "time"

type option func(m *Module)

// WithValidity configures how long the csrf tokens are valid for.
func WithValidity(t time.Duration) option {
	return func(m *Module) {
		m.csrfValidityDuration = t
	}
}
