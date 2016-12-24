package databaseauth

type option func(*Module)

func WithUserStore(u UserStore) option {
	return func(m *Module) {
		m.userStore = u
	}
}

func WithErrorHandler(e errorHandler) option {
	return func(m *Module) {
		m.errorHandler = e
	}
}

func WithRedirectPath(path string) option {
	return func(m *Module) {
		m.postAuthRedirectPath = path
	}
}

func WithLoginPath(path string) option {
	return func(m *Module) {
		m.loginPath = path
	}
}
