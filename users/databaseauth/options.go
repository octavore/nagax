package databaseauth

type option func(*Module)

func (m *Module) Configure(opts ...option) {
	for _, opt := range opts {
		opt(m)
	}
}

// WithUserStore sets the backing user store. Required.
func WithUserStore(u UserStore) option {
	return func(m *Module) {
		m.userStore = u
	}
}

// WithRedirectPath sets the path to redirect to after login.
// If set to blank, successful logins will not result in a redirect.
func WithRedirectPath(path string) option {
	return func(m *Module) {
		m.postAuthRedirectPath = path
	}
}

// WithLoginPath sets the path to listen for login requests. Optional.
func WithLoginPath(path string) option {
	return func(m *Module) {
		m.loginPath = path
	}
}

// WithoutLoginPath disables the login handler
func WithoutLoginPath() option {
	return func(m *Module) {
		m.loginPath = ""
	}
}
