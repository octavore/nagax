package oauth

import (
	"golang.org/x/oauth2"
)

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

func WithRedirectPath(url string) option {
	return func(m *Module) {
		m.postOAuthRedirectPath = url
	}
}

func WithOAuthConfig(config *oauth2.Config) option {
	return func(m *Module) {
		m.oauthConfig = config
	}
}

func WithAuthCodeOptions(authCodeOptions ...oauth2.AuthCodeOption) option {
	return func(m *Module) {
		m.oauthOptions = authCodeOptions
	}
}
