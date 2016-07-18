package users

import (
	"context"
	"net/http"
)

type UserTokenKey struct{}

func (m *Module) authenticatedFunc(next, unauthenticatedNext http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		userToken, err := m.Authenticate(req)
		if err != nil {
			unauthenticatedNext(rw, req)
		} else {
			ctx := context.WithValue(req.Context(), UserTokenKey{}, userToken)
			next(rw, req.WithContext(ctx))
		}
	}
}

func (m *Module) AuthenticatedFunc(next http.HandlerFunc) http.HandlerFunc {
	return m.authenticatedFunc(next, next)
}

func (m *Module) MustAuthenticateFunc(next http.HandlerFunc) http.HandlerFunc {
	return m.authenticatedFunc(next, func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, "/", http.StatusForbidden)
	})
}

func (m *Module) MustAuthenticateHandler(next http.Handler) http.HandlerFunc {
	return m.MustAuthenticateFunc(next.ServeHTTP)
}
